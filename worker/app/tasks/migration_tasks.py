from app.celery_app import celery_app


@celery_app.task(name="migration.import_legacy_blog")
def import_legacy_blog() -> dict[str, str]:
    return {
        "status": "pending-implementation",
    }

