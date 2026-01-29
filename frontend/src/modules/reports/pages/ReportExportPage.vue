<template>
  <div class="mx-auto max-w-full overflow-x-hidden px-2 sm:px-0">
    <div class="mb-4 space-y-3">
      <div class="flex flex-col md:flex-row md:items-end gap-2">
        <div class="md:flex-1">
          <label class="block text-sm font-medium mb-1">Sort By</label>
          <UiInput v-model="filters.sort_by" placeholder="e.g., code, title, created_at" class="w-full" />
        </div>
        <div class="md:w-40">
          <label class="block text-sm font-medium mb-1">Order</label>
          <UiSelect v-model="filters.sort_order" :options="[{ label: 'Ascending', value: 'asc' }, { label: 'Descending', value: 'desc' }]" />
        </div>
        <UiButton @click="applyFilters" size="sm">Apply Filters</UiButton>
        <UiButton @click="exportCSV" variant="secondary" size="sm">Export CSV</UiButton>
      </div>

      <div class="flex flex-col md:flex-row md:items-end gap-2">
        <div class="md:flex-1">
          <label class="block text-sm font-medium mb-1">Start Date</label>
          <UiInput type="date" v-model="filters.start" class="w-full" />
        </div>
        <div class="md:flex-1">
          <label class="block text-sm font-medium mb-1">End Date</label>
          <UiInput type="date" v-model="filters.end" class="w-full" />
        </div>
        <div class="md:w-32">
          <label class="block text-sm font-medium mb-1">Offset</label>
          <UiInput type="number" v-model.number="filters.offset" class="w-full" min="0" />
        </div>
        <div class="md:w-32">
          <label class="block text-sm font-medium mb-1">Limit</label>
          <UiInput type="number" v-model.number="filters.limit" class="w-full" min="1" />
        </div>
      </div>

      <UiCard>
        <button @click="showFields = !showFields" class="flex items-center gap-2 font-medium text-sm w-full">
          <span class="transition-transform" :class="{ 'rotate-90': showFields }">›</span>
          Field Selection ({{ selectedFields.length }})
        </button>
        <div v-if="showFields" class="mt-3 space-y-3">
          <div>
            <label class="block text-xs font-semibold text-gray-600 mb-2">Standard Fields</label>
            <div class="grid grid-cols-2 md:grid-cols-3 gap-2">
              <div v-for="field in availableFields" :key="field.key" class="flex items-center gap-1">
                <UiCheckbox
                  :model-value="selectedFields.includes(field.key)"
                  @update:model-value="toggleField(field.key)"
                  :label="getFieldLabel(field.key)"
                />
                <button
                  @click="startEditLabel(field.key)"
                  class="text-gray-400 hover:text-gray-600 text-xs"
                  title="Relabel"
                >
                  ✎
                </button>
              </div>
            </div>
            <div v-if="editingLabel" class="mt-3 flex gap-1">
              <UiInput v-model="labelInput" class="flex-1 text-sm" placeholder="Relabel column" @keyup.enter="saveLabel" />
              <UiButton @click="saveLabel" variant="secondary" size="sm">Save</UiButton>
              <UiButton @click="editingLabel = null" size="sm">Cancel</UiButton>
            </div>
          </div>
          <div class="border-t pt-3">
            <label class="block text-xs font-semibold text-gray-600 mb-2">Custom Fields</label>
            <div class="flex gap-1 mb-2">
              <UiInput
                v-model="customFieldInput"
                placeholder="Add custom field"
                class="flex-1 text-sm"
                @keyup.enter="addCustomField"
              />
              <UiButton @click="addCustomField" variant="secondary" size="sm">Add</UiButton>
            </div>
            <div v-if="customFields.length" class="flex flex-wrap gap-2">
              <span
                v-for="field in customFields"
                :key="field"
                class="inline-flex items-center gap-1 bg-blue-100 text-blue-800 px-2 py-1 rounded text-xs"
              >
                {{ field }}
                <button @click="removeCustomField(field)" class="font-bold hover:text-blue-900">×</button>
              </span>
            </div>
          </div>
        </div>
      </UiCard>
    </div>
    <UiCard :padded="false" class="relative min-h-31.25">
      <div class="w-full overflow-x-auto">
        <template v-if="previewDocs.length > 0">
          <!-- Mobile Table (Code & Title only) -->
          <div class="hidden md:block">
            <table class="text-sm w-full">
              <thead class="border-b border-gray-200 bg-gray-50">
                <tr>
                  <th v-for="col in tableColumns" :key="col.key" class="text-left px-4 py-3 font-semibold text-gray-900 whitespace-nowrap">
                    {{ col.label }}
                  </th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-200">
                <tr v-for="(row, idx) in previewDocs" :key="idx" class="hover:bg-gray-50 transition-colors">
                  <td v-for="col in tableColumns" :key="col.key" class="px-4 py-3 truncate max-w-xs">
                    <span v-if="Array.isArray(cellValue(row, col.key))" class="text-gray-600">
                      {{ cellValue(row, col.key).join(', ') }}
                    </span>
                    <span v-else class="text-gray-600">
                      {{ cellValue(row, col.key) }}
                    </span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- Desktop Table (All columns) -->
          <div class="md:hidden">
            <table class="text-sm w-full">
              <thead class="border-b border-gray-200 bg-gray-50">
                <tr>
                  <th v-for="col in mobileColumns" :key="col.key" class="text-left px-4 py-3 font-semibold text-gray-900 whitespace-nowrap">
                    {{ col.label }}
                  </th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-200">
                <tr v-for="(row, idx) in previewDocs" :key="idx" class="hover:bg-gray-50 transition-colors">
                  <td v-for="col in mobileColumns" :key="col.key" class="px-4 py-3 truncate max-w-xs">
                    <span v-if="Array.isArray(cellValue(row, col.key))" class="text-gray-600">
                      {{ cellValue(row, col.key).join(', ') }}
                    </span>
                    <span v-else class="text-gray-600">
                      {{ cellValue(row, col.key) }}
                    </span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </template>

        <div v-else class="p-8 text-center text-gray-600">
          No documents found
        </div>
      </div>
    </UiCard>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import UiButton from '@/core/ui/components/UiButton.vue'
import UiInput from '@/core/ui/components/UiInput.vue'
import UiCheckbox from '@/core/ui/components/UiCheckbox.vue'
import UiCard from '@/core/ui/components/UiCard.vue'
import UiSelect from '@/core/ui/components/UiSelect.vue'
import { DocumentService } from '@/modules/documents/services/documentService'

// Available fields from Document schema with display order
const fetchFields = async () => [
  { key: 'uuid', label: 'ID', visible: false },
  { key: 'code', label: 'Code', visible: true },
  { key: 'title', label: 'Title', visible: true },
  { key: 'folder_name', label: 'Folder', visible: false },
  { key: 'fields.nature', label: 'Nature', visible: true },
  { key: 'fields.status', label: 'Status', visible: true },
  { key: 'fields.complainants', label: 'Complainants', visible: true },
  { key: 'fields.respondents', label: 'Respondents', visible: true },
  { key: 'fields.complaint', label: 'Complaint', visible: true },
  { key: 'created_at', label: 'Date Filed', visible: true },
  { key: 'updated_at', label: 'Date Updated', visible: false },
]

const filters = ref({ 
  start: '', 
  end: '', 
  offset: 0, 
  limit: 100,
  sort_by: '',
  sort_order: 'asc' as 'asc' | 'desc'
})
const availableFields = ref<{ key: string, label: string, visible?: boolean }[]>([])
const selectedFields = ref<string[]>([])
const fieldLabels = ref<Record<string, string>>({}) // For custom labels
const previewDocs = ref<any[]>([])
const customFieldInput = ref('')
const customFields = ref<string[]>([])
const showFields = ref(false)
const editingLabel = ref<string | null>(null)
const labelInput = ref('')

const getFieldLabel = (key: string) => fieldLabels.value[key] || availableFields.value.find(f => f.key === key)?.label || key

const formatDate = (dateString: string) => {
  if (!dateString) return ''
  try {
    const date = new Date(dateString)
    return date.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' })
  } catch {
    return dateString
  }
}

const cellValue = (row: any, key: string) => {
  let value
  if (key.includes('.')) {
    const [obj, field] = key.split('.')
    value = obj && field ? row[obj]?.[field] ?? '' : ''
  } else {
    value = row[key] ?? ''
  }
  
  // Format date fields
  if ((key === 'created_at' || key === 'updated_at') && value) {
    return formatDate(value)
  }
  
  return value
}

const addCustomField = () => {
  const field = customFieldInput.value.trim()
  if (field && !customFields.value.includes(field)) {
    customFields.value.push(field)
    selectedFields.value = getOrderedFields([...selectedFields.value, field])
    customFieldInput.value = ''
  }
}

const removeCustomField = (field: string) => {
  customFields.value = customFields.value.filter(f => f !== field)
  selectedFields.value = selectedFields.value.filter(f => f !== field)
}

const getOrderedFields = (fields: string[]) => {
  const allFields = [...availableFields.value.map(f => f.key), ...customFields.value]
  return allFields.filter(f => fields.includes(f))
}

const startEditLabel = (key: string) => {
  editingLabel.value = key
  labelInput.value = getFieldLabel(key)
}

const saveLabel = () => {
  if (editingLabel.value && labelInput.value.trim()) {
    fieldLabels.value[editingLabel.value] = labelInput.value.trim()
  }
  editingLabel.value = null
}

const tableColumns = computed(() =>
  selectedFields.value.map(key => ({ key, label: getFieldLabel(key) }))
)

const mobileColumns = computed(() =>
  tableColumns.value.filter(col => col.key === 'code' || col.key === 'title')
)

const toggleField = (key: string) => {
  const idx = selectedFields.value.indexOf(key)
  if (idx > -1) {
    selectedFields.value.splice(idx, 1)
  } else {
    selectedFields.value.push(key)
  }
  // Reorder to match the master order
  selectedFields.value = getOrderedFields(selectedFields.value)
}

const applyFilters = async () => {
  try {
    const searchParams: any = {}
    if (filters.value.start) searchParams.date_from = filters.value.start
    if (filters.value.end) searchParams.date_to = filters.value.end
    if (filters.value.offset) searchParams.offset = filters.value.offset
    if (filters.value.limit) searchParams.limit = filters.value.limit
    if (filters.value.sort_by) searchParams.sort_by = filters.value.sort_by
    if (filters.value.sort_order === 'desc') searchParams.sort_desc = true
    
    previewDocs.value = (await DocumentService.search(searchParams)).data
  } catch (error) {
    console.error('Failed to load documents:', error)
    window.alert('An error occurred while loading documents. Please check your server.');
    previewDocs.value = []
  }
}

const load = async () => {
  availableFields.value = await fetchFields()
  selectedFields.value = availableFields.value
    .filter(f => f.visible !== false)
    .map(f => f.key)
  await applyFilters()
}

onMounted(load)

const exportCSV = () => {
  if (!previewDocs.value.length) return
  const rows = [selectedFields.value.map(getFieldLabel)]
  for (const doc of previewDocs.value) {
    rows.push(selectedFields.value.map(key => {
      const val = cellValue(doc, key)
      return Array.isArray(val) ? val.join(', ') : val ?? ''
    }))
  }
  const csv = rows.map(r => r.map(v => '"' + String(v).replace(/"/g, '""') + '"').join(',')).join('\n')
  const blob = new Blob([csv], { type: 'text/csv' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  const startDate = filters.value.start
  const endDate = filters.value.end
  const today = new Date().toISOString().split('T')[0]
  const filenameBase = startDate || endDate ? `report_${startDate || ''}${startDate && endDate ? '_to_' : ''}${endDate || ''}` : `report_${today}`
  a.href = url
  a.download = `${filenameBase}.csv`
  a.click()
  URL.revokeObjectURL(url)
}
</script>

<style scoped>
/* Minimal, rely on Tailwind */
</style>
