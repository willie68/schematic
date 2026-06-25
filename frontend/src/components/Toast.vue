<template>
  <div class="toast-container">
    <!-- Toasts -->
    <div
      v-for="toast in displayedToasts"
      :key="toast.id"
      :class="['toast', `toast-${toast.type}`]"
    >
      <div class="toast-content">
        <span>{{ toast.message }}</span>
        <button class="toast-close" @click="removeToast(toast.id)">✕</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { useToast } from '../composables/useToast'
import { computed } from 'vue'

const { toasts, removeToast } = useToast()

// Use computed to ensure Vue reactivity
const displayedToasts = computed(() => toasts.value)
</script>

<style scoped>
.toast-container {
  position: fixed;
  bottom: 1.5rem;
  right: 1.5rem;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 0.75rem;
  z-index: 9999;
  pointer-events: none;
}

.toast {
  min-width: 300px;
  max-width: min(420px, calc(100vw - 3rem));
  padding: 1rem;
  border-radius: 0.375rem;
  background-color: white;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  animation: slideIn 0.3s ease-in-out;
  pointer-events: auto;
  display: flex;
  align-items: center;
}

.toast-content {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  width: 100%;
  gap: 0.75rem;
  word-break: break-word;
}

.toast-close {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: inherit;
  opacity: 0.7;
  padding: 0;
  min-width: 2.5rem;
  text-align: center;
  margin-left: auto;
  flex-shrink: 0;
  line-height: 1;
}

.toast-close:hover {
  opacity: 1;
}

.toast-info {
  background-color: #dbeafe;
  color: #1e40af;
  border-left: 4px solid #3b82f6;
}

.toast-success {
  background-color: #dcfce7;
  color: #166534;
  border-left: 4px solid #10b981;
}

.toast-warning {
  background-color: #fef3c7;
  color: #92400e;
  border-left: 4px solid #f59e0b;
}

.toast-error {
  background-color: #fee2e2;
  color: #991b1b;
  border-left: 4px solid #ef4444;
}

@keyframes slideIn {
  from {
    transform: translateX(450px);
    opacity: 0;
  }
  to {
    transform: translateX(0);
    opacity: 1;
  }
}

@media (max-width: 767px) {
  .toast-container {
    bottom: 0.75rem;
    right: 0.5rem;
    left: 0.5rem;
    gap: 0.4rem;
  }

  .toast {
    min-width: 0;
    width: fit-content;
    max-width: 100%;
    padding: 0.45rem 0.6rem;
    border-radius: 0.25rem;
    font-size: 0.82rem;
    box-shadow: 0 2px 6px rgba(0, 0, 0, 0.1);
  }

  .toast-content {
    gap: 0.45rem;
  }

  .toast-close {
    font-size: 0.95rem;
    min-width: 1.1rem;
    opacity: 0.8;
  }
}
</style>
