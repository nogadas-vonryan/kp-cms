<template>
	<div class="space-y-4">
		<div class="flex items-center justify-between gap-3">
			<div>
				<h2 class="text-xl font-semibold text-gray-900">Katarungang Pambarangay Forms</h2>
				<p class="text-sm text-gray-600">Select a template, fill the fields, and generate a PDF.</p>
			</div>
			<UiButton variant="ghost" @click="resetForm" :disabled="!selectedTemplate">Reset</UiButton>
		</div>

		<div class="grid gap-4 lg:grid-cols-3">
			<!-- Template Picker -->
			<div class="space-y-3">
				<div class="flex gap-2">
					<UiInput v-model="search" placeholder="Search templates" class="flex-1" />
					<UiButton variant="ghost" @click="search = ''" :disabled="!search">Clear</UiButton>
				</div>

			<div class="rounded-lg overflow-hidden bg-white shadow-sm max-h-140 overflow-y-auto">
					<button
						v-for="template in filteredTemplates"
						:key="template.id"
						type="button"
						class="w-full text-left px-4 py-3 hover:bg-gray-50 transition flex flex-col gap-1"
						:class="template.id === selectedTemplateId ? 'bg-blue-50' : ''"
						@click="selectTemplate(template.id)"
					>
						<div class="flex items-center justify-between gap-2">
							<span class="font-medium text-gray-900">{{ template.name }}</span>
							<span class="text-xs text-gray-500">{{ template.fields.length }} fields</span>
						</div>
						<p class="text-sm text-gray-600 line-clamp-2">{{ template.description }}</p>
					</button>
				</div>
			</div>

			<!-- Form Renderer -->
			<div class="lg:col-span-2">
				<div v-if="!selectedTemplate" class="p-6 rounded-lg bg-white shadow-sm text-gray-600">
					No templates available.
				</div>

				<div v-else class="p-6 rounded-lg bg-white shadow-sm space-y-4">
					<div>
						<h3 class="text-lg font-semibold text-gray-900">{{ selectedTemplate.name }}</h3>
						<p class="text-sm text-gray-600">{{ selectedTemplate.description }}</p>
					</div>

					<form class="space-y-4" @submit.prevent="onGenerate">
						<div v-for="field in selectedTemplate.fields" :key="field.key" class="space-y-1">
							<label class="text-sm font-medium text-gray-800 flex items-center gap-1">
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

							<UiSelect
								v-else-if="field.type === 'select'"
								v-model="formData[field.key]"
								:options="(field.options || []).map((opt) => ({ label: opt, value: opt }))"
								:placeholder="field.placeholder || 'Select'"
							/>

							<!-- Array field for dynamic inputs -->
							<div v-else-if="field.type === 'array'" class="space-y-2">
								<div v-for="(_, index) in (formData[field.key] as string[])" :key="index" class="flex gap-2">
									<UiInput
										v-model="(formData[field.key] as string[])[index]"
										:placeholder="field.placeholder"
										:required="field.required && index === 0"
										class="flex-1"
									/>
									<UiButton
										type="button"
										variant="ghost"
										@click="removeArrayItem(field.key, index)"
										class="px-2 py-1 text-red-600 hover:bg-red-50"
									>
										Remove
									</UiButton>
								</div>
								<UiButton
									type="button"
									variant="ghost"
									@click="addArrayItem(field.key)"
									class="text-blue-600 hover:bg-blue-50"
								>
									+ Add
								</UiButton>
							</div>

							<input v-else v-model="formData[field.key]" class="hidden" />

							<p v-if="field.helpText" class="text-xs text-gray-500">{{ field.helpText }}</p>
							<p v-if="errors[field.key]" class="text-xs text-red-600">{{ errors[field.key] }}</p>
						</div>

						<div class="flex items-center gap-3">
							<UiButton type="submit" variant="primary" :loading="isGenerating" :disabled="!selectedTemplate">
								Generate PDF
							</UiButton>
							<UiButton
								type="button"
								variant="primary"
								:loading="isUploading"
								:disabled="!selectedTemplate || !(props.uuid || props.documentId || props.context?.document?.uuid)"
								@click="onUpload"
							>
								Upload PDF
							</UiButton>
							<UiButton type="button" variant="ghost" @click="resetForm">Reset</UiButton>
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
import UiAlert from '@/core/ui/components/UiAlert.vue';
import UiButton from '@/core/ui/components/UiButton.vue';
import UiInput from '@/core/ui/components/UiInput.vue';
import UiSelect from '@/core/ui/components/UiSelect.vue';
import UiTextarea from '@/core/ui/components/UiTextarea.vue';
import type { ArchivistPlugin, PluginContext } from '@/core/plugins/pluginRegistry';
import type { KPTemplate } from './types';
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
