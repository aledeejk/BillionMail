<template>
  <div class="workflow-viewer">
    <div v-if="orderedNodes.length === 0" class="viewer-empty">
      The scheme is not configured
    </div>
    <div v-else class="viewer-flow">
      <div v-for="(node, index) in orderedNodes" :key="node.id" class="viewer-step-wrap">
        <div class="viewer-step">
          <div class="viewer-step-title">{{ getNodeTitle(node) }}</div>
          <div v-if="getNodeDetails(node)" class="viewer-step-details">{{ getNodeDetails(node) }}</div>
        </div>
        <div v-if="index < orderedNodes.length - 1" class="viewer-arrow">↓</div>
      </div>
    </div>

    <div v-if="unorderedNodes.length > 0" class="viewer-section">
      <h3>Other nodes</h3>
      <ul class="viewer-list">
        <li v-for="node in unorderedNodes" :key="node.id">
          <strong>{{ getNodeTitle(node) }}</strong>
          <span v-if="getNodeDetails(node)"> — {{ getNodeDetails(node) }}</span>
        </li>
      </ul>
    </div>

    <div v-if="connections.length > 0" class="viewer-section">
      <h3>Connections</h3>
      <ul class="viewer-list">
        <li v-for="connection in connections" :key="connection.id">
          {{ connection.source }} → {{ connection.target }}
        </li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  nodes: any[]
  connections: any[]
}>()

const nodeById = computed(() => new Map(props.nodes.map(node => [node.id, node])))

const orderedNodes = computed(() => {
  if (props.nodes.length === 0) return []

  const targets = new Set(props.connections.map(connection => connection.target))
  const startNode = props.nodes.find(node => !targets.has(node.id)) || props.nodes[0]
  const result: any[] = []
  const visited = new Set<string>()
  let current: any | undefined = startNode

  while (current && !visited.has(current.id)) {
    result.push(current)
    visited.add(current.id)
    const nextConnection = props.connections.find(connection => connection.source === current.id && !visited.has(connection.target))
    current = nextConnection ? nodeById.value.get(nextConnection.target) : undefined
  }

  return result
})

const unorderedNodes = computed(() => {
  const orderedIds = new Set(orderedNodes.value.map(node => node.id))
  return props.nodes.filter(node => !orderedIds.has(node.id))
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

  if (node.type === 'trigger') {
    return config.triggerType || config.event || 'group subscription'
  }
  if (node.type === 'send-email') {
    return config.templateId || config.template || 'Welcome letter'
  }
  if (node.type === 'delay') {
    const duration = config.duration || 3
    const unit = config.unit || 'days'
    return `${duration} ${unit}`
  }
  if (node.type === 'condition') {
    return config.condition || config.operator || 'is the email open?'
  }
  if (node.type === 'action') {
    return config.action || config.actionType || 'add the onboarding tag'
  }

  return Object.keys(config).length > 0 ? JSON.stringify(config) : ''
}
</script>

<style scoped>
.workflow-viewer {
  padding: 0.5rem 0;
}

.viewer-empty {
  padding: 2rem;
  color: #666;
  text-align: center;
  background: #f8f9fa;
  border-radius: 8px;
}

.viewer-flow {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
}

.viewer-step-wrap {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 100%;
}

.viewer-step {
  min-width: 260px;
  max-width: 520px;
  padding: 1rem;
  background: #f8f9fa;
  border: 1px solid #dee2e6;
  border-radius: 8px;
  text-align: center;
}

.viewer-step-title {
  font-weight: 700;
  color: #333;
}

.viewer-step-details {
  margin-top: 0.35rem;
  color: #555;
}

.viewer-arrow {
  padding: 0.35rem 0;
  color: #007bff;
  font-size: 1.5rem;
  line-height: 1;
}

.viewer-section {
  margin-top: 1.5rem;
}

.viewer-section h3 {
  margin: 0 0 0.75rem;
  color: #333;
}

.viewer-list {
  margin: 0;
  padding-left: 1.25rem;
  color: #555;
}

.viewer-list li {
  margin-bottom: 0.5rem;
}
</style>
