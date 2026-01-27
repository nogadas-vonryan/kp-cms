<template>
  <div class="overflow-x-auto">
    <table class="min-w-full border border-gray-300 bg-white">
      <thead class="bg-gray-50">
        <tr>
          <th v-for="col in columns" :key="col.key" class="text-left border-b border-gray-300 px-3 py-2 text-gray-700">{{ col.label }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(row, idx) in rows" :key="idx" class="odd:bg-white even:bg-gray-50">
          <td v-for="col in columns" :key="col.key" class="border-t border-gray-200 px-3 py-2 text-gray-900">
            <slot :name="`cell:${col.key}`" :row="row" :value="row[col.key]">
              {{ row[col.key] }}
            </slot>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
type Column = { key: string; label: string }

const props = defineProps({
  columns: { type: Array as () => Column[], required: true },
  rows: { type: Array as () => Record<string, unknown>[], required: true }
})
</script>
