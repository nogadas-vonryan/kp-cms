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
      <!-- Frontend Section -->
      <section class="section">
        <div class="section-top">
          <div class="section-heading">
            <svg class="heading-icon" viewBox="0 0 24 24" fill="none" aria-hidden="true">
              <path d="M12 3.5a8.5 8.5 0 1 0 0 17 8.5 8.5 0 0 0 0-17z" stroke="currentColor" stroke-width="1.4"/>
              <path d="M3.6 9.5h16.8M3.6 14.5h16.8M12 3.5c-2 2.2-3.1 4.8-3.1 8s1.1 5.8 3.1 8c2-2.2 3.1-4.8 3.1-8s-1.1-5.8-3.1-8z" stroke="currentColor" stroke-width="1.4"/>
            </svg>
            <h2 class="section-title">Frontend Server</h2>
          </div>
          <span class="status-badge" :class="statusClass(frontendStatus)">
            <span class="status-dot"></span>
            {{ statusText(frontendStatus) }}
          </span>
        </div>

        <form class="form" @submit.prevent="startFrontend">
          <div class="form-grid">
            <div class="span-2">
              <label class="field-label">Host</label>
              <div class="input-wrap">
                <svg class="input-icon" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                  <path d="M4 6.5h16v11H4z" stroke="currentColor" stroke-width="1.4"/>
                  <path d="M9 15.5h6" stroke="currentColor" stroke-width="1.4" stroke-linecap="round"/>
                </svg>
                <input v-model="frontendHost" type="text" placeholder="localhost" />
              </div>
            </div>
            <div>
              <label class="field-label">Port</label>
              <div class="input-wrap">
                <svg class="input-icon" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                  <circle cx="12" cy="12" r="8" stroke="currentColor" stroke-width="1.4"/>
                  <path d="M12 8v4l2.5 2.5" stroke="currentColor" stroke-width="1.4" stroke-linecap="round"/>
                </svg>
                <input v-model="frontendPort" type="text" placeholder="3000" />
              </div>
            </div>
          </div>

          <div class="actions">
            <button type="button" class="btn ghost" @click="stopFrontend">Stop</button>
            <button type="submit" class="btn primary">Start</button>
          </div>
        </form>
      </section>

      <div class="divider"></div>

      <!-- Backend Section -->
      <section class="section">
        <div class="section-top">
          <div class="section-heading">
            <svg class="heading-icon" viewBox="0 0 24 24" fill="none" aria-hidden="true">
              <rect x="4" y="6.5" width="16" height="11" rx="1.4" stroke="currentColor" stroke-width="1.4"/>
              <path d="M8 10.5h8M8 13.5h8" stroke="currentColor" stroke-width="1.4" stroke-linecap="round"/>
            </svg>
            <h2 class="section-title">Backend Server</h2>
          </div>
          <span class="status-badge" :class="statusClass(backendStatus)">
            <span class="status-dot"></span>
            {{ statusText(backendStatus) }}
          </span>
        </div>

        <form class="form" @submit.prevent="startBackend">
          <div class="form-grid">
            <div class="span-2">
              <label class="field-label">Host</label>
              <div class="input-wrap">
                <svg class="input-icon" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                  <path d="M4 6.5h16v11H4z" stroke="currentColor" stroke-width="1.4"/>
                  <path d="M9 15.5h6" stroke="currentColor" stroke-width="1.4" stroke-linecap="round"/>
                </svg>
                <input v-model="backendHost" type="text" placeholder="127.0.0.1" />
              </div>
            </div>
            <div>
              <label class="field-label">Port</label>
              <div class="input-wrap">
                <svg class="input-icon" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                  <circle cx="12" cy="12" r="8" stroke="currentColor" stroke-width="1.4"/>
                  <path d="M12 8v4l2.5 2.5" stroke="currentColor" stroke-width="1.4" stroke-linecap="round"/>
                </svg>
                <input v-model="backendPort" type="text" placeholder="8080" />
              </div>
            </div>
            <div>
              <label class="field-label">User</label>
              <div class="input-wrap">
                <svg class="input-icon" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                  <circle cx="12" cy="8.5" r="3.2" stroke="currentColor" stroke-width="1.4"/>
                  <path d="M6.5 18.5c1.5-2 3.3-3 5.5-3s4 1 5.5 3" stroke="currentColor" stroke-width="1.4" stroke-linecap="round"/>
                </svg>
                <input v-model="user" type="text" placeholder="admin" />
              </div>
            </div>
            <div>
              <label class="field-label">Password</label>
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
            <button type="button" class="btn ghost" @click="stopBackend">Stop</button>
            <button type="submit" class="btn primary">Start</button>
          </div>
        </form>
      </section>
    </main>

    <footer class="panel-footer">
      <button type="button" class="btn ghost" @click="handleReset">Reset</button>
      <button type="button" class="btn primary" @click="handleSave">Save Changes</button>
    </footer>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const frontendHost = ref('0.0.0.0')
const frontendPort = ref('8081')
const backendHost = ref('0.0.0.0')
const backendPort = ref('8080')
const user = ref('admin')
const pass = ref('')
const dataPath = ref('')

const frontendStatus = ref('stopped')
const backendStatus = ref('stopped')

// simple applied flag for footer save button feedback
const isSaved = ref(false)

function handleSave() {
  isSaved.value = true
  setTimeout(() => {
    isSaved.value = false
  }, 2000)
}

function handleReset() {
  frontendHost.value = 'localhost'
  frontendPort.value = '3000'
  backendHost.value = '127.0.0.1'
  backendPort.value = '8080'
  user.value = 'admin'
  pass.value = ''
  dataPath.value = '/var/www/data'
  frontendStatus.value = 'stopped'
  backendStatus.value = 'stopped'
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
    alert('Folder picker not available in this environment')
  } catch (err) {
    console.error(err)
    alert('Failed to select folder')
  }
}

async function startFrontend() {
  try {
    const appNs = window && (window.go?.main?.App || window['go']?.['main']?.['App'])
    if (appNs && typeof appNs.StartWebServer === 'function') {
      await appNs.StartWebServer(frontendHost.value, parseInt(frontendPort.value) || 8081)
      alert(`Frontend server started at http://${frontendHost.value}:${frontendPort.value}`)
      frontendStatus.value = 'running'
    } else {
      alert('StartWebServer not available in this environment')
      frontendStatus.value = 'error'
    }
  } catch (err) {
    console.error(err)
    alert('Failed to start frontend server: ' + err.message)
    frontendStatus.value = 'error'
  }
}

async function stopFrontend() {
  try {
    const appNs = window && (window.go?.main?.App || window['go']?.['main']?.['App'])
    if (appNs && typeof appNs.StopWebServer === 'function') {
      await appNs.StopWebServer()
      alert('Frontend server stopped successfully')
      frontendStatus.value = 'stopped'
    } else {
      alert('StopWebServer not available in this environment')
      frontendStatus.value = 'error'
    }
  } catch (err) {
    console.error(err)
    alert('Failed to stop frontend server: ' + err.message)
    frontendStatus.value = 'error'
  }
}

async function startBackend() {
  try {
    const appNs = window && (window.go?.main?.App || window['go']?.['main']?.['App'])
    if (appNs && typeof appNs.StartBackendServer === 'function') {
      await appNs.StartBackendServer(
        backendHost.value,
        parseInt(backendPort.value) || 8080,
        user.value || 'admin',
        pass.value,
        dataPath.value
      )
      alert(`Backend server started at http://${backendHost.value}:${backendPort.value}\nCheck terminal for admin credentials if password was auto-generated.`)
      backendStatus.value = 'running'
    } else {
      alert('StartBackendServer not available in this environment')
      backendStatus.value = 'error'
    }
  } catch (err) {
    console.error(err)
    alert('Failed to start backend server: ' + err.message)
    backendStatus.value = 'error'
  }
}

async function stopBackend() {
  try {
    const appNs = window && (window.go?.main?.App || window['go']?.['main']?.['App'])
    if (appNs && typeof appNs.StopBackendServer === 'function') {
      await appNs.StopBackendServer()
      alert('Backend server stopped successfully')
      backendStatus.value = 'stopped'
    } else {
      alert('StopBackendServer not available in this environment')
      backendStatus.value = 'error'
    }
  } catch (err) {
    console.error(err)
    alert('Failed to stop backend server: ' + err.message)
    backendStatus.value = 'error'
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

.btn:hover{background:#f1f5f9}
.btn.primary:hover{background:#111827;border-color:#111827}
.btn.ghost:hover{color:#0f172a}

.btn:focus{outline:2px solid #bfdbfe;outline-offset:1px}

.panel-footer .btn.primary{
  background:linear-gradient(120deg, #0f172a, #111827);
  border-color:#0f172a;
}

.panel-footer .btn.primary.saving{
  background:#22c55e;
  border-color:#16a34a;
}
</style>
