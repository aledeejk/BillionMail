ALTER TABLE workflow_execution
    ADD COLUMN IF NOT EXISTS current_node_id VARCHAR(64)  NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS context         JSONB        NOT NULL DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS retry_count     INT          NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS idempotency_key VARCHAR(128) NOT NULL DEFAULT '';

CREATE UNIQUE INDEX IF NOT EXISTS idx_workflow_execution_idempotency
    ON workflow_execution (workflow_id, idempotency_key)
    WHERE idempotency_key <> '';
