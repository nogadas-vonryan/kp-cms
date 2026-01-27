<template>
  <div class="h-screen max-h-screen w-full max-w-4xl mx-auto flex flex-col gap-2.5 p-3 md:p-4 bg-slate-50 font-system overflow-hidden">
    <header class="flex items-center justify-between p-2 bg-white border border-slate-200 rounded-lg shadow-sm">
      <div class="flex items-center gap-2">
        <button @click="$emit('back')" title="Back to controls" class="w-8 h-8 p-0 border border-slate-200 bg-white rounded flex items-center justify-center text-slate-900 hover:bg-slate-100 hover:border-slate-300 transition-all">
          <svg viewBox="0 0 24 24" fill="none" aria-hidden="true" class="w-4 h-4">
            <path d="M19 12H5M5 12l6 6M5 12l6-6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </button>
        <h1 class="text-sm font-semibold tracking-tight">Application Logs</h1>
      </div>
      <span class="text-xs bg-slate-50 text-slate-700 px-2 py-0.5 rounded border border-slate-200 font-mono">v0.1</span>
    </header>

    <main class="flex-1 min-h-0 bg-white border border-slate-200 rounded-lg p-3 flex flex-col gap-2">
      <div class="flex justify-between items-center p-2 border-b border-slate-200">
        <span class="text-xs font-semibold text-slate-500">{{ logs.length }} entries</span>
        <button type="button" @click="clearLogs" class="px-3 py-2 text-xs font-semibold text-slate-600 bg-transparent border border-slate-200 rounded hover:bg-slate-100 hover:text-slate-900 transition-colors">Clear Logs</button>
      </div>

      <div class="flex-1 min-h-0 bg-slate-50 border border-slate-200 rounded p-3 overflow-y-auto flex flex-col">
        <div v-if="logs.length === 0" class="text-center text-slate-400 text-sm p-10 m-auto">No logs yet</div>
        <div v-else class="flex flex-col gap-1.5">
          <div v-for="(log, index) in logs" :key="index" class="flex gap-3 text-xs font-mono p-2 border-b border-slate-200 items-start last:border-b-0">
            <span class="text-slate-500 font-semibold min-w-21.25 shrink-0">{{ log.timestamp }}</span>
            <span class="text-slate-900 wrap-break-word flex-1">{{ log.message }}</span>
          </div>
        </div>
      </div>
    </main>

    <footer class="flex items-center gap-2 p-2 bg-white border border-slate-200 rounded-lg shadow-sm">
      <button type="button" @click="$emit('back')" class="px-3 py-2 text-xs font-semibold text-slate-600 bg-transparent border border-slate-200 rounded hover:bg-slate-100 hover:text-slate-900 transition-colors">Back</button>
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
