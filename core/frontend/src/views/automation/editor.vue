<template>
  <div class="workflow-editor">
    <div class="editor-header">
      <button @click="router.back()" class="btn-secondary">← Back</button>
      <h1>Workflow Editor: {{ workflow?.name }}</h1>
      <div class="toolbar-buttons">
        <button type="button" @click="clearCanvas" class="btn-danger">Clear All</button>
        <button type="button" @click="saveWorkflow()" class="btn-primary">Save</button>
      </div>
    </div>

    <div class="editor-content">
      <!-- Node Palette -->
      <div class="node-palette">
        <h3>Available Nodes</h3>
        <div class="node-item" draggable="true" @dragstart="onDragStart($event, 'trigger')">
          <strong>Trigger</strong>
          <p>Start event</p>
        </div>
        <div class="node-item" draggable="true" @dragstart="onDragStart($event, 'send-email')">
          <strong>Send Email</strong>
          <p>Deliver message</p>
        </div>
        <div class="node-item" draggable="true" @dragstart="onDragStart($event, 'delay')">
          <strong>Delay</strong>
          <p>Pause execution</p>
        </div>
        <div class="node-item" draggable="true" @dragstart="onDragStart($event, 'condition')">
          <strong>Condition</strong>
          <p>Check condition</p>
        </div>
        <div class="node-item" draggable="true" @dragstart="onDragStart($event, 'action')">
          <strong>Action</strong>
          <p>Run action</p>
        </div>

        <div class="help-box">
          <h4>How the workflow works</h4>
          <ul>
            <li>Workflow starts from Trigger node.</li>
            <li>Follows arrows to next nodes.</li>
            <li>Delay waits specified time.</li>
            <li>Condition checks data and branches.</li>
            <li>Action modifies contact tags/groups.</li>
          </ul>
        </div>
      </div>

      <!-- Vue Flow Canvas -->
      <div class="flow-canvas">
        <VueFlow
          v-model:nodes="nodes"
          v-model:edges="edges"
          @drop="onDrop"
          @dragover="onDragOver"
          @connect="onConnect"
          @node-click="onNodeClick"
          :fit-view-on-init="true"
          :nodes-draggable="true"
          :nodes-connectable="true"
          :edges-updatable="true"
        >
          <Background />
          <Controls />
          <MiniMap />
        </VueFlow>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { VueFlow } from '@vue-flow/core'
import { Background } from '@vue-flow/background'
import { Controls } from '@vue-flow/controls'
import { MiniMap } from '@vue-flow/minimap'
import '@vue-flow/core/dist/style.css'
import type { Node, Edge } from '@vue-flow/core'
import { workflowApi } from '@/api/workflow'
import type { Workflow } from '@/types/workflow'

const route = useRoute()
const router = useRouter()

const workflow = ref<Workflow | null>(null)
const nodes = ref<Node[]>([])
const edges = ref<Edge[]>([])
const draggedType = ref<string>('')
const selectedNodeId = ref<string | null>(null)

const loadWorkflowData = async () => {
  const id = route.params.id as string

  try {
    const response = await workflowApi.getWorkflowEditor(id)
    const data = response.workflow ? response : response.data || response

    workflow.value = data.workflow
    nodes.value = (data.nodes || []).map((n: any) => ({
      id: n.id,
      type: 'default',
      position: { x: n.position_x || 0, y: n.position_y || 0 },
      data: {
        label: n.type,
        nodeType: n.type,
        config: n.config || {},
      },
    }))
    edges.value = (data.connections || []).map((c: any) => ({
      id: c.id,
      source: c.source,
      target: c.target,
    }))
  } catch (error) {
    console.error('Failed to load workflow editor data:', error)
    alert('Failed to load workflow editor data')
  }
}

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

const createNodeData = (nodeType: string) => {
  const defaultConfig: any = {
    trigger: { event: 'subscribe' },
    'send-email': { template: '' },
    delay: { duration: 1, unit: 'days' },
    condition: { operator: 'equals', value: '' },
    action: { actionType: 'add-tag', tag: '' },
  }

  return {
    label: getNodeLabel(nodeType),
    nodeType,
    config: defaultConfig[nodeType] || {},
  }
}

const loadDefaultNode = (nodeType: string, position: { x: number; y: number }) => {
  return {
    id: `${nodeType}-${Date.now()}`,
    type: 'default',
    position,
    data: createNodeData(nodeType),
  }
}

const openNodeConfig = (node: Node) => {
  const type = node.data?.nodeType || node.type
  const config = { ...(node.data?.config || {}) }
  let label = getNodeLabel(type)

  if (type === 'delay') {
    const duration = prompt('Delay duration in days:', String(config.duration || 1))
    if (duration !== null) {
      config.duration = Number(duration) || 1
      config.unit = 'days'
      label = `Delay ${config.duration} ${config.unit}`
    }
  } else if (type === 'condition') {
    const condition = prompt('Condition expression:', config.condition || 'email_opened == true')
    if (condition !== null) {
      config.condition = condition
      label = `Condition: ${condition}`
    }
  } else if (type === 'action') {
    const action = prompt('Action (for example, add_tag:VIP):', config.action || 'add_tag:VIP')
    if (action !== null) {
      config.action = action
      label = `Action: ${action}`
    }
  } else if (type === 'send-email') {
    const template = prompt('Email template ID:', config.templateId || '')
    if (template !== null) {
      config.templateId = template
      label = template ? `Send Email: ${template}` : label
    }
  } else if (type === 'trigger') {
    const triggerType = prompt('Trigger type (group_subscription / date / event):', config.triggerType || 'group_subscription')
    if (triggerType !== null) {
      config.triggerType = triggerType
      label = `Trigger: ${triggerType}`
    }
  }

  node.data = {
    ...node.data,
    config,
    label,
  }
}

const onDragStart = (event: DragEvent, type: string) => {
  draggedType.value = type
  if (event.dataTransfer) {
    event.dataTransfer.setData('application/json', JSON.stringify({ type }))
    event.dataTransfer.effectAllowed = 'move'
  }
}

const onDragOver = (event: DragEvent) => {
  event.preventDefault()
}

const onDrop = (event: DragEvent) => {
  event.preventDefault()
  const payload = event.dataTransfer?.getData('application/json')
  let nodeType = draggedType.value || 'trigger'

  if (payload) {
    try {
      const data = JSON.parse(payload)
      nodeType = data.type || nodeType
    } catch {
      // ignore malformed data
    }
  }

  const target = event.currentTarget as HTMLElement
  const bounds = target.getBoundingClientRect()
  const position = {
    x: event.clientX - bounds.left,
    y: event.clientY - bounds.top,
  }

  const newNode: Node = loadDefaultNode(nodeType, position)
  nodes.value.push(newNode)
  openNodeConfig(newNode)
}

const onNodeClick = (event: any) => {
  const clickedNode = event.node || event
  const node = nodes.value.find((n: any) => n.id === clickedNode.id)
  if (!node) return

  selectedNodeId.value = node.id
  openNodeConfig(node)
}

const onConnect = (params: any) => {
  const newEdge: Edge = {
    id: `edge-${Date.now()}`,
    source: params.source,
    target: params.target,
  }
  edges.value.push(newEdge)
}

const deleteSelected = () => {
  const selectedNodes = nodes.value.filter(n => (n as any).selected)
  const selectedEdges = edges.value.filter(e => (e as any).selected)
  const selectedNodeIds = new Set(selectedNodes.map(n => n.id))
  const selectedEdgeIds = new Set(selectedEdges.map(e => e.id))

  nodes.value = nodes.value.filter(n => !selectedNodeIds.has(n.id))
  edges.value = edges.value.filter(
    e => !selectedEdgeIds.has(e.id)
      && !selectedNodeIds.has(e.source)
      && !selectedNodeIds.has(e.target),
  )
}

const onKeyDown = (event: KeyboardEvent) => {
  const target = event.target as HTMLElement | null
  const isEditingText = target?.tagName === 'INPUT'
    || target?.tagName === 'TEXTAREA'
    || target?.isContentEditable

  if (isEditingText) return

  if (event.key === 'Delete' || event.key === 'Backspace') {
    const selectedNodes = nodes.value.filter(n => (n as any).selected)
    const selectedEdges = edges.value.filter(e => (e as any).selected)

    if (selectedNodes.length === 0 && selectedEdges.length === 0) return

    deleteSelected()
    event.preventDefault()
  }
}

const clearCanvas = () => {
  if (confirm('Remove all nodes and edges? This cannot be undone.')) {
    nodes.value = []
    edges.value = []
  }
}

const saveWorkflow = async () => {
  if (!workflow.value) {
    alert('No workflow loaded')
    return
  }

  const payload = {
    nodes: nodes.value.map((n: any) => ({
      id: n.id,
      type: n.data.nodeType || n.type,
      config: n.data.config || {},
      position_x: n.position.x,
      position_y: n.position.y,
    })),
    connections: edges.value.map((e: any) => ({
      id: e.id,
      source: e.source,
      target: e.target,
    })),
  }

  try {
    await workflowApi.updateWorkflowEditor(workflow.value.id, payload)
    alert('Saved')
  } catch (error) {
    console.error('Failed to save workflow:', error)
    const message = error instanceof Error ? error.message : 'See console for details'
    alert(`Failed to save workflow: ${message}`)
  }
}

onMounted(() => {
  window.addEventListener('keydown', onKeyDown)
  loadWorkflowData()
})

onUnmounted(() => {
  window.removeEventListener('keydown', onKeyDown)
})
</script>

<style scoped>
.workflow-editor {
  height: 100vh;
  display: flex;
  flex-direction: column;
}

.editor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  background: var(--bg-primary, #fff);
  border-bottom: 1px solid var(--border-color, #ddd);
}

.editor-header h1 {
  color: var(--text-primary, #222);
  margin: 0;
}

.editor-content {
  flex: 1;
  display: flex;
}

.node-palette {
  width: 250px;
  padding: 1rem;
  background-color: var(--bg-secondary, #f5f5f5);
  color: var(--text-primary, #333);
  border-right: 1px solid var(--border-color, #ddd);
}

.node-item {
  padding: 1rem;
  margin-bottom: 0.5rem;
  background: var(--card-bg, #fff);
  color: var(--text-primary, #222);
  border: 1px solid var(--border-color, #ddd);
  border-radius: 4px;
  cursor: grab;
}

.node-item:hover {
  background: var(--hover-bg, #f0f0f0);
}

.help-box {
  margin-top: 1rem;
  padding: 0.75rem;
  background: var(--bg-secondary, #f9f9f9);
  color: var(--text-secondary, #555);
  border: 1px solid var(--border-color, #ddd);
  border-radius: 6px;
  font-size: 0.9rem;
  line-height: 1.4;
}

.help-box ul {
  padding-left: 1.2rem;
  margin: 0.5rem 0 0;
}

.flow-canvas {
  flex: 1;
  height: 100%;
}

.btn-primary, .btn-secondary, .btn-danger {
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.btn-primary {
  background: #007bff;
  color: white;
}

.btn-danger {
  background: #dc3545;
  color: white;
  margin-right: 0.5rem;
}

.btn-secondary {
  background: #6c757d;
  color: white;
}
</style>
