# UI 设计系统与实现规范

本文定义 Solitude Blog 公开站点与后台管理台的统一 UI 规范。它是设计评审、组件实现和验收的共同依据，不是仅供展示的视觉稿。

本文以 Mist UI 的语义 token、雾境氛围、组件状态和可访问性规范为基础，并按本项目需求扩展为“站点主题色 × 访客明暗模式”的二维架构。具体色值、状态权威和迁移规则以 [雾境二维主题色调系统](./05-theme-color-system.md) 为唯一事实来源。

## 适用范围与实现源

适用范围：

- 公开站点：主页、文章、归档、搜索等阅读型页面。
- 后台管理台：仪表盘、列表、表单、编辑器、媒体库等高密度页面。
- 项目内 `Base*` 基础组件及公开站点、后台管理台的业务组件。

当前实现源按以下顺序理解：

1. `web/src/styles/tokens/_base.scss`：与主题无关的字体、间距、圆角、阴影几何、动效和层级。
2. `web/src/styles/tokens/_mist-sea-salt.scss`：雾境海盐的 light/dark 语义映射。
3. `web/src/styles/tokens/_mist-forest.scss`：雾境青森的 light/dark 语义映射。
4. `web/src/styles/themes/_mist.scss`：八层光场、雾幕、玻璃层级与内容容器。
5. `web/src/styles/components/*.scss`：组件对语义 token 的消费方式。
6. `web/src/composables/useTheme.ts`：站点主题同步、访客明暗状态、持久化、旧偏好迁移与 DOM 属性同步。

旧 `_mist-violet.scss`、`_forest.scss`、`_strawberry.scss` 和 `themes/creamy.scss` 已从源码运行时删除，不得恢复引用。旧 `forest` 与 Mist UI 的 `mist-forest` 不是同一主题文件，只允许通过明确的数据迁移规则映射，禁止复用旧样式或静默别名命中。

本文与色调系统文档共同构成当前目标契约。颜色来源、具体值、透明度和颜色派生以色调系统为准；token 名称、组件 API、字体与尺寸、阴影几何、主题控制器和可访问性以本文为准；实时实现状态以 [当前进度](../../03-knowledge/wiki/01-design-progress.md) 和本文偏差清单为准。源码与目标不一致时记录偏差，不得在业务页面临时覆盖。

## 设计决策

### D1：采用二维主题，不把明暗模式混入主题名称

根节点统一使用：

```html
<html data-theme="mist-sea-salt" data-mode="light">
```

- `data-theme` 仅允许 `mist-sea-salt`、`mist-forest`，由后台站点设置控制。
- `data-mode` 仅允许 `light`、`dark`。
- 默认值为 `mist-sea-salt + light`。
- 公开前台只提供明暗控件，不提供主题色选择器；后台主题色变化不得重置访客 mode。
- 后台 `mode` 只定义首次访问默认值；一旦访客保存 `blog:mode`，服务端不得覆盖该偏好。
- Mist UI 通用规范默认一次装载一个主题包；本项目按产品需求使用带 `data-theme + data-mode` 作用域的项目级适配，不直接并列导入两个裸 `[data-mode]` 主题文件。

必须验证以下四种组合：

| 组合 | 使用重点 |
| --- | --- |
| `mist-sea-salt + light` | 默认阅读与清爽后台环境 |
| `mist-sea-salt + dark` | 低照度海蓝阅读环境 |
| `mist-forest + light` | 自然、平静的长时间阅读环境 |
| `mist-forest + dark` | 低照度青绿工作环境 |

### D2：组件只消费语义 token

- 业务组件和页面不得直接使用海盐或青森私有原色名，也不得按主题名称编写视觉分支。
- 组件 SCSS 中不得新增无解释的十六进制颜色、RGB 颜色或私有阴影。
- 主题差异只能在两个 `tokens/_mist-*.scss` 文件中映射；来源色到语义 token 的关系必须回写色调系统文档。
- `color-mix()` 只能基于语义 token 派生交互状态，不得绕过语义层引入新色板。
- SVG 图标默认使用 `currentColor`，由组件文字或状态 token 控制颜色。

### D3：阅读界面与管理界面共享语言，不共享密度

- 公开站点和长文阅读页面可以使用较大留白、柔和背景光斑、圆润卡片和轻微 hover 上浮。
- 后台管理台使用相同 token、焦点态和反馈方式，但减少装饰背景、缩短间距并保持表格可扫描性。
- 可折叠 Shell、玻璃内容壳、三栏节奏和自适应卡片网格均作为本项目统一布局语言，由语义 token 和响应式规则共同约束。
- 正文最大可读宽度建议为 `72ch`；包含代码、表格的长文主栏可按内容放宽，但不能挤压目录。

### D4：圆润和柔和层级不得损害可用性

- 圆角用于建立亲和感，不使用所有元素都为胶囊形的做法。
- 阴影只表达层级；边框仍需承担结构分隔，暗色模式不得只靠阴影。
- hover 允许轻微上浮，active 必须回落，禁用态不得响应 hover。
- 正文、表格和表单的可读性优先于装饰效果。

### D5：可访问性和状态完整性是组件 API 的组成部分

- 交互组件至少覆盖适用的 default、hover、focus、active、disabled、loading、empty、error 状态。
- 键盘焦点必须可见；不得用 `outline: none` 后不提供等价焦点环。
- 表单必须关联 label、hint 和 error；错误不得只依赖颜色表达。
- 图标使用 SVG/Icon 组件；emoji 不得作为生产界面的真实图标。
- 模态框、抽屉和菜单必须处理 ESC、焦点圈定、焦点归还和滚动锁定。
- Toast 使用 `aria-live` 或合适的 live region；错误反馈需要更高优先级的播报语义。

## Token 分层

### 第一层：基础 token

基础 token 与主题无关，只在 `_base.scss` 定义：

- 字体：`--font-*`
- 字号与行高：`--text-*`、`--leading-*`
- 间距：`--space-*`
- 圆角：`--radius-*`
- 阴影几何结构：`--shadow-*`；主题实现必须确保其中的颜色由当前主题映射，不得保留基础层的内置色
- 动效：`--transition-*`
- 层级：`--z-*`
- 布局：`--layout-*`

为兼容现有编辑式布局，基础层保留 `--layout-workbench-max`、`--layout-rail-width`、`--layout-toolbar-height`、`--layout-toc-width` 四个几何别名。它们不得承载主题差异；新布局优先使用 `--layout-admin-max`、`--layout-shell-*`、`--layout-*-header` 等项目主 token。

### 第二层：主题语义 token

每个主题模式必须让下列 token 全部解析为有效值；可在主题文件中显式映射，也可在基础层基于核心主题 token 派生，但组件不得自行补值：

- 背景：`--bg-primary`、`--bg-secondary`、`--bg-card`、`--bg-elevated`、`--bg-inset`、`--bg-hover`、`--bg-active`、`--bg-disabled`
- 文字：`--text-primary`、`--text-secondary`、`--text-muted`、`--text-disabled`、`--text-on-accent`、`--text-link`、`--text-inverse`
- 强调：`--accent`、`--accent-hover`、`--accent-active`、`--accent-soft`、`--accent-softer`、`--accent-gradient`、`--accent-gradient-hover`
- 状态：`--success`、`--warning`、`--danger`、`--info` 及对应的 `*-soft`
- 边框：`--border`、`--border-strong`、`--border-color`、`--border-color-strong`、`--border-focus`、`--focus-ring-color`、`--focus-ring`、`--divider-fade`、`--line-soft`、`--line-strong`
- 阴影：`--shadow-color` 与完整 `--shadow-*` 系列
- 展示表面：`--surface-glass-subtle`、`--surface-glass`、`--surface-glass-strong`、`--surface-glass-border`、`--surface-glass-highlight`、`--surface-glass-shadow`、`--surface-highlight`、`--nav-bg`、`--modal-overlay-bg`、`--scrim`
- 页面氛围：`--light-gradient-primary`、`--light-gradient-secondary`、`--fog-color-1..3`、`--fog-opacity`、`--fog-blend-mode`、`--body-bg-spot-1..8`、`--blob-color-1..2`
- 内容媒介：`--code-bg`、`--code-text`、`--figure-bg`
- 表单控件：`--select-arrow-color`、`--checkbox-check-color`、`--radio-dot-color`

### 第三层：组件别名

只有同一语义在多个组件中反复出现时才允许增加组件别名，例如：

```scss
:root {
  --control-height-sm: 34px;
  --control-height-md: 44px;
  --control-height-lg: 52px;
  --control-bg: var(--bg-elevated);
  --control-border: var(--border-strong);
  --control-border-hover: var(--border-strong);
  --control-border-focus: var(--border-focus);
}
```

`--border` 只用于非关键分隔和装饰边界。输入、选择、切换器等需要依靠轮廓识别的控件默认使用 `--border-strong`；聚焦时必须同时使用实色 `--border-focus` 与 `--focus-ring`，半透明 focus ring 只增强范围，不能单独承担 3:1 的关键边界对比。

组件别名不得包含页面或业务名称，例如禁止 `--article-card-green`、`--dashboard-pink-shadow`。

## 支持矩阵与选择器作用域

两个主题必须完整实现同一套语义 token，任何一组缺值都视为构建缺陷。项目样式必须以组合选择器隔离主题：

```scss
[data-theme='mist-sea-salt'][data-mode='light'] { /* 海盐浅色 */ }
[data-theme='mist-sea-salt'][data-mode='dark'] { /* 海盐深色 */ }
[data-theme='mist-forest'][data-mode='light'] { /* 青森浅色 */ }
[data-theme='mist-forest'][data-mode='dark'] { /* 青森深色 */ }
```

禁止只写 `[data-mode='dark']` 的主题色覆盖，否则两个主题同时打包时会由导入顺序决定结果。组件、布局和业务页面不得出现 `[data-theme]` 分支；主题选择器只属于 token 文件。冻结色值和对比度基线见 [雾境二维主题色调系统](./05-theme-color-system.md)。

## 字体与排版

### 字体族

沿用当前实现：

```scss
--font-sans: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont,
  "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif;
--font-serif: "Iowan Old Style", "Palatino Linotype", "Noto Serif SC",
  "Songti SC", Georgia, serif;
--font-mono: ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas,
  "Liberation Mono", monospace;
```

- 不以网络字体作为首屏可用的前提；Inter 未加载时必须自然回退。
- 中文正文优先系统中文字体，代码统一使用 `--font-mono`；`--font-serif` 只用于引文、题记等明确的编辑式强调，不替代正文默认字体。
- 正文不使用超细字重；常规正文 `400`，辅助信息 `500`，控件和小标题 `600/700`。

按 Mist UI 基础层补充到 `_base.scss` 的字号与行高：

| Token | 值 | 用途 |
| --- | --- | --- |
| `--text-xs` | `12px` | 时间、标签辅助信息 |
| `--text-sm` | `14px` | 表单说明、表格、次级导航 |
| `--text-base` | `16px` | 默认界面正文 |
| `--text-md` | `18px` | 阅读正文、引导文字 |
| `--text-lg` | `20px` | 卡片标题、三级标题 |
| `--text-xl` | `24px` | 二级标题 |
| `--text-2xl` | `32px` | 页面区块标题 |
| `--text-3xl` | `40px` | 桌面页面主标题 |
| `--text-hero` | `56px` | 仅首页 Hero；移动端降为 `35.2px` |
| `--leading-tight` | `1.25` | 标题 |
| `--leading-normal` | `1.6` | 界面正文 |
| `--leading-relaxed` | `1.75` | 文章与长文正文 |

## 间距、圆角与阴影

### 间距

采用 4px 基线，仅使用现有阶梯：

| Token | 值 | 常见用途 |
| --- | --- | --- |
| `--space-1` | `4px` | 紧凑内联间隔 |
| `--space-2` | `8px` | 图标与文字、紧凑控件 |
| `--space-3` | `12px` | 控件内边距、小组件间距 |
| `--space-4` | `16px` | 默认组件间距 |
| `--space-5` | `20px` | 面板紧凑内边距 |
| `--space-6` | `24px` | 卡片与区块内边距 |
| `--space-8` | `32px` | 页面区块间距 |
| `--space-10` | `40px` | 大区块间距 |
| `--space-12` | `48px` | 页面章节分隔 |

禁止在组件内随意出现 `13px`、`17px`、`26px` 等间距。确有排版需要时，先判断是否应补充全局阶梯。

### 圆角

| Token | 值 | 用途 |
| --- | --- | --- |
| `--radius-sm` | `8px` | 内联代码、小标签、小型控件 |
| `--radius-md` | `12px` | 按钮、输入框、菜单项 |
| `--radius-lg` | `20px` | 卡片、Toast、目录面板 |
| `--radius-xl` | `24px` | Hero、抽屉、大型浮层 |
| `--radius-pill` | `999px` | 标签、切换器、状态胶囊 |
| `--radius-full` | `9999px` | 圆形头像、圆点 |

同一组件嵌套时，内层圆角不得大于外层圆角。后台表格使用 `md/lg`，避免每个单元格胶囊化。

### 阴影

沿用基于 `--shadow-color` 的主题化阴影：

```scss
--shadow-xs: 0 1px 2px var(--shadow-color);
--shadow-sm: 0 8px 20px var(--shadow-color);
--shadow-md: 0 16px 36px var(--shadow-color);
--shadow-lg: 0 24px 48px var(--shadow-color);
--shadow-xl: 0 20px 32px -12px var(--shadow-color);
--shadow-inset: inset 0 1px 0 rgb(255 255 255 / 35%);
--shadow-glow: 0 0 0 3px color-mix(in srgb, var(--accent) 22%, transparent);
```

- `xs`：静态卡片或控件。
- `sm`：hover、下拉菜单。
- `md`：Toast、抽屉、小型浮层。
- `lg`：Modal 等高层级浮层。
- `xl`：仅用于需要与 Modal 拉开层级的全局浮层。
- 同一表面最多使用一个外阴影和一个 inset 高光。
- 暗色模式的结构边界必须同时有 border，不能只靠黑色阴影。

玻璃模糊属于与主题无关的表面几何，统一使用以下 token：

```scss
--glass-blur-sm: 14px;
--glass-blur-md: 24px;
--glass-blur-nav: 30px;
--glass-blur-lg: 32px;
--glass-saturation: 135%;
--fog-blur-sm: 40px;
--fog-blur-md: 60px;
--fog-blur-lg: 80px;
```

`sm` 用于 TOC 等小型粘性表面，`md` 用于后台侧栏与登录表单，`nav` 用于导航，`lg` 只用于首页等单一主玻璃壳；正文、表格单元格和玻璃壳内部的小卡不得重复使用 backdrop blur。旧 `--glass-saturate` 不再新增消费者。

## 动效

沿用当前时间与曲线：

```scss
--transition-fast: 150ms ease;
--transition-base: 240ms cubic-bezier(.2, .7, .3, 1);
--transition-slow: 360ms cubic-bezier(.2, .7, .3, 1);
--transition-spring: 420ms cubic-bezier(.22, 1, .36, 1);
```

- `fast`：颜色、边框、焦点、按钮按压。
- `base`：卡片上浮、菜单和 Toast 进入退出。
- `slow`：抽屉、Modal 等大范围位移；常规组件不得使用更长动效。
- `spring`：文章记录、归档条目等短距离方向反馈，不得用于布局尺寸。
- hover 上浮不超过 `2px`。
- active 回到 `translateY(0)`，形成按下反馈。
- 不对布局尺寸使用无目的的 `transition: all`。

必须提供降级：

```scss
@media (prefers-reduced-motion: reduce) {
  html {
    scroll-behavior: auto;
  }

  *,
  *::before,
  *::after {
    scroll-behavior: auto !important;
    animation-duration: 1ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 1ms !important;
  }
}
```

## 层级

沿用当前层级起点：

| Token | 值 | 用途 |
| --- | --- | --- |
| `--z-sticky` | `10` | 顶部导航、页内目录 |
| `--z-dropdown` | `50` | 菜单、选择器浮层 |
| `--z-modal` | `100` | Modal、Drawer、遮罩 |
| `--z-toast` | `200` | Toast 和全局通知 |

业务页面不得写 `9999`。新浮层必须进入这套层级关系，并验证嵌套浮层的焦点与遮罩行为。

## 布局规范

### 文章与阅读页

桌面端建议采用正文与页内目录结构：

```text
全宽顶部导航
└─ 页面容器
   ├─ 主内容：minmax(0, 760px)
   ├─ 栏间距：32px
   └─ 右侧页内目录：220px
```

- 页面最大宽度沿用 `1120px`，左右安全间距至少 `16px`。
- 右侧目录可 sticky，顶部偏移必须包含导航高度。
- 主栏必须设置 `min-width: 0`，代码块和表格使用自身横向滚动，不能撑破页面。
- 文章正文建议限制到 `72ch`；宽表格和架构图可以占满主栏。
- 背景光斑与雾幕由 `mist-page` 统一提供，正文背后不得使用高饱和装饰。

响应式规则：

- `> 860px`：正文与右侧目录同时展示。
- `<= 860px`：正文单栏，目录改为抽屉或折叠区；触发按钮点击区域至少 `44px × 44px`。
- 移动端抽屉关闭后必须把焦点归还给触发按钮。

### 后台页面

- 页面内容宽度使用布局层的 `--layout-admin-max`（当前设计为 `1280px`）与流式 gutter，具体结构以 [布局模式](./04-layout-patterns.md) 为准。
- 表单主次栏建议 `minmax(0, 1fr) 280px`，在 `860px` 以下改为单栏。
- 表格优先保持列对齐；窄屏可转为信息卡片，但字段标签不能丢失。
- 筛选、保存、危险操作必须有稳定位置；不得因主题或加载状态产生布局跳动。

## 组件命名与 Vue API

### 通用约定

- Vue 文件使用 `BaseButton.vue`、`BaseInput.vue` 等 `Base*` 命名。
- 基础组件 CSS 使用 Mist UI 的 `mist-*` BEM，例如 `mist-button mist-button--primary is-loading`；业务组件使用 `blog-navbar__actions` 等业务语义 BEM。
- 可替换内容使用 slot；表单值使用 `modelValue` / `update:modelValue`。
- 透传原生属性和事件，不重新发明 `onClick`、`isDisabled` 等非 Vue 约定 API。
- 每个表单组件应支持 `id`、`name`、`disabled`、`required`、`label`、`hint`、`error`。
- `error` 存在时设置 `aria-invalid="true"`，并通过 `aria-describedby` 连接错误和说明文本。
- 复杂行为放入 composable，例如 `useFocusTrap`、`useClickOutside`、`useToast`。

### 状态矩阵

| 状态 | 视觉要求 | 行为与语义要求 |
| --- | --- | --- |
| default | 使用正常表面、文字和边框 token | 原生语义正确，可被辅助技术识别 |
| hover | 边框增强或最多上浮 2px | 仅指针设备触发，不作为唯一提示 |
| focus | 使用 `--focus-ring` 与 `--border-focus` | 使用 `:focus-visible`，键盘焦点清晰 |
| active | 回落并使用 `--accent-active` 等 active token | 不造成布局跳动 |
| disabled | 使用 disabled token 或降低对比，禁止 hover | 设置原生 `disabled` 或 `aria-disabled`，不可触发操作 |
| loading | 保持组件宽度，显示非 emoji 进度图标 | 设置 `aria-busy="true"`，避免重复提交 |
| empty | 使用 `BaseEmpty`，提供原因和可选下一步 | 图标装饰时 `aria-hidden="true"` |
| error | 使用 danger 与 danger-soft，附带文字说明 | 表单连接 error id；全局错误使用适当 live region |

### 已有组件 API 契约

#### BaseButton

保留现有 API：

```ts
interface BaseButtonProps {
  variant?: 'primary' | 'secondary' | 'ghost'
  size?: 'sm' | 'md' | 'lg'
  loading?: boolean
  disabled?: boolean
  type?: 'button' | 'submit' | 'reset'
}
```

- `loading` 时同时禁用按钮并设置 `aria-busy="true"`。
- 加载图标使用 SVG 或 CSS spinner，按钮可访问名称在加载中不得消失。
- 危险操作后续增加 `danger` variant 时必须使用语义 danger token，不写死红色。

#### BaseInput / BaseTextarea / BaseSelect

统一目标 API：

```ts
interface BaseFieldProps {
  modelValue: string | number
  id?: string
  name?: string
  label: string
  hint?: string
  error?: string
  placeholder?: string
  disabled?: boolean
  required?: boolean
}
```

- Select 的 `options` 保持 `{ label, value, disabled? }[]`。
- placeholder prop 必须真正映射到原生元素；没有实现时不得保留无效 prop。
- Label、hint、error 必须通过唯一 id 与控件关联。
- Select 箭头改用统一 SVG Icon，不使用字符充当图标系统。

#### 按需组件边界

当前基线不保留无调用的 `BaseToggle` 与 `BaseCard`。明暗模式由 `BaseThemeControls` 提供明确按钮语义；内容区优先使用页面结构与语义表面，避免为统一外观重新形成卡片墙。

- 未来出现三个以上语义一致的开关场景时，再按按钮型 `aria-pressed` 或表单型 checkbox 语义提炼 Toggle。
- 未来确有跨页面复用的 Card 时，交互入口必须使用真实链接或按钮；视觉悬浮不能替代可点击语义。

#### BaseEmpty / BaseSkeleton / BaseToast

- BaseEmpty 图标改为 Icon slot，不提供 emoji 默认值。
- BaseSkeleton 在 reduced-motion 下停止 shimmer；加载容器设置 `aria-busy`，骨架条自身保持 `aria-hidden`。
- Toast variant 统一为 `info/success/warning/error`；不再保留含义重复的 `warn` 接口。
- Toast 容器使用 live region；error 使用 `role="alert"`，普通通知使用 `role="status"`。
- 自动关闭的 Toast 在 hover 或键盘聚焦时应暂停，关闭按钮必须有可见焦点态。

## 主题与明暗模式规范

### 状态与持久化

主题与模式必须拆分权威：

```ts
type ThemeName = 'mist-sea-salt' | 'mist-forest'
type ModeName = 'light' | 'dark'

interface ThemeState {
  theme: ThemeName
  mode: ModeName
}

const APPEARANCE_CACHE_KEY = 'blog:site-appearance'
const MODE_STORAGE_KEY = 'blog:mode'
const LEGACY_STORAGE_KEY = 'blog:theme'
```

主题解析优先级：

```text
setting/lobby.theme > 服务端外观缓存中的合法 theme > 合法 DOM theme > mist-sea-salt
```

模式解析优先级：

```text
合法 blog:mode > setting/lobby.mode > 服务端外观缓存中的合法 mode
  > 合法 DOM mode > prefers-color-scheme > light
```

- `blog:site-appearance` 仅缓存最近一次有效服务端响应或后台保存结果，不向访客暴露写入口；页面挂载后，合法 `setting/lobby.theme` 必须覆盖缓存。
- 运行时公开 API 只提供 `setMode()`、`cycleMode()` 和 `resetMode()`；公开前台不提供 `setTheme()`、`cycleTheme()`。
- `syncFromServer()` 必须分别处理两个轴：主题始终应用，默认 mode 只在没有 `blog:mode` 时应用。
- `resetMode()` 删除本地偏好并立即回到最近的服务端默认 mode。
- 旧 `blog:theme` 复合对象只一次性迁移其中合法的 `mode`，忽略旧 theme，随后删除旧键。
- 监听 `storage` 事件，同源多标签页分别同步访客 mode 与服务端外观缓存；主题变化不得清空或改写 mode。

### 首屏防闪烁

主题必须在 Vue mount 和主样式首次绘制前确定。`index.html` 静态提供 `data-theme="mist-sea-salt" data-mode="light"`；最小同步脚本在首帧读取服务端外观缓存与访客 mode，设置两个合法属性和对应的 `theme-color`。不能只在 `App.vue` 的 `onMounted()` 中初始化，也不能先绘制旧主题后再切换。

推荐在 `index.html` 中放置最小同步脚本，或在应用创建前执行 `initTheme()`：

```ts
initTheme()

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.mount('#app')
```

前台不显示主题色轮换控件；明暗控件应提供当前状态和动作名称，例如“当前：浅色模式，切换为深色模式”，并使用可理解的 SVG 图标与文字，不能只显示颜色圆点。后台未保存的主题预览使用嵌套 `data-theme + data-mode` 容器，不得提前改写全局状态。

## 内容与图标规则

- 标题层级按内容结构顺序使用，不因字号需求跳级。
- 链接必须可从正文颜色和下划线辨认，不能只依赖 hover。
- 代码使用 `--font-mono`；代码块、表格在窄屏自身滚动。
- 表格表头、行 hover、斑马纹均通过语义表面 token 表达，重要数据不能只靠颜色区分。
- 服务端提供的标签色是受控数据例外：必须校验颜色格式、计算文字对比度，并在非法或低对比值时回退到 `--accent-soft` 与 `--text-primary`；它不能成为组件默认色。
- Mermaid、图表和截图在 light/dark 两种模式下均需保持文字可读；无法主题化的图片应提供边框或稳定底色。
- 图标组件默认 `aria-hidden="true"`；图标本身承担操作含义时，按钮必须有 `aria-label` 或可见文字。
- 装饰性插图不得影响键盘顺序或占据辅助技术名称。

## 实现偏差清单

基础表单、按钮、空状态、Toast、Skeleton、主题控制与全局 reduced-motion 契约已完成本轮补齐；无调用的 Card、Toggle 以及 `BaseEmpty.icon`、`useToast().warn` 兼容入口已经移除。下表仅记录仍需继续收敛的工程债务；新功能不得扩大这些差异。

| 优先级 | 当前实现 | 目标与验收 |
| --- | --- | --- |
| P1 | 前台导航、文章目录与后台侧栏分别实现抽屉的焦点圈定、Esc 关闭、滚动锁定和焦点恢复 | 提炼共享 Drawer/Dialog 原语，并以键盘与读屏自动化用例防止行为漂移 |
| P2 | 服务端标签色目前只用于装饰色点，并仅校验十六进制格式 | 若未来用于文字背景或状态表达，必须加入对比度计算与语义回退，不能直接扩展现有用法 |
| P2 | 字号、行高和少量局部尺寸仍按页面语境直接定义 | 仅将跨三个以上组件重复且语义一致的数值提升为 token，避免把一次性构图参数过度抽象 |
| P2 | 两个主题 × 两个模式尚无自动化视觉回归与组件级无障碍测试 | 为首页、文章、归档、搜索、关于、后台壳层和仪表盘建立截图及 axe 基线 |

## Definition of Done

涉及 UI、主题或组件的改动只有同时满足以下条件才算完成。

### 主题与 token

- [ ] 在 `mist-sea-salt | mist-forest × light | dark` 四种组合中检查关键页面。
- [ ] 新组件只消费语义 token，没有新增无说明的主题色、圆角或阴影。
- [ ] `color-mix()`、`backdrop-filter` 等增强效果有可读的静态颜色、边框和表面回退。
- [ ] light/dark 切换无首屏闪烁，刷新后保持用户选择。
- [ ] 后台主题色保存后更新服务端、缓存与前台；访客 mode 独立切换，主题变化后保持不变。
- [ ] 损坏或旧版的本地存储、数据库主题和 Redis 缓存可安全归一化。
- [ ] 暗色模式的边框、表面和正文层级仍清晰。

### 组件状态

- [ ] 覆盖适用的 default、hover、focus、active、disabled、loading、empty、error。
- [ ] loading 不改变控件宽度，不允许重复提交，并暴露 `aria-busy`。
- [ ] 表单 label、hint、error 与控件正确关联。
- [ ] 所有图标来自 SVG/Icon 组件，生产 UI 不依赖 emoji。
- [ ] Modal、Drawer、Menu 完成 ESC、焦点圈定、焦点归还与滚动锁定。

### 响应式与内容

- [ ] 至少验证 `320px`、`390px`、`640px`、`768px`、`960px`、`1200px`、`1440px`，并抽查结构断点前后 1px。
- [ ] 代码块、表格、长 URL 和 Mermaid 图不会撑破主栏。
- [ ] 左右侧栏折叠后仍能访问全部导航和目录内容。
- [ ] 固定导航、抽屉和 Toast 不遮挡主要操作。

### 可访问性

- [ ] 只使用键盘可以完成页面核心流程，焦点顺序符合视觉顺序。
- [ ] 所有可交互元素有清晰的 `:focus-visible` 状态。
- [ ] 正文和常规控件文本对比度至少 `4.5:1`，大文本至少 `3:1`。
- [ ] 状态不只依赖颜色，错误和成功信息有文字说明。
- [ ] `prefers-reduced-motion: reduce` 下无持续 shimmer、平滑滚动或大幅位移动画。
- [ ] 使用读屏软件抽查表单错误、主题控件、抽屉和 Toast。

### 工程验证

- [ ] 在 `web` 目录执行 `npm run typecheck`。
- [ ] 涉及样式、组件或构建配置时执行 `npm run build`。
- [ ] 浏览器控制台没有 Vue warning、无效 ARIA 或主题解析错误。
- [ ] 更新或新增组件示例，能展示四种主题组合与主要状态。
- [ ] 若修正了本文偏差清单，必须同步更新对应条目。

## 评审规则

UI 评审按以下顺序进行：

1. 先检查信息架构、语义 HTML 和核心任务是否完整。
2. 再检查组件 API、状态和键盘行为。
3. 再检查两个主题 × 两个模式、响应式和内容溢出。
4. 最后检查圆角、阴影、背景光斑和微动效等视觉细节。

任何只追求表面视觉效果、却破坏语义 token、可访问性、信息密度或品牌表达的改动，都不应合入。最终交付必须保持 Solitude Blog 自主定义的内容结构与产品控制面；Mist UI 提供冻结色值、雾境氛围、语义组件和验收基线。
