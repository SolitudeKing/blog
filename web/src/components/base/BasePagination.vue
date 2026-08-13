<template>
  <nav class="base-pagination" :class="`base-pagination--${mode}`" aria-label="分页">
    <!-- 加载更多模式：单一"加载更多"按钮 -->
    <template v-if="mode === 'loadMore'">
      <button
        v-if="hasMore"
        type="button"
        class="base-pagination__load-more"
        :disabled="disabled || loading"
        :aria-busy="loading ? 'true' : undefined"
        @click="emit('loadMore')"
      >
        <span v-if="loading" class="base-pagination__spinner" aria-hidden="true" />
        <span class="base-pagination__label">{{ loading ? '加载中…' : label }}</span>
      </button>
      <p v-else-if="totalLabel" class="base-pagination__total" role="status" aria-live="polite">
        {{ totalLabel }}
      </p>
    </template>

    <!-- 页码模式：完整页码器（依赖 total 字段，本期 7 列表页未启用） -->
    <template v-else-if="mode === 'pages'">
      <button
        type="button"
        class="base-pagination__nav"
        :disabled="disabled || currentPage <= 1"
        aria-label="上一页"
        @click="goTo(currentPage - 1)"
      >
        ‹
      </button>
      <ul class="base-pagination__pages">
        <li v-for="item in pageItems" :key="item.key">
          <span
            v-if="item.kind === 'ellipsis'"
            class="base-pagination__ellipsis"
            aria-hidden="true"
          >…</span>
          <button
            v-else
            type="button"
            class="base-pagination__page"
            :class="{ 'is-current': item.value === currentPage }"
            :aria-current="item.value === currentPage ? 'page' : undefined"
            :aria-label="`第 ${item.value} 页`"
            @click="goTo(item.value ?? 0)"
          >
            {{ item.value }}
          </button>
        </li>
      </ul>
      <button
        type="button"
        class="base-pagination__nav"
        :disabled="disabled || currentPage >= totalPages"
        aria-label="下一页"
        @click="goTo(currentPage + 1)"
      >
        ›
      </button>
      <p v-if="totalLabel" class="base-pagination__total">{{ totalLabel }}</p>
    </template>
  </nav>
</template>

<script setup lang="ts">
import { computed } from 'vue'

type PageItem =
  | { kind: 'page'; key: number; value: number }
  | { kind: 'ellipsis'; key: string }

const props = withDefaults(
  defineProps<{
    mode?: 'loadMore' | 'pages'
    /** loadMore 模式 */
    loading?: boolean
    hasMore?: boolean
    label?: string
    /** pages 模式 */
    modelValue?: number
    total?: number
    pageSize?: number
    maxVisiblePages?: number
    /** 通用 */
    disabled?: boolean
  }>(),
  {
    mode: 'loadMore',
    loading: false,
    hasMore: false,
    label: '加载更多',
    modelValue: 1,
    total: 0,
    pageSize: 20,
    maxVisiblePages: 7,
    disabled: false,
  },
)

const emit = defineEmits<{
  loadMore: []
  'update:modelValue': [value: number]
  change: [value: number]
}>()

const currentPage = computed(() => Math.max(1, props.modelValue ?? 1))
const totalPages = computed(() =>
  Math.max(1, Math.ceil((props.total ?? 0) / Math.max(1, props.pageSize ?? 20))),
)

const pageItems = computed<PageItem[]>(() => {
  if (props.mode !== 'pages') return []
  const total = totalPages.value
  const max = Math.max(1, props.maxVisiblePages ?? 7)
  const current = currentPage.value

  if (total <= max) {
    return Array.from({ length: total }, (_, i) => ({
      kind: 'page',
      key: i + 1,
      value: i + 1,
    }))
  }

  const half = Math.floor(max / 2)
  let start = Math.max(1, current - half)
  let end = Math.min(total, start + max - 1)
  if (end - start < max - 1) {
    start = Math.max(1, end - max + 1)
  }

  const items: PageItem[] = []
  if (start > 1) {
    items.push({ kind: 'page', key: 1, value: 1 })
    if (start > 2) {
      items.push({ kind: 'ellipsis', key: 'start-ellipsis' })
    }
  }
  for (let i = start; i <= end; i++) {
    items.push({ kind: 'page', key: i, value: i })
  }
  if (end < total) {
    if (end < total - 1) {
      items.push({ kind: 'ellipsis', key: 'end-ellipsis' })
    }
    items.push({ kind: 'page', key: total, value: total })
  }
  return items
})

const totalLabel = computed(() => {
  if (props.mode === 'pages' && (props.total ?? 0) > 0) {
    return `共 ${props.total} 条`
  }
  return ''
})

function goTo(page: number) {
  if (props.disabled) return
  const target = Math.min(Math.max(1, page), totalPages.value)
  if (target === currentPage.value) return
  emit('update:modelValue', target)
  emit('change', target)
}
</script>

<style scoped lang="scss">
.base-pagination {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: center;
  gap: var(--space-3);
  min-width: 0;
}

.base-pagination--loadMore {
  padding: var(--space-3) 0;
}

.base-pagination__load-more {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  min-height: 44px;
  padding: 0 var(--space-5);
  border: 1px solid var(--border-strong);
  border-radius: var(--radius-md);
  background: var(--bg-elevated);
  color: var(--text-primary);
  font-weight: 600;
  cursor: pointer;
  transition:
    border-color 120ms ease,
    background-color 120ms ease,
    transform 120ms ease;
}

.base-pagination__load-more:hover:not(:disabled) {
  border-color: var(--accent);
  background: var(--bg-card);
  transform: translateY(-1px);
}

.base-pagination__load-more:focus-visible {
  outline: none;
  border-color: var(--border-focus);
  box-shadow: var(--focus-ring);
}

.base-pagination__load-more:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.base-pagination__spinner {
  width: 14px;
  height: 14px;
  border-radius: var(--radius-full);
  border: 2px solid var(--border);
  border-top-color: var(--accent);
  animation: base-pagination-spin 0.7s linear infinite;
}

@keyframes base-pagination-spin {
  to {
    transform: rotate(360deg);
  }
}

.base-pagination__label {
  min-width: 4em;
  text-align: center;
}

.base-pagination__total {
  margin: 0;
  color: var(--text-muted);
  font-size: var(--text-sm);
}

.base-pagination--pages {
  gap: var(--space-2);
}

.base-pagination__pages {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-1);
  margin: 0;
  padding: 0;
  list-style: none;
}

.base-pagination__page,
.base-pagination__nav {
  display: inline-grid;
  min-width: 36px;
  height: 36px;
  place-items: center;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  background: var(--bg-card);
  color: var(--text-primary);
  cursor: pointer;
  transition: border-color 120ms ease, background-color 120ms ease, color 120ms ease;
}

.base-pagination__page:hover:not(.is-current),
.base-pagination__nav:hover:not(:disabled) {
  border-color: var(--border-strong);
  background: var(--bg-elevated);
}

.base-pagination__page:focus-visible,
.base-pagination__nav:focus-visible {
  outline: none;
  border-color: var(--border-focus);
  box-shadow: var(--focus-ring);
}

.base-pagination__page.is-current {
  border-color: var(--accent);
  background: var(--accent-soft);
  color: var(--accent);
  font-weight: 700;
  cursor: default;
}

.base-pagination__nav:disabled,
.base-pagination__page:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.base-pagination__ellipsis {
  display: inline-grid;
  min-width: 36px;
  height: 36px;
  place-items: center;
  color: var(--text-muted);
  user-select: none;
}

@media (prefers-reduced-motion: reduce) {
  .base-pagination__spinner {
    animation: none;
  }
}
</style>
