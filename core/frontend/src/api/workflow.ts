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

const normalizeWorkflow = (workflow: any): Workflow => ({
  id: String(workflow.id),
  name: workflow.name || '',
  description: workflow.description || '',
  isActive: Boolean(workflow.is_active ?? workflow.isActive),
  version: Number(workflow.version || 0),
  createdAt: String(workflow.created_at ?? workflow.createdAt ?? ''),
  updatedAt: String(workflow.updated_at ?? workflow.updatedAt ?? ''),
})

const normalizeVersion = (version: any): WorkflowVersion => ({
  version: Number(version.version || 0),
  createdAt: String(version.created_at ?? version.createdAt ?? ''),
  status: version.status === 1 || version.status === 'active' ? 'active' : 'inactive',
})

const normalizeExecution = (execution: any): WorkflowExecution => ({
  id: String(execution.id),
  workflowId: String(execution.workflow_id ?? execution.workflowId ?? ''),
  status: execution.status === 2 || execution.status === 'completed' ? 'completed'
    : execution.status === 3 || execution.status === 'failed' ? 'failed'
      : execution.status === 1 || execution.status === 'running' ? 'running'
        : 'pending',
  startedAt: String(execution.started_at ?? execution.startedAt ?? ''),
  completedAt: execution.completed_at ?? execution.completedAt,
  errorMessage: execution.error_message ?? execution.errorMessage,
})

export const workflowApi = {
  async getWorkflows(params?: { page?: number; limit?: number; search?: string }): Promise<Workflow[]> {
    const response: any = await instance.get('/workflow', { params, timeout: 30000 })
    const data = response?.data ?? response
    const list = data?.list || (Array.isArray(data) ? data : [])
    return list.map(normalizeWorkflow)
  },

  async getWorkflow(id: string): Promise<Workflow | null> {
    const response: any = await instance.get(`/workflow/${id}`, { timeout: 30000 })
    return response ? normalizeWorkflow(response) : null
  },

  async createWorkflow(data: CreateWorkflowRequest): Promise<Workflow> {
    const response: any = await instance.post('/workflow', {
      name: data.name,
      description: data.description,
    }, { timeout: 30000 })
    return normalizeWorkflow(response)
  },

  async updateWorkflow(data: UpdateWorkflowRequest): Promise<Workflow> {
    if (data.nodes !== undefined || data.connections !== undefined) {
      await instance.put(`/workflow/${data.id}/editor`, {
        nodes: data.nodes?.map(n => ({
          id: n.id,
          type: n.type,
          config: n.config,
          position_x: n.position?.x || 0,
          position_y: n.position?.y || 0,
        })) || [],
        connections: data.connections?.map(c => ({
          id: c.id,
          source: c.sourceNodeId,
          target: c.targetNodeId,
        })) || [],
      }, { timeout: 30000 })
      const workflow = await this.getWorkflow(data.id)
      if (!workflow) throw new Error('Workflow not found')
      return workflow
    }

    const response: any = await instance.put(`/workflow/${data.id}`, {
      name: data.name,
      description: data.description,
    }, { timeout: 30000 })
    return normalizeWorkflow(response)
  },

  async deleteWorkflow(id: string): Promise<void> {
    await instance.delete(`/workflow/${id}`, { timeout: 30000 })
  },

  async duplicateWorkflow(id: string): Promise<Workflow> {
    const workflow = await this.getWorkflow(id)
    const response: any = await instance.post(`/workflow/${id}/duplicate`, {
      name: `${workflow?.name || 'Workflow'} (Copy)`,
    }, { timeout: 30000 })
    return normalizeWorkflow(response?.workflow || response)
  },

  async toggleWorkflow(id: string, isActive: boolean): Promise<Workflow> {
    await instance.post(`/workflow/${id}/toggle`, { is_active: isActive }, { timeout: 30000 })
    const workflow = await this.getWorkflow(id)
    if (!workflow) throw new Error('Workflow not found')
    return workflow
  },

  async getWorkflowStats(request: GetWorkflowStatsRequest): Promise<WorkflowStatistics> {
    const response: any = await instance.get(`/workflow/${request.id}/stats`, { timeout: 30000 })
    return {
      totalExecutions: Number(response.active_contacts || response.totalExecutions || 0),
      successfulExecutions: Number(response.emails_sent || response.successfulExecutions || 0),
      failedExecutions: Number(response.failedExecutions || 0),
      averageDuration: Number(response.averageDuration || 0),
      lastExecutionTime: response.lastExecutionTime,
    }
  },

  async getWorkflowVersions(id: string): Promise<WorkflowVersion[]> {
    const response: any = await instance.get(`/workflow/${id}/versions`, { timeout: 30000 })
    return (Array.isArray(response) ? response : []).map(normalizeVersion)
  },

  async deleteWorkflowVersion(id: string, version: number): Promise<void> {
    await instance.delete(`/workflow/${id}/versions/${version}`, { timeout: 30000 })
  },

  async getWorkflowExecutions(workflowId: string): Promise<WorkflowExecution[]> {
    const response: any = await instance.get(`/workflow/${workflowId}/executions`, { timeout: 30000 })
    return (Array.isArray(response) ? response : []).map(normalizeExecution)
  },

  async executeWorkflow(id: string, data?: any): Promise<WorkflowExecution> {
    const response: any = await instance.post(`/workflow/${id}/execute`, data || {}, { timeout: 30000 })
    return normalizeExecution(response)
  },

  async rollbackWorkflow(id: string, version: number): Promise<any> {
    const response: any = await instance.post(`/workflow/${id}/rollback/${version}`, undefined, { timeout: 30000 })
    const data = response?.data ?? response
    return {
      workflow: data.workflow ? normalizeWorkflow(data.workflow) : null,
      nodes: data.nodes || [],
      connections: data.connections || [],
    }
  },

  async getWorkflowEditor(id: string): Promise<any> {
    const response: any = await instance.get(`/workflow/${id}/editor`, { timeout: 30000 })
    const data = response?.data ?? response
    return {
      workflow: data.workflow ? normalizeWorkflow(data.workflow) : null,
      nodes: data.nodes || [],
      connections: data.connections || [],
    }
  },

  async updateWorkflowEditor(id: string, data: any): Promise<any> {
    return instance.put(`/workflow/${id}/editor`, data, { timeout: 30000 })
  },

  async createWorkflowExecutionLog(id: string, data: any): Promise<any> {
    return instance.post(`/workflow/${id}/execution-log`, data, { timeout: 30000 })
  },

  async getWorkflowExecutionLogs(id: string, params?: any): Promise<any> {
    return instance.get(`/workflow/${id}/execution-log`, { params, timeout: 30000 })
  },

  async getWorkflowNodeStats(id: string): Promise<any[]> {
    const response: any = await instance.get(`/workflow/${id}/node-stats`, { timeout: 30000 })
    return Array.isArray(response) ? response : response?.data || []
  },

  async getWorkflowReport(id: string): Promise<any> {
    return instance.get(`/workflow/${id}/report`, { timeout: 30000 })
  },

  async exportExecutionLogs(id: string, params?: { contact?: string; status?: string }): Promise<Blob> {
    const response: any = await instance.get(`/workflow/${id}/export`, {
      params,
      responseType: 'blob',
      timeout: 30000,
    })
    return response instanceof Blob ? response : response?.data
  },

  async exportWorkflowExecutionLogs(id: string, params?: { contact?: string; status?: string }): Promise<Blob> {
    return this.exportExecutionLogs(id, params)
  },
}
