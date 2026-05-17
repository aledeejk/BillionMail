<template>
  <div class="automation-container">
    <div class="automation-header">
      <div>
        <h1>Automation Workflows</h1>
        <p class="subtitle">Create and manage automated email workflows</p>
      </div>
      <button class="btn-primary" @click="showCreateDialog = true">
        <span>+ New Workflow</span>
      </button>
    </div>

    <div class="tabs">
      <button
        v-for="tab in tabs"
        :key="tab"
        class="tab"
        :class="{ active: activeTab === tab }"
        @click="activeTab = tab"
      >
        {{ tab }}
      </button>
    </div>

    <div v-if="loading" class="loading">
      <div class="spinner"></div>
      <p>Loading workflows...</p>
    </div>

    <div v-else class="workflows-grid">
      <div v-if="filteredWorkflows.length === 0" class="empty-state">
        <p>No workflows {{ activeTab !== 'All' ? 'in ' + activeTab.toLowerCase() : 'yet' }}</p>
      </div>

      <div v-for="workflow in filteredWorkflows" :key="workflow.id" class="workflow-card">
        <div class="card-header">
          <h3>{{ workflow.name }}</h3>
          <span class="status" :class="workflow.isActive ? 'active' : 'inactive'">
            {{ workflow.isActive ? 'Active' : 'Inactive' }}
          </span>
        </div>

        <p class="description">{{ workflow.description }}</p>

        <div class="card-meta">
          <span class="version">v{{ workflow.version }}</span>
          <span class="date">Updated {{ formatDate(workflow.updatedAt) }}</span>
        </div>

        <div class="card-actions">
          <button class="btn-sm btn-secondary" @click="viewStats(workflow.id)">
            📊 Stats
          </button>
          <button class="btn-sm btn-secondary" @click="viewVersions(workflow.id)">
            📜 Versions
          </button>
          <button class="btn-sm btn-secondary" @click="viewExecutionLog(workflow.id)">
            📝 Execution Log
          </button>
          <button class="btn-sm btn-secondary" @click="viewReport(workflow.id)">
            📈 Reports
          </button>
          <button class="btn-sm btn-secondary" @click="router.push(`/workflow-view/${workflow.id}`)">
            🧩 View Flow
          </button>
          <button class="btn-sm btn-secondary" @click="router.push(`/workflow-editor/${workflow.id}`)">
            🔧 Edit Flow
          </button>
          <button class="btn-sm btn-secondary" @click="editWorkflow(workflow)">
            ✏️ Edit
          </button>
          <button class="btn-sm btn-success" :disabled="executingWorkflows.includes(workflow.id)" @click="runWorkflow(workflow.id)">
            {{ executingWorkflows.includes(workflow.id) ? '⏳ Running...' : '▶️ Run' }}
          </button>
          <button
            class="btn-sm btn-secondary"
            @click="toggleWorkflow(workflow.id, !workflow.isActive)"
          >
            {{ workflow.isActive ? '⏸ Pause' : '▶ Resume' }}
          </button>
          <button class="btn-sm btn-secondary" @click="duplicateWorkflow(workflow.id)">
            📋 Duplicate
          </button>
          <button class="btn-sm btn-danger" @click="deleteWorkflow(workflow.id)">
            🗑 Delete
          </button>
        </div>
      </div>
    </div>

    <div v-if="showStatsModal" class="modal">
      <div class="modal-content">
        <div class="modal-header">
          <h2>Workflow Statistics</h2>
          <button class="close-btn" @click="showStatsModal = false">✕</button>
        </div>
        <div class="modal-body">
          <div class="tabs">
            <button class="tab" :class="{ active: statsTab === 'summary' }" @click="statsTab = 'summary'">Summary</button>
            <button class="tab" :class="{ active: statsTab === 'nodes' }" @click="statsTab = 'nodes'">Node Stats</button>
          </div>
          <template v-if="statsTab === 'summary'">
          <div v-if="selectedStats" class="stats-grid">
            <div class="stat-card">
              <span class="stat-label">Total Executions</span>
              <span class="stat-value">{{ selectedStats.totalExecutions }}</span>
            </div>
            <div class="stat-card success">
              <span class="stat-label">Successful</span>
              <span class="stat-value">{{ selectedStats.successfulExecutions }}</span>
            </div>
            <div class="stat-card error">
              <span class="stat-label">Failed</span>
              <span class="stat-value">{{ selectedStats.failedExecutions }}</span>
            </div>
            <div class="stat-card">
              <span class="stat-label">Avg Duration</span>
              <span class="stat-value">{{ (selectedStats.averageDuration / 1000).toFixed(1) }}s</span>
            </div>
          </div>
          </template>
          <template v-else>
            <table class="data-table">
              <thead>
                <tr>
                  <th>Node</th>
                  <th>Type</th>
                  <th>Reached</th>
                  <th>Conversion</th>
                  <th>Avg time</th>
                </tr>
              </thead>
              <tbody>
                <tr v-if="nodeStats.length === 0">
                  <td colspan="5">No node statistics yet</td>
                </tr>
                <tr v-for="node in nodeStats" :key="node.node_id">
                  <td>{{ node.node_id }}</td>
                  <td>{{ node.node_type }}</td>
                  <td>{{ node.entered_count ?? node.reached }}</td>
                  <td>{{ Number(node.conversion_rate ?? node.conversion ?? 0).toFixed(1) }}%</td>
                  <td>{{ formatNodeDuration(node) }}</td>
                </tr>
              </tbody>
            </table>
          </template>
        </div>
      </div>
    </div>

    <div v-if="showExecutionLogModal" class="modal execution-log-modal">
      <div class="modal-content modal-wide">
        <div class="modal-header">
          <h2>Execution Log</h2>
          <button class="close-btn" @click="showExecutionLogModal = false">✕</button>
        </div>
        <div class="modal-body">
          <div class="filter-bar">
            <input v-model="logFilters.contact" placeholder="Contact email" />
            <select v-model="logFilters.status">
              <option value="">All statuses</option>
              <option value="success">Success</option>
              <option value="failed">Failed</option>
              <option value="pending">Pending</option>
              <option value="completed">Completed</option>
            </select>
            <button class="btn-sm btn-secondary" @click="loadExecutionLogs">Apply</button>
            <button class="btn-sm btn-secondary" @click="exportLogs">Export CSV</button>
          </div>
          <table class="data-table log-table">
            <thead>
              <tr>
                <th>Contact</th>
                <th>Status</th>
                <th>Started</th>
                <th>Finished</th>
                <th>Nodes</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="logsList.length === 0">
                <td colspan="5">No execution logs yet</td>
              </tr>
              <tr v-for="log in logsList" :key="log.execution_id">
                <td>{{ log.contact_email || log.contact_id }}</td>
                <td><span class="status" :class="log.status">{{ log.status }}</span></td>
                <td>{{ formatDateTime(log.started_at) }}</td>
                <td>{{ formatDateTime(log.finished_at) }}</td>
                <td>{{ formatNodeSummary(log) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <div v-if="showReportModal" class="modal">
      <div class="modal-content">
        <div class="modal-header">
          <h2>Workflow Reports</h2>
          <button class="close-btn" @click="showReportModal = false">✕</button>
        </div>
        <div class="modal-body">
          <div v-if="reportData" class="stats-grid">
            <div class="stat-card"><span class="stat-label">Sent</span><span class="stat-value">{{ reportData.sent }}</span></div>
            <div class="stat-card"><span class="stat-label">Unique opens</span><span class="stat-value">{{ reportData.unique_opens }}</span></div>
            <div class="stat-card"><span class="stat-label">Unique clicks</span><span class="stat-value">{{ reportData.unique_clicks }}</span></div>
            <div class="stat-card"><span class="stat-label">Unsubscribes</span><span class="stat-value">{{ reportData.unsubscribes }}</span></div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showVersionsModal" class="modal">
      <div class="modal-content">
        <div class="modal-header">
          <h2>Workflow Versions</h2>
          <button class="close-btn" @click="showVersionsModal = false">✕</button>
        </div>
        <div class="modal-body">
          <div v-if="selectedVersions && selectedVersions.length > 0" class="versions-list">
            <div v-for="version in selectedVersions" :key="version.version" class="version-item">
              <div class="version-info">
                <span class="version-num">v{{ version.version }}</span>
                <span class="version-date">{{ formatDate(version.createdAt) }}</span>
                <span class="version-status" :class="version.status">{{ version.status }}</span>
              </div>
              <div class="version-actions">
                <button
                  v-if="version.status !== 'active'"
                  class="btn-sm btn-secondary"
                  @click="rollbackWorkflow(selectedWorkflowId, version.version)"
                >
                  Rollback
                </button>
                <button
                  v-if="version.status !== 'active'"
                  class="btn-sm btn-danger"
                  @click="deleteWorkflowVersion(selectedWorkflowId, version.version)"
                >
                  🗑 Delete
                </button>
              </div>
            </div>
          </div>
          <div v-else class="empty-state">
            <p>No versions found</p>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showCreateDialog" class="modal">
      <div class="modal-content">
        <div class="modal-header">
          <h2>{{ isEditMode ? 'Edit Workflow' : 'Create New Workflow' }}</h2>
          <button class="close-btn" @click="closeCreateDialog">✕</button>
        </div>
        <div class="modal-body">
          <form @submit.prevent="isEditMode ? updateWorkflow() : createWorkflow()">
            <div class="form-group">
              <label>Workflow Name</label>
              <input
                v-model="formData.name"
                type="text"
                placeholder="Enter workflow name"
                required
              />
            </div>
            <div class="form-group">
              <label>Description</label>
              <textarea
                v-model="formData.description"
                placeholder="Enter workflow description"
                rows="4"
              ></textarea>
            </div>
            <div class="form-actions">
              <button type="button" class="btn-secondary" @click="closeCreateDialog">
                Cancel
              </button>
              <button type="submit" class="btn-primary">{{ isEditMode ? 'Update Workflow' : 'Create Workflow' }}</button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import type { Workflow, WorkflowStatistics, WorkflowVersion } from '@/types/workflow';
import { workflowApi } from '@/api/workflow';

const router = useRouter();

const workflows = ref<Workflow[]>([]);
const executingWorkflows = ref<string[]>([]);
const loading = ref(false);
const activeTab = ref('All');
const tabs = ref(['All', 'Active', 'Inactive']);

const showCreateDialog = ref(false);
const showStatsModal = ref(false);
const showVersionsModal = ref(false);
const showExecutionLogModal = ref(false);
const showReportModal = ref(false);
const isEditMode = ref(false);
const selectedWorkflowForEdit = ref<Workflow | null>(null);

const selectedStats = ref<WorkflowStatistics | null>(null);
const selectedVersions = ref<WorkflowVersion[]>([]);
const selectedWorkflowId = ref<string>('');
const statsTab = ref('summary');
const nodeStats = ref<any[]>([]);
const logsList = ref<any[]>([]);
const reportData = ref<any>(null);
const logFilters = ref({
  contact: '',
  status: '',
});

const formData = ref({
  name: '',
  description: '',
});

const filteredWorkflows = computed(() => {
  if (activeTab.value === 'Active') {
    return workflows.value.filter((w) => w.isActive);
  } else if (activeTab.value === 'Inactive') {
    return workflows.value.filter((w) => !w.isActive);
  }
  return workflows.value;
});

const loadWorkflows = async () => {
  loading.value = true;
  const data = await workflowApi.getWorkflows();
  workflows.value = data;
  loading.value = false;
};

const editWorkflow = (workflow: Workflow) => {
  isEditMode.value = true;
  selectedWorkflowForEdit.value = workflow;
  formData.value = {
    name: workflow.name,
    description: workflow.description,
  };
  showCreateDialog.value = true;
};

const closeCreateDialog = () => {
  showCreateDialog.value = false;
  isEditMode.value = false;
  selectedWorkflowForEdit.value = null;
  formData.value = { name: '', description: '' };
};

const createWorkflow = async () => {
  try {
    await workflowApi.createWorkflow({
      name: formData.value.name,
      description: formData.value.description,
      nodes: [],
      connections: [],
    });
    await loadWorkflows();
    closeCreateDialog();
  } catch (error) {
    console.error('Failed to create workflow:', error);
  }
};

const updateWorkflow = async () => {
  if (!selectedWorkflowForEdit.value) return;
  try {
    const updated = await workflowApi.updateWorkflow({
      id: selectedWorkflowForEdit.value.id,
      name: formData.value.name,
      description: formData.value.description,
    });
    workflows.value = workflows.value.map(w => 
      w.id === updated.id ? updated : w
    );
    closeCreateDialog();
  } catch (error) {
    console.error('Failed to update workflow:', error);
  }
};

const deleteWorkflow = async (id: string) => {
  if (!confirm('Are you sure you want to delete this workflow?')) return;
  try {
    await workflowApi.deleteWorkflow(id);
    workflows.value = workflows.value.filter((w) => w.id !== id);
  } catch (error) {
    console.error('Failed to delete workflow:', error);
  }
};

const toggleWorkflow = async (id: string, isActive: boolean) => {
  try {
    const updated = await workflowApi.toggleWorkflow(id, isActive);
    workflows.value = workflows.value.map(w => 
      w.id === id ? updated : w
    )
  } catch (error) {
    console.error('Failed to toggle workflow:', error);
  }
};

const duplicateWorkflow = async (id: string) => {
  try {
    const duplicate = await workflowApi.duplicateWorkflow(id);
    workflows.value.push(duplicate);
  } catch (error) {
    console.error('Failed to duplicate workflow:', error);
  }
};

const runWorkflow = async (id: string) => {
  if (executingWorkflows.value.includes(id)) return;
  executingWorkflows.value.push(id);
  try {
    const email = prompt('Enter contact email for test run:', 'test@example.com');
    if (!email) { executingWorkflows.value = executingWorkflows.value.filter(x => x !== id); return; }
    
    const inputDataStr = prompt('Enter test data (JSON):', '{"email_opened": true}');
    let inputData: Record<string, any> = {};
    if (inputDataStr) {
      try {
        inputData = JSON.parse(inputDataStr);
      } catch(e) {
        console.warn('Invalid JSON, using empty object');
      }
    }
    
    inputData.email = email;
    inputData.contact_email = email;
    
    const result = await workflowApi.executeWorkflow(id, {
      trigger: 'manual',
      contact_email: email,
      input_data: inputData
    });
    
    alert(`✅ Workflow executed successfully!\n📋 Execution ID: ${result.id}\n📊 Status: ${result.status}`);
    console.log('Execution result:', result);
    
    await loadWorkflows();
  } catch (error) {
    console.error('Failed to run workflow:', error);
    const errorMessage = error instanceof Error ? error.message : 'Unknown error';
    alert(`❌ Failed to run workflow: ${errorMessage}`);
  } finally {
    executingWorkflows.value = executingWorkflows.value.filter(x => x !== id);
  }
};

const viewStats = async (id: string) => {
  try {
    selectedWorkflowId.value = id;
    statsTab.value = 'summary';
    selectedStats.value = await workflowApi.getWorkflowStats({ id });
    nodeStats.value = await workflowApi.getWorkflowNodeStats(id);
    showStatsModal.value = true;
  } catch (error) {
    console.error('Failed to load statistics:', error);
  }
};

const formatUnix = (value: number) => {
  return value ? new Date(value * 1000).toLocaleString() : '-';
};

const viewExecutionLog = async (workflowId: string) => {
  try {
    selectedWorkflowId.value = workflowId;
    await loadExecutionLogs();
    showExecutionLogModal.value = true;
  } catch (error) {
    console.error('Failed to load execution logs:', error);
  }
};

const loadExecutionLogs = async () => {
  if (!selectedWorkflowId.value) return;
  const data = await workflowApi.getWorkflowExecutionLogs(selectedWorkflowId.value, {
    contact: logFilters.value.contact,
    status: logFilters.value.status,
  });
  logsList.value = data?.list || [];
};

const viewReport = async (workflowId: string) => {
  try {
    selectedWorkflowId.value = workflowId;
    reportData.value = await workflowApi.getWorkflowReport(workflowId);
    showReportModal.value = true;
  } catch (error) {
    console.error('Failed to load report:', error);
  }
};

const exportLogs = async () => {
  if (!selectedWorkflowId.value) return;
  try {
    const blob = await workflowApi.exportExecutionLogs(selectedWorkflowId.value, {
      contact: logFilters.value.contact,
      status: logFilters.value.status,
    });
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.setAttribute('download', `execution_log_${selectedWorkflowId.value}.csv`);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    window.URL.revokeObjectURL(url);
  } catch (error) {
    console.error('Failed to export logs:', error);
  }
};

const formatDateTime = (value: string | number | null) => {
  if (!value) return '-';
  if (typeof value === 'number') return formatUnix(value);
  return new Date(value).toLocaleString();
};

const formatNodeSummary = (log: any) => {
  if (!Array.isArray(log.nodes) || log.nodes.length === 0) return '-';
  return log.nodes.map((node: any) => `${node.node_type || node.node_id} (${node.status})`).join(' → ');
};

const formatNodeDuration = (node: any) => {
  if (node.avg_duration_ms !== undefined) {
    return `${(Number(node.avg_duration_ms) / 1000).toFixed(1)}s`;
  }
  return `${Number(node.average_duration_sec || 0).toFixed(1)}s`;
};

const viewVersions = async (id: string) => {
  try {
    selectedWorkflowId.value = id;
    selectedVersions.value = await workflowApi.getWorkflowVersions(id);
    showVersionsModal.value = true;
  } catch (error) {
    console.error('Failed to load versions:', error);
  }
};

const deleteWorkflowVersion = async (id: string, version: number) => {
  if (!confirm(`Delete version ${version}? This action is irreversible.`)) return;
  try {
    await workflowApi.deleteWorkflowVersion(id, version);
    selectedVersions.value = await workflowApi.getWorkflowVersions(id);
  } catch (error) {
    console.error('Failed to delete workflow version:', error);
  }
};

const rollbackWorkflow = async (id: string, version: number) => {
  if (!confirm(`Rollback to version ${version}?`)) return;
  try {
    await workflowApi.rollbackWorkflow(id, version);
    selectedVersions.value = await workflowApi.getWorkflowVersions(id);
    workflows.value = await workflowApi.getWorkflows();
    showVersionsModal.value = false;
  } catch (error) {
    console.error('Failed to rollback workflow:', error);
    alert('Failed to rollback workflow');
  }
};

const formatDate = (dateString: string) => {
  const date = new Date(dateString);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);
  const diffDays = Math.floor(diffMs / 86400000);

  if (diffMins < 1) return 'just now';
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  if (diffDays < 7) return `${diffDays}d ago`;

  return date.toLocaleDateString();
};

onMounted(() => {
  loadWorkflows();
});
</script>

<style scoped>
.automation-container {
  padding: 2rem;
  background: #f5f5f5;
  min-height: 100vh;
}

.automation-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
  background: white;
  padding: 1.5rem;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.automation-header h1 {
  margin: 0;
  font-size: 1.75rem;
  color: #333;
}

.subtitle {
  margin: 0.5rem 0 0;
  color: #666;
  font-size: 0.9rem;
}

.btn-primary {
  padding: 0.75rem 1.5rem;
  background: #007bff;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 1rem;
  cursor: pointer;
  transition: background 0.3s;
}

.btn-primary:hover {
  background: #0056b3;
}

.btn-success {
  background: #28a745;
  color: white;
}

.btn-success:hover {
  background: #1e7e34;
}

.tabs {
  display: flex;
  gap: 1rem;
  margin-bottom: 2rem;
  border-bottom: 2px solid #e0e0e0;
}

.tab {
  padding: 0.75rem 1.5rem;
  background: none;
  border: none;
  color: #666;
  cursor: pointer;
  font-size: 1rem;
  border-bottom: 3px solid transparent;
  transition: all 0.3s;
}

.tab.active {
  color: #007bff;
  border-bottom-color: #007bff;
}

.tab:hover {
  color: #333;
}

.loading {
  text-align: center;
  padding: 3rem;
  background: white;
  border-radius: 8px;
}

.spinner {
  display: inline-block;
  width: 40px;
  height: 40px;
  border: 4px solid #e0e0e0;
  border-top-color: #007bff;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.workflows-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 1.5rem;
}

.workflow-card {
  background: white;
  border-radius: 8px;
  padding: 1.5rem;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  transition: transform 0.3s, box-shadow 0.3s;
}

.workflow-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.card-header h3 {
  margin: 0;
  font-size: 1.25rem;
  color: #333;
}

.status {
  padding: 0.25rem 0.75rem;
  border-radius: 20px;
  font-size: 0.85rem;
  font-weight: 500;
}

.status.active {
  background: #d4edda;
  color: #155724;
}

.status.inactive {
  background: #f8d7da;
  color: #721c24;
}

.description {
  margin: 1rem 0;
  color: #666;
  font-size: 0.95rem;
  line-height: 1.5;
}

.card-meta {
  display: flex;
  gap: 1rem;
  margin: 1rem 0;
  font-size: 0.85rem;
  color: #999;
}

.version {
  font-weight: 600;
  color: #007bff;
}

.card-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid #e0e0e0;
}

.btn-sm {
  padding: 0.4rem 0.8rem;
  font-size: 0.85rem;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.3s;
  flex: 1;
  min-width: 80px;
}

.btn-secondary {
  background: #f0f0f0;
  color: #333;
}

.btn-secondary:hover {
  background: #e0e0e0;
}

.btn-danger {
  background: #f8d7da;
  color: #721c24;
}

.btn-danger:hover {
  background: #f5c6cb;
}

.empty-state {
  grid-column: 1 / -1;
  text-align: center;
  padding: 3rem;
  background: white;
  border-radius: 8px;
  color: #999;
}

.modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: white;
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
  max-width: 500px;
  width: 90%;
  max-height: 80vh;
  overflow-y: auto;
}

.modal-wide {
  max-width: 1000px;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid #e0e0e0;
}

.modal-header h2 {
  margin: 0;
  color: #333;
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: #999;
}

.close-btn:hover {
  color: #333;
}

.modal-body {
  padding: 1.5rem;
}

.form-group {
  margin-bottom: 1.5rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 500;
  color: #333;
}

.form-group input,
.form-group textarea {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 1rem;
  font-family: inherit;
}

.form-group input:focus,
.form-group textarea:focus {
  outline: none;
  border-color: #007bff;
  box-shadow: 0 0 0 3px rgba(0, 123, 255, 0.1);
}

.form-actions {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1rem;
}

.stat-card {
  background: #f9f9f9;
  padding: 1.5rem;
  border-radius: 6px;
  text-align: center;
  border-left: 4px solid #007bff;
}

.stat-card.success {
  border-left-color: #28a745;
}

.stat-card.error {
  border-left-color: #dc3545;
}

.stat-label {
  display: block;
  color: #999;
  font-size: 0.85rem;
  margin-bottom: 0.5rem;
}

.stat-value {
  display: block;
  font-size: 1.75rem;
  font-weight: bold;
  color: #333;
}

.filter-bar {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.filter-bar input,
.filter-bar select {
  padding: 0.5rem;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
}

.data-table th,
.data-table td {
  padding: 0.65rem;
  border-bottom: 1px solid #eee;
  text-align: left;
  font-size: 0.9rem;
}

.execution-log-modal,
.execution-log-modal *,
.execution-log-modal .log-table,
.execution-log-modal .log-table td,
.execution-log-modal .log-table th,
.execution-log-modal .log-table span,
.execution-log-modal .log-table div {
  color: #000000 !important;
}

.execution-log-modal .modal-content,
.execution-log-modal .log-table,
.execution-log-modal .log-table td,
.execution-log-modal .log-table th {
  background-color: #ffffff;
}

:global(.dark) .execution-log-modal,
:global(.dark) .execution-log-modal .modal-content,
:global(.dark) .execution-log-modal .modal-header,
:global(.dark) .execution-log-modal .modal-body,
:global(.dark) .execution-log-modal .data-table,
:global(.dark) .execution-log-modal .data-table th,
:global(.dark) .execution-log-modal .data-table td,
:global(.dark) .execution-log-modal .filter-bar,
:global([data-theme="dark"]) .execution-log-modal,
:global([data-theme="dark"]) .execution-log-modal .modal-content,
:global([data-theme="dark"]) .execution-log-modal .modal-header,
:global([data-theme="dark"]) .execution-log-modal .modal-body,
:global([data-theme="dark"]) .execution-log-modal .data-table,
:global([data-theme="dark"]) .execution-log-modal .data-table th,
:global([data-theme="dark"]) .execution-log-modal .data-table td,
:global([data-theme="dark"]) .execution-log-modal .filter-bar {
  color: #1f2937 !important;
}

:global(.dark) .execution-log-modal .modal-content,
:global([data-theme="dark"]) .execution-log-modal .modal-content {
  background: #ffffff;
}

:global(.dark) .execution-log-modal input,
:global(.dark) .execution-log-modal select,
:global([data-theme="dark"]) .execution-log-modal input,
:global([data-theme="dark"]) .execution-log-modal select {
  color: #111827 !important;
  background: #ffffff;
}

.status.success {
  background: #d4edda;
  color: #155724;
}

.status.failed {
  background: #f8d7da;
  color: #721c24;
}

.status.pending {
  background: #fff3cd;
  color: #856404;
}

.contact-path {
  margin-top: 1.25rem;
  padding: 1rem;
  background: #f8f9fa;
  border-radius: 8px;
}

.path-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.path-node {
  padding: 0.65rem 0.9rem;
  border-radius: 999px;
  background: #e9ecef;
}

.path-node.success {
  background: #d1e7dd;
  color: #0f5132;
}

.path-node.failed {
  background: #f8d7da;
  color: #842029;
}

.stub-note {
  padding: 0.75rem;
  background: #fff3cd;
  color: #664d03;
  border-radius: 6px;
}

.versions-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.version-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  background: #f9f9f9;
  border-radius: 6px;
  border-left: 4px solid #007bff;
}

.version-info {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.version-actions {
  display: flex;
  gap: 0.5rem;
}

.version-num {
  font-weight: 600;
  color: #007bff;
}

.version-date {
  color: #999;
  font-size: 0.9rem;
}

.version-status {
  padding: 0.25rem 0.5rem;
  border-radius: 3px;
  font-size: 0.8rem;
  font-weight: 500;
}

.version-status.active {
  background: #d4edda;
  color: #155724;
}

.version-status.inactive,
.version-status.archived {
  background: #f8d7da;
  color: #721c24;
}

.diagram-header {
  margin-bottom: 1rem;
  color: #333;
  font-weight: 600;
}

.diagram-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.diagram-item {
  min-width: 120px;
  flex: 1;
  padding: 1rem;
  background: #f7fbff;
  border: 1px solid #dbeafe;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.06);
  border-radius: 10px;
  text-align: center;
}

.diagram-item strong {
  display: block;
  margin-bottom: 0.5rem;
  font-size: 1rem;
}

.diagram-item p {
  margin: 0;
  color: #555;
  font-size: 0.9rem;
}

.diagram-connector {
  font-size: 2rem;
  color: #007bff;
  min-width: 36px;
  text-align: center;
}
</style>