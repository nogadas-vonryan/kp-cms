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
              <label class="mb-1 block text-[10px] font-bold uppercase tracking-[0.08em] text-slate-500">Admin User</label>
              <div class="relative flex items-center gap-2 rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 focus-within:border-blue-500 focus-within:bg-white">
                <svg class="absolute left-3 h-4 w-4 text-slate-400" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                  <circle cx="12" cy="8.5" r="3.2" stroke="currentColor" stroke-width="1.4" />
                  <path d="M6.5 18.5c1.5-2 3.3-3 5.5-3s4 1 5.5 3" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" />
                </svg>
                <input v-model="user" type="text" placeholder="admin" class="w-full border-none bg-transparent pl-6 text-sm text-slate-900 outline-none" />
              </div>
            </div>

            <div>
              <label class="mb-1 block text-[10px] font-bold uppercase tracking-[0.08em] text-slate-500">Admin Password</label>
              <div class="relative flex items-center gap-2 rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 focus-within:border-blue-500 focus-within:bg-white">
                <svg class="absolute left-3 h-4 w-4 text-slate-400" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                  <rect x="5" y="10" width="14" height="9" rx="2" stroke="currentColor" stroke-width="1.4" />
                  <path d="M9 10V8a3 3 0 1 1 6 0v2" stroke="currentColor" stroke-width="1.4" />
                </svg>
                <input v-model="pass" type="password" placeholder="Leave empty to use saved" class="w-full border-none bg-transparent pl-6 text-sm text-slate-900 outline-none" />
              </div>
            </div>

            <div class="col-span-2 -mt-1">
              <span v-if="savedPassword && !pass" class="text-[10px] font-medium text-green-600 block">
                (using saved password)
              </span>
              <p class="text-[9px] text-slate-500">
                Leave empty to use saved password, or will auto-generate if none saved.
              </p>
            </div>

            <details class="col-span-2 mt-2 group border-t border-slate-100 pt-3">
              <summary class="flex cursor-pointer list-none items-center gap-2 text-[10px] font-bold uppercase tracking-[0.12em] text-slate-400 hover:text-slate-600 transition-colors">
                <svg class="h-3 w-3 transform transition-transform group-open:rotate-90" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                  <polyline points="9 18 15 12 9 6"></polyline>
                </svg>
                Advanced Connectivity & Paths
              </summary>
              
              <div class="grid grid-cols-2 gap-3 mt-4 animate-in fade-in slide-in-from-top-1">
                <div>
                  <label class="mb-1 block text-[10px] font-bold uppercase tracking-[0.08em] text-slate-500">Frontend Host</label>
                  <div class="relative flex items-center gap-2 rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 focus-within:border-blue-500 focus-within:bg-white">
                    <input v-model="frontendHost" type="text" placeholder="0.0.0.0" class="w-full border-none bg-transparent text-sm text-slate-900 outline-none" />
                  </div>
                </div>
                <div>
                  <label class="mb-1 block text-[10px] font-bold uppercase tracking-[0.08em] text-slate-500">Frontend Port</label>
                  <div class="relative flex items-center gap-2 rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 focus-within:border-blue-500 focus-within:bg-white">
                    <input v-model="frontendPort" type="text" placeholder="8081" class="w-full border-none bg-transparent text-sm text-slate-900 outline-none" />
                  </div>
                </div>
                <div>
                  <label class="mb-1 block text-[10px] font-bold uppercase tracking-[0.08em] text-slate-500">Backend Host</label>
                  <div class="relative flex items-center gap-2 rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 focus-within:border-blue-500 focus-within:bg-white">
                    <input v-model="backendHost" type="text" placeholder="0.0.0.0" class="w-full border-none bg-transparent text-sm text-slate-900 outline-none" />
                  </div>
                </div>
                <div>
                  <label class="mb-1 block text-[10px] font-bold uppercase tracking-[0.08em] text-slate-500">Backend Port</label>
                  <div class="relative flex items-center gap-2 rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 focus-within:border-blue-500 focus-within:bg-white">
                    <input v-model="backendPort" type="text" placeholder="8080" class="w-full border-none bg-transparent text-sm text-slate-900 outline-none" />
                  </div>
                </div>
                <div class="col-span-2">
                  <label class="mb-1 block text-[10px] font-bold uppercase tracking-[0.08em] text-slate-500">Data Path</label>
                  <div class="relative flex items-center gap-2 rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 focus-within:border-blue-500 focus-within:bg-white">
                    <input v-model="dataPath" type="text" placeholder="Current folder" class="w-full border-none bg-transparent text-sm text-slate-900 outline-none" />
                    <button type="button" class="ml-2 shrink-0 rounded-md border border-slate-200 px-3 py-1 text-xs font-semibold text-slate-700 hover:bg-slate-100" @click="selectFolder">Browse</button>
                  </div>
                </div>
                <div class="col-span-2">
                  <label class="mb-1 block text-[10px] font-bold uppercase tracking-[0.08em] text-slate-500">Backup Data Folder</label>
                  <div class="relative flex items-center gap-2 rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 focus-within:border-blue-500 focus-within:bg-white">
                    <input v-model="backupPath" type="text" placeholder="Optional backup folder" class="w-full border-none bg-transparent text-sm text-slate-900 outline-none" />
                    <button type="button" class="ml-2 shrink-0 rounded-md border border-slate-200 px-3 py-1 text-xs font-semibold text-slate-700 hover:bg-slate-100" @click="selectBackupFolder">Browse</button>
                  </div>
                </div>
                <div class="col-span-2 pt-2">
                  <label class="inline-flex items-center gap-2 text-[11px] font-semibold text-slate-700">
                    <input v-model="useRemoteBackend" type="checkbox" class="h-4 w-4 rounded border-slate-300 text-slate-900 focus:ring-slate-900" />
                    <span>Use remote backend (only start frontend proxy)</span>
                  </label>
                </div>
              </div>
            </details>
          </div>

          <div class="flex flex-wrap justify-end gap-2 mt-4">
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
const savedPassword = ref('')
const dataPath = ref('')
const backupPath = ref('')
const useRemoteBackend = ref(false)
const networkIP = ref('')

const showModal = ref(false)
const modalTitle = ref('')
const modalMessage = ref('')
const networkInfoRef = ref(null)
const NETWORK_IP_STORAGE_KEY = 'archivist.networkIP'

// Watch for network info and scroll
watch([networkIP, () => props.frontendStatus], ([ip, status], [prevIp, prevStatus]) => {
  if (ip && status === 'running') {
    nextTick(() => {
      if (networkInfoRef.value) {
        networkInfoRef.value.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
      }
    })
  } else if (prevIp && (!ip || status !== 'running')) {
    nextTick(() => {
      window.scrollTo({ top: 0, behavior: 'smooth' })
    })
  }
})

const getPasswordToUse = () => pass.value || savedPassword.value

onMounted(async () => {
  try {
    const appNs = window && (window.go?.main?.App || window['go']?.['main']?.['App'])
    if (appNs && typeof appNs.LoadSavedConfig === 'function') {
      const config = await appNs.LoadSavedConfig()
      if (config) {
        if (config.frontendHost) frontendHost.value = config.frontendHost
        if (config.frontendPort) frontendPort.value = config.frontendPort.toString()
        if (config.backendHost) backendHost.value = config.backendHost
        if (config.backendPort) backendPort.value = config.backendPort.toString()
        if (config.username) user.value = config.username
        if (config.password) savedPassword.value = config.password
        if (config.dataPath) dataPath.value = config.dataPath
        if (config.backupPath) backupPath.value = config.backupPath
      }
    }
    
    if (!dataPath.value && appNs && typeof appNs.GetExecutableDir === 'function') {
      const execDir = await appNs.GetExecutableDir()
      if (execDir) dataPath.value = execDir + '/data'
    }

    const cachedIP = window?.localStorage?.getItem(NETWORK_IP_STORAGE_KEY)
    if (cachedIP) networkIP.value = cachedIP

  } catch (err) {
    console.error('Failed to load configuration:', err)
  }
})

watch(networkIP, (ip) => {
  if (ip) window?.localStorage?.setItem(NETWORK_IP_STORAGE_KEY, ip)
  else window?.localStorage?.removeItem(NETWORK_IP_STORAGE_KEY)
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
      showMessage('Error', 'Wails API not available')
      return
    }

    if (pass.value === '' && savedPassword.value !== '') {
      if (!confirm('Remove password? Auto-generated one will be used on next start.')) return
    }

    if (typeof appNs.SaveConfiguration === 'function') {
      const passwordToSave = getPasswordToUse()
      await appNs.SaveConfiguration(
        frontendHost.value,
        parseInt(frontendPort.value) || 8081,
        backendHost.value,
        parseInt(backendPort.value) || 8080,
        user.value || 'admin',
        passwordToSave,
        dataPath.value,
        backupPath.value || ''
      )
      
      if (pass.value) {
        savedPassword.value = pass.value
        pass.value = ''
      }
      showMessage('Success', 'Configuration saved successfully.')
    }
  } catch (err) {
    showMessage('Error', 'Failed to save configuration: ' + extractErrorMessage(err))
  }
}

function viewLogs() {
  emit('view-logs')
}

async function handleReset() {
  try {
    const appNs = window && (window.go?.main?.App || window['go']?.['main']?.['App'])
    if (appNs && typeof appNs.LoadSavedConfig === 'function') {
      const config = await appNs.LoadSavedConfig()
      if (config) {
        frontendHost.value = config.frontendHost || '0.0.0.0'
        frontendPort.value = (config.frontendPort || 8081).toString()
        backendHost.value = config.backendHost || '0.0.0.0'
        backendPort.value = (config.backendPort || 8080).toString()
        user.value = config.username || 'admin'
        savedPassword.value = config.password || ''
        pass.value = ''
        dataPath.value = config.dataPath || ''
        backupPath.value = config.backupPath || ''
        useRemoteBackend.value = false
        networkIP.value = ''
        emit('update-frontend-status', 'stopped')
        emit('update-backend-status', 'stopped')
        showMessage('Reset', 'Configuration reset to saved values.')
        return
      }
    }
  } catch (err) {
    showMessage('Error', 'Failed to reset configuration')
  }
}

async function selectFolder() {
  const appNs = window && (window.go?.main?.App || window['go']?.['main']?.['App'])
  if (appNs && typeof appNs.SelectFolder === 'function') {
    const selection = await appNs.SelectFolder()
    if (selection) dataPath.value = selection
  }
}

async function selectBackupFolder() {
  const appNs = window && (window.go?.main?.App || window['go']?.['main']?.['App'])
  if (appNs && typeof appNs.SelectFolder === 'function') {
    const selection = await appNs.SelectFolder()
    if (selection) backupPath.value = selection
  }
}

async function startServers() {
  const appNs = window && (window.go?.main?.App || window['go']?.['main']?.['App'])
  if (!appNs) return

  let frontendStarted = false
  let backendStarted = false
  let errors = []
  let generatedPassword = null
  const remoteOnly = useRemoteBackend.value
  const passwordToUse = getPasswordToUse()

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
        dataPath.value,
        backupPath.value
      )
      backendStarted = true
      emit('update-backend-status', 'running')
      if (result && result.generated) {
        generatedPassword = result.password
        savedPassword.value = result.password
        pass.value = ''
      }
    } catch (err) {
      errors.push(extractErrorMessage(err, 'Backend'))
      emit('update-backend-status', 'error')
    }
  }

  if (typeof appNs.StartWebServer === 'function') {
    try {
      await appNs.StartWebServer(
        frontendHost.value,
        parseInt(frontendPort.value) || 8081,
        backendHost.value,
        parseInt(backendPort.value) || 8080
      )
      frontendStarted = true
      emit('update-frontend-status', 'running')
    } catch (err) {
      errors.push(extractErrorMessage(err, 'Frontend'))
      emit('update-frontend-status', 'error')
    }
  }

  if (frontendStarted && frontendHost.value === '0.0.0.0' && typeof appNs.GetNetworkInfo === 'function') {
    const info = await appNs.GetNetworkInfo()
    if (info?.localIP) networkIP.value = info.localIP
  }

  if (frontendStarted && backendStarted) {
    let msg = `Frontend started.\nBackend: http://${backendHost.value}:${backendPort.value}`
    if (generatedPassword) msg += `\n\nGenerated password: ${generatedPassword}`
    showMessage('Success', msg)
  } else if (errors.length) {
    showMessage('Error', errors.join('\n'))
  }
}

function openFrontendInBrowser() {
  let host = frontendHost.value === '0.0.0.0' ? '127.0.0.1' : frontendHost.value
  const url = `http://${host}:${frontendPort.value || 8081}`
  if (typeof BrowserOpenURL === 'function') BrowserOpenURL(url)
  else window.open(url, '_blank')
}

async function stopServers() {
  const appNs = window && (window.go?.main?.App || window['go']?.['main']?.['App'])
  if (!appNs) return
  
  if (typeof appNs.StopWebServer === 'function') {
    await appNs.StopWebServer()
    emit('update-frontend-status', 'stopped')
  }
  if (!useRemoteBackend.value && typeof appNs.StopBackendServer === 'function') {
    await appNs.StopBackendServer()
    emit('update-backend-status', 'stopped')
  } else if (useRemoteBackend.value) {
    emit('update-backend-status', 'stopped')
  }
  networkIP.value = ''
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
  return { running: 'Running', stopped: 'Stopped', remote: 'Remote', error: 'Error' }[status] || 'Unknown'
}

function dotClass(status) {
  return { running: 'bg-green-500', stopped: 'bg-slate-300', remote: 'bg-blue-500', error: 'bg-amber-500' }[status] || 'bg-slate-300'
}

function extractErrorMessage(error, context = '') {
  if (!error) return 'Unknown error'
  const msg = error.message || (typeof error === 'string' ? error : 'Check logs for details')
  return context ? `${context}: ${msg}` : msg
}
</script>

<style scoped>
summary::-webkit-details-marker { display: none; }
.animate-in { animation-duration: 300ms; }
</style>