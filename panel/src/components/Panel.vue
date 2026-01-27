<template>
  <div class="min-h-screen max-h-screen w-full max-w-5xl mx-auto flex flex-col gap-3 p-4 bg-slate-50 text-slate-900 font-sans">
    <header class="flex items-center justify-between rounded-xl border border-slate-200 bg-white px-3 py-2 shadow-sm">
      <div class="flex items-center gap-2 text-sm font-semibold">
        <span class="h-2 w-2 rounded-full bg-green-500"></span>
        <h1 class="text-sm font-semibold tracking-tight">Katarungang Pambarangay</h1>
      </div>
      <span class="rounded-md border border-slate-200 bg-slate-50 px-2 py-1 text-[10px] font-mono text-slate-600">v0.1</span>
    </header>

    <main class="flex-1 rounded-xl border border-slate-200 bg-white p-3 shadow-sm overflow-auto flex flex-col gap-3">
      <section class="flex flex-col gap-3">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div class="flex items-center gap-2 text-slate-500">
            <svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" aria-hidden="true">
              <rect x="4" y="6.5" width="16" height="11" rx="1.4" stroke="currentColor" stroke-width="1.4" />
              <path d="M8 10.5h8M8 13.5h8" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" />
            </svg>
            <h2 class="text-[11px] font-bold uppercase tracking-[0.12em]">Server Configuration</h2>
          </div>
          <div class="flex flex-wrap items-center gap-2">
            <span class="inline-flex items-center gap-2 rounded-full border px-3 py-1 text-xs font-semibold" :class="statusClass(frontendStatus)">
              <span class="h-2 w-2 rounded-full" :class="dotClass(frontendStatus)"></span>
              Frontend: {{ statusText(frontendStatus) }}
            </span>
            <span class="inline-flex items-center gap-2 rounded-full border px-3 py-1 text-xs font-semibold" :class="statusClass(backendStatus)">
              <span class="h-2 w-2 rounded-full" :class="dotClass(backendStatus)"></span>
              Backend: {{ statusText(backendStatus) }}
            </span>
          </div>
        </div>

        <form class="flex flex-col gap-3" @submit.prevent="startServers">
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="mb-1 block text-[10px] font-bold uppercase tracking-[0.08em] text-slate-500">Frontend Host</label>
              <div class="relative flex items-center gap-2 rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 focus-within:border-blue-500 focus-within:bg-white">
                <svg class="absolute left-3 h-4 w-4 text-slate-400" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                  <path d="M4 6.5h16v11H4z" stroke="currentColor" stroke-width="1.4" />
                  <path d="M9 15.5h6" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" />
                </svg>
                <input v-model="frontendHost" type="text" placeholder="0.0.0.0" class="w-full border-none bg-transparent pl-6 text-sm text-slate-900 outline-none" />
              </div>
            </div>
            <div>
              <label class="mb-1 block text-[10px] font-bold uppercase tracking-[0.08em] text-slate-500">Frontend Port</label>
              <div class="relative flex items-center gap-2 rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 focus-within:border-blue-500 focus-within:bg-white">
                <svg class="absolute left-3 h-4 w-4 text-slate-400" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                  <circle cx="12" cy="12" r="8" stroke="currentColor" stroke-width="1.4" />
                  <path d="M12 8v4l2.5 2.5" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" />
                </svg>
                <input v-model="frontendPort" type="text" placeholder="8081" class="w-full border-none bg-transparent pl-6 text-sm text-slate-900 outline-none" />
              </div>
            </div>
            <div class="sm:col-span-2">
              <label class="mb-1 block text-[10px] font-bold uppercase tracking-[0.08em] text-slate-500">Backend Host</label>
              <div class="relative flex items-center gap-2 rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 focus-within:border-blue-500 focus-within:bg-white">
                <svg class="absolute left-3 h-4 w-4 text-slate-400" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                  <path d="M4 6.5h16v11H4z" stroke="currentColor" stroke-width="1.4" />
                  <path d="M9 15.5h6" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" />
                </svg>
                <input v-model="backendHost" type="text" placeholder="0.0.0.0" class="w-full border-none bg-transparent pl-6 text-sm text-slate-900 outline-none" />
              </div>
            </div>
            <div>
              <label class="mb-1 block text-[10px] font-bold uppercase tracking-[0.08em] text-slate-500">Backend Port</label>
              <div class="relative flex items-center gap-2 rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 focus-within:border-blue-500 focus-within:bg-white">
                <svg class="absolute left-3 h-4 w-4 text-slate-400" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                  <circle cx="12" cy="12" r="8" stroke="currentColor" stroke-width="1.4" />
                  <path d="M12 8v4l2.5 2.5" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" />
                </svg>
                <input v-model="backendPort" type="text" placeholder="8080" class="w-full border-none bg-transparent pl-6 text-sm text-slate-900 outline-none" />
              </div>
            </div>
            <div>
              <label class="mb-1 block text-[10px] font-bold uppercase tracking-[0.08em] text-slate-500">Admin User</label>
              <div class="relative flex items-center gap-2 rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 focus-within:border-blue-500 focus-within:bg-white">
                <svg class="absolute left-3 h-4 w-4 text-slate-400" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                  <circle cx="12" cy="8.5" r="3.2" stroke="currentColor" stroke-width="1.4" />
                  <path d="M6.5 18.5c1.5-2 3.3-3 5.5-3s4 1 5.5 3" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" />
                </svg>
                <input v-model="user" type="text" placeholder="admin" class="w-full border-none bg-transparent pl-6 text-sm text-slate-900 outline-none" />
              </div>
            </div>
            <div class="sm:col-span-2">
              <label class="mb-1 block text-[10px] font-bold uppercase tracking-[0.08em] text-slate-500">Admin Password</label>
              <div class="relative flex items-center gap-2 rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 focus-within:border-blue-500 focus-within:bg-white">
                <svg class="absolute left-3 h-4 w-4 text-slate-400" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                  <rect x="5" y="10" width="14" height="9" rx="2" stroke="currentColor" stroke-width="1.4" />
                  <path d="M9 10V8a3 3 0 1 1 6 0v2" stroke="currentColor" stroke-width="1.4" />
                </svg>
                <input v-model="pass" type="password" placeholder="••••••" class="w-full border-none bg-transparent pl-6 text-sm text-slate-900 outline-none" />
              </div>
            </div>
            <div class="col-span-2">
              <label class="mb-1 block text-[10px] font-bold uppercase tracking-[0.08em] text-slate-500">Data Path</label>
              <div class="relative flex items-center gap-2 rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 focus-within:border-blue-500 focus-within:bg-white">
                <svg class="absolute left-3 h-4 w-4 text-slate-400" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                  <path d="M4.5 7.5h6l1.5 2h7.5v8a1 1 0 0 1-1 1h-14a1 1 0 0 1-1-1v-9z" stroke="currentColor" stroke-width="1.4" />
                </svg>
                <input v-model="dataPath" type="text" placeholder="/var/www/data" class="w-full border-none bg-transparent pl-6 text-sm text-slate-900 outline-none" />
                <button type="button" class="ml-2 shrink-0 rounded-md border border-slate-200 px-3 py-1 text-xs font-semibold text-slate-700 hover:bg-slate-100" @click="selectFolder">Browse</button>
              </div>
            </div>
          </div>

          <div class="flex flex-wrap justify-end gap-2">
            <button type="button" class="rounded-lg border border-slate-200 px-3 py-2 text-xs font-semibold text-slate-700 hover:bg-slate-100" @click="stopServers">Stop All</button>
            <button type="button" class="rounded-lg border border-slate-200 px-3 py-2 text-xs font-semibold text-slate-900 hover:bg-slate-100" @click="openFrontendInBrowser">Open Frontend</button>
            <button type="submit" class="rounded-lg bg-slate-900 px-3 py-2 text-xs font-semibold text-white hover:bg-slate-800">Start Servers</button>
          </div>
        </form>
      </section>
    </main>

    <footer class="flex items-center gap-2 rounded-xl border border-slate-200 bg-white px-3 py-2 shadow-sm">
      <button type="button" class="rounded-lg border border-slate-200 px-3 py-2 text-xs font-semibold text-slate-700 hover:bg-slate-100" @click="viewLogs">View Logs</button>
      <div class="ml-auto flex gap-2">
        <button type="button" class="rounded-lg border border-slate-200 px-3 py-2 text-xs font-semibold text-slate-700 hover:bg-slate-100" @click="handleReset">Reset</button>
        <button type="button" class="rounded-lg bg-linear-to-r from-slate-900 to-slate-800 px-3 py-2 text-xs font-semibold text-white hover:from-slate-800 hover:to-slate-700" @click="handleSave">Save Changes</button>
      </div>
    </footer>

    <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 p-4" @click="closeModal">
      <div class="w-full max-w-md overflow-hidden rounded-xl border border-slate-200 bg-white shadow-2xl" @click.stop>
        <div class="border-b border-slate-200 px-4 py-3">
          <h3 class="text-sm font-semibold text-slate-900">{{ modalTitle }}</h3>
        </div>
        <div class="max-h-[60vh] overflow-y-auto px-4 py-3">
          <pre class="whitespace-pre-wrap wrap-break-word text-sm text-slate-700">{{ modalMessage }}</pre>
        </div>
        <div class="flex justify-end gap-2 border-t border-slate-200 px-4 py-3">
          <button type="button" class="rounded-lg bg-slate-900 px-3 py-2 text-xs font-semibold text-white hover:bg-slate-800" @click="closeModal">OK</button>
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
  showMessage('Saved', 'Settings saved locally.')
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
      const errorMsg = extractErrorMessage(err, 'Backend')
      errors.push(errorMsg)
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
      const errorMsg = extractErrorMessage(err, 'Frontend')
      errors.push(errorMsg)
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
        const errorMsg = extractErrorMessage(err, 'Frontend')
        errors.push(errorMsg)
        emit('update-frontend-status', 'error')
      }
    }

    // Stop backend server
    if (typeof appNs.StopBackendServer === 'function') {
      try {
        await appNs.StopBackendServer()
        emit('update-backend-status', 'stopped')
      } catch (err) {
        const errorMsg = extractErrorMessage(err, 'Backend')
        errors.push(errorMsg)
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
    const errorMsg = extractErrorMessage(err)
    showMessage('Error', 'Failed to stop servers: ' + errorMsg)
  }
}

function statusClass(status) {
  return {
    running: 'border-green-200 bg-green-50 text-green-700',
    stopped: 'border-slate-200 bg-slate-50 text-slate-600',
    error: 'border-amber-200 bg-amber-50 text-amber-700'
  }[status] || 'border-slate-200 bg-slate-50 text-slate-600'
}

function statusText(status) {
  return {
    running: 'Running',
    stopped: 'Stopped',
    error: 'Unavailable'
  }[status] || 'Unknown'
}

function dotClass(status) {
  return {
    running: 'bg-green-500',
    stopped: 'bg-slate-300',
    error: 'bg-amber-500'
  }[status] || 'bg-slate-300'
}

function extractErrorMessage(error, context = '') {
  // Handle null or undefined
  if (error === null || error === undefined) {
    return context ? `${context}: No error details available` : 'No error details available'
  }

  // Handle string errors
  if (typeof error === 'string') {
    return context ? `${context}: ${error}` : error
  }

  // Handle error objects with message property
  if (error.message && typeof error.message === 'string' && error.message.trim()) {
    const msg = error.message.trim()
    // Add context-aware suggestions for common errors
    let suggestion = ''
    if (msg.toLowerCase().includes('address already in use') || msg.toLowerCase().includes('bind')) {
      suggestion = ' (Port may already be in use - try a different port or stop other services)'
    } else if (msg.toLowerCase().includes('permission denied') || msg.toLowerCase().includes('eacces')) {
      suggestion = ' (Permission denied - try running with appropriate privileges)'
    } else if (msg.toLowerCase().includes('connection refused') || msg.toLowerCase().includes('econnrefused')) {
      suggestion = ' (Connection refused - backend may not be running or port is incorrect)'
    } else if (msg.toLowerCase().includes('no such file') || msg.toLowerCase().includes('enoent')) {
      suggestion = ' (File or directory not found - check data path)'
    } else if (msg.toLowerCase().includes('timeout')) {
      suggestion = ' (Operation timed out - server may be unresponsive)'
    }
    return context ? `${context}: ${msg}${suggestion}` : `${msg}${suggestion}`
  }

  // Handle error objects with toString method
  if (typeof error.toString === 'function') {
    const str = error.toString().trim()
    if (str && str !== '[object Object]') {
      return context ? `${context}: ${str}` : str
    }
  }

  // Handle error objects with status or code properties
  if (error.code || error.status) {
    const code = error.code || error.status
    return context ? `${context}: Error code ${code}` : `Error code ${code}`
  }

  // Fallback for any other object
  return context ? `${context}: Unknown error occurred` : 'Unknown error occurred'
}
</script>
