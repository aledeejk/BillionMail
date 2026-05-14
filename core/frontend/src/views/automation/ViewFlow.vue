<template>
  <div class="view-flow-page">
    <div class="view-flow-header">
      <button @click="router.back()" class="btn-secondary">← Back</button>
      <h1>Workflow Flow: {{ workflow?.name || 'Workflow' }}</h1>
    </div>

    <div class="view-flow-content">
      <div v-if="loading" class="state-box">Loading workflow scheme...</div>
      <div v-else-if="nodes.length === 0" class="state-box">The scheme is not configured</div>
      <div v-else class="flow-list">
        <div v-for="(node, index) in orderedNodes" :key="node.id" class="flow-step-wrap">
          <div class="flow-step">
            <strong>{{ getNodeTitle(node) }}</strong>
            <span v-if="getNodeDetails(node)">: {{ getNodeDetails(node) }}</span>
          </div>
          <div v-if="index < orderedNodes.length - 1" class="flow-arrow">↓</div>
        </div>
      </div>

      <div v-if="unorderedNodes.length > 0" class="flow-section">
        <h2>Other nodes</h2>
        <ul>
          <li v-for="node in unorderedNodes" :key="node.id">
            <strong>{{ getNodeTitle(node) }}</strong>
            <span v-if="getNodeDetails(node)">: {{ getNodeDetails(node) }}</span>
          </li>
        </ul>
      </div>

      <div v-if="connections.length > 0" class="flow-section">
        <h2>Connections</h2>
        <ul>
          <li v-for="connection in connections" :key="connection.id">
            {{ connection.source }} → {{ connection.target }}
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { workflowApi } from '@/api/workflow'
import type { Workflow } from '@/types/workflow'

const route = useRoute()
const router = useRouter()

const workflow = ref<Workflow | null>(null)
const loading = ref(false)
const nodes = ref<any[]>([])
const connections = ref<any[]>([])

const nodeById = computed(() => new Map(nodes.value.map(node => [node.id, node])))

const orderedNodes = computed(() => {
  if (nodes.value.length === 0) return []

  const targets = new Set(connections.value.map(connection => connection.target))
  const startNode = nodes.value.find(node => !targets.has(node.id)) || nodes.value[0]
  const result: any[] = []
  const visited = new Set<string>()
  let current: any | undefined = startNode

  while (current && !visited.has(current.id)) {
    result.push(current)
    visited.add(current.id)
    const nextConnection = connections.value.find(connection => connection.source === current.id && !visited.has(connection.target))
    current = nextConnection ? nodeById.value.get(nextConnection.target) : undefined
  }

  return result
})

const unorderedNodes = computed(() => {
  const orderedIds = new Set(orderedNodes.value.map(node => node.id))
  return nodes.value.filter(node => !orderedIds.has(node.id))
})

const getNodeLabel = (type: string) => {
  switch (type) {
    case 'trigger':
      return 'Trigger'
    case 'send-email':
      return 'Send Email'
    case 'delay':
      return 'Delay'
    case 'condition':
      return 'Condition'
    case 'action':
      return 'Action'
    default:
      return type
  }
}

const getNodeTitle = (node: any) => getNodeLabel(node.type)

const getNodeDetails = (node: any) => {
  const config = node.config || {}

  if (node.type === 'trigger') return config.triggerType || config.event || 'group subscription'
  if (node.type === 'send-email') return config.templateId || config.template || 'Welcome letter'
  if (node.type === 'delay') return `${config.duration || 3} ${config.unit || 'days'}`
  if (node.type === 'condition') return config.condition || config.operator || 'is the email open?'
  if (node.type === 'action') return config.action || config.actionType || 'add the onboarding tag'

  return Object.keys(config).length > 0 ? JSON.stringify(config) : ''
}

const loadWorkflowFlow = async () => {
  loading.value = true
  try {
    const data = await workflowApi.getWorkflowEditor(route.params.id as string)
    workflow.value = data.workflow
    nodes.value = data.nodes || []
    connections.value = data.connections || []
  } catch (error) {
    console.error('Failed to load workflow flow:', error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadWorkflowFlow()
})
</script>

<style scoped>
.view-flow-page {
  min-height: 100vh;
  background: #f5f5f5;
}

.view-flow-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem;
  background: #fff;
  border-bottom: 1px solid #ddd;
}

.view-flow-header h1 {
  margin: 0;
  color: #222;
}

.view-flow-content {
  max-width: 860px;
  margin: 2rem auto;
  padding: 2rem;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.btn-secondary {
  padding: 0.5rem 1rem;
  border: 1px solid #ccc;
  background: #f8f9fa;
  border-radius: 4px;
  cursor: pointer;
}

.state-box {
  padding: 2rem;
  text-align: center;
  color: #666;
  background: #f8f9fa;
  border-radius: 8px;
}

.flow-list {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.flow-step-wrap {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 100%;
}

.flow-step {
  min-width: 300px;
  max-width: 620px;
  padding: 1rem;
  background: #f8f9fa;
  border: 1px solid #dee2e6;
  border-radius: 8px;
  text-align: center;
  color: #333;
}

.flow-arrow {
  padding: 0.5rem 0;
  color: #007bff;
  font-size: 1.5rem;
}

.flow-section {
  margin-top: 2rem;
}

.flow-section h2 {
  margin: 0 0 0.75rem;
  color: #333;
}

.flow-section ul {
  margin: 0;
  padding-left: 1.25rem;
  color: #555;
}

.flow-section li {
  margin-bottom: 0.5rem;
}
</style>
