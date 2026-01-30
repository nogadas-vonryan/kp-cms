<template>
	<div class="space-y-4">
		<div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
			<div>
				<h2 class="text-lg sm:text-xl font-semibold text-gray-900">Katarungang Pambarangay Forms</h2>
				<p class="text-xs sm:text-sm text-gray-600">Select a template, fill the fields, and generate a PDF.</p>
			</div>
			<button 
				@click="resetForm" 
				:disabled="!selectedTemplate"
				class="px-3 py-1.5 text-xs font-medium text-gray-600 hover:bg-gray-100 rounded transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-1"
			>
				<RotateCcw :size="14" />
				<span>Reset</span>
			</button>
		</div>

		<!-- Mobile: Toggle between template list and form -->
		<div class="lg:hidden">
			<div class="flex gap-0 border border-gray-300 rounded bg-white">
				<button
					@click="mobileView = 'templates'"
					:class="[
						'flex-1 px-3 py-2 text-xs font-bold transition-colors flex items-center justify-center gap-1',
						mobileView === 'templates'
							? 'bg-gray-900 text-white'
							: 'text-gray-600 hover:bg-gray-100'
					]"
				>
					<FileText :size="14" />
					<span>Templates</span>
				</button>
				<button
					@click="mobileView = 'form'"
					:class="[
						'flex-1 px-3 py-2 text-xs font-bold transition-colors flex items-center justify-center gap-1 border-l border-gray-300',
						mobileView === 'form'
							? 'bg-gray-900 text-white'
							: 'text-gray-600 hover:bg-gray-100'
					]"
				>
					<Edit :size="14" />
					<span>Form</span>
				</button>
			</div>
		</div>

		<div class="grid gap-4 lg:grid-cols-3">
			<!-- Template Picker -->
			<div class="space-y-3" :class="mobileView === 'templates' ? 'block' : 'hidden lg:block'">
				<div class="flex gap-2">
					<div class="relative flex-1">
						<Search class="absolute left-2 top-1/2 -translate-y-1/2 text-gray-400" :size="16" />
						<input 
							v-model="search" 
							placeholder="Search templates" 
							class="w-full pl-9 pr-3 py-2 text-sm border border-gray-300 rounded bg-white focus:ring-1 focus:ring-blue-500 outline-none"
						/>
					</div>
					<button 
						@click="search = ''" 
						:disabled="!search"
						class="px-3 py-2 text-xs font-medium text-gray-600 hover:bg-gray-100 rounded transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-1"
					>
						<X :size="14" />
					</button>
				</div>

			<div class="rounded-lg overflow-hidden bg-white shadow-sm max-h-[60vh] lg:max-h-[70vh] overflow-y-auto border border-gray-200">
					<button
						v-for="template in filteredTemplates"
						:key="template.id"
						type="button"
						class="w-full text-left px-3 sm:px-4 py-2.5 sm:py-3 hover:bg-gray-50 transition flex flex-col gap-1 border-b border-gray-100 last:border-b-0"
						:class="template.id === selectedTemplateId ? 'bg-blue-50 border-blue-200' : ''"
						@click="selectTemplate(template.id); mobileView = 'form'"
					>
						<div class="flex items-center justify-between gap-2">
							<span class="text-sm font-medium text-gray-900">{{ template.name }}</span>
							<span class="text-[10px] sm:text-xs text-gray-500 shrink-0">{{ template.fields.length }} fields</span>
						</div>
						<p class="text-xs sm:text-sm text-gray-600 line-clamp-2">{{ template.description }}</p>
					</button>
				</div>
			</div>

			<!-- Form Renderer -->
			<div class="lg:col-span-2" :class="mobileView === 'form' ? 'block' : 'hidden lg:block'">
				<div v-if="!selectedTemplate" class="p-4 sm:p-6 rounded-lg bg-white shadow-sm border border-gray-200 text-center">
					<FileQuestion class="mx-auto mb-2 text-gray-400" :size="32" />
					<p class="text-sm text-gray-600">No templates available.</p>
				</div>

				<div v-else class="p-4 sm:p-6 rounded-lg bg-white shadow-sm border border-gray-200 space-y-4">
					<div class="flex items-start justify-between gap-2">
						<div class="flex-1">
							<h3 class="text-base sm:text-lg font-semibold text-gray-900">{{ selectedTemplate.name }}</h3>
							<p class="text-xs sm:text-sm text-gray-600">{{ selectedTemplate.description }}</p>
						</div>
						<button
							@click="mobileView = 'templates'"
							class="lg:hidden p-1.5 text-gray-600 hover:bg-gray-100 rounded transition-colors"
							title="Back to templates"
						>
							<ChevronLeft :size="20" />
						</button>
					</div>

					<form class="space-y-3 sm:space-y-4" @submit.prevent="onGenerate">
						<div v-for="field in selectedTemplate.fields" :key="field.key" class="space-y-1">
							<label class="text-xs sm:text-sm font-medium text-gray-800 flex items-center gap-1">
								<span>{{ field.label }}</span>
								<span v-if="field.required" class="text-red-500">*</span>
							</label>

							<UiInput
								v-if="field.type === 'text' || field.type === 'date'"
								v-model="formData[field.key]"
								:type="field.type === 'date' ? 'date' : 'text'"
								:placeholder="field.placeholder"
								:required="field.required"
							/>

							<UiTextarea
								v-else-if="field.type === 'textarea'"
								v-model="formData[field.key]"
								:placeholder="field.placeholder"
								:required="field.required"
								:rows="4"
							/>

							<UiComboBox
								v-else-if="field.type === 'select'"
								v-model="formData[field.key]"
								:options="getFieldOptions(field)"
								:placeholder="field.placeholder || 'Select'"
							/>
							<div v-else-if="field.type === 'array'" class="space-y-2">
								<div v-for="(_, index) in (formData[field.key] as string[])" :key="index" class="flex gap-1.5">
									<input
										v-model="(formData[field.key] as string[])[index]"
										:placeholder="field.placeholder"
										:required="field.required && index === 0"
										class="flex-1 text-sm p-2 border border-gray-300 rounded bg-white focus:ring-1 focus:ring-blue-500 outline-none"
									/>
									<button
										type="button"
										@click="removeArrayItem(field.key, index)"
										class="p-2 text-red-600 hover:bg-red-600 hover:text-white rounded transition-colors shrink-0"
										title="Remove"
									>
										<Trash2 :size="16" />
									</button>
								</div>
								<button
									type="button"
									@click="addArrayItem(field.key)"
									class="w-full px-3 py-2 text-xs font-bold text-blue-600 hover:bg-blue-100 rounded transition-colors flex items-center justify-center gap-1"
								>
									<Plus :size="14" />
									<span>Add {{ field.label }}</span>
								</button>
							</div>

							<input v-else v-model="formData[field.key]" class="hidden" />

							<p v-if="field.helpText" class="text-xs text-gray-500">{{ field.helpText }}</p>
							<p v-if="errors[field.key]" class="text-xs text-red-600 flex items-center gap-1">
								<AlertCircle :size="12" />
								<span>{{ errors[field.key] }}</span>
							</p>
						</div>

						<div class="flex flex-col sm:flex-row items-stretch sm:items-center gap-2 pt-2 border-t border-gray-200">
							<button
								type="submit"
								:disabled="isGenerating || !selectedTemplate"
								class="flex-1 px-4 py-2.5 bg-blue-600 text-white text-sm font-medium rounded hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
							>
								<Download :size="16" :class="isGenerating ? 'animate-pulse' : ''" />
								<span>{{ isGenerating ? 'Generating...' : 'Generate PDF' }}</span>
							</button>
							<button
								type="button"
								@click="onUpload"
								:disabled="isUploading || !selectedTemplate || !(props.uuid || props.documentId || props.context?.document?.uuid)"
								class="flex-1 px-4 py-2.5 bg-green-600 text-white text-sm font-medium rounded hover:bg-green-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
							>
								<Upload :size="16" :class="isUploading ? 'animate-pulse' : ''" />
								<span>{{ isUploading ? 'Uploading...' : 'Upload PDF' }}</span>
							</button>
							<button
								type="button"
								@click="resetForm"
								class="px-4 py-2.5 text-sm font-medium text-gray-600 hover:bg-gray-100 rounded transition-colors flex items-center justify-center gap-2"
							>
								<RotateCcw :size="16" />
								<span class="hidden sm:inline">Reset</span>
							</button>
						</div>
					</form>

					<UiAlert v-if="statusMessage" :type="statusMessage.type" class="mt-2">
						{{ statusMessage.text }}
					</UiAlert>
				</div>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import { Download, Upload, Edit, Trash2, X, Plus, RotateCcw, Search, FileText, FileQuestion, AlertCircle, ChevronLeft } from 'lucide-vue-next';
import UiAlert from '@/core/ui/components/UiAlert.vue';
import UiInput from '@/core/ui/components/UiInput.vue';
import UiComboBox from '@/core/ui/components/UiComboBox.vue';
import UiTextarea from '@/core/ui/components/UiTextarea.vue';
import type { ArchivistPlugin, PluginContext } from '@/core/plugins/pluginRegistry';
import type { KPTemplate, KPFormField } from './types';
import { kpTemplates } from './templates';
import { DocumentService } from '@/modules/documents/services/documentService';

const pdfMake = window.pdfMake;

const props = defineProps<{ plugin?: ArchivistPlugin; context?: PluginContext; documentId?: string; uuid?: string }>();

const search = ref('');
const selectedTemplateId = ref(kpTemplates[0]?.id || '');
const formData = reactive<Record<string, any>>({});
const errors = reactive<Record<string, string>>({});
const isGenerating = ref(false);
const isUploading = ref(false);
const statusMessage = ref<{ type: 'success' | 'error' | 'info'; text: string } | null>(null);
const mobileView = ref<'templates' | 'form'>('templates');

const filteredTemplates = computed(() => {
	const term = search.value.toLowerCase().trim();
	if (!term) return kpTemplates;
	return kpTemplates.filter(
		(t) => t.name.toLowerCase().includes(term) || t.description.toLowerCase().includes(term)
	);
});

const selectedTemplate = computed<KPTemplate | undefined>(() =>
	kpTemplates.find((t) => t.id === selectedTemplateId.value)
);

watch(selectedTemplate, (tpl) => {
	if (tpl) initForm(tpl);
}, { immediate: true });

function getFieldOptions(field: KPFormField): string[] {
	// For punongBarangay, get options from localStorage
	if (field.key === 'punongBarangay') {
		const stored = localStorage.getItem('punongBarangay');
		if (stored) {
			try {
				const list = JSON.parse(stored);
				if (Array.isArray(list) && list.length > 0) {
					return list;
				}
			} catch (e) {
				console.error('Failed to parse punongBarangay from localStorage:', e);
			}
		}
	}
	// Return field's static options or empty array
	return field.options || [];
}

function derivePrefill(key: string): string | string[] {
	const doc = props.context?.document;
	if (!doc) return key === 'complainants' || key === 'respondents' ? [''] : '';
	if (key === 'caseNo') return doc.code || '';
	if ((key === 'complainant' || key === 'complainants') && Array.isArray(doc.fields?.complainants)) {
		return (doc.fields.complainants as string[]);
	}
	if ((key === 'respondent' || key === 'respondents') && Array.isArray(doc.fields?.respondents)) {
		return (doc.fields.respondents as string[]);
	}
	if (key === 'title' || key === 'subject') return doc.title || '';
	if (key === 'complaint') return doc.fields.complaint || '';
	if (key === 'punongBarangay') {
		// Get from localStorage for select field
		const stored = localStorage.getItem('punongBarangay');
		let punongBarangayList: string[] = [];
		
		if (stored) {
			try {
				punongBarangayList = JSON.parse(stored);
			} catch (e) {
				console.error('Failed to parse punongBarangay from localStorage:', e);
			}
		}
		
		// Return the current value from doc or first from list or empty string
		const currentValue = doc.fields?.punongBarangay;
		return currentValue || (punongBarangayList.length > 0 ? punongBarangayList[0] : '');
	}
	return '';
}

function initForm(tpl: KPTemplate) {
	statusMessage.value = null;
	Object.keys(formData).forEach((k) => delete formData[k]);
	Object.keys(errors).forEach((k) => delete errors[k]);

	const today = new Date().toISOString().split('T')[0];

	tpl.fields.forEach((field) => {
		const prefill = derivePrefill(field.key);
		
		if (field.type === 'array') {
			// For array fields, initialize with prefilled data or empty array with one input
			if (Array.isArray(prefill) && prefill.length > 0) {
				formData[field.key] = prefill;
			} else {
				formData[field.key] = [''];
			}
		} else {
			const baseValue = field.type === 'date' && !field.defaultValue ? today : field.defaultValue || '';
			formData[field.key] = typeof prefill === 'string' ? prefill || baseValue : baseValue;
		}
	});
}

function selectTemplate(id: string) {
	selectedTemplateId.value = id;
}

function validate(): boolean {
	if (!selectedTemplate.value) return false;
	Object.keys(errors).forEach((k) => delete errors[k]);

	selectedTemplate.value.fields.forEach((field) => {
		const value = formData[field.key];
		
		if (field.type === 'array') {
			const items = (value as string[]).filter(item => item && String(item).trim() !== '');
			if (field.required && items.length === 0) {
				errors[field.key] = 'Please add at least one entry';
			}
		} else {
			if (field.required && (!value || String(value).trim() === '')) {
				errors[field.key] = 'This field is required';
			}
		}
	});

	return Object.keys(errors).length === 0;
}

function addArrayItem(fieldKey: string) {
	if (!Array.isArray(formData[fieldKey])) {
		formData[fieldKey] = [''];
	} else {
		(formData[fieldKey] as string[]).push('');
	}
}

function removeArrayItem(fieldKey: string, index: number) {
	if (Array.isArray(formData[fieldKey])) {
		(formData[fieldKey] as string[]).splice(index, 1);
		// Ensure at least one empty field remains
		if ((formData[fieldKey] as string[]).length === 0) {
			(formData[fieldKey] as string[]).push('');
		}
	}
}

function fileName(): string {
	const tpl = selectedTemplate.value;
	const slug = tpl ? tpl.id : 'kp-form';
	const docRef = props.context?.document?.code || props.documentId || props.uuid || 'document';
	return `${slug}-${docRef}.pdf`;
}

async function onGenerate() {
	if (!selectedTemplate.value) return;
	if (!validate()) return;

	try {
		isGenerating.value = true;
		
		// Save punongBarangay to localStorage if it exists and not already saved
		if (formData.punongBarangay && typeof formData.punongBarangay === 'string' && formData.punongBarangay.trim()) {
			const stored = localStorage.getItem('punongBarangay');
			let punongBarangayList: string[] = [];
			
			if (stored) {
				try {
					punongBarangayList = JSON.parse(stored);
				} catch (e) {
					console.error('Failed to parse punongBarangay from localStorage:', e);
				}
			}
			
			if (!punongBarangayList.includes(formData.punongBarangay)) {
				punongBarangayList.push(formData.punongBarangay);
				localStorage.setItem('punongBarangay', JSON.stringify(punongBarangayList));
			}
		}
		
		// Convert array fields to newline-separated strings for PDF generation
		const dataForPdf = { ...formData };
		selectedTemplate.value.fields.forEach((field) => {
			if (field.type === 'array' && Array.isArray(dataForPdf[field.key])) {
				dataForPdf[field.key] = (dataForPdf[field.key] as string[])
					.filter(item => item && String(item).trim() !== '')
					.join('\n');
			}
		});
		
		const docDef = selectedTemplate.value.generatePdf(dataForPdf);
		pdfMake.createPdf(docDef).download(fileName());
		statusMessage.value = { type: 'success', text: 'PDF generated and downloaded.' };
		props.context?.events.onFileAdd?.(null);
	} catch (err: any) {
		console.error('Failed to generate PDF', err);
		statusMessage.value = { type: 'error', text: err?.message || 'Failed to generate PDF.' };
	} finally {
		isGenerating.value = false;
	}
}

function getPdfBlob(docDef: any): Promise<Blob> {
	const pdf: any = pdfMake.createPdf(docDef);
	try {
		const result = (pdf.getBlob as any)();
		if (result && typeof result.then === 'function') {
			return result as Promise<Blob>;
		}
	} catch {}
	return new Promise((resolve, reject) => {
		try {
			pdf.getBlob((blob: Blob) => resolve(blob));
		} catch (e) {
			reject(e);
		}
	});
}

async function onUpload() {
	if (!selectedTemplate.value) return;
	if (!validate()) return;

	const uuid = props.uuid || props.documentId || props.context?.document?.uuid;
	if (!uuid) {
		statusMessage.value = { type: 'error', text: 'No document ID found to upload.' };
		return;
	}

	try {
		isUploading.value = true;

		// Save punongBarangay to localStorage if it exists and not already saved
		if (formData.punongBarangay && typeof formData.punongBarangay === 'string' && formData.punongBarangay.trim()) {
			const stored = localStorage.getItem('punongBarangay');
			let punongBarangayList: string[] = [];
			
			if (stored) {
				try {
					punongBarangayList = JSON.parse(stored);
				} catch (e) {
					console.error('Failed to parse punongBarangay from localStorage:', e);
				}
			}
			
			if (!punongBarangayList.includes(formData.punongBarangay)) {
				punongBarangayList.push(formData.punongBarangay);
				localStorage.setItem('punongBarangay', JSON.stringify(punongBarangayList));
			}
		}

		const dataForPdf = { ...formData };
		selectedTemplate.value.fields.forEach((field) => {
			if (field.type === 'array' && Array.isArray(dataForPdf[field.key])) {
				dataForPdf[field.key] = (dataForPdf[field.key] as string[])
					.filter(item => item && String(item).trim() !== '')
					.join('\n');
			}
		});

		const docDef = selectedTemplate.value.generatePdf(dataForPdf);
		const blob = await getPdfBlob(docDef);
		const file = new File([blob], fileName(), { type: 'application/pdf' });

		await DocumentService.uploadFile(uuid, file);
		statusMessage.value = { type: 'success', text: 'PDF uploaded successfully.' };
		props.context?.events.onFileAdd?.(file.name);
	} catch (err: any) {
		console.error('Failed to upload PDF', err);
		statusMessage.value = { type: 'error', text: err?.message || 'Failed to upload PDF.' };
	} finally {
		isUploading.value = false;
	}
}

function resetForm() {
	if (selectedTemplate.value) initForm(selectedTemplate.value);
}
</script>
