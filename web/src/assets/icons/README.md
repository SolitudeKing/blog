# SVG 图标维护

Web 端图标统一通过 `SvgIcon` 组件引用，页面和业务组件中不要直接编写 `<svg>`、`<path>` 等图形代码。

## 新增图标

1. 在当前目录新增语义化命名的 `.svg` 文件。
2. 使用 `viewBox`，颜色使用 `currentColor`，避免写入主题色值。
3. 将文件名加入 `src/config/svgIcons.ts` 的 `svgIconNames`。
4. 在 Vue 模板中引用：

```vue
<SvgIcon name="arrow-right" />
```

图标默认作为装饰元素隐藏读屏语义；图标独立传达信息时提供 `label`：

```vue
<SvgIcon name="info" label="文章信息" />
```

尺寸优先沿用所在组件的 CSS，也可以通过 `size` 传入数字像素值或 CSS 长度。
