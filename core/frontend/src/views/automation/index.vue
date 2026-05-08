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

    <!-- Tabs -->
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

    <!-- Loading State -->
    <div v-if="loading" class="loading">
      <div class="spinner"></div>
      <p>Loading workflows...</p>
    </div>

    <!-- Workflows List -->
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
          <button class="btn-sm btn-secondary" @click="viewDiagram(workflow)">
            🧩 View Flow
          </button>
          <button class="btn-sm btn-secondary" @click="$router.push(`/workflow-editor/${workflow.id}`)">
            🔧 Edit Flow
          </button>
          <button class="btn-sm btn-secondary" @click="editWorkflow(workflow)">
            ✏️ Edit
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

    <!-- Stats Modal -->
    <div v-if="showStatsModal" class="modal">
      <div class="modal-content">
        <div class="modal-header">
          <h2>Workflow Statistics</h2>
          <button class="close-btn" @click="showStatsModal = false">✕</button>
        </div>
        <div class="modal-body">
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
        </div>
      </div>
    </div>

    <!-- Versions Modal -->
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
              <button
                v-if="version.status !== 'active'"
                class="btn-sm btn-secondary"
                @click="rollbackWorkflow(selectedWorkflowId, version.version)"
              >
                Rollback
              </button>
            </div>
          </div>
          <div v-else class="empty-state">
            <p>No versions found</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Diagram Modal -->
    <div v-if="showDiagramModal" class="modal">
      <div class="modal-content">
        <div class="modal-header">
          <h2>Workflow Diagram</h2>
          <button class="close-btn" @click="showDiagramModal = false">✕</button>
        </div>
        <div class="modal-body">
          <p class="diagram-header">Example flow for {{ selectedDiagram?.name || 'this workflow' }}</p>
          <div class="diagram-row">
            <div class="diagram-item">
              <strong>Trigger</strong>
              <p>Start event</p>
            </div>
            <div class="diagram-connector">→</div>
            <div class="diagram-item">
              <strong>Send Email</strong>
              <p>Deliver message</p>
            </div>
            <div class="diagram-connector">→</div>
            <div class="diagram-item">
              <strong>Delay</strong>
              <p>Pause before next step</p>
            </div>
            <div class="diagram-connector">→</div>
            <div class="diagram-item">
              <strong>Condition</strong>
              <p>Check recipient response</p>
            </div>
            <div class="diagram-connector">→</div>
            <div class="diagram-item">
              <strong>Action</strong>
              <p>Run next task</p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Create/Edit Dialog -->
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
import type { Workflow, WorkflowStatistics, WorkflowVersion } from '@/types/workflow';
import { workflowApi } from '@/api/workflow';

// State
const workflows = ref<Workflow[]>([]);
const loading = ref(false);
const activeTab = ref('All');
const tabs = ref(['All', 'Active', 'Inactive']);

const showCreateDialog = ref(false);
const showStatsModal = ref(false);
const showVersionsModal = ref(false);
const showDiagramModal = ref(false);
const isEditMode = ref(false);
const selectedWorkflowForEdit = ref<Workflow | null>(null);

const selectedStats = ref<WorkflowStatistics | null>(null);
const selectedVersions = ref<WorkflowVersion[]>([]);
const selectedWorkflowId = ref<string>('');
const selectedDiagram = ref<Workflow | null>(null);

const formData = ref({
  name: '',
  description: '',
});

// Computed
const filteredWorkflows = computed(() => {
  if (activeTab.value === 'Active') {
    return workflows.value.filter((w) => w.isActive);
  } else if (activeTab.value === 'Inactive') {
    return workflows.value.filter((w) => !w.isActive);
  }
  return workflows.value;
});

// Methods
const loadWorkflows = async () => {
  loading.value = true;
  try {
    workflows.value = await workflowApi.getWorkflows();
  } catch (error) {
    console.error('Failed to load workflows:', error);
  } finally {
    loading.value = false;
  }
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
    const newWorkflow = await workflowApi.createWorkflow({
      name: formData.value.name,
      description: formData.value.description,
      nodes: [],
      connections: [],
    });
    workflows.value.push(newWorkflow);
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
  console.log('[UI Toggle] Toggling workflow', id, 'to active:', isActive)
  try {
    const updated = await workflowApi.toggleWorkflow(id, isActive);
    console.log('[UI Toggle] Received updated workflow:', updated)
    // Update the workflow in the array using map to ensure reactivity
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

const viewStats = async (id: string) => {
  try {
    selectedStats.value = await workflowApi.getWorkflowStats({ id });
    showStatsModal.value = true;
  } catch (error) {
    console.error('Failed to load statistics:', error);
  }
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

const viewDiagram = (workflow: Workflow) => {
  selectedDiagram.value = workflow;
  showDiagramModal.value = true;
};

const rollbackWorkflow = async (id: string, version: number) => {
  if (!confirm(`Rollback to version ${version}?`)) return;
  try {
    const updated = await workflowApi.rollbackWorkflow(id, version);
    const index = workflows.value.findIndex((w) => w.id === id);
    if (index !== -1) {
      workflows.value[index] = updated;
    }
    showVersionsModal.value = false;
  } catch (error) {
    console.error('Failed to rollback workflow:', error);
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

// Lifecycle
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
