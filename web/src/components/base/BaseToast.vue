<template>
  <Teleport to="body">
    <div
      class="mist-toast-container"
      role="region"
      aria-label="通知"
      aria-live="polite"
      aria-relevant="additions"
    >
      <TransitionGroup name="mist-toast">
        <div
          v-for="item in items"
          :key="item.id"
          class="mist-toast"
          :class="`mist-toast--${item.variant}`"
          :role="item.variant === 'error' ? 'alert' : 'status'"
          :aria-live="item.variant === 'error' ? 'assertive' : 'polite'"
          aria-atomic="true"
          @pointerenter="pause(item.id)"
          @pointerleave="resume(item.id)"
          @pointercancel="resume(item.id)"
          @focusin="pause(item.id)"
          @focusout="onFocusOut($event, item.id)"
        >
          <span class="mist-toast__dot" aria-hidden="true" />
          <p class="mist-toast__message">{{ item.message }}</p>
          <button
            class="mist-toast__close"
            type="button"
            aria-label="关闭通知"
            @click="dismiss(item.id)"
          >
            <svg viewBox="0 0 16 16" aria-hidden="true">
              <path d="m4 4 8 8M12 4l-8 8" fill="none" stroke="currentColor" stroke-linecap="round" stroke-width="1.8" />
            </svg>
          </button>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { useToast } from '@/composables/useToast'

const { items, dismiss, pause, resume } = useToast()

function onFocusOut(event: FocusEvent, id: number) {
  const toast = event.currentTarget as HTMLElement
  if (event.relatedTarget instanceof Node && toast.contains(event.relatedTarget)) {
    return
  }
  resume(id)
}
</script>
