#!/usr/bin/env python3
"""Export the legacy BlogMini SQLite/Markdown/picbed data into a portable package."""

from __future__ import annotations

import argparse
import base64
import hashlib
import json
import mimetypes
import re
import shutil
import sqlite3
from datetime import datetime, timezone
from pathlib import Path
from typing import Any


DEFAULT_SOURCE = Path("blog-mini-serve")
DEFAULT_OUTPUT = Path("migration-output/legacy-export")


def main() -> int:
    parser = argparse.ArgumentParser(description="Export legacy BlogMini data.")
    parser.add_argument("--source", default=str(DEFAULT_SOURCE), help="Legacy project root.")
    parser.add_argument("--output", default=str(DEFAULT_OUTPUT), help="Export package directory.")
    args = parser.parse_args()

    source = Path(args.source)
    output = Path(args.output)
    db_path = source / "instance" / "BlogMini.sqlite"
    articles_dir = source / "articles"
    pics_dir = source / "pics"
    assets_dir = output / "assets"

    report = new_report(source, db_path, output)
    if not db_path.exists():
        report["failures"].append(f"SQLite database not found: {db_path}")
        write_json(output / "legacy-export-report.json", report)
        return 1

    output.mkdir(parents=True, exist_ok=True)
    assets_dir.mkdir(parents=True, exist_ok=True)

    with sqlite3.connect(db_path) as conn:
        conn.row_factory = sqlite3.Row
        tables = set(load_table_names(conn))
        categories = export_categories(conn, tables, report)
        tags = export_tags(conn, tables, report)
        settings = export_settings(conn, tables, report)
        notices = export_notices(conn, tables, report)
        assets = export_assets(conn, tables, pics_dir, assets_dir, report)
        articles = export_articles(conn, tables, articles_dir, categories, tags, report)

    payload = {
        "version": 1,
        "generated_at": now_iso(),
        "source": {
            "root": str(source),
            "database": str(db_path),
            "articles_dir": str(articles_dir),
            "pics_dir": str(pics_dir),
        },
        "settings": settings,
        "categories": categories,
        "tags": tags,
        "articles": articles,
        "assets": assets,
        "notices": notices,
        "report": report,
    }

    report["counts"] = {
        "categories": len(categories),
        "tags": len(tags),
        "articles": len(articles),
        "assets": len(assets),
        "notices": len(notices),
    }
    write_json(output / "legacy-export.json", payload)
    write_json(output / "legacy-export-report.json", report)
    print(f"Exported legacy package: {output / 'legacy-export.json'}")
    print(json.dumps(report["counts"], ensure_ascii=False, indent=2))
    return 0 if not report["failures"] else 2


def new_report(source: Path, db_path: Path, output: Path) -> dict[str, Any]:
    return {
        "generated_at": now_iso(),
        "source": str(source),
        "database": str(db_path),
        "output": str(output),
        "counts": {},
        "warnings": [],
        "failures": [],
    }


def load_table_names(conn: sqlite3.Connection) -> list[str]:
    rows = conn.execute("SELECT name FROM sqlite_master WHERE type = 'table'").fetchall()
    return [row["name"] for row in rows]


def export_categories(conn: sqlite3.Connection, tables: set[str], report: dict[str, Any]) -> list[dict[str, Any]]:
    if "ArticleTypes" not in tables:
        report["warnings"].append("ArticleTypes table not found; uncategorized fallback will be used.")
        return []
    rows = conn.execute("SELECT * FROM ArticleTypes ORDER BY id ASC").fetchall()
    categories = []
    used_slugs: set[str] = set()
    for row in rows:
        name = str(row["article_type"] or "").strip()
        if not name:
            report["warnings"].append(f"Skip empty category name, legacy id={row['id']}.")
            continue
        slug = unique_slug(name, used_slugs, f"category-{row['id']}")
        categories.append(
            {
                "legacy_id": str(row["id"]),
                "name": name,
                "slug": slug,
                "description": "",
                "sort_order": len(categories) + 1,
                "created_at": parse_time(row["create_time"]),
                "updated_at": parse_time(row["last_editors"] or row["create_time"]),
            }
        )
    return categories


def export_tags(conn: sqlite3.Connection, tables: set[str], report: dict[str, Any]) -> list[dict[str, Any]]:
    if "ArticleLabels" not in tables:
        report["warnings"].append("ArticleLabels table not found; article tags will be empty.")
        return []
    rows = conn.execute("SELECT * FROM ArticleLabels ORDER BY id ASC").fetchall()
    tags = []
    used_slugs: set[str] = set()
    for row in rows:
        name = str(row["article_label"] or "").strip()
        if not name:
            report["warnings"].append(f"Skip empty tag name, legacy id={row['id']}.")
            continue
        slug = unique_slug(name, used_slugs, f"tag-{row['id']}")
        tags.append(
            {
                "legacy_id": str(row["id"]),
                "name": name,
                "slug": slug,
                "description": "",
                "color": "",
                "created_at": parse_time(row["create_time"]),
                "updated_at": parse_time(row["last_editors"] or row["create_time"]),
            }
        )
    return tags


def export_settings(conn: sqlite3.Connection, tables: set[str], report: dict[str, Any]) -> dict[str, Any]:
    defaults = {
        "site_name": "Solitude Blog",
        "author": "Solitude King",
        "essay": "Keep writing, keep shipping.",
        "theme": "forest",
        "mode": "light",
        "social_links": {},
        "about_content": "",
    }
    if "BlogConfig" not in tables:
        report["warnings"].append("BlogConfig table not found; default settings will be used.")
        return defaults
    row = conn.execute("SELECT * FROM BlogConfig WHERE id = 1").fetchone()
    if row is None:
        report["warnings"].append("BlogConfig id=1 not found; default settings will be used.")
        return defaults
    about_content = decode_text(row["about_content"], report, "BlogConfig.about_content")
    return {
        "site_name": "Solitude Blog",
        "author": row["blog_author"] or defaults["author"],
        "essay": row["blog_essay"] or defaults["essay"],
        "theme": "forest",
        "mode": "light",
        "social_links": {
            "gitee": row["gitee_link"] or "",
            "bilibili": row["bilibili_link"] or "",
            "douyin": row["douyin_link"] or "",
            "qq_img": row["qq_img"] or "",
            "wechat_pay": row["wecat_pay"] or "",
            "ali_pay": row["ali_pay"] or "",
        },
        "about_content": about_content,
    }


def export_notices(conn: sqlite3.Connection, tables: set[str], report: dict[str, Any]) -> list[dict[str, Any]]:
    if "BlogNotice" not in tables:
        return []
    rows = conn.execute("SELECT * FROM BlogNotice ORDER BY id ASC").fetchall()
    notices = []
    for row in rows:
        title = str(row["notice_title"] or "").strip()
        content = str(row["notice_content"] or "").strip()
        if not title and not content:
            report["warnings"].append(f"Skip empty notice, legacy id={row['id']}.")
            continue
        notices.append(
            {
                "legacy_id": str(row["id"]),
                "title": title or "旧公告",
                "content": content,
                "enabled": True,
                "sort_order": len(notices) + 1,
                "created_at": parse_time(row["create_time"]),
                "updated_at": parse_time(row["last_editors"] or row["create_time"]),
            }
        )
    return notices


def export_articles(
    conn: sqlite3.Connection,
    tables: set[str],
    articles_dir: Path,
    categories: list[dict[str, Any]],
    tags: list[dict[str, Any]],
    report: dict[str, Any],
) -> list[dict[str, Any]]:
    if "Articles" not in tables:
        report["failures"].append("Articles table not found.")
        return []
    category_ids = {item["legacy_id"] for item in categories}
    tag_ids = {item["legacy_id"] for item in tags}
    rows = conn.execute("SELECT * FROM Articles ORDER BY create_time ASC, id ASC").fetchall()
    articles = []
    used_slugs: set[str] = set()
    for row in rows:
        title = str(row["article_title"] or "").strip()
        if not title:
            report["warnings"].append(f"Skip article without title, legacy id={row['id']}.")
            continue
        content = read_article_content(row, articles_dir, report)
        category_legacy_id = str(row["article_type"] or "")
        if category_legacy_id not in category_ids:
            report["warnings"].append(f"Article {title} references unknown category {category_legacy_id}.")
            category_legacy_id = ""
        article_tag_ids = split_legacy_ids(row["article_label"])
        missing_tags = [tag_id for tag_id in article_tag_ids if tag_id not in tag_ids]
        for tag_id in missing_tags:
            report["warnings"].append(f"Article {title} references unknown tag {tag_id}.")
        article_tag_ids = [tag_id for tag_id in article_tag_ids if tag_id in tag_ids]
        slug = unique_slug(title, used_slugs, "article-" + short_hash(str(row["id"]) + title))
        created_at = parse_time(row["create_time"])
        updated_at = parse_time(row["last_editors"] or row["create_time"])
        status = map_article_status(row["article_status"])
        articles.append(
            {
                "legacy_id": str(row["id"]),
                "title": title,
                "slug": slug,
                "summary": make_summary(content),
                "content_md": content,
                "status": status,
                "category_legacy_id": category_legacy_id,
                "tag_legacy_ids": article_tag_ids,
                "view_count": int(row["browse_num"] or 0),
                "created_at": created_at,
                "updated_at": updated_at,
                "published_at": created_at if status == "published" else "",
            }
        )
    return articles


def export_assets(
    conn: sqlite3.Connection,
    tables: set[str],
    pics_dir: Path,
    assets_dir: Path,
    report: dict[str, Any],
) -> list[dict[str, Any]]:
    assets = []
    seen_names: set[str] = set()
    if "PicBed" in tables:
        rows = conn.execute("SELECT * FROM PicBed ORDER BY create_time ASC, id ASC").fetchall()
        for row in rows:
            exported = export_picbed_row(row, pics_dir, assets_dir, report)
            if exported:
                assets.append(exported)
                seen_names.add(Path(exported["file_path"]).name)
    if pics_dir.exists():
        for path in sorted(pics_dir.iterdir()):
            if not path.is_file() or path.name in seen_names:
                continue
            exported = copy_asset_file(path, assets_dir / path.name, path.name, parse_time(""), parse_time(""))
            assets.append(exported)
    return assets


def export_picbed_row(row: sqlite3.Row, pics_dir: Path, assets_dir: Path, report: dict[str, Any]) -> dict[str, Any] | None:
    legacy_id = str(row["id"] or "").strip()
    suffix = str(row["img_suffix"] or Path(legacy_id).suffix or "").strip()
    display_name = str(row["img_name"] or Path(legacy_id).stem or legacy_id).strip()
    filename = legacy_id if Path(legacy_id).suffix else f"{display_name}{suffix}"
    filename = sanitize_filename(filename or f"legacy-asset-{short_hash(display_name)}{suffix}")
    target = assets_dir / filename
    source = pics_dir / legacy_id
    try:
        if row["img_base64"]:
            target.write_bytes(base64.b64decode(row["img_base64"], validate=False))
        elif source.exists():
            shutil.copy2(source, target)
        else:
            report["warnings"].append(f"PicBed asset missing binary data and file: {legacy_id}")
            return None
    except Exception as exc:
        report["failures"].append(f"Export asset {legacy_id} failed: {exc}")
        return None
    return asset_payload(
        target,
        legacy_id,
        display_name,
        parse_time(row["create_time"]),
        parse_time(row["create_time"]),
        active=bool(row["pic_status"] != 0),
    )


def copy_asset_file(source: Path, target: Path, legacy_id: str, created_at: str, updated_at: str) -> dict[str, Any]:
    shutil.copy2(source, target)
    return asset_payload(target, legacy_id, source.stem, created_at, updated_at, active=True)


def asset_payload(
    target: Path,
    legacy_id: str,
    display_name: str,
    created_at: str,
    updated_at: str,
    active: bool,
) -> dict[str, Any]:
    content = target.read_bytes()
    mime_type = mimetypes.guess_type(target.name)[0] or "application/octet-stream"
    storage_key = "legacy/" + target.name
    return {
        "legacy_id": legacy_id,
        "display_name": display_name or target.stem,
        "alt_text": display_name or target.stem,
        "storage_key": storage_key,
        "url": "/uploads/" + storage_key,
        "thumb_url": "",
        "mime_type": mime_type,
        "ext": target.suffix,
        "size": len(content),
        "sha256": hashlib.sha256(content).hexdigest(),
        "status": "active" if active else "temporary",
        "created_at": created_at,
        "updated_at": updated_at,
        "file_path": str(Path("assets") / target.name),
    }


def read_article_content(row: sqlite3.Row, articles_dir: Path, report: dict[str, Any]) -> str:
    encoded = row["article_content"]
    title = str(row["article_title"] or "").strip()
    if encoded:
        return decode_text(encoded, report, f"Articles.article_content:{title}")
    md_path = articles_dir / f"{title}.md"
    if md_path.exists():
        return md_path.read_text(encoding="utf-8")
    report["warnings"].append(f"Article markdown file not found: {md_path}")
    return ""


def decode_text(value: Any, report: dict[str, Any], label: str) -> str:
    if value is None or value == "":
        return ""
    text = str(value)
    try:
        return base64.b64decode(text, validate=False).decode("utf-8")
    except Exception:
        report["warnings"].append(f"Base64 decode failed for {label}; value kept as plain text.")
        return text


def map_article_status(value: Any) -> str:
    mapping = {0: "draft", 1: "published", 2: "private"}
    try:
        return mapping.get(int(value), "draft")
    except (TypeError, ValueError):
        return "draft"


def split_legacy_ids(value: Any) -> list[str]:
    if value is None:
        return []
    return [part.strip() for part in str(value).split("|") if part.strip()]


def make_summary(content: str) -> str:
    text = re.sub(r"!\[[^\]]*]\([^)]*\)", "", content)
    text = re.sub(r"\[[^\]]+]\([^)]*\)", "", text)
    text = re.sub(r"[#>*_`~\-]+", " ", text)
    text = re.sub(r"\s+", " ", text).strip()
    return text[:180]


def unique_slug(value: str, used: set[str], fallback: str) -> str:
    base = slugify(value) or fallback
    slug = base[:180].strip("-") or fallback
    current = slug
    index = 2
    while current in used:
        suffix = f"-{index}"
        current = (slug[: 180 - len(suffix)] + suffix).strip("-")
        index += 1
    used.add(current)
    return current


def slugify(value: str) -> str:
    normalized = value.lower()
    normalized = re.sub(r"[^a-z0-9]+", "-", normalized)
    return normalized.strip("-")


def sanitize_filename(value: str) -> str:
    path = Path(value)
    stem = re.sub(r"[^A-Za-z0-9._-]+", "-", path.stem).strip("-") or short_hash(value)
    return stem[:120] + path.suffix.lower()


def short_hash(value: str) -> str:
    return hashlib.sha1(value.encode("utf-8")).hexdigest()[:10]


def parse_time(value: Any) -> str:
    if value is None or value == "":
        return ""
    text = str(value).strip()
    try:
        timestamp = int(float(text))
        return datetime.fromtimestamp(timestamp, timezone.utc).isoformat()
    except ValueError:
        return text


def now_iso() -> str:
    return datetime.now(timezone.utc).isoformat()


def write_json(path: Path, payload: Any) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(payload, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


if __name__ == "__main__":
    raise SystemExit(main())
