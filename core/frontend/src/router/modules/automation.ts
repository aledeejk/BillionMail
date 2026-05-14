import { RouteRecordRaw } from 'vue-router'
import { Layout } from '@/router/constant'

const route: RouteRecordRaw = {
	path: '/automation',
	component: Layout,
	meta: {
		sort: 4,
		key: 'automation',
		title: 'Automation',
		icon: 'i-mdi-playlist-check',
		hidden: false,
	},
	children: [
		{
			path: '/automation',
			name: 'Automation',
			component: () => import('@/views/automation/index.vue'),
		},
		{
			path: '/workflow-editor/:id',
			name: 'WorkflowEditor',
			component: () => import('@/views/automation/editor.vue'),
			meta: { title: 'Workflow Editor' },
		},
		{
			path: '/workflow-view/:id',
			name: 'WorkflowViewer',
			component: () => import('@/views/automation/ViewFlow.vue'),
			meta: { title: 'Workflow Viewer' },
		},
		{
			path: '/executions',
			name: 'WorkflowExecutions',
			component: () => import('@/views/automation/Executions.vue'),
			meta: { title: 'Workflow Runs' },
		},
	],
}

export default route
