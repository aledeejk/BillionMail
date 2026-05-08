import { instance } from '@/api'
import type {
  Workflow,
  WorkflowExecution,
  WorkflowStatistics,
  WorkflowVersion,
  CreateWorkflowRequest,
  UpdateWorkflowRequest,
  GetWorkflowStatsRequest,
} from '@/types/workflow'

// Mock data storage
let workflows: Workflow[] = [
  {
    id: '1',
    name: 'Newsletter Campaign',
    description: 'Monthly newsletter automation',
    isActive: true,
    version: 2,
    createdAt: '2024-01-15T10:00:00Z',
    updatedAt: '2024-05-01T14:30:00Z',
  },
  {
    id: '2',
    name: 'Welcome Email Series',
    description: 'Automated welcome emails for new subscribers',
    isActive: false,
    version: 1,
    createdAt: '2024-02-20T09:15:00Z',
    updatedAt: '2024-03-10T11:45:00Z',
  },
  {
    id: '3',
    name: 'Abandoned Cart Reminder',
    description: 'Remind customers about abandoned carts',
    isActive: true,
    version: 3,
    createdAt: '2024-03-05T16:20:00Z',
    updatedAt: '2024-05-07T08:00:00Z',
  },
]

let nextId = 4

const workflowStats: Record<string, WorkflowStatistics> = {
  '1': {
    totalExecutions: 150,
    successfulExecutions: 145,
    failedExecutions: 5,
    averageDuration: 1200,
    lastExecutionTime: '2024-05-06T12:00:00Z',
  },
  '2': {
    totalExecutions: 75,
    successfulExecutions: 70,
    failedExecutions: 5,
    averageDuration: 800,
    lastExecutionTime: '2024-04-15T10:30:00Z',
  },
  '3': {
    totalExecutions: 200,
    successfulExecutions: 190,
    failedExecutions: 10,
    averageDuration: 1500,
    lastExecutionTime: '2024-05-07T07:45:00Z',
  },
}

const workflowVersions: Record<string, WorkflowVersion[]> = {
  '1': [
    { version: 1, createdAt: '2024-01-15T10:00:00Z', status: 'active' },
    { version: 2, createdAt: '2024-05-01T14:30:00Z', status: 'active' },
  ],
  '2': [
    { version: 1, createdAt: '2024-02-20T09:15:00Z', status: 'inactive' },
  ],
  '3': [
    { version: 1, createdAt: '2024-03-05T16:20:00Z', status: 'active' },
    { version: 2, createdAt: '2024-04-10T13:00:00Z', status: 'active' },
    { version: 3, createdAt: '2024-05-07T08:00:00Z', status: 'active' },
  ],
}

const workflowExecutions: Record<string, WorkflowExecution[]> = {
  '1': [
    {
      id: 'exec1',
      workflowId: '1',
      status: 'completed',
      startedAt: '2024-05-06T12:00:00Z',
      completedAt: '2024-05-06T12:20:00Z',
      errorMessage: undefined,
    },
  ],
  '2': [
    {
      id: 'exec2',
      workflowId: '2',
      status: 'completed',
      startedAt: '2024-04-15T10:30:00Z',
      completedAt: '2024-04-15T10:38:00Z',
      errorMessage: undefined,
    },
  ],
  '3': [
    {
      id: 'exec3',
      workflowId: '3',
      status: 'completed',
      startedAt: '2024-05-07T07:45:00Z',
      completedAt: '2024-05-07T08:00:00Z',
      errorMessage: undefined,
    },
  ],
}

// Simulate delay
const delay = (ms: number) => new Promise(resolve => setTimeout(resolve, ms))

export const workflowApi = {
  async getWorkflows(params?: { page?: number; limit?: number; search?: string }): Promise<Workflow[]> {
    await delay(400)
    let result = [...workflows]
    if (params?.search) {
      result = result.filter(w => w.name.toLowerCase().includes(params.search!.toLowerCase()) || w.description.toLowerCase().includes(params.search!.toLowerCase()))
    }
    if (params?.limit) {
      result = result.slice(0, params.limit)
    }
    return result
  },

  async getWorkflow(id: string): Promise<Workflow | null> {
    await delay(300)
    return workflows.find(w => w.id === id) || null
  },

  async createWorkflow(data: CreateWorkflowRequest): Promise<Workflow> {
    await delay(500)
    const newWorkflow: Workflow = {
      id: nextId.toString(),
      name: data.name,
      description: data.description,
      isActive: false,
      version: 1,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    }
    workflows.push(newWorkflow)
    // Add default stats for new workflow
    workflowStats[newWorkflow.id] = {
      totalExecutions: 0,
      successfulExecutions: 0,
      failedExecutions: 0,
      averageDuration: 0,
      lastExecutionTime: undefined,
    }
    nextId++
    return newWorkflow
  },

  async updateWorkflow(data: UpdateWorkflowRequest): Promise<Workflow> {
    await delay(400)
    const workflow = workflows.find(w => w.id === data.id)
    if (!workflow) throw new Error('Workflow not found')
    if (data.name !== undefined) workflow.name = data.name
    if (data.description !== undefined) workflow.description = data.description
    if (data.nodes !== undefined) workflow.nodes = data.nodes
    if (data.connections !== undefined) workflow.connections = data.connections
    workflow.updatedAt = new Date().toISOString()
    return workflow
  },

  async deleteWorkflow(id: string): Promise<void> {
    await delay(300)
    const index = workflows.findIndex(w => w.id === id)
    if (index === -1) throw new Error('Workflow not found')
    workflows.splice(index, 1)
  },

  async duplicateWorkflow(id: string): Promise<Workflow> {
    await delay(600)
    const original = workflows.find(w => w.id === id)
    if (!original) throw new Error('Workflow not found')
    const newWorkflow: Workflow = {
      id: nextId.toString(),
      name: `${original.name} (Copy)`,
      description: original.description,
      isActive: original.isActive,
      version: 1,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    }
    workflows.push(newWorkflow)
    // Add default stats for duplicated workflow
    workflowStats[newWorkflow.id] = {
      totalExecutions: 0,
      successfulExecutions: 0,
      failedExecutions: 0,
      averageDuration: 0,
      lastExecutionTime: undefined,
    }
    nextId++
    return newWorkflow
  },

  async toggleWorkflow(id: string, isActive: boolean): Promise<Workflow> {
    console.log('[Toggle API] Toggling workflow', id, 'to active:', isActive)
    await delay(400)
    const workflow = workflows.find(w => w.id === id)
    if (!workflow) throw new Error('Workflow not found')
    workflow.isActive = isActive
    workflow.updatedAt = new Date().toISOString()
    console.log('[Toggle API] Updated workflow:', workflow)
    return workflow
  },

  async getWorkflowStats(request: GetWorkflowStatsRequest): Promise<WorkflowStatistics> {
    await delay(350)
    const stats = workflowStats[request.id]
    if (!stats) throw new Error('Workflow stats not found')
    return stats
  },

  async getWorkflowVersions(id: string): Promise<WorkflowVersion[]> {
    await delay(400)
    const versions = workflowVersions[id] || []
    const workflow = workflows.find(w => w.id === id)
    if (workflow) {
      return versions.map(v => ({
        ...v,
        status: v.version === workflow.version ? 'active' : 'inactive'
      }))
    }
    return versions
  },

  async getWorkflowExecutions(workflowId: string): Promise<WorkflowExecution[]> {
    await delay(450)
    return workflowExecutions[workflowId] || []
  },

  async executeWorkflow(id: string): Promise<WorkflowExecution> {
    await delay(800)
    const execution: WorkflowExecution = {
      id: `exec${Date.now()}`,
      workflowId: id,
      status: 'completed',
      startedAt: new Date().toISOString(),
      completedAt: new Date(Date.now() + 2000).toISOString(),
      errorMessage: undefined,
    }
    if (!workflowExecutions[id]) workflowExecutions[id] = []
    workflowExecutions[id].push(execution)
    return execution
  },

  async rollbackWorkflow(id: string, version: number): Promise<Workflow> {
    await delay(600)
    const workflow = workflows.find(w => w.id === id)
    if (!workflow) throw new Error('Workflow not found')
    workflow.version = version
    workflow.updatedAt = new Date().toISOString()
    return workflow
  },
}