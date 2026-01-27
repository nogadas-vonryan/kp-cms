<template>
  <div class="panel">
    <header class="panel-header">
      <h1 class="title">Startup Panel</h1>
      <p class="subtitle">Configure and start your servers.</p>
    </header>

    <!-- Frontend Server Section -->
    <section class="server-section">
      <h2 class="section-title">Frontend Server</h2>
      <form class="form" @submit.prevent="startFrontend">
        <label class="field">Host
          <input v-model="frontendHost" type="text" placeholder="0.0.0.0" />
        </label>

        <label class="field">Port
          <input v-model="frontendPort" type="text" placeholder="8081" />
        </label>

        <div class="actions">
          <button type="submit" class="button primary">
            <svg class="icon" viewBox="0 0 24 24" fill="none" aria-hidden="true">
              <path d="M8 5.5v13l10-6.5-10-6.5z" stroke="currentColor" stroke-width="1.5"/>
            </svg>
            <span>Start</span>
          </button>
          <button type="button" class="button danger" @click="stopFrontend">
            <svg class="icon" viewBox="0 0 24 24" fill="none" aria-hidden="true">
              <rect x="6" y="6" width="12" height="12" stroke="currentColor" stroke-width="1.5"/>
            </svg>
            <span>Stop</span>
          </button>
        </div>
      </form>
    </section>

    <!-- Backend Server Section -->
    <section class="server-section">
      <h2 class="section-title">Backend Server</h2>
      <form class="form" @submit.prevent="startBackend">
        <label class="field">Host
          <input v-model="backendHost" type="text" placeholder="0.0.0.0" />
        </label>

        <label class="field">Port
          <input v-model="backendPort" type="text" placeholder="8080" />
        </label>

        <label class="field">User
          <input v-model="user" type="text" placeholder="admin" />
        </label>

        <label class="field">Password
          <input v-model="pass" type="password" placeholder="Leave empty to auto-generate" />
        </label>

        <label class="field">Data Path
          <div class="inline-row">
            <input v-model="dataPath" type="text" placeholder="/path/to/folder" />
            <button type="button" class="button" @click="selectFolder">
              <svg class="icon" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                <path d="M3.5 6.5h6l1.5 2h9.5v9.5a1.5 1.5 0 0 1-1.5 1.5H5a1.5 1.5 0 0 1-1.5-1.5V6.5z" stroke="currentColor" stroke-width="1.5"/>
              </svg>
              <span>Browse</span>
            </button>
          </div>
        </label>

        <div class="actions">
          <button type="submit" class="button primary">
            <svg class="icon" viewBox="0 0 24 24" fill="none" aria-hidden="true">
              <path d="M8 5.5v13l10-6.5-10-6.5z" stroke="currentColor" stroke-width="1.5"/>
            </svg>
            <span>Start</span>
          </button>
          <button type="button" class="button danger" @click="stopBackend">
            <svg class="icon" viewBox="0 0 24 24" fill="none" aria-hidden="true">
              <rect x="6" y="6" width="12" height="12" stroke="currentColor" stroke-width="1.5"/>
            </svg>
            <span>Stop</span>
          </button>
        </div>
      </form>
    </section>
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
    } else {
      alert('StartWebServer not available in this environment')
    }
  } catch (err) {
    console.error(err)
    alert('Failed to start frontend server: ' + err.message)
  }
}

async function stopFrontend() {
  try {
    const appNs = window && (window.go?.main?.App || window['go']?.['main']?.['App'])
    if (appNs && typeof appNs.StopWebServer === 'function') {
      await appNs.StopWebServer()
      alert('Frontend server stopped successfully')
    } else {
      alert('StopWebServer not available in this environment')
    }
  } catch (err) {
    console.error(err)
    alert('Failed to stop frontend server: ' + err.message)
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
    } else {
      alert('StartBackendServer not available in this environment')
    }
  } catch (err) {
    console.error(err)
    alert('Failed to start backend server: ' + err.message)
  }
}

async function stopBackend() {
  try {
    const appNs = window && (window.go?.main?.App || window['go']?.['main']?.['App'])
    if (appNs && typeof appNs.StopBackendServer === 'function') {
      await appNs.StopBackendServer()
      alert('Backend server stopped successfully')
    } else {
      alert('StopBackendServer not available in this environment')
    }
  } catch (err) {
    console.error(err)
    alert('Failed to stop backend server: ' + err.message)
  }
}
</script>

<style scoped>
/* Minimal, modern, clean (white, no shadows or gradients) */
.panel{
  max-width:560px;
  margin:40px auto;
  padding:24px;
  background:#fff;
  border:1px solid #e5e7eb;
  border-radius:10px;
  color:#111827;
  font-family:system-ui, -apple-system, 'Segoe UI', Roboto, 'Helvetica Neue', Arial;
}
.panel-header{margin-bottom:20px}
.title{font-size:20px;line-height:28px;margin:0;font-weight:600}
.subtitle{margin:6px 0 0;color:#6b7280;font-size:13px}

.server-section{
  margin-bottom:24px;
  padding-bottom:24px;
  border-bottom:1px solid #f3f4f6;
}
.server-section:last-child{
  margin-bottom:0;
  padding-bottom:0;
  border-bottom:none;
}
.section-title{
  font-size:16px;
  font-weight:600;
  margin:0 0 16px 0;
  color:#374151;
}

.form{display:flex;flex-direction:column;gap:16px}
.field{display:flex;flex-direction:column;gap:6px;font-size:13px;color:#334155}
.field input{
  padding:10px 12px;
  border:1px solid #e5e7eb;
  border-radius:8px;
  background:#fff;
  font-size:14px;
}
.field input:focus{
  outline:none;
  border-color:#111827;
}

.inline-row{display:flex;gap:10px;align-items:center}
.inline-row input{flex:1}

.button{
  display:inline-flex;
  align-items:center;
  gap:8px;
  padding:10px 12px;
  background:#fff;
  color:#111827;
  border:1px solid #e5e7eb;
  border-radius:8px;
  cursor:pointer;
}
.button:hover{background:#f9fafb}
.button:focus{outline:none;border-color:#111827}
.button .icon{width:16px;height:16px;display:block}

.actions{display:flex;justify-content:flex-end;gap:10px;margin-top:4px}
.primary{border-color:#111827}
.danger{border-color:#dc2626;color:#dc2626}
.danger:hover{background:#fef2f2}
</style>
