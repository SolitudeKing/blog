import { reactive } from 'vue'

export type ToastVariant = 'info' | 'success' | 'warning' | 'error'

export interface ToastItem {
  id: number
  variant: ToastVariant
  message: string
  duration: number
}

export interface ToastOptions {
  duration?: number
}

const state = reactive({
  items: [] as ToastItem[],
})

let nextId = 1
const timers = new Map<number, { timeoutId: number | null; remaining: number; startedAt: number }>()

function push(variant: ToastVariant, message: string, options: ToastOptions = {}) {
  if (!message) {
    return
  }
  const id = nextId++
  const item: ToastItem = {
    id,
    variant,
    message,
    duration: options.duration ?? 3200,
  }
  state.items.push(item)
  timers.set(id, { timeoutId: null, remaining: item.duration, startedAt: 0 })
  schedule(id)
  return id
}

function schedule(id: number) {
  const timer = timers.get(id)
  if (!timer || timer.remaining <= 0 || typeof window === 'undefined') {
    return
  }
  timer.startedAt = Date.now()
  timer.timeoutId = window.setTimeout(() => dismiss(id), timer.remaining)
}

function pause(id: number) {
  const timer = timers.get(id)
  if (!timer || timer.timeoutId === null) {
    return
  }
  window.clearTimeout(timer.timeoutId)
  timer.timeoutId = null
  timer.remaining = Math.max(0, timer.remaining - (Date.now() - timer.startedAt))
}

function resume(id: number) {
  const timer = timers.get(id)
  if (!timer || timer.timeoutId !== null || timer.remaining <= 0) {
    return
  }
  schedule(id)
}

function dismiss(id: number) {
  const timer = timers.get(id)
  if (timer?.timeoutId !== null && timer?.timeoutId !== undefined && typeof window !== 'undefined') {
    window.clearTimeout(timer.timeoutId)
  }
  timers.delete(id)
  const idx = state.items.findIndex((t) => t.id === id)
  if (idx >= 0) {
    state.items.splice(idx, 1)
  }
}

export function useToast() {
  return {
    items: state.items,
    info: (msg: string, opts?: ToastOptions) => push('info', msg, opts),
    success: (msg: string, opts?: ToastOptions) => push('success', msg, opts),
    warning: (msg: string, opts?: ToastOptions) => push('warning', msg, opts),
    /** @deprecated Use warning(). */
    warn: (msg: string, opts?: ToastOptions) => push('warning', msg, opts),
    error: (msg: string, opts?: ToastOptions) => push('error', msg, opts),
    pause,
    resume,
    dismiss,
  }
}
