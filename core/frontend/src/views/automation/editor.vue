<template>
  <div class="workflow-editor">
    <div class="editor-header">
      <h1>Workflow Editor: {{ workflow?.name }}</h1>
      <div class="actions">
        <button @click="$router.back()" class="btn-secondary">Back</button>
        <button @click="saveWorkflow" class="btn-primary">Save</button>
      </div>
    </div>

    <div class="editor-content">
      <!-- Node Palette -->
      <div class="node-palette">
        <h3>Available Nodes</h3>
        <div class="node-item" draggable @dragstart="onDragStart($event, 'trigger')">
          <strong>Trigger</strong>
          <p>Start event</p>
        </div>
        <div class="node-item" draggable @dragstart="onDragStart($event, 'send-email')">
          <strong>Send Email</strong>
          <p>Deliver message</p>
        </div>
        <div class="node-item" draggable @dragstart="onDragStart($event, 'delay')">
          <strong>Delay</strong>
          <p>Pause execution</p>
        </div>
        <div class="node-item" draggable @dragstart="onDragStart($event, 'condition')">
          <strong>Condition</strong>
          <p>Check condition</p>
        </div>
        <div class="node-item" draggable @dragstart="onDragStart($event, 'action')">
          <strong>Action</strong>
          <p>Run action</p>
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
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { VueFlow, Background, Controls, MiniMap } from '@vue-flow/core'
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

const loadWorkflow = async () => {
  const id = route.params.id as string
  try {
    workflow.value = await workflowApi.getWorkflow(id)
    if (workflow.value) {
      // Load nodes and edges from workflow
      nodes.value = workflow.value.nodes?.map((node: any) => ({
        id: node.id,
        type: 'default',
        position: { x: node.position?.x || node.position_x || 0, y: node.position?.y || node.position_y || 0 },
        data: { label: node.type },
      })) || []
      edges.value = workflow.value.connections?.map((conn: any) => ({
        id: conn.id,
        source: conn.source,
        target: conn.target,
      })) || []
    }
  } catch (error) {
    console.error('Failed to load workflow:', error)
  }
}

const onDragStart = (event: DragEvent, type: string) => {
  draggedType.value = type
}

const onDragOver = (event: DragEvent) => {
  event.preventDefault()
}

const onDrop = (event: DragEvent) => {
  event.preventDefault()
  const position = {
    x: event.offsetX,
    y: event.offsetY,
  }
  const newNode: Node = {
    id: `${draggedType.value}-${Date.now()}`,
    type: 'default',
    position,
    data: { label: draggedType.value },
  }
  nodes.value.push(newNode)
}

const onConnect = (params: any) => {
  const newEdge: Edge = {
    id: `edge-${Date.now()}`,
    source: params.source,
    target: params.target,
  }
  edges.value.push(newEdge)
}

const saveWorkflow = async () => {
  if (!workflow.value) return
  try {
    const updatedNodes = nodes.value.map(node => ({
      id: node.id,
      type: node.data.label,
      position_x: node.position.x,
      position_y: node.position.y,
    }))
    const updatedConnections = edges.value.map(edge => ({
      id: edge.id,
      source: edge.source,
      target: edge.target,
    }))
    await workflowApi.updateWorkflow({
      id: workflow.value.id,
      nodes: updatedNodes,
      connections: updatedConnections,
    })
    alert('Workflow saved successfully!')
  } catch (error) {
    console.error('Failed to save workflow:', error)
  }
}

onMounted(() => {
  loadWorkflow()
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
  background: white;
  border-bottom: 1px solid #ddd;
}

.editor-content {
  flex: 1;
  display: flex;
}

.node-palette {
  width: 250px;
  padding: 1rem;
  background: #f5f5f5;
  border-right: 1px solid #ddd;
}

.node-item {
  padding: 1rem;
  margin-bottom: 0.5rem;
  background: white;
  border: 1px solid #ddd;
  border-radius: 4px;
  cursor: grab;
}

.node-item:hover {
  background: #f0f0f0;
}

.flow-canvas {
  flex: 1;
  height: 100%;
}

.btn-primary, .btn-secondary {
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.btn-primary {
  background: #007bff;
  color: white;
}

.btn-secondary {
  background: #6c757d;
  color: white;
}
</style>