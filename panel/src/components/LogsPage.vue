<template>
  <div class="app-shell">
    <header class="panel-header">
      <div class="status">
        <button class="back-btn" @click="$emit('back')" title="Back to controls">
          <svg viewBox="0 0 24 24" fill="none" aria-hidden="true">
            <path d="M19 12H5M5 12l6 6M5 12l6-6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </button>
        <h1 class="panel-title">Application Logs</h1>
      </div>
      <span class="badge">v0.1</span>
    </header>

    <main class="panel-body logs-page">
      <div class="logs-toolbar">
        <div class="toolbar-info">
          <span class="log-count">{{ logs.length }} entries</span>
        </div>
        <button type="button" class="btn ghost" @click="clearLogs">Clear Logs</button>
      </div>

      <div class="logs-container-full">
        <div v-if="logs.length === 0" class="logs-empty">
          No logs yet
        </div>
        <div v-else class="logs-list">
          <div v-for="(log, index) in logs" :key="index" class="log-entry">
            <span class="log-time">{{ log.timestamp }}</span>
            <span class="log-message">{{ log.message }}</span>
          </div>
        </div>
      </div>
    </main>

    <footer class="panel-footer">
      <button type="button" class="btn ghost" @click="$emit('back')">Back</button>
      
    </footer>
  </div>
</template>

<script setup>
import { defineProps, defineEmits } from 'vue'

defineProps({
  logs: Array
})

const emit = defineEmits(['back', 'clear-logs'])

function clearLogs() {
  if (confirm('Are you sure you want to clear all logs?')) {
    emit('clear-logs')
  }
}
</script>

<style scoped>
.app-shell{
  height:100vh;
  max-height:100vh;
  width:min(1100px, 100%);
  margin:0 auto;
  display:flex;
  flex-direction:column;
  gap:10px;
  padding:12px 16px 16px;
  background:#f8fafc;
  font-family:system-ui, -apple-system, 'Segoe UI', Roboto, 'Helvetica Neue', Arial;
  box-sizing:border-box;
  overflow:hidden;
}

.panel-header,
.panel-footer{
  display:flex;
  align-items:center;
  justify-content:space-between;
  padding:8px 10px;
  background:#fff;
  border:1px solid #e2e8f0;
  border-radius:10px;
  box-shadow:0 1px 2px rgba(15, 23, 42, 0.04);
}

.panel-footer{gap:8px}

.panel-body{
  flex:1;
  min-height:0;
  background:#fff;
  border:1px solid #e2e8f0;
  border-radius:10px;
  padding:12px;
  display:flex;
  flex-direction:column;
  gap:12px;
  box-shadow:0 1px 2px rgba(15, 23, 42, 0.04);
  overflow:auto;
}

.panel-body.logs-page {
  gap: 8px;
}

.status{
  display:flex;
  align-items:center;
  gap:8px;
}

.back-btn {
  width: 32px;
  height: 32px;
  padding: 0;
  border: 1px solid #e2e8f0;
  background: #fff;
  border-radius: 6px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #0f172a;
  transition: all 0.15s ease;
}

.back-btn:hover {
  background: #f1f5f9;
  border-color: #cbd5e1;
}

.back-btn svg {
  width: 16px;
  height: 16px;
}

.panel-title{
  margin:0;
  font-size:14px;
  font-weight:600;
  letter-spacing:-0.01em;
}

.badge{
  font-size:10px;
  background:#f8fafc;
  color:#475569;
  padding:2px 8px;
  border-radius:6px;
  border:1px solid #e2e8f0;
  font-family:"SFMono-Regular", ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
}

.logs-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  border-bottom: 1px solid #e2e8f0;
  margin-bottom: 8px;
}

.toolbar-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.log-count {
  font-size: 12px;
  color: #64748b;
  font-weight: 600;
}

.logs-container-full {
  flex: 1;
  min-height: 0;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 12px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}

.logs-empty {
  text-align: center;
  color: #94a3b8;
  font-size: 13px;
  padding: 40px 20px;
  margin: auto;
}

.logs-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.log-entry {
  display: flex;
  gap: 12px;
  font-size: 12px;
  font-family: 'Courier New', monospace;
  padding: 8px 0;
  border-bottom: 1px solid #e2e8f0;
  align-items: flex-start;
}

.log-entry:last-child {
  border-bottom: none;
}

.log-time {
  color: #64748b;
  font-weight: 600;
  min-width: 85px;
  flex-shrink: 0;
}

.log-message {
  color: #0f172a;
  word-break: break-word;
  flex: 1;
}

.btn{
  border:1px solid #e2e8f0;
  background:#fff;
  color:#0f172a;
  padding:8px 12px;
  border-radius:8px;
  font-size:12px;
  font-weight:600;
  cursor:pointer;
  transition:background-color 0.15s ease, color 0.15s ease, border-color 0.15s ease;
}

.btn.ghost{
  background:transparent;
  color:#475569;
}

.btn:hover{background:#f1f5f9}
.btn.ghost:hover{color:#0f172a}
.btn:focus{outline:2px solid #bfdbfe;outline-offset:1px}
</style>
