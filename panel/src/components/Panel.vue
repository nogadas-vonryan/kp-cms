<template>
  <div class="app-shell">
    <header class="panel-header">
      <div class="status">
        <span class="status-dot"></span>
        <h1 class="panel-title">Katarungang Pambarangay</h1>
      </div>
      <span class="badge">v0.1</span>
    </header>

    <main class="panel-body">
      <!-- Unified Server Section -->
      <section class="section">
        <div class="section-top">
          <div class="section-heading">
            <svg class="heading-icon" viewBox="0 0 24 24" fill="none" aria-hidden="true">
              <rect x="4" y="6.5" width="16" height="11" rx="1.4" stroke="currentColor" stroke-width="1.4"/>
              <path d="M8 10.5h8M8 13.5h8" stroke="currentColor" stroke-width="1.4" stroke-linecap="round"/>
            </svg>
            <h2 class="section-title">Server Configuration</h2>
          </div>
          <div style="display: flex; gap: 8px;">
            <span class="status-badge" :class="statusClass(frontendStatus)">
              <span class="status-dot"></span>
              Frontend: {{ statusText(frontendStatus) }}
            </span>
            <span class="status-badge" :class="statusClass(backendStatus)">
              <span class="status-dot"></span>
              Backend: {{ statusText(backendStatus) }}
            </span>
          </div>
        </div>

        <form class="form" @submit.prevent="startServers">
          <div class="form-grid">
            <div class="span-2">
              <label class="field-label">Frontend Host</label>
              <div class="input-wrap">
                <svg class="input-icon" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                  <path d="M4 6.5h16v11H4z" stroke="currentColor" stroke-width="1.4"/>
                  <path d="M9 15.5h6" stroke="currentColor" stroke-width="1.4" stroke-linecap="round"/>
                </svg>
                <input v-model="frontendHost" type="text" placeholder="0.0.0.0" />
              </div>
            </div>
            <div>
              <label class="field-label">Frontend Port</label>
              <div class="input-wrap">
                <svg class="input-icon" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                  <circle cx="12" cy="12" r="8" stroke="currentColor" stroke-width="1.4"/>
                  <path d="M12 8v4l2.5 2.5" stroke="currentColor" stroke-width="1.4" stroke-linecap="round"/>
                </svg>
                <input v-model="frontendPort" type="text" placeholder="8081" />
              </div>
            </div>
            <div class="span-2">
              <label class="field-label">Backend Host</label>
              <div class="input-wrap">
                <svg class="input-icon" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                  <path d="M4 6.5h16v11H4z" stroke="currentColor" stroke-width="1.4"/>
                  <path d="M9 15.5h6" stroke="currentColor" stroke-width="1.4" stroke-linecap="round"/>
                </svg>
                <input v-model="backendHost" type="text" placeholder="0.0.0.0" />
              </div>
            </div>
            <div>
              <label class="field-label">Backend Port</label>
              <div class="input-wrap">
                <svg class="input-icon" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                  <circle cx="12" cy="12" r="8" stroke="currentColor" stroke-width="1.4"/>
                  <path d="M12 8v4l2.5 2.5" stroke="currentColor" stroke-width="1.4" stroke-linecap="round"/>
                </svg>
                <input v-model="backendPort" type="text" placeholder="8080" />
              </div>
            </div>
            <div>
              <label class="field-label">Admin User</label>
              <div class="input-wrap">
                <svg class="input-icon" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                  <circle cx="12" cy="8.5" r="3.2" stroke="currentColor" stroke-width="1.4"/>
                  <path d="M6.5 18.5c1.5-2 3.3-3 5.5-3s4 1 5.5 3" stroke="currentColor" stroke-width="1.4" stroke-linecap="round"/>
                </svg>
                <input v-model="user" type="text" placeholder="admin" />
              </div>
            </div>
            <div class="span-2">
              <label class="field-label">Admin Password</label>
              <div class="input-wrap">
                <svg class="input-icon" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                  <rect x="5" y="10" width="14" height="9" rx="2" stroke="currentColor" stroke-width="1.4"/>
                  <path d="M9 10V8a3 3 0 1 1 6 0v2" stroke="currentColor" stroke-width="1.4"/>
                </svg>
                <input v-model="pass" type="password" placeholder="••••••" />
              </div>
            </div>
            <div class="span-3">
              <label class="field-label">Data Path</label>
              <div class="input-wrap">
                <svg class="input-icon" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                  <path d="M4.5 7.5h6l1.5 2h7.5v8a1 1 0 0 1-1 1h-14a1 1 0 0 1-1-1v-9z" stroke="currentColor" stroke-width="1.4"/>
                </svg>
                <input v-model="dataPath" type="text" placeholder="/var/www/data" />
                <button type="button" class="btn compact" @click="selectFolder">Browse</button>
              </div>
            </div>
          </div>

          <div class="actions">
            <button type="button" class="btn ghost" @click="stopServers">Stop All</button>
            <button type="button" class="btn" @click="openFrontendInBrowser">Open Frontend</button>
            <button type="submit" class="btn primary">Start Servers</button>
          </div>
        </form>
      </section>

    </main>

    <footer class="panel-footer">
      <button type="button" class="btn ghost" @click="viewLogs">View Logs</button>
      <div style="display: flex; gap: 8px; margin-left: auto;">
        <button type="button" class="btn ghost" @click="handleReset">Reset</button>
        <button type="button" class="btn primary" @click="handleSave">Save Changes</button>
      </div>
    </footer>

    <!-- Modal Dialog -->
    <div v-if="showModal" class="modal-overlay" @click="closeModal">
      <div class="modal-dialog" @click.stop>
        <div class="modal-header">
          <h3 class="modal-title">{{ modalTitle }}</h3>
        </div>
        <div class="modal-body">
          <pre class="modal-message">{{ modalMessage }}</pre>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn primary" @click="closeModal">OK</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, defineProps, defineEmits } from 'vue'
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime'

defineProps({
  frontendStatus: String,
  backendStatus: String
})

const emit = defineEmits(['view-logs', 'update-frontend-status', 'update-backend-status'])

const frontendHost = ref('0.0.0.0')
const frontendPort = ref('8081')
const backendHost = ref('0.0.0.0')
const backendPort = ref('8080')
const user = ref('admin')
const pass = ref('')
const dataPath = ref('')

// simple applied flag for footer save button feedback
const isSaved = ref(false)

// Modal state
const showModal = ref(false)
const modalTitle = ref('')
const modalMessage = ref('')

function showMessage(title, message) {
  modalTitle.value = title
  modalMessage.value = message
  showModal.value = true
}

function closeModal() {
  showModal.value = false
}

function handleSave() {
  isSaved.value = true
  setTimeout(() => {
    isSaved.value = false
  }, 2000)
}

function viewLogs() {
  emit('view-logs')
}

function handleReset() {
  frontendHost.value = '0.0.0.0'
  frontendPort.value = '8081'
  backendHost.value = '0.0.0.0'
  backendPort.value = '8080'
  user.value = 'admin'
  pass.value = ''
  dataPath.value = ''
  emit('update-frontend-status', 'stopped')
  emit('update-backend-status', 'stopped')
}

async function selectFolder() {
  try {
    // Try canonical Wails bridge paths with safe existence checks
    const appNs = window && (window.go?.main?.App || window['go']?.['main']?.['App'])
    if (appNs && typeof appNs.SelectFolder === 'function') {
      const selection = await appNs.SelectFolder()
      if (selection) dataPath.value = selection
      return
    }
    showMessage('Not Available', 'Folder picker not available in this environment')
  } catch (err) {
    console.error(err)
    showMessage('Error', 'Failed to select folder')
  }
}

async function startServers() {
  const appNs = window && (window.go?.main?.App || window['go']?.['main']?.['App'])
  if (!appNs) {
    showMessage('Error', 'Wails API not available in this environment')
    return
  }

  let frontendStarted = false
  let backendStarted = false
  let errors = []

  // Start backend server first
  if (typeof appNs.StartBackendServer === 'function') {
    try {
      const result = await appNs.StartBackendServer(
        backendHost.value,
        parseInt(backendPort.value) || 8080,
        user.value || 'admin',
        pass.value,
        dataPath.value
      )
      // If we get here without exception, it succeeded
      backendStarted = true
      emit('update-backend-status', 'running')
    } catch (err) {
      console.error('Backend start error:', err)
      if (err && (err.message || err.toString())) {
        errors.push('Backend: ' + (err.message || err.toString()))
      } else {
        errors.push('Backend: Unknown error')
      }
      emit('update-backend-status', 'error')
    }
  }

  // Start frontend server with backend configuration
  if (typeof appNs.StartWebServer === 'function') {
    try {
      const result = await appNs.StartWebServer(
        frontendHost.value,
        parseInt(frontendPort.value) || 8081,
        backendHost.value,
        parseInt(backendPort.value) || 8080
      )
      // If we get here without exception, it succeeded
      frontendStarted = true
      emit('update-frontend-status', 'running')
    } catch (err) {
      console.error('Frontend start error:', err)
      if (err && (err.message || err.toString())) {
        errors.push('Frontend: ' + (err.message || err.toString()))
      } else {
        errors.push('Frontend: Unknown error')
      }
      emit('update-frontend-status', 'error')
    }
  }

  // Show appropriate message
  if (frontendStarted && backendStarted) {
    showMessage('Success', `Servers started successfully!\nFrontend: http://${frontendHost.value}:${frontendPort.value}\nBackend: http://${backendHost.value}:${backendPort.value}\n\nCheck the Logs tab for admin credentials if password was auto-generated.`)
  } else if (frontendStarted || backendStarted) {
    const started = []
    if (frontendStarted) started.push(`Frontend: http://${frontendHost.value}:${frontendPort.value}`)
    if (backendStarted) started.push(`Backend: http://${backendHost.value}:${backendPort.value}`)
    showMessage('Partial Success', `Partially started:\n${started.join('\n')}\n\nErrors:\n${errors.join('\n')}`)
  } else if (errors.length > 0) {
    showMessage('Error', 'Failed to start servers:\n' + errors.join('\n'))
  }
}

function openFrontendInBrowser() {
  const host = frontendHost.value || 'localhost'
  const port = parseInt(frontendPort.value, 10) || 8081
  const url = `http://${host}${port ? `:${port}` : ''}`

  try {
    if (typeof BrowserOpenURL === 'function') {
      BrowserOpenURL(url)
    } else if (window?.open) {
      window.open(url, '_blank', 'noopener,noreferrer')
    } else {
      showMessage('Not Available', 'BrowserOpenURL not available in this environment')
    }
  } catch (err) {
    console.error(err)
    showMessage('Error', 'Failed to open browser: ' + err.message)
  }
}

async function stopServers() {
  try {
    const appNs = window && (window.go?.main?.App || window['go']?.['main']?.['App'])
    if (!appNs) {
      showMessage('Error', 'Wails API not available in this environment')
      return
    }

    let errors = []

    // Stop frontend server
    if (typeof appNs.StopWebServer === 'function') {
      try {
        await appNs.StopWebServer()
        emit('update-frontend-status', 'stopped')
      } catch (err) {
        errors.push('Frontend: ' + err.message)
        emit('update-frontend-status', 'error')
      }
    }

    // Stop backend server
    if (typeof appNs.StopBackendServer === 'function') {
      try {
        await appNs.StopBackendServer()
        emit('update-backend-status', 'stopped')
      } catch (err) {
        errors.push('Backend: ' + err.message)
        emit('update-backend-status', 'error')
      }
    }

    if (errors.length > 0) {
      showMessage('Partial Success', 'Some servers failed to stop:\n' + errors.join('\n'))
    } else {
      showMessage('Success', 'All servers stopped successfully')
    }
  } catch (err) {
    console.error(err)
    showMessage('Error', 'Failed to stop servers: ' + err.message)
  }
}

function statusClass(status) {
  return {
    running: 'status-ok',
    stopped: 'status-muted',
    error: 'status-warn'
  }[status] || 'status-muted'
}

function statusText(status) {
  return {
    running: 'Running',
    stopped: 'Stopped',
    error: 'Unavailable'
  }[status] || 'Unknown'
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

.status{display:flex;align-items:center;gap:8px}
.status-dot{width:8px;height:8px;border-radius:50%;background:#22c55e;display:inline-block}
.panel-title{margin:0;font-size:14px;font-weight:600;letter-spacing:-0.01em}
.badge{
  font-size:10px;
  background:#f8fafc;
  color:#475569;
  padding:2px 8px;
  border-radius:6px;
  border:1px solid #e2e8f0;
  font-family:"SFMono-Regular", ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
}

.section{display:flex;flex-direction:column;gap:10px}
.section-top{display:flex;justify-content:space-between;align-items:center;gap:12px}
.section-heading{display:flex;align-items:center;gap:8px}
.heading-icon{width:16px;height:16px;color:#94a3b8}
.section-title{margin:0;font-size:11px;font-weight:700;letter-spacing:0.12em;text-transform:uppercase;color:#94a3b8}

.status-badge{
  display:inline-flex;
  align-items:center;
  gap:6px;
  padding:6px 10px;
  border-radius:999px;
  font-size:11px;
  font-weight:600;
  border:1px solid #e2e8f0;
  color:#334155;
  background:#f8fafc;
  text-transform:capitalize;
}

.status-badge .status-dot{width:8px;height:8px;border-radius:999px;background:#cbd5e1}
.status-badge.status-ok{border-color:#22c55e1a;background:#f0fdf4;color:#166534}
.status-badge.status-ok .status-dot{background:#22c55e}
.status-badge.status-warn{border-color:#f973161a;background:#fff7ed;color:#9a3412}
.status-badge.status-warn .status-dot{background:#f97316}
.status-badge.status-muted{border-color:#e2e8f0;background:#f8fafc;color:#475569}
.status-badge.status-muted .status-dot{background:#cbd5e1}

.divider{height:1px;background:#e2e8f0}

.form{display:flex;flex-direction:column;gap:12px}
.form-grid{display:grid;grid-template-columns:repeat(3, minmax(0, 1fr));gap:12px;align-items:start}
.span-2{grid-column:span 2}
.span-3{grid-column:span 3}

.field-label{
  display:block;
  font-size:10px;
  font-weight:700;
  text-transform:uppercase;
  letter-spacing:0.08em;
  color:#64748b;
  margin:0 0 4px 0;
}

.input-wrap{
  position:relative;
  display:flex;
  align-items:center;
  gap:8px;
  background:#f8fafc;
  border:1px solid #e2e8f0;
  border-radius:8px;
  padding:0 12px 0 36px;
  min-height:38px;
  transition:border-color 0.15s ease, background-color 0.15s ease;
}

.input-wrap:focus-within{border-color:#3b82f6;background:#fff}
.input-icon{
  position:absolute;
  left:12px;
  width:14px;
  height:14px;
  color:#94a3b8;
}

.input-wrap input{
  width:100%;
  border:none;
  outline:none;
  background:transparent;
  font-size:13px;
  color:#0f172a;
  padding:0;
  height:20px;
}

.actions{display:flex;justify-content:flex-end;gap:8px;margin-top:4px}

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

.btn.primary{
  background:#0f172a;
  border-color:#0f172a;
  color:#fff;
}

.btn.ghost{
  background:transparent;
  color:#475569;
}

.btn.compact{
  padding:6px 10px;
  font-size:11px;
  margin-left:8px;
  height:28px;
  flex-shrink:0;
}


/* Modal styles */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(15, 23, 42, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 16px;
}

.modal-dialog {
  background: #fff;
  border-radius: 10px;
  box-shadow: 0 10px 25px rgba(15, 23, 42, 0.2);
  width: 100%;
  max-width: 380px;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
  border: 1px solid #e2e8f0;
}

.modal-header {
  padding: 16px 20px;
  border-bottom: 1px solid #e2e8f0;
}

.modal-title {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: #0f172a;
}

.modal-body {
  padding: 16px 20px;
  overflow-y: auto;
  flex: 1;
  min-height: 0;
}

.modal-message {
  margin: 0;
  font-size: 13px;
  color: #334155;
  white-space: pre-wrap;
  word-wrap: break-word;
  font-family: system-ui, -apple-system, 'Segoe UI', Roboto, 'Helvetica Neue', Arial;
  line-height: 1.5;
}

.modal-footer {
  padding: 12px 20px;
  border-top: 1px solid #e2e8f0;
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
.btn:hover{background:#f1f5f9}
.btn.primary:hover{background:#111827;border-color:#111827}
.btn.ghost:hover{color:#0f172a}

.btn:focus{outline:2px solid #bfdbfe;outline-offset:1px}

.panel-footer .btn.primary{
  background:linear-gradient(120deg, #0f172a, #111827);
  flex-shrink: 0;
}

.log-message {
  color: #0f172a;
  word-break: break-word;
}
</style>
