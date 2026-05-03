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

const mapWorkflow = (data: any): Workflow => ({
  id: data.id,
  name: data.name,
  description: data.description,
  isActive: data.is_active,
  version: data.version,
  createdAt: data.created_at,
  updatedAt: data.updated_at,
})

const mapWorkflowVersion = (data: any): WorkflowVersion => ({
  version: data.version,
  createdAt: data.created_at,
  status: data.status,
})

const mapWorkflowExecution = (data: any): WorkflowExecution => ({
  id: data.id,
  workflowId: data.workflow_id,
  status: data.status,
  startedAt: data.started_at,
  completedAt: data.completed_at,
  errorMessage: data.error_message,
})

const mapWorkflowStats = (data: any): WorkflowStatistics => ({
  totalExecutions: data.total_executions ?? data.active_contacts ?? 0,
  successfulExecutions: data.successful_executions ?? data.emails_sent ?? 0,
  failedExecutions: data.failed_executions ?? 0,
  averageDuration: data.average_duration ?? 0,
  lastExecutionTime: data.last_execution_time,
})

export const workflowApi = {
  async getWorkflows(params?: { page?: number; limit?: number; search?: string }): Promise<Workflow[]> {
    const response = await instance.get('/workflow', { params })
    const data = response.data
    return (data.list || data || []).map(mapWorkflow)
  },

  async getWorkflow(id: string): Promise<Workflow | null> {
    const response = await instance.get(`/workflow/${id}`)
    const data = response.data
    return data ? mapWorkflow(data) : null
  },

  async createWorkflow(data: CreateWorkflowRequest): Promise<Workflow> {
    const payload = {
      name: data.name,
      description: data.description,
      is_active: true,  // по умолчанию активен
    }
    const response = await instance.post('/workflow', payload)
    return mapWorkflow(response.data)
  },

  async updateWorkflow(data: UpdateWorkflowRequest): Promise<Workflow> {
    const payload: Record<string, any> = {}
    if (data.name !== undefined) payload.name = data.name
    if (data.description !== undefined) payload.description = data.description
    // isActive обрабатывается отдельно через toggle

    const response = await instance.put(`/workflow/${data.id}`, payload)
    return mapWorkflow(response.data)
  },

  async deleteWorkflow(id: string): Promise<void> {
    await instance.delete(`/workflow/${id}`)
  },

  async duplicateWorkflow(id: string): Promise<Workflow> {
    const workflow = await this.getWorkflow(id)
    if (!workflow) {
      throw new Error('Workflow not found')
    }

    const response = await instance.post(`/workflow/${id}/duplicate`, {
      name: `${workflow.name} Copy`,
    })
    return mapWorkflow(response.data.workflow || response.data)
  },

  async toggleWorkflow(id: string, isActive: boolean): Promise<Workflow> {
    const response = await instance.post(`/workflow/${id}/toggle`)
    // После toggle получаем обновлённый workflow
    const updated = await this.getWorkflow(id)
    if (!updated) {
      throw new Error('Workflow not found after toggle')
    }
    return updated
  },

  async getWorkflowStats(request: GetWorkflowStatsRequest): Promise<WorkflowStatistics> {
    const response = await instance.get(`/workflow/${request.id}/stats`)
    return mapWorkflowStats(response.data)
  },

  async getWorkflowVersions(id: string): Promise<WorkflowVersion[]> {
    const response = await instance.get(`/workflow/${id}/versions`)
    const data = response.data
    return (data.list || data || []).map(mapWorkflowVersion)
  },

  async getWorkflowExecutions(workflowId: string): Promise<WorkflowExecution[]> {
    const response = await instance.get(`/workflow/${workflowId}/executions`)
    const data = response.data
    return (data.list || data || []).map(mapWorkflowExecution)
  },

  async executeWorkflow(id: string): Promise<WorkflowExecution> {
    const response = await instance.post(`/workflow/${id}/execute`)
    return mapWorkflowExecution(response.data)
  },
}