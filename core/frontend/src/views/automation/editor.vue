<template>
  <div class="workflow-editor">
    <div class="editor-header">
      <button @click="router.back()" class="btn-secondary">← Back</button>
      <h1>Workflow Editor: {{ workflow?.name }}</h1>
      <div class="toolbar-buttons">
        <button type="button" @click="showTestModal = true" class="btn-secondary">Test</button>
        <button type="button" @click="runWorkflowExecution" class="btn-secondary">Test execution</button>
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
        <div class="node-item" draggable="true" @dragstart="onDragStart($event, 'split')">
          <strong>Split / A/B Test</strong>
          <p>Split traffic into branches</p>
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
          <div v-if="workflow" class="webhook-box">
            <strong>Webhook URL</strong>
            <code>{{ webhookUrl }}</code>
            <small>POST JSON payload to trigger this workflow externally.</small>
          </div>
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
          @edge-click="onEdgeClick"
          :fit-view-on-init="true"
          :nodes-draggable="true"
          :nodes-connectable="true"
          :edges-updatable="true"
          :delete-key-code="['Delete', 'Backspace']"
        >
          <Background />
          <Controls />
          <MiniMap />
        </VueFlow>
      </div>
    </div>

    <div v-if="showNodeModal" class="modal">
      <div class="modal-content">
        <div class="modal-header">
          <h2>{{ getNodeLabel(nodeForm.type) }} settings</h2>
          <button class="close-btn" @click="closeNodeModal">✕</button>
        </div>
        <div class="modal-body">
          <template v-if="nodeForm.type === 'trigger'">
            <label>Trigger type</label>
            <select v-model="nodeForm.config.triggerType">
              <option value="group_subscription">Group subscription</option>
              <option value="date">Date</option>
              <option value="event">Event</option>
              <option value="field_change">Field change</option>
              <option value="api">API</option>
              <option value="recurring">Recurring</option>
            </select>

            <label v-if="nodeForm.config.triggerType === 'group_subscription'">Group</label>
            <select v-if="nodeForm.config.triggerType === 'group_subscription'" v-model="nodeForm.config.groupName">
              <option value="">Select group</option>
              <option v-for="group in demoGroups" :key="group.id" :value="group.name">{{ group.name }}</option>
            </select>

            <template v-if="nodeForm.config.triggerType === 'date'">
              <label>Event type</label>
              <select v-model="nodeForm.config.eventType">
                <option value="birthday">Birthday</option>
                <option value="subscription_anniversary">Subscription anniversary</option>
                <option value="custom_date">Custom date</option>
              </select>
              <label>Date</label>
              <input v-model="nodeForm.config.date" type="date" />
            </template>

            <label v-if="nodeForm.config.triggerType === 'event'">Event type</label>
            <select v-if="nodeForm.config.triggerType === 'event'" v-model="nodeForm.config.eventType">
              <option value="">Select event</option>
              <option value="email_opened">Email opened</option>
              <option value="link_clicked">Link clicked</option>
            </select>
            <label v-if="nodeForm.config.triggerType === 'event' && nodeForm.config.eventType === 'link_clicked'">Link URL</label>
            <input v-if="nodeForm.config.triggerType === 'event' && nodeForm.config.eventType === 'link_clicked'" v-model="nodeForm.config.linkUrl" type="text" placeholder="https://example.com/page" />

            <template v-if="nodeForm.config.triggerType === 'field_change'">
              <label>Field name</label>
              <input v-model="nodeForm.config.fieldName" type="text" placeholder="status" />
              <label>Operator</label>
              <select v-model="nodeForm.config.operator">
                <option value="equals">Equals</option>
                <option value="changed">Changed</option>
              </select>
              <label v-if="nodeForm.config.operator === 'equals'">Value</label>
              <input v-if="nodeForm.config.operator === 'equals'" v-model="nodeForm.config.fieldValue" type="text" placeholder="active" />
            </template>

            <template v-if="nodeForm.config.triggerType === 'api'">
              <label>Webhook URL</label>
              <input v-model="nodeForm.config.webhookUrl" type="text" placeholder="https://example.com/webhook" />
              <label>Method</label>
              <input v-model="nodeForm.config.method" type="text" disabled />
              <label>Request body</label>
              <textarea v-model="nodeForm.config.requestBody" rows="3" placeholder="{ &quot;source&quot;: &quot;demo&quot; }"></textarea>
            </template>

            <label v-if="nodeForm.config.triggerType === 'recurring'">Interval</label>
            <select v-if="nodeForm.config.triggerType === 'recurring'" v-model="nodeForm.config.interval">
              <option value="">Select interval</option>
              <option value="daily">Daily</option>
              <option value="weekly">Weekly</option>
              <option value="monthly">Monthly</option>
            </select>
            <label v-if="nodeForm.config.triggerType === 'recurring' && nodeForm.config.interval === 'weekly'">Day of week</label>
            <select v-if="nodeForm.config.triggerType === 'recurring' && nodeForm.config.interval === 'weekly'" v-model="nodeForm.config.weekday">
              <option value="">Select day</option>
              <option value="monday">Monday</option>
              <option value="tuesday">Tuesday</option>
              <option value="wednesday">Wednesday</option>
              <option value="thursday">Thursday</option>
              <option value="friday">Friday</option>
              <option value="saturday">Saturday</option>
              <option value="sunday">Sunday</option>
            </select>
            <label v-if="nodeForm.config.triggerType === 'recurring' && nodeForm.config.interval === 'monthly'">Day of month</label>
            <select v-if="nodeForm.config.triggerType === 'recurring' && nodeForm.config.interval === 'monthly'" v-model="nodeForm.config.monthDay">
              <option value="">Select day</option>
              <option v-for="day in monthDays" :key="day" :value="String(day)">{{ day }}</option>
              <option value="last">Last day</option>
            </select>
          </template>

          <template v-else-if="nodeForm.type === 'send-email'">
            <label>Email template ID</label>
            <input v-model="nodeForm.config.templateId" type="text" placeholder="welcome-template" />
          </template>

          <template v-else-if="nodeForm.type === 'delay'">
            <label>Duration</label>
            <input v-model.number="nodeForm.config.duration" type="number" min="1" />
            <label>Unit</label>
            <select v-model="nodeForm.config.unit">
              <option value="minutes">Minutes</option>
              <option value="hours">Hours</option>
              <option value="days">Days</option>
            </select>
          </template>

          <template v-else-if="nodeForm.type === 'condition'">
            <div class="condition-builder">
              <button type="button" class="btn-secondary" @click="ensureConditionTree">Initialize condition builder</button>
              <div v-if="nodeForm.config.logic" class="condition-group">
                <div class="condition-group-header">
                  <label>Top-level logic</label>
                  <select v-model="nodeForm.config.logic">
                    <option value="AND">AND</option>
                    <option value="OR">OR</option>
                  </select>
                </div>

                <div v-for="(condition, index) in nodeForm.config.conditions" :key="index" class="condition-row">
                  <select v-model="condition.field">
                    <option v-for="field in conditionFields" :key="field.value" :value="field.value">{{ field.label }}</option>
                  </select>
                  <select v-model="condition.operator">
                    <option v-for="operator in conditionOperators" :key="operator.value" :value="operator.value">{{ operator.label }}</option>
                  </select>
                  <input v-model="condition.value" type="text" placeholder="Value" />
                  <input v-if="condition.operator === 'between'" v-model="condition.valueTo" type="text" placeholder="To" />
                  <button type="button" class="btn-danger" @click="removeCondition(nodeForm.config, index)">Delete</button>
                </div>

                <ConditionGroupEditor
                  v-for="(group, index) in nodeForm.config.groups"
                  :key="index"
                  :group="group"
                  :condition-fields="conditionFields"
                  :condition-operators="conditionOperators"
                  @remove="removeGroup(nodeForm.config, index)"
                />

                <div class="condition-actions">
                  <button type="button" class="btn-secondary" @click="addCondition(nodeForm.config)">+ Add condition</button>
                  <button type="button" class="btn-secondary" @click="addGroup(nodeForm.config)">+ Add group</button>
                </div>

                <label>JSON preview</label>
                <pre class="condition-preview">{{ JSON.stringify(nodeForm.config, null, 2) }}</pre>
              </div>
            </div>
          </template>

          <template v-else-if="nodeForm.type === 'action'">
            <label>Action type</label>
            <select v-model="nodeForm.config.actionType">
              <option value="add-tag">Add tag</option>
              <option value="remove-tag">Remove tag</option>
              <option value="move-to-group">Move to group</option>
            </select>
            <label>Action value</label>
            <input v-model="nodeForm.config.value" type="text" placeholder="onboarding" />
          </template>

          <template v-else-if="nodeForm.type === 'split'">
            <label>Branch A traffic percentage</label>
            <input v-model.number="nodeForm.config.branchA" type="number" min="0" max="100" />
            <label>Branch B traffic percentage</label>
            <input v-model.number="nodeForm.config.branchB" type="number" min="0" max="100" />
          </template>

          <div v-if="nodeConfigError" class="form-error">{{ nodeConfigError }}</div>

          <div class="form-actions">
            <button type="button" class="btn-secondary" @click="closeNodeModal">Cancel</button>
            <button type="button" class="btn-primary" @click="saveNodeConfig">Save node</button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showEdgeModal" class="modal">
      <div class="modal-content">
        <div class="modal-header">
          <h2>Transition condition</h2>
          <button class="close-btn" @click="closeEdgeModal">✕</button>
        </div>
        <div class="modal-body">
          <label>Condition / branch label</label>
          <input v-model="edgeConditionForm" type="text" placeholder="email_opened == true, A, B..." />
          <div class="form-actions">
            <button type="button" class="btn-secondary" @click="closeEdgeModal">Cancel</button>
            <button type="button" class="btn-primary" @click="saveEdgeCondition">Save condition</button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showTestModal" class="modal">
      <div class="modal-content">
        <div class="modal-header">
          <h2>Test workflow</h2>
          <button class="close-btn" @click="showTestModal = false">✕</button>
        </div>
        <div class="modal-body">
          <label>Contact email</label>
          <input v-model="testEmail" type="email" placeholder="contact@example.com" />
          <label>Events / attributes</label>
          <textarea v-model="testEvents" rows="4" placeholder="{ &quot;email_opened&quot;: true }"></textarea>
          <div v-if="testResult.length > 0" class="test-result">
            <h3>Backend execution log</h3>
            <ol>
              <li v-for="item in testResult" :key="item">{{ item }}</li>
            </ol>
          </div>
          <div class="form-actions">
            <button type="button" class="btn-secondary" @click="showTestModal = false">Close</button>
            <button type="button" class="btn-primary" @click="runWorkflowTest">Run test</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, defineComponent, ref, onMounted, onUnmounted } from 'vue'
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

const ConditionGroupEditor = defineComponent({
  name: 'ConditionGroupEditor',
  props: {
    group: { type: Object, required: true },
    conditionFields: { type: Array, required: true },
    conditionOperators: { type: Array, required: true },
  },
  emits: ['remove'],
  setup(props, { emit }) {
    const addCondition = (group: any) => {
      group.conditions.push({ field: 'email', operator: 'contains', value: '' })
    }
    const addGroup = (group: any) => {
      group.groups.push({ logic: 'AND', conditions: [], groups: [] })
    }
    const removeCondition = (group: any, index: number) => {
      group.conditions.splice(index, 1)
    }
    const removeGroup = (group: any, index: number) => {
      group.groups.splice(index, 1)
    }
    return {
      props,
      emit,
      addCondition,
      addGroup,
      removeCondition,
      removeGroup,
    }
  },
  template: `
    <div class="condition-group nested">
      <div class="condition-group-header">
        <label>Group logic</label>
        <select v-model="group.logic">
          <option value="AND">AND</option>
          <option value="OR">OR</option>
        </select>
        <button type="button" class="btn-danger" @click="emit('remove')">Delete group</button>
      </div>
      <div v-for="(condition, index) in group.conditions" :key="index" class="condition-row">
        <select v-model="condition.field">
          <option v-for="field in conditionFields" :key="field.value" :value="field.value">{{ field.label }}</option>
        </select>
        <select v-model="condition.operator">
          <option v-for="operator in conditionOperators" :key="operator.value" :value="operator.value">{{ operator.label }}</option>
        </select>
        <input v-model="condition.value" type="text" placeholder="Value" />
        <input v-if="condition.operator === 'between'" v-model="condition.valueTo" type="text" placeholder="To" />
        <button type="button" class="btn-danger" @click="removeCondition(group, index)">Delete</button>
      </div>
      <ConditionGroupEditor
        v-for="(nestedGroup, index) in group.groups"
        :key="index"
        :group="nestedGroup"
        :condition-fields="conditionFields"
        :condition-operators="conditionOperators"
        @remove="removeGroup(group, index)"
      />
      <div class="condition-actions">
        <button type="button" class="btn-secondary" @click="addCondition(group)">+ Add condition</button>
        <button type="button" class="btn-secondary" @click="addGroup(group)">+ Add group</button>
      </div>
    </div>
  `,
})

const workflow = ref<Workflow | null>(null)
const webhookUrl = computed(() => workflow.value ? `${window.location.origin}/api/webhook/trigger/${workflow.value.id}` : '')
const nodes = ref<Node[]>([])
const edges = ref<Edge[]>([])
const draggedType = ref<string>('')
const selectedNodeId = ref<string | null>(null)
const showNodeModal = ref(false)
const nodeForm = ref({
  id: '',
  type: '',
  config: {} as any,
})
const nodeConfigError = ref('')
const demoGroups = [
  { id: 'vip', name: 'VIP' },
  { id: 'new-subscribers', name: 'New Subscribers' },
  { id: 'active', name: 'Active' },
]
const monthDays = Array.from({ length: 28 }, (_, index) => index + 1)
const conditionFields = [
  { value: 'email', label: 'Email address' },
  { value: 'group', label: 'Group membership' },
  { value: 'tag', label: 'Tags' },
  { value: 'custom_attribute', label: 'Custom attribute' },
  { value: 'last_open_date', label: 'Last open date' },
  { value: 'last_click_date', label: 'Last click date' },
  { value: 'open_count', label: 'Open count' },
  { value: 'click_count', label: 'Click count' },
  { value: 'opened_in_last_days', label: 'Has opened any email in last X days' },
  { value: 'clicked_in_last_days', label: 'Has clicked any link in last X days' },
  { value: 'last_activity', label: 'Last activity date' },
]
const conditionOperators = [
  { value: 'equals', label: 'Equals (==)' },
  { value: 'not_equals', label: 'Not equals (!=)' },
  { value: 'contains', label: 'Contains' },
  { value: 'greater_than', label: 'Greater than (>)' },
  { value: 'less_than', label: 'Less than (<)' },
  { value: 'between', label: 'Between' },
]
const showEdgeModal = ref(false)
const selectedEdgeId = ref<string | null>(null)
const edgeConditionForm = ref('')
const showTestModal = ref(false)
const testEmail = ref('')
const testEvents = ref('{\n  "email_opened": true\n}')
const testResult = ref<string[]>([])
let workflowEventsChannel: BroadcastChannel | null = null

const loadWorkflowData = async () => {
  const id = route.params.id as string

  try {
    const data = await workflowApi.getWorkflowEditor(id)

    workflow.value = data.workflow
    nodes.value = (data.nodes || []).map((n: any) => ({
      id: n.id,
      type: 'default',
      position: { x: n.position_x ?? n.positionX ?? 0, y: n.position_y ?? n.positionY ?? 0 },
      data: {
        config: n.config || {},
        label: getNodeConfigLabel(n.type, n.config || {}),
        nodeType: n.type,
      },
    }))
    edges.value = (data.connections || []).map((c: any) => ({
      id: c.id,
      source: c.source,
      target: c.target,
      label: c.condition || '',
      data: {
        condition: c.condition || '',
      },
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
    case 'split':
      return 'Split / A/B Test'
    default:
      return type
  }
}

const createNodeData = (nodeType: string) => {
  const defaultConfig: any = {
    trigger: { triggerType: 'group_subscription', groupName: '' },
    'send-email': { templateId: '' },
    delay: { duration: 1, unit: 'days' },
    condition: {
      logic: 'AND',
      conditions: [{ field: 'email', operator: 'contains', value: '@gmail.com' }],
      groups: [],
    },
    action: { actionType: 'add-tag', value: '' },
    split: { branchA: 50, branchB: 50 },
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
  nodeConfigError.value = ''
  selectedNodeId.value = node.id
  nodeForm.value = {
    id: node.id,
    type,
    config,
  }
  showNodeModal.value = true
}

const getNodeConfigLabel = (type: string, config: any) => {
  if (type === 'trigger') return getTriggerLabel(config)
  if (type === 'send-email') return config.templateId ? `Send Email: ${config.templateId}` : getNodeLabel(type)
  if (type === 'delay') return `Delay ${config.duration || 1} ${config.unit || 'days'}`
  if (type === 'condition') return `Condition: ${formatConditionGroup(config)}`
  if (type === 'action') return `Action: ${config.actionType || 'add-tag'} ${config.value || ''}`.trim()
  if (type === 'split') return `Split: A ${config.branchA || 50}% / B ${config.branchB || 50}%`
  return getNodeLabel(type)
}

const closeNodeModal = () => {
  showNodeModal.value = false
  selectedNodeId.value = null
  nodeConfigError.value = ''
  nodeForm.value = {
    id: '',
    type: '',
    config: {},
  }
}

const saveNodeConfig = () => {
  if (!selectedNodeId.value) return
  const error = validateNodeConfig(nodeForm.value.type, nodeForm.value.config)
  if (error) {
    nodeConfigError.value = error
    return
  }
  nodes.value = nodes.value.map((node: any) => {
    if (node.id !== selectedNodeId.value) return node
    return {
      ...node,
      data: {
        ...node.data,
        nodeType: nodeForm.value.type,
        config: { ...nodeForm.value.config },
        label: getNodeConfigLabel(nodeForm.value.type, nodeForm.value.config),
      },
    }
  })
  closeNodeModal()
}

const ensureConditionTree = () => {
  if (nodeForm.value.type !== 'condition') return
  if (!nodeForm.value.config.logic) {
    nodeForm.value.config = {
      logic: 'AND',
      conditions: [{ field: 'email', operator: 'contains', value: '@gmail.com' }],
      groups: [],
    }
  }
  if (!Array.isArray(nodeForm.value.config.conditions)) nodeForm.value.config.conditions = []
  if (!Array.isArray(nodeForm.value.config.groups)) nodeForm.value.config.groups = []
}

const addCondition = (group: any) => {
  group.conditions.push({ field: 'email', operator: 'contains', value: '' })
}

const addGroup = (group: any) => {
  group.groups.push({ logic: 'AND', conditions: [], groups: [] })
}

const removeCondition = (group: any, index: number) => {
  group.conditions.splice(index, 1)
}

const removeGroup = (group: any, index: number) => {
  group.groups.splice(index, 1)
}

const formatCondition = (condition: any) => {
  const operatorMap: Record<string, string> = {
    equals: '==',
    not_equals: '!=',
    contains: 'contains',
    greater_than: '>',
    less_than: '<',
    between: 'between',
  }
  if (condition.operator === 'between') {
    return `${condition.field} between ${condition.value || '?'} and ${condition.valueTo || '?'}`
  }
  return `${condition.field || 'field'} ${operatorMap[condition.operator] || condition.operator || '=='} ${condition.value || '?'}`
}

const formatConditionGroup = (group: any): string => {
  if (!group?.logic) return 'not configured'
  const parts = [
    ...(group.conditions || []).map(formatCondition),
    ...(group.groups || []).map((child: any) => `(${formatConditionGroup(child)})`),
  ].filter(Boolean)
  return parts.length > 0 ? parts.join(` ${group.logic} `) : 'empty'
}

const getTriggerLabel = (config: any) => {
  if (config.triggerType === 'group_subscription') return `Group Subscription: ${config.groupName || 'Select group'}`
  if (config.triggerType === 'date') return `Date: ${config.eventType || 'custom date'} ${config.date || ''}`.trim()
  if (config.triggerType === 'event') {
    const eventLabel = config.eventType === 'link_clicked' ? 'clicking on a link' : 'opening an email'
    return `Event: ${eventLabel}${config.linkUrl ? ` (${config.linkUrl})` : ''}`
  }
  if (config.triggerType === 'field_change') {
    return `Field change: ${config.fieldName || 'field'} ${config.operator || 'changed'} ${config.operator === 'equals' ? config.fieldValue || '' : ''}`.trim()
  }
  if (config.triggerType === 'api') return `API call: ${config.webhookUrl || 'URL'}`
  if (config.triggerType === 'recurring') {
    if (config.interval === 'weekly') return `Repeat: weekly (${config.weekday || 'day'})`
    if (config.interval === 'monthly') return `Repeat: monthly (${config.monthDay || 'day'})`
    return `Repeat: ${config.interval || 'interval'}`
  }
  return 'Trigger'
}

const isValidDate = (value: string) => {
  if (!value) return false
  const date = new Date(value)
  return !Number.isNaN(date.getTime())
}

const validateNodeConfig = (type: string, config: any) => {
  if (type !== 'trigger') return ''

  if (!config.triggerType) return 'Trigger type is required'
  if (config.triggerType === 'group_subscription') {
    return demoGroups.some(group => group.name === config.groupName) ? '' : 'Select an existing group'
  }
  if (config.triggerType === 'date') {
    if (!isValidDate(config.date)) return 'Enter a valid date'
    if (config.eventType !== 'birthday') {
      const selectedDate = new Date(config.date)
      const today = new Date()
      today.setHours(0, 0, 0, 0)
      if (selectedDate < today) return 'Date cannot be in the past'
    }
    return ''
  }
  if (config.triggerType === 'event') return config.eventType ? '' : 'Event type is required'
  if (config.triggerType === 'field_change') {
    if (!config.fieldName) return 'Field name is required'
    if (config.operator === 'equals' && !config.fieldValue) return 'Value is required for equals operator'
    return ''
  }
  if (config.triggerType === 'api') {
    return /^https?:\/\//.test(config.webhookUrl || '') ? '' : 'Webhook URL must start with http:// or https://'
  }
  if (config.triggerType === 'recurring') {
    if (!config.interval) return 'Interval is required'
    if (config.interval === 'weekly' && !config.weekday) return 'Day of week is required'
    if (config.interval === 'monthly' && !config.monthDay) return 'Day of month is required'
  }
  return ''
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
    label: '',
    data: {
      condition: '',
    },
  }
  edges.value.push(newEdge)
}

const onEdgeClick = (event: any) => {
  const clickedEdge = event.edge || event
  selectedEdgeId.value = clickedEdge.id
  edgeConditionForm.value = clickedEdge.data?.condition || clickedEdge.label || ''
  showEdgeModal.value = true
}

const closeEdgeModal = () => {
  showEdgeModal.value = false
  selectedEdgeId.value = null
  edgeConditionForm.value = ''
}

const saveEdgeCondition = () => {
  if (!selectedEdgeId.value) return
  edges.value = edges.value.map((edge: any) => {
    if (edge.id !== selectedEdgeId.value) return edge
    return {
      ...edge,
      label: edgeConditionForm.value,
      data: {
        ...(edge.data || {}),
        condition: edgeConditionForm.value,
      },
    }
  })
  closeEdgeModal()
}

const runWorkflowTest = async () => {
  if (!workflow.value) {
    alert('No workflow loaded')
    return
  }
  let inputData: Record<string, any> = {}
  try {
    inputData = testEvents.value ? JSON.parse(testEvents.value) : {}
  } catch {
    alert('Events / attributes must be valid JSON')
    return
  }
  if (testEmail.value) {
    inputData.email = testEmail.value
    inputData.contact_email = testEmail.value
  }
  try {
    await saveWorkflow()
    const execution = await workflowApi.executeWorkflow(workflow.value.id, {
      trigger: 'test',
      contact_email: testEmail.value,
      input_data: inputData,
    })
    const logData = await workflowApi.getWorkflowExecutionLogs(workflow.value.id)
    const logs = Array.isArray(logData?.list) ? logData.list : []
    const currentExecution = logs.find((item: any) => String(item.execution_id) === String(execution.id))
    const nodeLogs = Array.isArray(currentExecution?.nodes) ? currentExecution.nodes : []
    testResult.value = [
      `Execution #${execution.id} finished with status ${execution.status}`,
      ...nodeLogs.map((item: any) => `${item.node_type || 'node'} ${item.node_id || ''}: ${item.status}`),
    ]
  } catch (error) {
    console.error('Workflow test execution failed:', error)
    const message = error instanceof Error ? error.message : 'See console for details'
    alert(`Workflow test execution failed: ${message}`)
  }
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
      condition: e.data?.condition || e.label || '',
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

const runWorkflowExecution = async () => {
  if (!workflow.value) {
    alert('No workflow loaded')
    return
  }
  await saveWorkflow()
  const execution = await workflowApi.executeWorkflow(workflow.value.id, {
    trigger: 'manual',
    input_data: {},
  })
  alert(`Workflow executed by backend. Execution #${execution.id} status: ${execution.status}`)
}

const onWorkflowRollbackCompleted = async (event: Event) => {
  const detail = (event as CustomEvent<{ workflowId?: string }>).detail
  const currentWorkflowId = route.params.id as string
  if (detail?.workflowId && detail.workflowId === currentWorkflowId) {
    await loadWorkflowData()
  }
}

const reloadAfterRollback = async (workflowId?: string) => {
  const currentWorkflowId = route.params.id as string
  if (workflowId && workflowId === currentWorkflowId) {
    await loadWorkflowData()
  }
}

const onWorkflowRollbackStorage = async (event: StorageEvent) => {
  if (event.key !== 'workflow-rollback-completed' || !event.newValue) return
  try {
    const detail = JSON.parse(event.newValue) as { workflowId?: string }
    await reloadAfterRollback(detail.workflowId)
  } catch (error) {
    console.error('Failed to process workflow rollback event:', error)
  }
}

onMounted(() => {
  window.addEventListener('keydown', onKeyDown)
  window.addEventListener('workflow-rollback-completed', onWorkflowRollbackCompleted)
  window.addEventListener('storage', onWorkflowRollbackStorage)
  if ('BroadcastChannel' in window) {
    workflowEventsChannel = new BroadcastChannel('workflow-events')
    workflowEventsChannel.onmessage = async event => {
      if (event.data?.type === 'workflow-rollback-completed') {
        await reloadAfterRollback(event.data.workflowId)
      }
    }
  }
  loadWorkflowData()
})

onUnmounted(() => {
  window.removeEventListener('keydown', onKeyDown)
  window.removeEventListener('workflow-rollback-completed', onWorkflowRollbackCompleted)
  window.removeEventListener('storage', onWorkflowRollbackStorage)
  workflowEventsChannel?.close()
  workflowEventsChannel = null
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

.modal {
  position: fixed;
  inset: 0;
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.45);
}

.modal-content {
  width: min(560px, 92vw);
  max-height: 90vh;
  overflow: auto;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.2);
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem;
  border-bottom: 1px solid #ddd;
}

.modal-header h2 {
  margin: 0;
  color: #222;
}

.modal-body {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  padding: 1rem;
}

.modal-body input,
.modal-body textarea {
  width: 100%;
  padding: 0.65rem;
  border: 1px solid #ccc;
  border-radius: 4px;
}

.close-btn {
  border: none;
  background: transparent;
  cursor: pointer;
  font-size: 1.25rem;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}

.form-error {
  padding: 0.65rem;
  color: #842029;
  background: #f8d7da;
  border: 1px solid #f5c2c7;
  border-radius: 4px;
}

.condition-builder {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.condition-group {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  padding: 0.75rem;
  border: 1px solid #dee2e6;
  border-radius: 6px;
  background: #f8f9fa;
}

.condition-group.nested {
  margin-left: 1rem;
  background: #fff;
}

.condition-group-header,
.condition-row,
.condition-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
}

.condition-row select,
.condition-row input,
.condition-group-header select {
  width: auto;
  min-width: 140px;
}

.condition-preview {
  max-height: 220px;
  overflow: auto;
  padding: 0.75rem;
  background: #212529;
  color: #f8f9fa;
  border-radius: 6px;
  font-size: 0.85rem;
}

.test-result {
  padding: 0.75rem;
  background: #f8f9fa;
  border-radius: 6px;
}
</style>
