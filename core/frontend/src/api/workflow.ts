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

const getResponseData = (response: any) => response?.data ?? response

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
    const data = getResponseData(response)
    return (data.list || data || []).map(mapWorkflow)
  },

  async getWorkflow(id: string): Promise<Workflow | null> {
    const response = await instance.get(`/workflow/${id}`)
    const data = getResponseData(response)
    return data ? mapWorkflow(data) : null
  },

  async createWorkflow(data: CreateWorkflowRequest): Promise<Workflow> {
    const payload = {
      name: data.name,
      description: data.description,
      is_active: true,
      nodes: data.nodes || [],
      connections: data.connections || [],
    }
    const response = await instance.post('/workflow', payload)
    const responseData = getResponseData(response)
    return mapWorkflow(responseData)
  },

  async updateWorkflow(data: UpdateWorkflowRequest): Promise<Workflow> {
    const payload: Record<string, any> = {}
    if (data.name !== undefined) payload.name = data.name
    if (data.description !== undefined) payload.description = data.description
    if (data.nodes !== undefined) payload.nodes = data.nodes
    if (data.connections !== undefined) payload.connections = data.connections

    const response = await instance.put(`/workflow/${data.id}`, payload)
    const responseData = getResponseData(response)
    return mapWorkflow(responseData)
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
    const responseData = getResponseData(response)
    return mapWorkflow(responseData.workflow || responseData)
  },

  async toggleWorkflow(id: string, isActive: boolean): Promise<Workflow> {
    await instance.post(`/workflow/${id}/toggle`, {
      is_active: isActive,
    })
    const updated = await this.getWorkflow(id)
    if (!updated) {
      throw new Error('Workflow not found after toggle')
    }
    return updated
  },

  async getWorkflowStats(request: GetWorkflowStatsRequest): Promise<WorkflowStatistics> {
    const response = await instance.get(`/workflow/${request.id}/stats`)
    const responseData = getResponseData(response)
    return mapWorkflowStats(responseData)
  },

  async getWorkflowVersions(id: string): Promise<WorkflowVersion[]> {
    const response = await instance.get(`/workflow/${id}/versions`)
    const data = getResponseData(response)
    return (data.list || data || []).map(mapWorkflowVersion)
  },

  async getWorkflowExecutions(workflowId: string): Promise<WorkflowExecution[]> {
    const response = await instance.get(`/workflow/${workflowId}/executions`)
    const data = getResponseData(response)
    return (data.list || data || []).map(mapWorkflowExecution)
  },

  async executeWorkflow(id: string): Promise<WorkflowExecution> {
    const response = await instance.post(`/workflow/${id}/execute`)
    const responseData = getResponseData(response)
    return mapWorkflowExecution(responseData)
  },
}