import { defineStore } from 'pinia'
import { ref } from 'vue'
import { workflowApi } from '@/api/workflow'

export type WorkflowRunStatus = 'queued' | 'running' | 'paused' | 'completed' | 'failed'
export type WorkflowRunPriority = 'low' | 'normal' | 'high'

export interface WorkflowRunLog {
  time: string
  message: string
  level: 'info' | 'success' | 'warning' | 'error'
}

export interface WorkflowRun {
  id: string
  workflowId: string
  workflowName: string
  priority: WorkflowRunPriority
  status: WorkflowRunStatus
  contactId: string
  currentNodeId?: string
  errorMessage?: string
  createdAt: string
  startedAt?: string
  completedAt?: string
  logs: WorkflowRunLog[]
}

const STORAGE_KEY = 'workflow_execution_runs'
const priorityWeight: Record<WorkflowRunPriority, number> = {
  low: 1,
  normal: 2,
  high: 3,
}

const wait = (ms: number) => new Promise(resolve => setTimeout(resolve, ms))

export const useWorkflowExecutionStore = defineStore('workflowExecutionStore', () => {
  const runs = ref<WorkflowRun[]>(loadRuns())
  const activeWorkflowLocks = ref<Record<string, string>>({})
  const pauseResolvers = new Map<string, () => void>()

  function loadRuns(): WorkflowRun[] {
    try {
      const raw = localStorage.getItem(STORAGE_KEY)
      return raw ? JSON.parse(raw) : []
    } catch {
      return []
    }
  }

  function persistRuns() {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(runs.value))
  }

  function addLog(runId: string, message: string, level: WorkflowRunLog['level'] = 'info') {
    const run = runs.value.find(item => item.id === runId)
    if (!run) return
    const entry = {
      time: new Date().toLocaleTimeString(),
      message,
      level,
    }
    run.logs.push(entry)
    console.log(`[Workflow ${run.workflowName}] ${message}`)
    persistRuns()
  }

  function setStatus(runId: string, status: WorkflowRunStatus, errorMessage = '') {
    const run = runs.value.find(item => item.id === runId)
    if (!run) return
    run.status = status
    run.errorMessage = errorMessage || run.errorMessage
    if (status === 'running' && !run.startedAt) run.startedAt = new Date().toISOString()
    if (status === 'completed' || status === 'failed') {
      run.completedAt = new Date().toISOString()
      delete activeWorkflowLocks.value[run.workflowId]
    }
    persistRuns()
  }

  async function startWorkflow(workflowId: string, workflowName: string, priority: WorkflowRunPriority = 'normal') {
    if (activeWorkflowLocks.value[workflowId]) {
      alert('Workflow already running')
      return
    }

    const run: WorkflowRun = {
      id: `run-${Date.now()}-${Math.random().toString(16).slice(2)}`,
      workflowId,
      workflowName,
      priority,
      status: 'queued',
      contactId: `demo-contact-${Date.now()}@example.com`,
      createdAt: new Date().toISOString(),
      logs: [],
    }

    runs.value.unshift(run)
    activeWorkflowLocks.value[workflowId] = run.id
    persistRuns()
    addLog(run.id, `Queued with ${priority.toUpperCase()} priority`, 'info')
    processQueue()
  }

  async function processQueue() {
    const queuedRuns = runs.value
      .filter(run => run.status === 'queued')
      .sort((a, b) => priorityWeight[b.priority] - priorityWeight[a.priority] || a.createdAt.localeCompare(b.createdAt))

    await Promise.all(queuedRuns.map(run => executeRun(run.id)))
  }

  async function executeRun(runId: string) {
    const run = runs.value.find(item => item.id === runId)
    if (!run || run.status !== 'queued') return

    setStatus(runId, 'running')
    addLog(runId, 'Background processor started', 'success')
    addLog(runId, 'Lock acquired to prevent racing', 'info')

    try {
      const editorData = await workflowApi.getWorkflowEditor(run.workflowId)
      const nodes = editorData.nodes || []
      const connections = editorData.connections || []
      if (nodes.length === 0) throw new Error('Workflow has no nodes')

      const targets = new Set(connections.map((connection: any) => connection.target))
      let current = nodes.find((node: any) => !targets.has(node.id)) || nodes[0]
      const visited = new Set<string>()

      while (current && !visited.has(current.id)) {
        await waitIfPaused(runId)
        visited.add(current.id)
        await executeNode(runId, current)
        const nextEdge = connections.find((connection: any) => connection.source === current.id)
        current = nextEdge ? nodes.find((node: any) => node.id === nextEdge.target) : null
      }

      setStatus(runId, 'completed')
      addLog(runId, 'Workflow completed successfully', 'success')
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Unknown execution error'
      setStatus(runId, 'failed', message)
      addLog(runId, `Workflow failed: ${message}`, 'error')
    }
  }

  async function executeNode(runId: string, node: any) {
    const run = runs.value.find(item => item.id === runId)
    if (!run) return

    run.currentNodeId = node.id
    persistRuns()

    const nodeType = node.type || node.nodeType || node.data?.nodeType
    const config = node.config || node.data?.config || {}
    const label = describeNode(nodeType, config)

    for (let attempt = 1; attempt <= 3; attempt += 1) {
      await waitIfPaused(runId)
      addLog(runId, `${label} — attempt ${attempt}`, 'info')
      await wait(1000)

      if (shouldFail(nodeType)) {
        addLog(runId, `Simulated error in ${label}`, 'error')
        if (attempt < 3) {
          addLog(runId, 'Retrying in 1 second...', 'warning')
          await wait(1000)
          continue
        }
        throw new Error(`${label} failed after 3 retries`)
      }

      addLog(runId, `${label} completed`, 'success')
      return
    }
  }

  function shouldFail(nodeType: string) {
    return ['send-email', 'condition', 'action'].includes(nodeType) && Math.random() < 0.2
  }

  async function waitIfPaused(runId: string) {
    const run = runs.value.find(item => item.id === runId)
    if (!run || run.status !== 'paused') return
    addLog(runId, 'Execution paused', 'warning')
    await new Promise<void>(resolve => pauseResolvers.set(runId, resolve))
    addLog(runId, 'Execution resumed', 'success')
  }

  function pauseRun(runId: string) {
    const run = runs.value.find(item => item.id === runId)
    if (!run || run.status !== 'running') return
    setStatus(runId, 'paused')
  }

  function resumeRun(runId: string) {
    const run = runs.value.find(item => item.id === runId)
    if (!run || run.status !== 'paused') return
    setStatus(runId, 'running')
    const resolver = pauseResolvers.get(runId)
    if (resolver) {
      resolver()
      pauseResolvers.delete(runId)
    }
  }

  function clearRuns() {
    runs.value = []
    activeWorkflowLocks.value = {}
    pauseResolvers.clear()
    persistRuns()
  }

  function describeNode(type: string, config: any) {
    if (type === 'trigger') return `Trigger activated (${config.triggerType || 'start'})`
    if (type === 'send-email') return `Email sent (${config.templateId || 'template'})`
    if (type === 'delay') return `Delay ${config.duration || 1} ${config.unit || 'days'}`
    if (type === 'condition') return 'Condition evaluated (true)'
    if (type === 'action') return `Action applied (${config.actionType || 'action'})`
    if (type === 'split') return `A/B split (${config.branchA || 50}/${config.branchB || 50})`
    return `Node ${type}`
  }

  return {
    runs,
    startWorkflow,
    pauseRun,
    resumeRun,
    clearRuns,
  }
})
