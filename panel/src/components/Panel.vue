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
              <label class="mb-1 block text-[10px] font-bold uppercase tracking-[0.08em] text-slate-500">
                Admin Password
              </label>
              <div class="relative flex items-center gap-2 rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 focus-within:border-blue-500 focus-within:bg-white">
                <svg class="absolute left-3 h-4 w-4 text-slate-400" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                  <rect x="5" y="10" width="14" height="9" rx="2" stroke="currentColor" stroke-width="1.4" />
                  <path d="M9 10V8a3 3 0 1 1 6 0v2" stroke="currentColor" stroke-width="1.4" />
                </svg>
                <input v-model="pass" type="password" placeholder="Leave empty to use saved or auto-generate" class="w-full border-none bg-transparent pl-6 text-sm text-slate-900 outline-none" />
              </div>
              <span v-if="savedPassword && !pass" class="ml-2 text-[10px] font-normal text-green-600">(using saved password)</span>
              <p class="mt-1 text-[9px] text-slate-500">Leave empty to use saved password, or will auto-generate if none saved.</p>
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
            <div class="col-span-2">
              <label class="inline-flex items-center gap-2 text-[11px] font-semibold text-slate-700">
                <input v-model="useRemoteBackend" type="checkbox" class="h-4 w-4 rounded border-slate-300 text-slate-900 focus:ring-slate-900" />
                <span>Use remote backend (only start frontend proxy)</span>
              </label>
              <p class="mt-1 text-[10px] text-slate-500">Skips starting the local backend and proxies to the backend host/port above.</p>
            </div>
          </div>

          <div class="flex flex-wrap justify-end gap-2">
            <button type="button" class="rounded-lg border border-slate-200 px-3 py-2 text-xs font-semibold text-slate-700 hover:bg-slate-100" @click="stopServers">Stop All</button>
            <button type="button" class="rounded-lg border border-slate-200 px-3 py-2 text-xs font-semibold text-slate-900 hover:bg-slate-100" @click="openFrontendInBrowser">Open Frontend</button>
            <button type="submit" class="rounded-lg bg-slate-900 px-3 py-2 text-xs font-semibold text-white hover:bg-slate-800">Start Servers</button>
          </div>
        </form>

        <div v-if="networkIP && frontendStatus === 'running'" ref="networkInfoRef" class="rounded-lg border border-blue-200 bg-blue-50 p-3">
          <div class="flex items-start gap-2">
            <svg class="h-5 w-5 shrink-0 text-blue-600" viewBox="0 0 24 24" fill="none" aria-hidden="true">
              <circle cx="12" cy="12" r="9" stroke="currentColor" stroke-width="1.5" />
              <path d="M12 8v4m0 4h.01" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
            </svg>
            <div class="flex-1">
              <h3 class="text-xs font-bold text-blue-900">Network Access</h3>
              <p class="mt-1 text-xs text-blue-700">Connect from other devices on your local network:</p>
              <div class="mt-2 rounded-md bg-white px-3 py-2 font-mono text-sm font-semibold text-blue-900">http://{{ networkIP }}:{{ frontendPort }}</div>
            </div>
          </div>
        </div>
      </section>
    </main>

    <footer class="flex items-center gap-2 rounded-xl border border-slate-200 bg-white px-3 py-2 shadow-sm">
      <button type="button" class="rounded-lg border border-slate-200 px-3 py-2 text-xs font-semibold text-slate-700 hover:bg-slate-100" @click="viewLogs">View Logs</button>
      <div class="ml-auto flex gap-2">
        <button type="button" class="rounded-lg border border-slate-200 px-3 py-2 text-xs font-semibold text-slate-700 hover:bg-slate-100" @click="handleReset">Reset</button>
        <button type="button" class="rounded-lg border border-slate-200 px-3 py-2 text-xs font-semibold text-slate-700 hover:bg-slate-100" @click="handleSave">Save Changes</button>
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
import { ref, defineProps, defineEmits, onMounted, watch, nextTick } from 'vue'
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime'

const props = defineProps({
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
const savedPassword = ref('') // Stored separately for security
const dataPath = ref('')
const useRemoteBackend = ref(false)
const networkIP = ref('')

const showModal = ref(false)
const modalTitle = ref('')
const modalMessage = ref('')
const networkInfoRef = ref(null)
const NETWORK_IP_STORAGE_KEY = 'archivist.networkIP'

// Watch for network info appearing and scroll to it
watch([networkIP, () => props.frontendStatus], ([ip, status], [prevIp, prevStatus]) => {
  if (ip && status === 'running') {
    nextTick(() => {
      if (networkInfoRef.value) {
        networkInfoRef.value.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
      }
    })
  } else if (prevIp && (!ip || status !== 'running')) {
    // Network message disappeared, scroll back to top
    nextTick(() => {
      window.scrollTo({ top: 0, behavior: 'smooth' })
    })
  }
})

// Computed property to select the effective password (entered or saved)
const getPasswordToUse = () => pass.value || savedPassword.value

// Load saved configuration on mount
onMounted(async () => {
  try {
    const appNs = window && (window.go?.main?.App || window['go']?.['main']?.['App'])
    if (appNs && typeof appNs.LoadSavedConfig === 'function') {
      const config = await appNs.LoadSavedConfig()
      if (config) {
        // Load all configuration values
        if (config.frontendHost) frontendHost.value = config.frontendHost
        if (config.frontendPort) frontendPort.value = config.frontendPort.toString()
        if (config.backendHost) backendHost.value = config.backendHost
        if (config.backendPort) backendPort.value = config.backendPort.toString()
        if (config.username) user.value = config.username
        // Store password separately without displaying it for security
        if (config.password) {
          savedPassword.value = config.password
          // Don't populate the password field - keep it empty for security
        }
        if (config.dataPath) dataPath.value = config.dataPath
      }
    }
    
    // If dataPath is still empty, populate with executable directory + '/data'
    if (!dataPath.value && appNs && typeof appNs.GetExecutableDir === 'function') {
      const execDir = await appNs.GetExecutableDir()
      if (execDir) {
        dataPath.value = execDir + '/data'
      }
    }

    try {
      const cachedIP = window?.localStorage?.getItem(NETWORK_IP_STORAGE_KEY)
      if (cachedIP && !networkIP.value) {
        networkIP.value = cachedIP
      }
    } catch (err) {
      console.error('Failed to restore cached network IP:', err)
    }
  } catch (err) {
    console.error('Failed to load configuration:', err)
    // Try to at least get the executable directory
    try {
      const appNs = window && (window.go?.main?.App || window['go']?.['main']?.['App'])
      if (appNs && typeof appNs.GetExecutableDir === 'function') {
        const execDir = await appNs.GetExecutableDir()
        if (execDir) {
          dataPath.value = execDir + '/data'
        }
      }
    } catch (err2) {
      console.error('Failed to get executable directory:', err2)
    }
  }
})

  watch(networkIP, (ip) => {
    try {
      if (ip) {
        window?.localStorage?.setItem(NETWORK_IP_STORAGE_KEY, ip)
      } else {
        window?.localStorage?.removeItem(NETWORK_IP_STORAGE_KEY)
      }
    } catch (err) {
      console.error('Failed to persist network IP:', err)
    }
  })

function showMessage(title, message) {
  modalTitle.value = title
  modalMessage.value = message
  showModal.value = true
}

function closeModal() {
  showModal.value = false
}

async function handleSave() {
  try {
    const appNs = window && (window.go?.main?.App || window['go']?.['main']?.['App'])
    if (!appNs) {
      showMessage('Error', 'Wails API not available in this environment')
      return
    }

    // Warn if removing password
    if (pass.value === '' && savedPassword.value !== '') {
      if (!confirm('Remove password? Auto-generated one will be used on next backend start.')) {
        return
      }
    }

    if (typeof appNs.SaveConfiguration === 'function') {
      // Use the entered password if provided, otherwise keep the saved one
      const passwordToSave = getPasswordToUse()
      
      await appNs.SaveConfiguration(
        frontendHost.value,
        parseInt(frontendPort.value) || 8081,
        backendHost.value,
        parseInt(backendPort.value) || 8080,
        user.value || 'admin',
        passwordToSave,
        dataPath.value
      )
      
      // Update saved password if a new one was entered
      if (pass.value) {
        savedPassword.value = pass.value
        pass.value = '' // Clear the input field after saving for security
      }
      
      showMessage('Success', 'Configuration saved successfully.')
    } else {
      showMessage('Error', 'SaveConfiguration function not available')
    }
  } catch (err) {
    console.error('Failed to save configuration:', err)
    const errorMsg = extractErrorMessage(err)
    showMessage('Error', 'Failed to save configuration: ' + errorMsg)
  }
}

function viewLogs() {
  emit('view-logs')
}

async function handleReset() {
  try {
    const appNs = window && (window.go?.main?.App || window['go']?.['main']?.['App'])
    
    // Try to reload saved config
    if (appNs && typeof appNs.LoadSavedConfig === 'function') {
      try {
        const config = await appNs.LoadSavedConfig()
        if (config) {
          frontendHost.value = config.frontendHost || '0.0.0.0'
          frontendPort.value = (config.frontendPort || 8081).toString()
          backendHost.value = config.backendHost || '0.0.0.0'
          backendPort.value = (config.backendPort || 8080).toString()
          user.value = config.username || 'admin'
          savedPassword.value = config.password || ''
          pass.value = '' // Clear the visible password field for security
          dataPath.value = config.dataPath || ''
          useRemoteBackend.value = false
          networkIP.value = ''
          emit('update-frontend-status', 'stopped')
          emit('update-backend-status', 'stopped')
          showMessage('Reset', 'Configuration reset to saved values.')
          return
        }
      } catch (err) {
        console.error('Failed to load saved config:', err)
      }
    }
    
    // Fallback to defaults if loading saved config fails
    frontendHost.value = '0.0.0.0'
    frontendPort.value = '8081'
    backendHost.value = '0.0.0.0'
    backendPort.value = '8080'
    user.value = 'admin'
    pass.value = ''
    savedPassword.value = ''
    dataPath.value = ''
    useRemoteBackend.value = false
    networkIP.value = ''
    emit('update-frontend-status', 'stopped')
    emit('update-backend-status', 'stopped')
    showMessage('Reset', 'Configuration reset to defaults.')
  } catch (err) {
    console.error('Reset failed:', err)
    showMessage('Error', 'Failed to reset configuration')
  }
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

  // If dataPath is empty, populate it with executable directory + '/data'
  if (!dataPath.value && typeof appNs.GetExecutableDir === 'function') {
    try {
      const execDir = await appNs.GetExecutableDir()
      if (execDir) {
        dataPath.value = execDir + '/data'
      }
    } catch (err) {
      console.error('Failed to get executable directory:', err)
    }
  }

  let frontendStarted = false
  let backendStarted = false
  let errors = []
  let generatedPassword = null
  const remoteOnly = useRemoteBackend.value
  
  // Use the helper function for password selection
  const passwordToUse = getPasswordToUse()

  // Start backend server first
  if (remoteOnly) {
    backendStarted = true
    emit('update-backend-status', 'remote')
  } else if (typeof appNs.StartBackendServer === 'function') {
    try {
      const result = await appNs.StartBackendServer(
        backendHost.value,
        parseInt(backendPort.value) || 8080,
        user.value || 'admin',
        passwordToUse,
        dataPath.value
      )
      // If we get here without exception, it succeeded
      backendStarted = true
      emit('update-backend-status', 'running')
      
      // Handle auto-generated password
      if (result && result.generated) {
        generatedPassword = result.password
        savedPassword.value = result.password
        pass.value = '' // Clear the input field since we now have saved password
      }
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

  // Get network info if frontend host is 0.0.0.0
  if (frontendStarted && frontendHost.value === '0.0.0.0' && typeof appNs.GetNetworkInfo === 'function') {
    try {
      const info = await appNs.GetNetworkInfo()
      if (info && info.localIP) {
        networkIP.value = info.localIP
      }
    } catch (err) {
      console.error('Failed to get network info:', err)
    }
  }

  // Show appropriate message
  if (frontendStarted && backendStarted) {
    const backendLabel = remoteOnly ? 'Remote Backend' : 'Backend'
    let successMsg = `Frontend started successfully.\n${backendLabel}: http://${backendHost.value}:${backendPort.value}`
    
    // Add note about auto-generated password
    if (generatedPassword) {
      successMsg += `\n\nAuto-generated admin password: ${generatedPassword}\nPassword has been saved to configuration.`
    }
    
    showMessage('Success', successMsg)
  } else if (frontendStarted || backendStarted) {
    const started = []
    if (frontendStarted) started.push(`Frontend: http://${frontendHost.value}:${frontendPort.value}`)
    if (backendStarted) started.push(`${remoteOnly ? 'Remote Backend proxy' : 'Backend'}: http://${backendHost.value}:${backendPort.value}`)
    const errorText = errors.length ? `\n\nErrors:\n${errors.join('\n')}` : ''
    showMessage('Partial Success', `Partially started:\n${started.join('\n')}${errorText}`)
  } else if (errors.length > 0) {
    showMessage('Error', 'Failed to start servers:\n' + errors.join('\n'))
  }
}

function openFrontendInBrowser() {
  let host = frontendHost.value || 'localhost'
  // Convert 0.0.0.0 to 127.0.0.1 for browser compatibility (especially on Windows)
  if (host === '0.0.0.0') {
    host = '127.0.0.1'
  }
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
    const remoteOnly = useRemoteBackend.value

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
    if (!remoteOnly && typeof appNs.StopBackendServer === 'function') {
      try {
        await appNs.StopBackendServer()
        emit('update-backend-status', 'stopped')
      } catch (err) {
        const errorMsg = extractErrorMessage(err, 'Backend')
        errors.push(errorMsg)
        emit('update-backend-status', 'error')
      }
    } else if (remoteOnly) {
      emit('update-backend-status', 'stopped')
    }

    // Clear network IP
    networkIP.value = ''

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
    remote: 'border-blue-200 bg-blue-50 text-blue-700',
    error: 'border-amber-200 bg-amber-50 text-amber-700'
  }[status] || 'border-slate-200 bg-slate-50 text-slate-600'
}

function statusText(status) {
  return {
    running: 'Running',
    stopped: 'Stopped',
    remote: 'Remote',
    error: 'Unavailable'
  }[status] || 'Unknown'
}

function dotClass(status) {
  return {
    running: 'bg-green-500',
    stopped: 'bg-slate-300',
    remote: 'bg-blue-500',
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
