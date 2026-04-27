/**
 * Workflow type definitions
 */

export interface Workflow {
  id: string;
  name: string;
  description: string;
  isActive: boolean;
  version: number;
  createdAt: string;
  updatedAt: string;
}

export interface WorkflowNode {
  id: string;
  type: 'trigger' | 'action' | 'condition' | 'delay';
  config: Record<string, any>;
  position: {
    x: number;
    y: number;
  };
}

export interface WorkflowConnection {
  id: string;
  sourceNodeId: string;
  targetNodeId: string;
  condition?: string;
}

export interface WorkflowVersion {
  version: number;
  createdAt: string;
  status: 'active' | 'inactive' | 'archived';
}

export interface WorkflowExecution {
  id: string;
  workflowId: string;
  status: 'pending' | 'running' | 'completed' | 'failed';
  startedAt: string;
  completedAt?: string;
  errorMessage?: string;
}

export interface WorkflowLog {
  id: string;
  executionId: string;
  nodeId: string;
  message: string;
  level: 'info' | 'warning' | 'error';
  timestamp: string;
}

export interface WorkflowStatistics {
  totalExecutions: number;
  successfulExecutions: number;
  failedExecutions: number;
  averageDuration: number;
  lastExecutionTime?: string;
}

export interface CreateWorkflowRequest {
  name: string;
  description: string;
  nodes: WorkflowNode[];
  connections: WorkflowConnection[];
}

export interface UpdateWorkflowRequest {
  id: string;
  name?: string;
  description?: string;
  nodes?: WorkflowNode[];
  connections?: WorkflowConnection[];
}

export interface GetWorkflowRequest {
  id: string;
}

export interface DuplicateWorkflowRequest {
  id: string;
}

export interface DeleteWorkflowRequest {
  id: string;
}

export interface ToggleWorkflowRequest {
  id: string;
  isActive: boolean;
}

export interface GetWorkflowStatsRequest {
  id: string;
}
