<template>
  <Teleport to="body">
    <Transition :name="transitionName" appear>
      <div
        v-if="modelValue"
        class="base-modal"
        role="dialog"
        aria-modal="true"
        :aria-labelledby="titleId"
        @keydown.esc.stop="onEscape"
        @mousedown.self="onBackdrop"
      >
        <div
          ref="panelRef"
          class="base-modal__panel"
          :class="`base-modal__panel--${size}`"
          tabindex="-1"
          @keydown.tab="onTab"
        >
          <header class="base-modal__header">
            <div>
              <p v-if="eyebrow" class="base-modal__eyebrow">{{ eyebrow }}</p>
              <div :id="titleId" class="base-modal__title" role="heading" aria-level="2">
                {{ title }}
              </div>
            </div>
            <button
              type="button"
              class="base-modal__close"
              :disabled="loading"
              aria-label="关闭"
              @click="close"
            >
              <SvgIcon name="close" :size="20" />
            </button>
          </header>

          <p v-if="error" class="base-modal__error" role="alert" aria-live="assertive">
            {{ error }}
          </p>

          <div class="base-modal__body">
            <slot />
          </div>

          <footer v-if="$slots.footer" class="base-modal__footer">
            <slot name="footer" />
          </footer>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, useId, watch } from 'vue'
import SvgIcon from '@/components/base/SvgIcon.vue'

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    title: string
    eyebrow?: string
    size?: 'sm' | 'md' | 'lg'
    closeOnBackdrop?: boolean
    closeOnEsc?: boolean
    loading?: boolean
    error?: string
  }>(),
  {
    eyebrow: '',
    size: 'md',
    closeOnBackdrop: true,
    closeOnEsc: true,
    loading: false,
    error: '',
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  close: []
}>()

const panelRef = ref<HTMLDivElement | null>(null)
const titleId = useId()
const transitionName = 'base-modal'

let lockedScrollY = 0
let previousBodyStyle: Partial<CSSStyleDeclaration> = {}
let triggerElement: HTMLElement | null = null

const isOpen = computed(() => props.modelValue)

function close() {
  if (props.loading) return
  emit('update:modelValue', false)
  emit('close')
}

function onEscape() {
  if (!props.closeOnEsc || props.loading) return
  close()
}

function onBackdrop(event: MouseEvent) {
  if (!props.closeOnBackdrop || props.loading) return
  // 仅当点击事件 target 本身是 backdrop 容器（不是 panel 内部冒泡）时关闭
  if (event.target === event.currentTarget) {
    close()
  }
}

function onTab(event: KeyboardEvent) {
  if (event.key !== 'Tab') return
  const focusables = getFocusables()
  if (focusables.length === 0) {
    event.preventDefault()
    panelRef.value?.focus()
    return
  }
  const first = focusables[0]
  const last = focusables[focusables.length - 1]
  const active = document.activeElement as HTMLElement | null
  if (event.shiftKey && (active === first || !panelRef.value?.contains(active))) {
    event.preventDefault()
    last.focus()
  } else if (!event.shiftKey && active === last) {
    event.preventDefault()
    first.focus()
  }
}

function getFocusables(): HTMLElement[] {
  if (!panelRef.value) return []
  const selector = [
    'a[href]',
    'button:not([disabled])',
    'input:not([disabled])',
    'select:not([disabled])',
    'textarea:not([disabled])',
    '[tabindex]:not([tabindex="-1"])',
  ].join(',')
  return Array.from(panelRef.value.querySelectorAll<HTMLElement>(selector)).filter(
    (node) => !node.hasAttribute('inert') && node.offsetParent !== null,
  )
}

function lockPage() {
  const { body } = document
  lockedScrollY = window.scrollY
  previousBodyStyle = {
    position: body.style.position,
    top: body.style.top,
    left: body.style.left,
    right: body.style.right,
    width: body.style.width,
    overflow: body.style.overflow,
  }
  body.style.position = 'fixed'
  body.style.top = `-${lockedScrollY}px`
  body.style.left = '0'
  body.style.right = '0'
  body.style.width = '100%'
  body.style.overflow = 'hidden'
  const app = document.getElementById('app')
  if (app) {
    app.inert = true
  }
}

function unlockPage() {
  const { body } = document
  body.style.position = previousBodyStyle.position ?? ''
  body.style.top = previousBodyStyle.top ?? ''
  body.style.left = previousBodyStyle.left ?? ''
  body.style.right = previousBodyStyle.right ?? ''
  body.style.width = previousBodyStyle.width ?? ''
  body.style.overflow = previousBodyStyle.overflow ?? ''
  const app = document.getElementById('app')
  if (app) {
    app.inert = false
  }
  window.scrollTo(0, lockedScrollY)
}

function focusFirst() {
  nextTick(() => {
    const focusables = getFocusables()
    if (focusables.length > 0) {
      focusables[0].focus()
    } else {
      panelRef.value?.focus()
    }
  })
}

watch(isOpen, async (open) => {
  if (open) {
    triggerElement = document.activeElement as HTMLElement | null
    lockPage()
    focusFirst()
  } else {
    unlockPage()
    nextTick(() => {
      triggerElement?.focus()
      triggerElement = null
    })
  }
})

onBeforeUnmount(() => {
  if (isOpen.value) {
    unlockPage()
  }
})
</script>

<style scoped lang="scss">
.base-modal {
  position: fixed;
  inset: 0;
  z-index: var(--z-modal);
  display: grid;
  place-items: center;
  overflow-y: auto;
  padding: var(--space-5);
  background: var(--scrim);
}

.base-modal__panel {
  display: grid;
  width: 100%;
  max-height: min(760px, calc(100dvh - var(--space-10)));
  gap: var(--space-5);
  overflow-y: auto;
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  padding: clamp(var(--space-4), 3vw, var(--space-6));
  color: var(--text-primary);
  background: var(--bg-elevated);
  box-shadow: var(--shadow-lg);
}

.base-modal__panel:focus {
  outline: none;
}

.base-modal__panel--sm {
  max-width: 480px;
}

.base-modal__panel--md {
  max-width: 640px;
}

.base-modal__panel--lg {
  max-width: 880px;
}

.base-modal__header {
  display: flex;
  min-width: 0;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
  border-bottom: 1px solid var(--border);
  padding-bottom: var(--space-4);
}

.base-modal__header > div {
  min-width: 0;
}

.base-modal__eyebrow {
  margin: 0 0 var(--space-1);
  color: var(--text-muted);
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.base-modal__title {
  margin: 0;
  font-size: clamp(20px, 3vw, 26px);
  line-height: 1.25;
}

.base-modal__close {
  display: inline-grid;
  width: 44px;
  height: 44px;
  flex: none;
  place-items: center;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  color: var(--text-primary);
  background: var(--bg-card);
  cursor: pointer;
  transition: border-color 120ms ease, background-color 120ms ease;
}

.base-modal__close:hover:not(:disabled) {
  border-color: var(--border-strong);
  background: var(--bg-elevated);
}

.base-modal__close:focus-visible {
  border-color: var(--border-focus);
  outline: none;
  box-shadow: var(--focus-ring);
}

.base-modal__close:disabled {
  cursor: not-allowed;
  opacity: 0.56;
  pointer-events: none;
}

.base-modal__error {
  margin: 0;
  border: 1px solid var(--danger);
  border-left-width: 4px;
  border-radius: var(--radius-md);
  padding: var(--space-3);
  color: var(--danger);
  background: var(--bg-card);
  font-size: var(--text-sm);
}

.base-modal__body {
  display: grid;
  min-width: 0;
  gap: var(--space-4);
}

.base-modal__footer {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: flex-end;
  gap: var(--space-3);
  border-top: 1px solid var(--border);
  padding-top: var(--space-4);
}

.base-modal-enter-active,
.base-modal-leave-active {
  transition: opacity var(--transition-base);
}

.base-modal-enter-active .base-modal__panel,
.base-modal-leave-active .base-modal__panel {
  transition: transform var(--transition-base);
}

.base-modal-enter-from,
.base-modal-leave-to {
  opacity: 0;
}

.base-modal-enter-from .base-modal__panel,
.base-modal-leave-to .base-modal__panel {
  transform: translateY(var(--space-4));
}
</style>
