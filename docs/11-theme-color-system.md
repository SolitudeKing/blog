# 雾境二维主题色调系统

本文定义 Solitude Blog 的主题色调、明暗模式、状态权威、首屏初始化与后台配置契约。主题色值采用 Mist UI 的“雾境海盐”和“雾境青森”语义映射；组件只消费共享 token，不感知具体主题名称。

## 1. 主题模型

| 项目 | 约定 |
| --- | --- |
| 主题轴 | `mist-sea-salt`（雾境海盐）、`mist-forest`（雾境青森） |
| 模式轴 | `light`、`dark` |
| 默认组合 | `mist-sea-salt + light` |
| DOM 契约 | `data-theme="<theme>" data-mode="<mode>"` |
| 主题权威 | 后台站点设置，经 `setting/lobby` 向前台发布 |
| 模式权威 | 访客本地偏好优先；未设置时使用后台默认模式 |
| 主题元素 | 按主题 ID 持久化的少量编辑性纯文本，经 `setting/lobby` 发布 |
| 色值来源 | Mist UI 主题包的冻结 sRGB / alpha 映射 |
| 运行时矩阵 | 2 个主题 × 2 个模式，共 4 种合法组合 |

Mist UI 的通用技能默认一次只加载一个主题包。本项目因产品需求需要由后台切换站点主题，因此在消费层建立二维矩阵；每个主题包自身仍保持完整、独立的 light/dark 映射，不混合两个主题的私有原色，也不建立包含具体主题名的组件 token。

合法组合如下：

| 组合 | 用途 |
| --- | --- |
| `mist-sea-salt + light` | 默认阅读与清爽后台环境 |
| `mist-sea-salt + dark` | 低照度海蓝阅读环境 |
| `mist-forest + light` | 自然、平静的长时间阅读环境 |
| `mist-forest + dark` | 低照度青绿工作环境 |

## 2. 两个轴的权威与优先级

### 2.1 主题轴

主题是站点级配置，不是访客偏好。

```text
有效的 setting/lobby.theme
  > 上次服务端外观缓存（仅用于首屏防闪）
  > 合法 DOM data-theme
  > mist-sea-salt
```

- 前台导航不提供主题色选择器。
- `localStorage` 中的站点主题缓存只能由服务端响应或后台保存结果更新。
- 访客切换明暗模式时不得改写主题。
- 服务端返回新主题时必须无条件更新当前主题；不能因访客已有本地 mode 而拒绝主题同步。

### 2.2 模式轴

模式是访客级偏好，后台的 `mode` 只定义站点默认值。

```text
合法访客本地 mode
  > 有效的 setting/lobby.mode
  > 上次服务端默认模式缓存
  > 合法 DOM data-mode
  > prefers-color-scheme
  > light
```

- 前台和后台壳层均可提供明暗按钮；该按钮只更新访客本地 mode。
- “恢复默认”删除访客 mode 后，回到服务端默认模式；服务端不可用时再按 DOM、系统和 light 降级。
- 多标签页通过 `storage` 事件同步访客 mode 与服务端外观缓存。

### 2.3 首屏缓存不是配置权威

静态 Vite 页面无法在 HTML 返回前读取数据库，因此使用两类存储：

| Key | 内容 | 写入者 | 权威 |
| --- | --- | --- | --- |
| `blog:site-appearance` | 最近一次服务端返回的 `{ theme, mode }` | `setting` store / 后台保存流程 | 仅用于下次首屏防闪 |
| `blog:mode` | `light` 或 `dark` | 前台明暗控件 | 访客 mode 权威 |
| `blog:theme` | 旧版 `{ theme, mode }` | 只读迁移 | 仅迁移合法 mode，忽略旧 theme |

页面挂载后，`setting/lobby` 的主题始终覆盖缓存；本地 mode 仍可覆盖响应中的默认 mode。

## 3. 视觉定位

### 3.1 雾境海盐 `mist-sea-salt`

模拟海岸清晨薄雾中的透明蓝玻璃，关键词为清爽、安静、纯净和空气感。

| 角色 | Light | Dark |
| --- | --- | --- |
| 页面 | `#F5FAFF` | `#101B22` |
| 卡片 | `#F0F8FC` | `#162630` |
| 浮层 | `#FFFFFF` | `#1C303C` |
| 主文字 | `#203846` | `#EAF6FC` |
| 次文字 | `#3D5C6C` | `#BCD0DA` |
| muted | `#587485` | `#91ACB9` |
| accent | `#2F6F92` | `#8BCBE6` |
| accent hover | `#255E7C` | `#A5D9ED` |
| accent active | `#1B4D68` | `#6BB5D5` |
| strong border | `#6E93A6` | `#6E91A3` |

色彩比例为雾蓝 60%、透明白 30%、深海光影 10%。主光从左上进入，右下以海盐蓝建立景深；深海蓝始终承担主要交互，不让辅助浅蓝替代按钮 hover 色相。

### 3.2 雾境青森 `mist-forest`

模拟森林晨雾中的绿色玻璃和露水冷光，关键词为自然、呼吸、生命和平静。

| 角色 | Light | Dark |
| --- | --- | --- |
| 页面 | `#F7FCF8` | `#101B16` |
| 卡片 | `#F2FAF4` | `#17261F` |
| 浮层 | `#FFFFFF` | `#1D3028` |
| 主文字 | `#29473A` | `#EDF8F1` |
| 次文字 | `#476557` | `#C7DCCF` |
| muted | `#5D766A` | `#96B1A2` |
| accent | `#357258` | `#8FD0AD` |
| accent hover | `#2B604A` | `#A7DEC0` |
| accent active | `#224F3D` | `#79B89A` |
| strong border | `#759688` | `#789487` |

色彩比例为森林绿雾 50%、青色空气感 30%、米白暖光 20%。深森绿承担主要交互，露水青只用于 info 与环境折射；绿色状态语义必须仍用文字或图标区分，不能把所有绿色内容理解为 success。

## 4. 共享语义 token

主题文件必须完整定义以下颜色语义；组件不得设置私有原色或按主题分支。

| 分组 | 必需 token |
| --- | --- |
| 背景 | `--bg-primary`、`--bg-secondary`、`--bg-card`、`--bg-elevated`、`--bg-inset`、`--bg-hover`、`--bg-active`、`--bg-disabled` |
| 文字 | `--text-primary`、`--text-secondary`、`--text-muted`、`--text-disabled`、`--text-on-accent`、`--text-link`、`--text-inverse` |
| 交互 | `--accent`、`--accent-hover`、`--accent-active`、`--accent-soft`、`--accent-softer`、`--accent-gradient`、`--accent-gradient-hover` |
| 状态 | `--success(-soft)`、`--warning(-soft)`、`--danger(-soft)`、`--info(-soft)` |
| 边界 | `--border`、`--border-strong`、`--border-focus`、`--focus-ring-color`、`--focus-ring`、`--divider-fade`、`--line-soft`、`--line-strong` |
| 阴影 | `--shadow-color`、`--shadow-xs..xl`、`--shadow-inset(-color)`、`--shadow-glow` |
| 玻璃 | `--surface-glass-subtle`、`--surface-glass`、`--surface-glass-strong`、`--surface-glass-border`、`--surface-glass-highlight`、`--surface-glass-shadow` |
| 光场 | `--surface-highlight`、`--light-gradient-primary`、`--light-gradient-secondary` |
| 雾幕 | `--fog-color-1..3`、`--fog-opacity`、`--fog-blend-mode`、`--body-bg-spot-1..8`、`--blob-color-1..2` |
| 浮层 | `--nav-bg`、`--modal-overlay-bg`、`--scrim` |
| 内容媒介 | `--code-bg`、`--code-text`、`--figure-bg` |
| 表单图形 | `--select-arrow-color`、`--checkbox-check-color`、`--radio-dot-color` |

主题私有原色只允许出现在：

```text
web/src/styles/tokens/_mist-sea-salt.scss
web/src/styles/tokens/_mist-forest.scss
```

核心冻结值以 Mist UI 同名主题 token 为来源快照；`web/src/styles/tokens/_mist-sea-salt.scss` 与 `_mist-forest.scss` 是当前项目映射源，主题作用域由 `web/src/styles/themes/_mist.scss` 统一组织。组件不得复制这些具体色值。

| Token | 海盐 Light | 海盐 Dark | 青森 Light | 青森 Dark |
| --- | --- | --- | --- | --- |
| `--text-inverse` | `#EAF6FC` | `#203846` | `#EDF8F1` | `#29473A` |
| `--accent-softer` | `rgb(215 236 247 / 58%)` | `rgb(139 190 216 / 10%)` | `rgb(220 239 228 / 58%)` | `rgb(121 184 154 / 10%)` |
| `--divider-fade` 中点 | `#B6CFDD` | `#3D5D6E` | `#B8D6C0` | `#3D5F50` |
| `--code-bg` | `#132630` | `#081116` | `#13271E` | `#08120D` |
| `--code-text` | `#DCEFF7` | `#DCEFF7` | `#DFF3E7` | `#DFF3E7` |
| `--figure-bg` | `#DCEEF8` | `#172D38` | `#DFF1E6` | `#183128` |
| `--line-soft` | `rgb(78 145 181 / 28%)` | `rgb(139 190 216 / 20%)` | `rgb(63 128 101 / 28%)` | `rgb(121 184 154 / 20%)` |
| `--line-strong` | `#4E91B5` | `#8BBED8` | `#3F8065` | `#79B89A` |

`--divider-fade` 的实际值是 `linear-gradient(90deg, transparent, <中点>, transparent)`。这些别名必须在四种组合中同构存在；组件选择器不得把海盐值直接写入业务 SCSS。

## 5. 五层雾境氛围

页面按固定顺序构图：

1. `--bg-primary` 建立背景空间。
2. 八层 `--body-bg-spot-*` 与 `--light-gradient-primary` 建立主题光场。
3. 页面伪元素使用 `--fog-color-*` 建立雾幕。
4. 导航、首页主壳、侧栏与浮层使用分级玻璃表面。
5. 一个视口最多一个重点表面使用 `--light-gradient-secondary` 微光。

约束：

- 雾幕和光场不承载正文，统一 `pointer-events: none`。
- blur 保持静态，只动画 transform 与小范围 opacity。
- 一个视口最多两个大面积 blur 雾层；后台编辑器可启用 quiet 模式减少装饰。
- 不支持 `backdrop-filter` 时先使用 `--bg-card` / `--bg-elevated` 实色回退。
- 玻璃表面同时声明标准 `backdrop-filter` 与 `-webkit-backdrop-filter`；只有两者都不支持时才切换到实色回退。
- 避免玻璃套玻璃；正文、表格单元格和密集表单不重复 backdrop blur。
- `<= 640px` 停止 fixed background 和持续雾漂移，并降低模糊半径。

## 6. 基础几何与组件适配

颜色不进入基础层。两个主题共享以下 Mist UI 几何：

```scss
--glass-blur-sm: 14px;
--glass-blur-md: 24px;
--glass-blur-nav: 30px;
--glass-blur-lg: 32px;
--glass-saturation: 135%;
--fog-blur-sm: 40px;
--fog-blur-md: 60px;
--fog-blur-lg: 80px;
--aurora-drift: 10s;
--fog-drift-slow: 24s;
--fog-drift-slower: 30s;
```

为兼容现有编辑式排版与布局命名，基础层另提供以下不含颜色的别名：

```scss
--font-serif: "Iowan Old Style", "Palatino Linotype", "Noto Serif SC",
  "Songti SC", Georgia, serif;
--layout-workbench-max: var(--layout-admin-max);
--layout-rail-width: 248px;
--layout-toolbar-height: var(--layout-public-header);
--layout-toc-width: 220px;
```

这些别名不参与主题切换；海盐、青森与 light/dark 四种组合下的几何必须完全一致。

Mist UI 原始 `--focus-ring` 是颜色值，而现有项目把它作为完整 `box-shadow` 使用。项目适配层固定采用：

```scss
--focus-ring-color: <主题 focus 色>;
--focus-ring: 0 0 0 3px var(--focus-ring-color);
```

这是组件契约适配，不改变主题色值。新组件优先使用 `--shadow-glow`；迁移完成前两者必须保持同一主题色相。

## 7. 后台设置与公开 API

站点设置保存主题、默认模式，以及按主题 ID 隔离的主题元素：

```json
{
  "theme": "mist-sea-salt",
  "mode": "light",
  "theme_elements": {
    "mist-sea-salt": {
      "home_latest_empty_description": "第一篇文章正在潮汐之外酝酿。",
      "home_latest_end_text": "已经读到潮汐尽头"
    },
    "mist-forest": {
      "home_latest_empty_description": "第一篇文章正在林雾之间酝酿。",
      "home_latest_end_text": "已经走到林径尽头"
    }
  }
}
```

| 字段 | 合法值 | 管理方式 | 前台行为 |
| --- | --- | --- | --- |
| `theme` | `mist-sea-salt`、`mist-forest` | 后台“站点设置 → 外观” | 服务端主题无条件生效并更新首屏缓存 |
| `mode` | `light`、`dark` | 后台设置访客默认值 | 仅在访客没有本地 mode 时生效 |
| `theme_elements` | 以合法主题 ID 为键的元素对象 | 后台按主题分别编辑 | 前台只消费当前 `theme` 对应的元素组 |

当前主题元素的范围刻意保持最小，只覆盖首页“最近发布”区：

| 字段 | 雾境海盐默认值 | 雾境青森默认值 | 最大长度 |
| --- | --- | --- | ---: |
| `home_latest_empty_description` | 第一篇文章正在潮汐之外酝酿。 | 第一篇文章正在林雾之间酝酿。 | 160 个 Unicode 字符 |
| `home_latest_end_text` | 已经读到潮汐尽头 | 已经走到林径尽头 | 80 个 Unicode 字符 |

- 两个字段都是编辑性纯文本，不允许携带或解释为 HTML、SVG、CSS，也不承担颜色、布局、插图或组件结构配置。
- 加载、失败、重试、按钮、表单提示、状态播报和无障碍名称属于功能性文案，必须保持跨主题一致且语义中性。
- `mode` 只影响明暗视觉，不选择另一套文案；主题元素只由当前 `theme` 决定。
- 切换主题后读取对应元素组，未启用主题的已保存值仍须保留。

三个端点的契约如下：

- `GET setting/lobby`：公开返回当前 `theme`、默认 `mode` 和两个主题的完整、归一化 `theme_elements`；公开端点只读。
- `GET setting/detail`：后台返回与 lobby 相同的完整外观数据，用于编辑两个主题的元素。
- `PUT setting/update`：接受完整或兼容旧客户端的局部 `theme_elements`，成功后返回完整归一化结果并失效站点设置缓存。

更新合并规则固定为：

1. 请求省略 `theme_elements` 或传 `null` 时保留当前持久化映射；当前存储也不可用时才回退内置默认值。
2. 显式提交局部映射时，未出现的主题键保留当前值。
3. 一旦提交某个主题对象，该对象内缺失、`null` 或去除首尾空白后为空的字段回退到该主题内置默认值。
4. 非空值超过 160/80 字符上限时返回参数错误，非字符串类型作为畸形 JSON 请求拒绝；不得截断后伪装保存成功。

后台外观区必须：

- 展示两个主题的中文名、气质说明和语义预览。
- 在每个主题上下文中编辑该主题的首页最近发布空状态说明与列表结束文案，并明确显示长度上限。
- 保存真实 `form.theme`，禁止硬编码默认主题覆盖提交值。
- 将 mode 标注为“访客默认模式”，避免与当前管理员本地模式混淆。
- 保存成功后更新 setting store、DOM、首屏缓存与当前预览；保存失败不得伪装成已发布。

服务端必须校验主题、模式以及主题元素的类型与长度，不能截断无效输入或把所有输入强制改成默认值后返回成功；只有契约明确规定的缺失、`null` 或空白字段允许回退内置文案。主题元素即使包含类似标记的字符，也只能作为转义后的纯文本展示。

## 8. 兼容迁移

| 旧值 | 新值 |
| --- | --- |
| `forest` | `mist-forest` |
| `mist-violet` | `mist-sea-salt` |
| `strawberry` | `mist-sea-salt` |
| 空值或未知值 | `mist-sea-salt` |

- 数据库启动迁移只处理旧值或非法值，不得覆盖已经合法的海盐/青森选择。
- 旧数据库没有 `theme_elements_json`、值为空、JSON 无效或只含部分字段时，读取与迁移流程按主题逐字段补齐内置默认值；已有合法自定义值不得被覆盖。
- Redis 旧缓存缺少 `theme_elements` 或只含部分字段时，读取流程按相同规则归一化；下一次站点设置更新会主动失效缓存并写入完整结构。
- 旧版 `PUT setting/update` 省略 `theme_elements` 或传 `null` 时保留当前值，不能因旧客户端保存基础信息而重置主题文案。
- 旧 `blog:theme` 只迁移合法 mode；旧主题不得成为访客级权威。
- 旧 `_mist-violet.scss`、`_forest.scss`、`_strawberry.scss` 与 `themes/creamy.scss` 已从运行时源码删除。
- 旧数据导入器使用同一映射规则，避免导入后写回无效主题。

## 9. 可访问性与对比度基线

Mist UI 已校准的关键组合：

| 组合 | 海盐 | 青森 |
| --- | ---: | ---: |
| Light 主文字 / 页面 | `11.66:1` | `9.84:1` |
| Light muted / 页面 | `4.70:1` | `4.74:1` |
| 白色 / Light accent | `5.51:1` | `5.68:1` |
| Light strong border / 页面 | `3.13:1` | `3.13:1` |
| Dark 主文字 / 页面 | `15.88:1` | `16.21:1` |
| Dark muted / 浮层 | `5.73:1` | `6.05:1` |
| Dark on-accent / accent | `8.37:1` | `9.15:1` |
| Dark strong border / 浮层 | `4.06:1` | `4.24:1` |

- 普通文字目标不低于 `4.5:1`，大文字与关键边界不低于 `3:1`。
- 输入、选择器、切换器使用 `--border-strong`；focus 同时改变 `--border-focus` 并绘制 ring。
- 状态必须同时包含文字或图标，不能只靠蓝色或绿色区分。
- Dark 玻璃始终保留结构边框，不能只依赖黑色阴影。

## 10. 文件结构与加载顺序

```text
web/src/styles/
├── tokens/
│   ├── _base.scss
│   ├── _mist-sea-salt.scss
│   └── _mist-forest.scss
├── themes/
│   └── _mist.scss
└── index.scss
```

`index.scss` 的顺序固定为：reset → base token → 两个限定作用域的主题 token → Mist 氛围 → 共享组件 → 页面样式。

两个主题文件可以同时进入构建产物，但每个主题的选择器必须同时限定 `data-theme` 与 `data-mode`。禁止直接导入 Mist UI 原始的 `:root, [data-mode]` 单主题文件，否则后导入主题会覆盖前一个主题。

## 11. 验收矩阵

每次主题系统变更必须检查：

- [ ] 海盐 light / dark 与青森 light / dark 四种组合 token 全部有值。
- [ ] 编辑式语义 token 与基础布局别名均已定义，构建后不存在未解析的 `var()`。
- [ ] 后台切换主题后，当前页面、公开页面与新标签页都使用新主题。
- [ ] 已有本地 dark 的访客仍会接收后台主题更新。
- [ ] 前台明暗按钮只改变 mode，不改变 theme。
- [ ] 首页最近发布空状态与列表结束文案来自当前主题元素组，海盐与青森切换后使用各自文案。
- [ ] 功能性文案不随主题变化，主题元素中没有 HTML、SVG、CSS 或任意视觉代码。
- [ ] 清除本地 mode 后回到后台默认模式。
- [ ] 首屏脚本、Vue 状态、DOM、setting store 与 `theme-color` 一致。
- [ ] 旧存储、旧数据库值、旧 Redis 缓存、旧 PUT 请求和旧导入包安全归一化且不覆盖合法主题元素。
- [ ] 不支持 backdrop blur 与 reduced motion 时仍可读、可操作。
- [ ] 320、390、640、768、960、1200、1440px 无主题导致的布局变化。
- [ ] 组件样式没有新增具体色值或主题私有分支。

主题系统只有在文档、服务端持久化、首屏初始化、Vue 状态和四组合构建验收同时一致时才算完成。
