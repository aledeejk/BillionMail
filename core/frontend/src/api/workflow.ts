/**
 * Workflow API client with mock data
 */

import type {
  Workflow,
  WorkflowExecution,
  WorkflowStatistics,
  WorkflowVersion,
  CreateWorkflowRequest,
  UpdateWorkflowRequest,
  GetWorkflowStatsRequest,
} from '@/types/workflow';

// Mock data
const MOCK_WORKFLOWS: Workflow[] = [
  {
    id: '1',
    name: 'Newsletter Campaign',
    description: 'Send weekly newsletter to subscribers',
    isActive: true,
    version: 3,
    createdAt: '2025-04-01T10:00:00Z',
    updatedAt: '2025-04-25T15:30:00Z',
  },
  {
    id: '2',
    name: 'Welcome Email Series',
    description: 'Send welcome emails to new subscribers',
    isActive: true,
    version: 2,
    createdAt: '2025-03-15T08:00:00Z',
    updatedAt: '2025-04-20T12:00:00Z',
  },
  {
    id: '3',
    name: 'Abandoned Cart Reminder',
    description: 'Remind users about items left in cart',
    isActive: false,
    version: 1,
    createdAt: '2025-04-10T09:00:00Z',
    updatedAt: '2025-04-10T09:00:00Z',
  },
];

const MOCK_EXECUTIONS: WorkflowExecution[] = [
  {
    id: 'exec-1',
    workflowId: '1',
    status: 'completed',
    startedAt: '2025-04-27T08:00:00Z',
    completedAt: '2025-04-27T08:15:00Z',
  },
  {
    id: 'exec-2',
    workflowId: '1',
    status: 'running',
    startedAt: '2025-04-27T16:00:00Z',
  },
  {
    id: 'exec-3',
    workflowId: '2',
    status: 'completed',
    startedAt: '2025-04-27T07:00:00Z',
    completedAt: '2025-04-27T07:10:00Z',
  },
];

const MOCK_STATISTICS: Record<string, WorkflowStatistics> = {
  '1': {
    totalExecutions: 52,
    successfulExecutions: 50,
    failedExecutions: 2,
    averageDuration: 15000,
    lastExecutionTime: '2025-04-27T16:00:00Z',
  },
  '2': {
    totalExecutions: 28,
    successfulExecutions: 27,
    failedExecutions: 1,
    averageDuration: 10000,
    lastExecutionTime: '2025-04-27T07:00:00Z',
  },
  '3': {
    totalExecutions: 5,
    successfulExecutions: 3,
    failedExecutions: 2,
    averageDuration: 20000,
  },
};

const MOCK_VERSIONS: Record<string, WorkflowVersion[]> = {
  '1': [
    { version: 3, createdAt: '2025-04-25T15:30:00Z', status: 'active' },
    { version: 2, createdAt: '2025-04-20T10:00:00Z', status: 'inactive' },
    { version: 1, createdAt: '2025-04-01T10:00:00Z', status: 'archived' },
  ],
  '2': [
    { version: 2, createdAt: '2025-04-20T12:00:00Z', status: 'active' },
    { version: 1, createdAt: '2025-03-15T08:00:00Z', status: 'archived' },
  ],
};

/**
 * Workflow API Service
 */
export const workflowApi = {
  /**
   * Get all workflows
   */
  async getWorkflows(): Promise<Workflow[]> {
    // Simulate API call
    await new Promise((resolve) => setTimeout(resolve, 500));
    return MOCK_WORKFLOWS;
  },

  /**
   * Get workflow by ID
   */
  async getWorkflow(id: string): Promise<Workflow | null> {
    await new Promise((resolve) => setTimeout(resolve, 300));
    return MOCK_WORKFLOWS.find((w) => w.id === id) || null;
  },

  /**
   * Create workflow
   */
  async createWorkflow(data: CreateWorkflowRequest): Promise<Workflow> {
    await new Promise((resolve) => setTimeout(resolve, 800));
    const newWorkflow: Workflow = {
      id: String(MOCK_WORKFLOWS.length + 1),
      name: data.name,
      description: data.description,
      isActive: false,
      version: 1,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };
    MOCK_WORKFLOWS.push(newWorkflow);
    return newWorkflow;
  },

  /**
   * Update workflow
   */
  async updateWorkflow(data: UpdateWorkflowRequest): Promise<Workflow> {
    await new Promise((resolve) => setTimeout(resolve, 800));
    const workflow = MOCK_WORKFLOWS.find((w) => w.id === data.id);
    if (!workflow) throw new Error('Workflow not found');

    if (data.name) workflow.name = data.name;
    if (data.description) workflow.description = data.description;
    workflow.updatedAt = new Date().toISOString();
    workflow.version += 1;

    return workflow;
  },

  /**
   * Delete workflow
   */
  async deleteWorkflow(id: string): Promise<void> {
    await new Promise((resolve) => setTimeout(resolve, 500));
    const index = MOCK_WORKFLOWS.findIndex((w) => w.id === id);
    if (index !== -1) {
      MOCK_WORKFLOWS.splice(index, 1);
    }
  },

  /**
   * Duplicate workflow
   */
  async duplicateWorkflow(id: string): Promise<Workflow> {
    await new Promise((resolve) => setTimeout(resolve, 800));
    const original = MOCK_WORKFLOWS.find((w) => w.id === id);
    if (!original) throw new Error('Workflow not found');

    const duplicate: Workflow = {
      ...original,
      id: String(MOCK_WORKFLOWS.length + 1),
      name: `${original.name} (Copy)`,
      version: 1,
      isActive: false,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };

    MOCK_WORKFLOWS.push(duplicate);
    return duplicate;
  },

  /**
   * Toggle workflow active status
   */
  async toggleWorkflow(id: string, isActive: boolean): Promise<Workflow> {
    await new Promise((resolve) => setTimeout(resolve, 500));
    const workflow = MOCK_WORKFLOWS.find((w) => w.id === id);
    if (!workflow) throw new Error('Workflow not found');

    workflow.isActive = isActive;
    workflow.updatedAt = new Date().toISOString();
    return workflow;
  },

  /**
   * Get workflow statistics
   */
  async getWorkflowStats(request: GetWorkflowStatsRequest): Promise<WorkflowStatistics> {
    await new Promise((resolve) => setTimeout(resolve, 400));
    return (
      MOCK_STATISTICS[request.id] || {
        totalExecutions: 0,
        successfulExecutions: 0,
        failedExecutions: 0,
        averageDuration: 0,
      }
    );
  },

  /**
   * Get workflow versions
   */
  async getWorkflowVersions(id: string): Promise<WorkflowVersion[]> {
    await new Promise((resolve) => setTimeout(resolve, 400));
    return MOCK_VERSIONS[id] || [];
  },

  /**
   * Get workflow executions
   */
  async getWorkflowExecutions(workflowId: string): Promise<WorkflowExecution[]> {
    await new Promise((resolve) => setTimeout(resolve, 400));
    return MOCK_EXECUTIONS.filter((e) => e.workflowId === workflowId);
  },

  /**
   * Execute workflow
   */
  async executeWorkflow(id: string): Promise<WorkflowExecution> {
    await new Promise((resolve) => setTimeout(resolve, 600));
    const execution: WorkflowExecution = {
      id: `exec-${Date.now()}`,
      workflowId: id,
      status: 'running',
      startedAt: new Date().toISOString(),
    };
    MOCK_EXECUTIONS.push(execution);
    return execution;
  },

  /**
   * Rollback workflow to previous version
   */
  async rollbackWorkflow(id: string, version: number): Promise<Workflow> {
    await new Promise((resolve) => setTimeout(resolve, 800));
    const workflow = MOCK_WORKFLOWS.find((w) => w.id === id);
    if (!workflow) throw new Error('Workflow not found');

    workflow.version = version;
    workflow.updatedAt = new Date().toISOString();
    return workflow;
  },
};
