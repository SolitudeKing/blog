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

5. 检查导入报告：

   ```bash
   cat migration-output/legacy-import-report.json
   ```

## 映射规则

- `ArticleTypes` 迁移为 `categories`。
- `ArticleLabels` 迁移为 `tags`。
- `Articles.article_type` 按旧分类 ID 映射到新分类。
- `Articles.article_label` 按 `|` 拆分为旧标签 ID 列表。
- `article_status`：`0` 为草稿、`1` 为已发布、`2` 为私有。
- 文章正文优先读取 `article_content` Base64；为空时读取同名 Markdown 文件。
- PicBed 图片导出到迁移包 `assets/`，导入后写入 `STORAGE_LOCAL_ROOT/legacy/`。
- 迁移导入按 `slug` 和 `storage_key` 幂等更新，可重复执行。

## 验收检查

- 导出报告中 `failures` 为空。
- 导入报告中 `articles_imported` 与旧文章数量一致。
- 导入报告中 `assets_imported` 与旧图片数量一致。
- 后台文章列表能看到迁移文章。
- 前台已发布文章详情可打开。
- `/uploads/legacy/{filename}` 图片可访问。
