<template>
  <UiModal
    v-model:open="isOpen"
    title="Add New Inhabitant"
    size="lg"
    :prevent-close="submitting"
  >
    <InhabitantForm
      ref="formRef"
      :loading="submitting"
      :quick-mode="true"
    />

    <template #footer>
      <div class="flex w-full flex-col gap-2">
        <UiAlert v-if="error" type="error">{{ error }}</UiAlert>
        <div class="flex w-full flex-col-reverse gap-2 sm:flex-row sm:justify-end">
          <UiButton @click="handleCancel" :disabled="submitting">Cancel</UiButton>
          <UiButton
            @click="handleSubmit"
            :disabled="submitting"
            variant="secondary"
            class="flex items-center gap-2"
          >
            <Check :size="16" />
            <span>{{ submitting ? 'Saving...' : 'Save' }}</span>
          </UiButton>
        </div>
      </div>
    </template>
  </UiModal>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue';
import { Check } from 'lucide-vue-next';
import UiModal from '@/core/ui/components/UiModal.vue';
import UiButton from '@/core/ui/components/UiButton.vue';
import UiAlert from '@/core/ui/components/UiAlert.vue';
import InhabitantForm from './InhabitantForm.vue';
import { InhabitantService, type Inhabitant } from '@/modules/inhabitants/services/inhabitantService';

const props = defineProps<{
  open: boolean;
}>();

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void;
  (e: 'created', inhabitant: Inhabitant): void;
}>();

const isOpen = ref(props.open);
const submitting = ref(false);
const error = ref('');
const formRef = ref<InstanceType<typeof InhabitantForm> | null>(null);

// Sync open state with parent
watch(() => props.open, (val) => {
  isOpen.value = val;
});

watch(isOpen, (val) => {
  emit('update:open', val);
  if (!val) {
    // Reset on close
    error.value = '';
  }
});

async function handleSubmit() {
  if (!formRef.value) return;

  error.value = '';
  submitting.value = true;

  try {
    const formData = formRef.value.formData;

    // Validate birthdate if provided
    if (formData.birthdate) {
      const birthDate = new Date(formData.birthdate);
      const today = new Date();
      const age = today.getFullYear() - birthDate.getFullYear();
      if (age > 120 || birthDate > today) {
        error.value = 'Please enter a valid birthdate';
        return;
      }
    }

    // Validate email if provided
    if (formData.email_address) {
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      if (!emailRegex.test(formData.email_address)) {
        error.value = 'Please enter a valid email address';
        return;
      }
    }

    const response = await InhabitantService.create(formData);
    emit('created', response.data);
    isOpen.value = false;
  } catch (err: any) {
    error.value = err.response?.data?.error || 'Failed to create inhabitant';
  } finally {
    submitting.value = false;
  }
}

function handleCancel() {
  if (submitting.value) return;
  isOpen.value = false;
}
</script>
