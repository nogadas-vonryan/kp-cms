<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import Panel from './components/Panel.vue'
import LogsPage from './components/LogsPage.vue'

// Declare the shape of window.runtime so TypeScript knows about it
declare global {
  interface Runtime {
    EventsOn?: (event: string, handler: (...args: any[]) => void) => void
    EventsOff?: (event: string) => void
    // add other runtime methods/properties as needed
  }

  interface Window {
    runtime?: Runtime
  }
}

const currentPage = ref<'panel' | 'logs'>('panel')

// Shared state
type LogEntry = { timestamp: string; message: string }
type ServerStatus = 'stopped' | 'running' | 'error' | 'remote'
const logs = ref<LogEntry[]>([])
const maxLogs = 1000
const frontendStatus = ref<ServerStatus>('stopped')
const backendStatus = ref<ServerStatus>('stopped')

function showLogs() {
  currentPage.value = 'logs'
}

function backToPanel() {
  currentPage.value = 'panel'
}

function clearLogs() {
  logs.value = []
}

// Update status functions
function updateFrontendStatus(status: ServerStatus) {
  frontendStatus.value = status
}

function updateBackendStatus(status: ServerStatus) {
  backendStatus.value = status
}

// Listen for log events from backend - set up once at app level
onMounted(() => {
  try {
    const runtime = window.runtime
    if (runtime && typeof runtime.EventsOn === 'function') {
      runtime.EventsOn('log', (message: string) => {
        const timestamp = new Date().toLocaleTimeString()
        logs.value.push({ timestamp, message })
        // Keep only the last maxLogs entries
        if (logs.value.length > maxLogs) {
          logs.value.shift()
        }
      })
    }
  } catch (err) {
    console.error('Failed to setup log listener:', err)
  }
})

onUnmounted(() => {
  try {
    const runtime = window.runtime
    if (runtime && typeof runtime.EventsOff === 'function') {
      runtime.EventsOff('log')
    }
  } catch (err) {
    console.error('Failed to cleanup log listener:', err)
  }
})
</script>

<template>
  <Panel 
    v-if="currentPage === 'panel'" 
    :frontend-status="frontendStatus"
    :backend-status="backendStatus"
    @view-logs="showLogs"
    @update-frontend-status="updateFrontendStatus"
    @update-backend-status="updateBackendStatus"
  />
  <LogsPage 
    v-else
    :logs="logs"
    @back="backToPanel"
    @clear-logs="clearLogs"
  />
</template>
