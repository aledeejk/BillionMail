<template>
  <div class="executions-page">
    <div class="page-header">
      <div>
        <h1>Workflow Runs</h1>
        <p>Frontend simulation of background processor, priority queue, locks, retries, pause/resume.</p>
      </div>
      <div class="header-actions">
        <button class="btn-secondary" @click="router.push('/automation')">← Workflows</button>
        <button class="btn-danger" @click="executionStore.clearRuns()">Clear logs</button>
      </div>
    </div>

    <div class="runs-table">
      <table>
        <thead>
          <tr>
            <th>Workflow</th>
            <th>Priority</th>
            <th>Status</th>
            <th>Started</th>
            <th>Completed</th>
            <th>Controls</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="executionStore.runs.length === 0">
            <td colspan="6" class="empty">No simulated workflow runs yet.</td>
          </tr>
          <tr v-for="run in executionStore.runs" :key="run.id">
            <td>{{ run.workflowName }}</td>
            <td><span class="priority" :class="run.priority">{{ run.priority }}</span></td>
            <td><span class="status" :class="run.status">{{ run.status }}</span></td>
            <td>{{ formatDate(run.startedAt || run.createdAt) }}</td>
            <td>{{ run.completedAt ? formatDate(run.completedAt) : '-' }}</td>
            <td>
              <button v-if="run.status === 'running'" class="btn-sm btn-secondary" @click="executionStore.pauseRun(run.id)">Pause</button>
              <button v-if="run.status === 'paused'" class="btn-sm btn-secondary" @click="executionStore.resumeRun(run.id)">Resume</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="run-logs">
      <div v-for="run in executionStore.runs" :key="`${run.id}-logs`" class="log-card">
        <div class="log-header">
          <h3>{{ run.workflowName }}</h3>
          <span class="status" :class="run.status">{{ run.status }}</span>
        </div>
        <p v-if="run.errorMessage" class="error-message">{{ run.errorMessage }}</p>
        <div class="logs">
          <div v-for="(log, index) in run.logs" :key="index" class="log-line" :class="log.level">
            <span class="log-time">{{ log.time }}</span>
            <span>{{ log.message }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useWorkflowExecutionStore } from '@/stores/workflowExecution'

const router = useRouter()
const executionStore = useWorkflowExecutionStore()

const formatDate = (value: string) => {
  return value ? new Date(value).toLocaleString() : '-'
}
</script>

<style scoped>
.executions-page {
  min-height: 100vh;
  padding: 2rem;
  background: #f5f7fb;
}

.page-header {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.page-header h1 {
  margin: 0;
}

.header-actions {
  display: flex;
  gap: 0.5rem;
}

.runs-table,
.log-card {
  margin-bottom: 1.5rem;
  padding: 1rem;
  background: #fff;
  border-radius: 10px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

table {
  width: 100%;
  border-collapse: collapse;
}

th,
td {
  padding: 0.75rem;
  border-bottom: 1px solid #eee;
  text-align: left;
}

.empty {
  text-align: center;
  color: #777;
}

.status,
.priority {
  display: inline-block;
  padding: 0.25rem 0.5rem;
  border-radius: 999px;
  font-size: 0.85rem;
  text-transform: uppercase;
}

.status.queued { background: #e2e3e5; color: #41464b; }
.status.running { background: #cff4fc; color: #055160; }
.status.paused { background: #fff3cd; color: #664d03; }
.status.completed { background: #d1e7dd; color: #0f5132; }
.status.failed { background: #f8d7da; color: #842029; }
.priority.high { background: #f8d7da; color: #842029; }
.priority.normal { background: #cff4fc; color: #055160; }
.priority.low { background: #e2e3e5; color: #41464b; }

.log-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.log-header h3 {
  margin: 0 0 0.75rem;
}

.logs {
  max-height: 280px;
  overflow: auto;
  padding: 0.75rem;
  background: #212529;
  border-radius: 8px;
}

.log-line {
  display: flex;
  gap: 0.75rem;
  padding: 0.25rem 0;
  color: #f8f9fa;
  font-family: Consolas, monospace;
}

.log-line.success { color: #75b798; }
.log-line.warning { color: #ffda6a; }
.log-line.error { color: #ea868f; }
.log-time { color: #adb5bd; }
.error-message { color: #842029; }

.btn-secondary,
.btn-danger,
.btn-sm {
  padding: 0.5rem 0.75rem;
  border: none;
  border-radius: 6px;
  cursor: pointer;
}

.btn-secondary { background: #6c757d; color: #fff; }
.btn-danger { background: #dc3545; color: #fff; }
.btn-sm { font-size: 0.85rem; }
</style>
