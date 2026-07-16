# 旧博客迁移运行手册

本文档用于 M4 阶段把 `blog-mini-serve` 旧项目数据迁移到新个人博客系统。

## 迁移对象

- SQLite：`blog-mini-serve/instance/BlogMini.sqlite`
- Markdown：`blog-mini-serve/articles/*.md`
- 图片文件：`blog-mini-serve/pics/*`
- PicBed Base64 图片：`PicBed.img_base64`

## 迁移步骤

1. 导出旧数据迁移包：

   ```bash
   python scripts/export_legacy_blog.py --source blog-mini-serve --output migration-output/legacy-export
   ```

2. 检查导出报告：

   ```bash
   cat migration-output/legacy-export/legacy-export-report.json
   ```

3. 预检导入计划：

   ```bash
   cd server
   go run ./cmd/import_legacy --dry-run --input ../migration-output/legacy-export/legacy-export.json
   ```

4. 导入到新系统 MySQL：

   ```bash
   cd server
   go run ./cmd/import_legacy --input ../migration-output/legacy-export/legacy-export.json
   ```

   导入器会读取 `ADMIN_USERNAME`、`ADMIN_PASSWORD`、`ADMIN_NICKNAME`：若管理员不存在则在同一事务中创建，并将迁移文章关联到该用户；若同名管理员已存在则复用其真实 ID。

5. 检查导入报告：

   ```bash
   cat migration-output/legacy-import-report.json
   ```

## 映射规则

- `ArticleTypes` 直接迁移为 `topics`；若目标库曾使用中间态 `categories`，先把其名称、slug、描述与排序迁入 `topics`，并补齐非空 `label`。
- `ArticleLabels` 迁移为 `tags`。
- `Articles.article_type`（或中间态 `articles.category_id`）按旧分类 ID 映射到新 `articles.topic_id`。
- 专题 `name` 沿用旧分类名称；`label` 优先使用可用短标识，否则由名称生成；`slug` 保持稳定且唯一；`description`、`cover_url` 可为空；`sort_order` 沿用旧顺序或默认 `0`。
- `Articles.article_label` 按 `|` 拆分为旧标签 ID 列表。
- `article_status`：`0` 为草稿、`1` 为已发布、`2` 为私有。
- 文章正文优先读取 `article_content` Base64；为空时读取同名 Markdown 文件。
- PicBed 图片导出到迁移包 `assets/`，导入后写入 `STORAGE_LOCAL_ROOT/legacy/`。
- 迁移导入按 `slug` 和 `storage_key` 幂等更新，可重复执行。

## 验收检查

- 导出报告中 `failures` 为空。
- 导入报告中 `articles_imported` 与旧文章数量一致。
- 导入报告中 `topics_imported` 与旧分类数量一致，且没有无法解析的专题映射。
- 导入报告中 `assets_imported` 与旧图片数量一致。
- 后台文章列表能看到迁移文章。
- 后台专题列表数量与旧分类数量一致，文章的 `topic_id` 引用均有效。
- 前台已发布文章详情可打开。
- `/uploads/legacy/{filename}` 图片可访问。
