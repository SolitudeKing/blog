from celery import Celery

from app.config import settings

celery_app = Celery(
    "solitude_blog_worker",
    broker=settings.celery_broker_url,
    backend=settings.celery_result_backend,
    include=[
        "app.tasks.health_tasks",
        "app.tasks.article_tasks",
        "app.tasks.asset_tasks",
        "app.tasks.search_tasks",
        "app.tasks.sitemap_tasks",
        "app.tasks.backup_tasks",
        "app.tasks.migration_tasks",
    ],
)

celery_app.conf.update(
    task_track_started=True,
    task_serializer="json",
    result_serializer="json",
    accept_content=["json"],
    timezone="UTC",
    enable_utc=True,
)

