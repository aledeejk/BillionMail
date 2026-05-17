package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	goredislib "github.com/redis/go-redis/v9"
)

const (
	StreamName     = "workflow:execution:stream"
	ConsumerGroup  = "workflow:workers"
	deadLetterKey  = "workflow:execution:dead"
	lockTTL        = 30 * time.Second
	claimIdleAfter = 60 * time.Second
	maxRetries     = 3
)

type ExecutionTask struct {
	WorkflowId     int64                  `json:"workflow_id"`
	ContactId      int64                  `json:"contact_id"`
	InputData      map[string]interface{} `json:"input_data"`
	IdempotencyKey string                 `json:"idempotency_key"`
	EnqueuedAt     int64                  `json:"enqueued_at"`
}

type TaskHandler func(ctx context.Context, task ExecutionTask) error

type Metrics struct {
	Enqueued      atomic.Int64
	Processed     atomic.Int64
	Failed        atomic.Int64
	ActiveWorkers atomic.Int64
}

var GlobalMetrics = &Metrics{}

type RedisQueue struct {
	client     *goredislib.Client
	rs         *redsync.Redsync
	workerName string
}

func NewRedisQueue(addr string) (*RedisQueue, error) {
	client := goredislib.NewClient(&goredislib.Options{
		Addr:         addr,
		DialTimeout:  3 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis unavailable at %s: %w", addr, err)
	}

	pool := goredis.NewPool(client)
	rs := redsync.New(pool)

	hostname, _ := os.Hostname()
	workerName := fmt.Sprintf("%s:%d", hostname, os.Getpid())

	q := &RedisQueue{client: client, rs: rs, workerName: workerName}

	if err := q.ensureConsumerGroup(context.Background()); err != nil {
		return nil, err
	}

	return q, nil
}

func (q *RedisQueue) ensureConsumerGroup(ctx context.Context) error {
	err := q.client.XGroupCreateMkStream(ctx, StreamName, ConsumerGroup, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("failed to create consumer group: %w", err)
	}
	return nil
}

func (q *RedisQueue) Enqueue(ctx context.Context, task ExecutionTask) error {
	if task.IdempotencyKey == "" {
		task.IdempotencyKey = fmt.Sprintf("%d:%d:%d", task.WorkflowId, task.ContactId, time.Now().UnixNano())
	}
	task.EnqueuedAt = time.Now().Unix()

	payload, err := json.Marshal(task)
	if err != nil {
		return err
	}

	if err := q.client.XAdd(ctx, &goredislib.XAddArgs{
		Stream: StreamName,
		Values: map[string]interface{}{"payload": string(payload)},
	}).Err(); err != nil {
		return fmt.Errorf("enqueue failed: %w", err)
	}

	GlobalMetrics.Enqueued.Add(1)
	return nil
}

func (q *RedisQueue) AcquireWorkflowLock(workflowId int64) (*redsync.Mutex, error) {
	return q.AcquireWorkflowLockWithKey(workflowId, "")
}

func (q *RedisQueue) AcquireWorkflowLockWithKey(workflowId int64, idempotencyKey string) (*redsync.Mutex, error) {
	var lockKey string
	if idempotencyKey != "" {
		lockKey = fmt.Sprintf("workflow:%d:exec:%s", workflowId, idempotencyKey)
	} else {
		lockKey = fmt.Sprintf("workflow:%d:execution", workflowId)
	}
	mutex := q.rs.NewMutex(lockKey,
		redsync.WithExpiry(lockTTL),
		redsync.WithTries(1),
	)
	if err := mutex.Lock(); err != nil {
		return nil, err
	}
	return mutex, nil
}

func (q *RedisQueue) QueueDepth(ctx context.Context) (int64, error) {
	info, err := q.client.XInfoStream(ctx, StreamName).Result()
	if err != nil {
		return 0, err
	}
	return info.Length, nil
}

func (q *RedisQueue) RunWorkers(ctx context.Context, count int, handler TaskHandler) {
	var wg sync.WaitGroup

	for i := 0; i < count; i++ {
		wg.Add(1)
		workerIdx := i
		go func() {
			defer wg.Done()
			q.runWorkerLoop(ctx, workerIdx, handler)
		}()
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigCh:
		fmt.Println("[queue] shutdown signal received, draining workers…")
	case <-ctx.Done():
	}

	wg.Wait()
	fmt.Println("[queue] all workers stopped cleanly")
}

func (q *RedisQueue) runWorkerLoop(ctx context.Context, idx int, handler TaskHandler) {
	consumer := fmt.Sprintf("%s:w%d", q.workerName, idx)
	fmt.Printf("[worker %d] started as %s\n", idx, consumer)

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("[worker %d] shutting down\n", idx)
			return
		default:
		}

		if err := q.claimStaleMessages(ctx, consumer, handler); err != nil {
			fmt.Printf("[worker %d] claim error: %v\n", idx, err)
		}

		msgs, err := q.client.XReadGroup(ctx, &goredislib.XReadGroupArgs{
			Group:    ConsumerGroup,
			Consumer: consumer,
			Streams:  []string{StreamName, ">"},
			Count:    1,
			Block:    2 * time.Second,
		}).Result()

		if err != nil {
			if err == goredislib.Nil || err.Error() == "redis: nil" {
				continue
			}
			if ctx.Err() != nil {
				return
			}
			fmt.Printf("[worker %d] read error: %v\n", idx, err)
			time.Sleep(time.Second)
			continue
		}

		for _, stream := range msgs {
			for _, msg := range stream.Messages {
				q.processMessage(ctx, consumer, msg, handler)
			}
		}
	}
}

func (q *RedisQueue) processMessage(ctx context.Context, consumer string, msg goredislib.XMessage, handler TaskHandler) {
	GlobalMetrics.ActiveWorkers.Add(1)
	defer GlobalMetrics.ActiveWorkers.Add(-1)

	payload, ok := msg.Values["payload"].(string)
	if !ok {
		q.ackAndDiscard(ctx, msg.ID, "missing payload")
		return
	}

	var task ExecutionTask
	if err := json.Unmarshal([]byte(payload), &task); err != nil {
		q.ackAndDiscard(ctx, msg.ID, fmt.Sprintf("unmarshal error: %v", err))
		return
	}

	retryCount := q.getRetryCount(ctx, msg.ID)

	err := handler(ctx, task)
	if err != nil {
		GlobalMetrics.Failed.Add(1)
		fmt.Printf("[queue] task failed (workflow %d, attempt %d): %v\n", task.WorkflowId, retryCount+1, err)

		if retryCount >= maxRetries {
			q.moveToDead(ctx, msg.ID, payload, err)
			q.ack(ctx, msg.ID)
		}
		return
	}

	GlobalMetrics.Processed.Add(1)
	q.ack(ctx, msg.ID)
	q.clearRetryCount(ctx, msg.ID)
}

func (q *RedisQueue) claimStaleMessages(ctx context.Context, consumer string, handler TaskHandler) error {
	msgs, _, err := q.client.XAutoClaim(ctx, &goredislib.XAutoClaimArgs{
		Stream:   StreamName,
		Group:    ConsumerGroup,
		Consumer: consumer,
		MinIdle:  claimIdleAfter,
		Start:    "0-0",
		Count:    10,
	}).Result()
	if err != nil {
		return nil
	}
	for _, msg := range msgs {
		q.processMessage(ctx, consumer, msg, handler)
	}
	return nil
}

func (q *RedisQueue) ack(ctx context.Context, msgID string) {
	q.client.XAck(ctx, StreamName, ConsumerGroup, msgID)
}

func (q *RedisQueue) ackAndDiscard(ctx context.Context, msgID, reason string) {
	fmt.Printf("[queue] discarding message %s: %s\n", msgID, reason)
	q.client.XAck(context.Background(), StreamName, ConsumerGroup, msgID)
}

func (q *RedisQueue) moveToDead(ctx context.Context, msgID, payload string, err error) {
	q.client.RPush(ctx, deadLetterKey, fmt.Sprintf(`{"id":%q,"payload":%q,"error":%q}`, msgID, payload, err.Error()))
	fmt.Printf("[queue] message %s moved to dead letter queue\n", msgID)
}

func (q *RedisQueue) getRetryCount(ctx context.Context, msgID string) int {
	key := fmt.Sprintf("workflow:retry:%s", msgID)
	val, err := q.client.Incr(ctx, key).Result()
	if err != nil {
		return 0
	}
	if val == 1 {
		q.client.Expire(ctx, key, 24*time.Hour)
	}
	return int(val) - 1
}

func (q *RedisQueue) clearRetryCount(ctx context.Context, msgID string) {
	q.client.Del(ctx, fmt.Sprintf("workflow:retry:%s", msgID))
}

func (q *RedisQueue) PrintMetrics(ctx context.Context) {
	depth, _ := q.QueueDepth(ctx)
	fmt.Printf(
		"[metrics] enqueued=%d processed=%d failed=%d active_workers=%d queue_depth=%d\n",
		GlobalMetrics.Enqueued.Load(),
		GlobalMetrics.Processed.Load(),
		GlobalMetrics.Failed.Load(),
		GlobalMetrics.ActiveWorkers.Load(),
		depth,
	)
}
