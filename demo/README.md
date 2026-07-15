# 潮汐之间 · 个人博客静态站点

基于 Mist UI「雾境海盐」构建的单主题静态博客与创作后台，包含 Light / Dark 两种模式，不包含额外主题矩阵。

## 页面

- `index.html`：非对称编辑式首页与潮汐观测站
- `article.html`：无全站导航的专注阅读流、左侧返回通道、阅读进度、浮动目录与上一篇 / 下一篇
- `archives.html`：按年份展开的连续时间轴
- `search.html`：实时搜索工作台与空状态
- `about.html`：个人叙事、Now、原则与联系入口
- `admin.html`：博客创作与内容管理后台工作台

## 本地预览

在当前目录启动任意静态文件服务器，例如：

```powershell
python -m http.server 4173 --bind 127.0.0.1
```

然后打开 `http://127.0.0.1:4173/index.html`。

## 结构

主题入口位于 `assets/styles/themes/mist-sea-salt.css`，只导入基础 token、海盐 Light/Dark token 与共享组件。前台构图位于 `assets/styles/pages.css`；后台构图位于 `assets/styles/admin.css`，交互位于 `assets/app.js` 与 `assets/admin.js`。
