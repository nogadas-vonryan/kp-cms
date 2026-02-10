<template>
  <UiCard class="relative">
    <div 
      v-if="isProcessing" 
      class="absolute inset-0 z-20 flex items-center justify-center bg-white/60 backdrop-blur-[1px] rounded-lg"
    >
      <div class="flex flex-col items-center gap-2">
        <span class="text-sm font-medium text-gray-600 animate-pulse">Processing import...</span>
      </div>
    </div>

    <div class="space-y-6">
      <div class="space-y-3">
        <label class="text-xs font-semibold text-gray-600 uppercase tracking-wider block">
          Select Source File
        </label>
        
        <div 
          class="border-2 border-dashed border-gray-300 rounded-lg p-8 text-center hover:bg-gray-50 transition-colors cursor-pointer relative"
          @dragover.prevent
          @drop.prevent="handleDrop"
        >
          <input 
            type="file" 
            ref="fileInput"
            accept=".csv"
            class="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
            @change="handleFileUpload"
          />
          <div class="flex flex-col items-center gap-2 text-gray-500">
            <FileSpreadsheet :size="32" class="text-gray-400" />
            <div v-if="selectedFile" class="font-medium text-gray-900">
              {{ selectedFile.name }}
            </div>
            <div v-else class="text-sm">
              <span class="text-blue-600 font-medium">Click to upload</span> or drag and drop CSV
            </div>
          </div>
        </div>
      </div>

      <details v-if="selectedFile" class="group" open>
        <summary class="text-sm text-blue-600 hover:text-blue-700 cursor-pointer font-medium list-none flex items-center gap-2 select-none p-2 hover:bg-blue-50 rounded transition-colors w-fit">
          <ChevronRight :size="16" class="group-open:rotate-90 transition-transform" />
          <span>Column Mapping Configuration</span>
        </summary>
        
        <div class="pt-4 mt-2 border-t border-gray-100 space-y-8">
          <div>
            <h3 class="text-xs font-bold text-gray-400 uppercase tracking-widest mb-4">Core Inhabitant Fields</h3>
            <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
              <div v-for="key in coreFieldKeys" :key="key">
                <label class="text-[10px] font-bold text-gray-500 uppercase mb-1.5 block">
                  {{ formatLabel(key) }}
                </label>
                <select
                  v-model="columnMap[key]"
                  class="w-full text-sm p-2 border border-gray-300 rounded-lg bg-white focus:ring-2 focus:ring-blue-500 outline-none transition-all"
                >
                  <option value="">(Ignore Field)</option>
                  <option v-for="header in detectedHeaders" :key="header" :value="header">
                    {{ header }}
                  </option>
                </select>
              </div>
            </div>
          </div>

          <div v-if="customFieldKeys.length > 0">
            <h3 class="text-xs font-bold text-gray-400 uppercase tracking-widest mb-4">Detected Custom Fields</h3>
            <p class="text-[11px] text-gray-500 mb-4 italic">These fields are unmapped by default and will be ignored.</p>
            <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
              <div v-for="key in customFieldKeys" :key="key">
                <label class="text-[10px] font-bold text-gray-500 uppercase mb-1.5 block">
                  Field: {{ formatLabel(key) }}
                </label>
                <select
                  v-model="columnMap[key]"
                  class="w-full text-sm p-2 border border-blue-200 bg-blue-50/30 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none transition-all"
                  disabled
                >
                  <option value="">(Unmapped / Ignore)</option>
                </select>
              </div>
            </div>
          </div>
        </div>
      </details>

      <div class="flex items-center justify-end gap-3 pt-4 border-t border-gray-100">
        <UiButton 
          v-if="selectedFile"
          variant="secondary" 
          @click="resetForm"
          :disabled="isProcessing"
        >
          Cancel
        </UiButton>
        <UiButton 
          @click="processImport" 
          :disabled="!selectedFile || isProcessing"
          class="flex items-center gap-2"
        >
          <Upload :size="16" />
          <span>Start Import</span>
        </UiButton>
      </div>

      <div v-if="importResult" class="animate-in fade-in slide-in-from-bottom-2 duration-300">
         <UiSystemNotice 
          :type="importResult.type === 'success' ? 'success' : 'error'" 
          :label="importResult.type === 'success' ? 'Import Task Complete' : 'Import Partially Failed'"
          :modelValue="true"
          class="w-full"
        >
          <template #title>
            <div class="px-2 text-sm leading-relaxed w-full min-w-0">
              <p class="font-medium text-gray-900 mb-4">{{ importResult.message }}</p>
              
              <div v-if="importResult.details && importResult.details.length" class="mt-4 w-full">
                <div class="flex items-center justify-between mb-2 border-b border-gray-100 pb-2">
                  <span class="text-[10px] uppercase tracking-widest font-bold text-gray-400">
                    Issue Log
                  </span>
                  <span class="text-[10px] font-medium px-1.5 py-0.5 bg-gray-100 text-gray-500 rounded">
                    {{ importResult.details.length }} items need attention
                  </span>
                </div>
                
                <div class="max-h-60 overflow-y-auto rounded-md border border-gray-100 bg-gray-50/50">
                  <ul class="divide-y divide-gray-100 font-mono text-[11px] leading-normal text-gray-600">
                    <li 
                      v-for="(detail, idx) in importResult.details" 
                      :key="idx"
                      class="px-3 py-2 hover:bg-white transition-colors flex gap-3"
                    >
                      <span class="text-gray-300 select-none">{{ String(idx + 1).padStart(2, '0') }}</span>
                      <span class="break-all">{{ detail }}</span>
                    </li>
                  </ul>
                </div>
              </div>
            </div>
          </template>
        </UiSystemNotice>
      </div>
    </div>
  </UiCard>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue';
import { FileSpreadsheet, ChevronRight, Upload } from 'lucide-vue-next';
import { InhabitantService } from '@/modules/inhabitants/services/inhabitantService';
import UiCard from '@/core/ui/components/UiCard.vue';
import UiButton from '@/core/ui/components/UiButton.vue';
import UiSystemNotice from '@/core/ui/components/UiSystemNotice.vue';

const selectedFile = ref<File | null>(null);
const isProcessing = ref(false);
const fileInput = ref<HTMLInputElement | null>(null);
const detectedHeaders = ref<string[]>([]);

type ImportResult = {
  type: 'success' | 'error';
  message: string;
  details?: string[];
} | null;

const importResult = ref<ImportResult>(null);

const coreFieldKeys = [
  'first_name', 'last_name', 'middle_name', 'suffix', 'birthdate', 'contact_no', 'address',
  'civil_status', 'citizenship', 'inhabitant_type', 'sex', 'birth_place', 'occupation',
  'email_address', 'highest_educational_attainment', 'mother_first_name', 'mother_middle_name', 'mother_last_name'
];

const columnMap = reactive<Record<string, string>>({});

const customFieldKeys = computed(() => {
  return Object.keys(columnMap).filter(key => !coreFieldKeys.includes(key));
});

const parseCSVLine = (line: string) => {
  const result = [];
  let cur = '';
  let inQuotes = false;
  for (let i = 0; i < line.length; i++) {
    const char = line[i];
    if (char === '"') {
      inQuotes = !inQuotes;
    } else if (char === ',' && !inQuotes) {
      result.push(cur.trim().replace(/^"|"$/g, ''));
      cur = '';
    } else {
      cur += char;
    }
  }
  result.push(cur.trim().replace(/^"|"$/g, ''));
  return result;
};

const extractHeaders = (file: File) => {
  const reader = new FileReader();
  reader.onload = (e) => {
    const text = e.target?.result as string;
    const firstLine = text.split(/\r?\n/)[0];
    if (firstLine) {
      detectedHeaders.value = parseCSVLine(firstLine);
      initializeMapping();
    }
  };
  reader.readAsText(file);
};

const initializeMapping = () => {
  Object.keys(columnMap).forEach(key => delete columnMap[key]);

  coreFieldKeys.forEach(key => { columnMap[key] = ''; });

  detectedHeaders.value.forEach(header => {
    const cleanHeader = header.toLowerCase().trim().replace(/\s+/g, '_');
    if (coreFieldKeys.includes(cleanHeader)) {
      columnMap[cleanHeader] = header;
    } else {
      columnMap[cleanHeader] = '';
    }
  });
};

const handleFileUpload = (event: Event) => {
  const target = event.target as HTMLInputElement;
  if (target.files?.[0]) {
    selectedFile.value = target.files[0];
    extractHeaders(target.files[0]);
  }
};

const handleDrop = (event: DragEvent) => {
  const file = event.dataTransfer?.files?.[0];
  if (file && (file.type === 'text/csv' || file.name.endsWith('.csv'))) {
    selectedFile.value = file;
    extractHeaders(file);
  }
};

const formatLabel = (key: string) => {
  return key.charAt(0).toUpperCase() + key.slice(1).replace(/_/g, ' ');
};

const resetForm = () => {
  selectedFile.value = null;
  importResult.value = null;
  detectedHeaders.value = [];
  Object.keys(columnMap).forEach(key => delete columnMap[key]);
  if (fileInput.value) fileInput.value.value = '';
};

const convertToDateString = (dateStr: string): string => {
  if (!dateStr) return '';
  const timestamp = Date.parse(dateStr);
  if (!isNaN(timestamp)) {
    const date = new Date(timestamp);
    return date.toISOString().split('T')[0] || ''; // YYYY-MM-DD
  }
  return dateStr;
};

const mapRowToInhabitant = (row: Record<string, string>) => {
  const getVal = (key: string) => {
    const header = columnMap[key];
    return header ? (row[header] || '').trim() : '';
  };

  const inhabitant: any = {
    first_name: getVal('first_name'),
    last_name: getVal('last_name'),
    middle_name: getVal('middle_name'),
    suffix: getVal('suffix'),
    birthdate: convertToDateString(getVal('birthdate')),
    contact_no: getVal('contact_no'),
    address: getVal('address'),
    civil_status: getVal('civil_status'),
    citizenship: getVal('citizenship'),
    inhabitant_type: getVal('inhabitant_type'),
    sex: getVal('sex'),
    birth_place: getVal('birth_place'),
    occupation: getVal('occupation'),
    email_address: getVal('email_address'),
    highest_educational_attainment: getVal('highest_educational_attainment'),
    mother_first_name: getVal('mother_first_name'),
    mother_middle_name: getVal('mother_middle_name'),
    mother_last_name: getVal('mother_last_name'),
  };

  // Remove empty fields
  Object.keys(inhabitant).forEach(key => {
    if (!inhabitant[key]) delete inhabitant[key];
  });

  return inhabitant;
};

const processImport = async () => {
  if (!selectedFile.value) return;

  isProcessing.value = true;
  importResult.value = null;
  const errorDetails = ref<string[]>([]); 
  let successCount = 0;

  const reader = new FileReader();

  reader.onload = async (e) => {
    try {
      const text = e.target?.result as string;
      const lines = text.split(/\r?\n/).filter(line => line.trim() !== '');
      if (lines.length < 2) throw new Error("CSV contains no data rows");

      const headers = parseCSVLine(lines[0] || '');
      const dataRows = lines.slice(1).map(line => {
        const values = parseCSVLine(line);
        return headers.reduce((obj, header, index) => {
          obj[header] = values[index] || '';
          return obj;
        }, {} as Record<string, string>);
      });

      for (let i = 0; i < dataRows.length; i++) {
        const rowNum = i + 2;
        try {
          const row = dataRows[i];
          if (!row) continue;

          const payload = mapRowToInhabitant(row);
          if (!payload.first_name || !payload.last_name) {
            errorDetails.value.push(`Row ${rowNum}: First Name or Last Name column is unmapped.`);
            continue; 
          }

          await InhabitantService.create(payload);
          successCount++;
        } catch (err: any) {
          const contextError = err.response?.data?.error || err.response?.data?.message || err.message || 'Unknown Error';
          errorDetails.value.push(`Row ${rowNum}: ${contextError}`);
        }
      }

      importResult.value = {
        type: errorDetails.value.length === 0 ? 'success' : 'error',
        message: `Import finished. Successfully created ${successCount} of ${dataRows.length} inhabitants.`,
        details: errorDetails.value
      };

    } catch (err: any) {
      importResult.value = {
        type: 'error',
        message: 'Critical error processing file.',
        details: [err.message]
      };
    } finally {
      isProcessing.value = false;
    }
  };

  reader.readAsText(selectedFile.value);
};
</script>