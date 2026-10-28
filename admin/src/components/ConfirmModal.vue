<!-- src/components/ConfirmModal.vue (CoreUI v3 / Vue 2) -->
<template>
  <CModal :show="isOpen" @update:show="handleCancel" centered title="Confirm Action">
    <template #header>
      <h5 class="modal-title">{{ options.title || 'Confirm Action' }}</h5>
    </template>

    {{ options.message }}

    <template #footer>
      <CButton :color= "options.cancelColor || 'secondary'" @click="handleCancel">
        {{ options.cancelText || 'Cancel' }}
      </CButton>
      <CButton v-if="options.confirmText && options.confirmText.length > 0" :color="options.confirmColor || 'primary'" @click="handleConfirm">
        {{ options.confirmText || 'Proceed' }}
      </CButton>
    </template>
  </CModal>
</template>

<script>
export default {
  name: 'ConfirmModal',
  props: {
    isOpen: {
      type: Boolean,
      default: false,
    },
    options: {
      type: Object,
      default: () => ({ message: '' }),
    },
  },
  methods: {
    handleConfirm() {
      this.$emit('confirm')
    },
    handleCancel() {
      this.$emit('cancel')
    },
  },
}
</script>
