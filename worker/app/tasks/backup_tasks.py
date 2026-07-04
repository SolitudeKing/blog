from app.celery_app import celery_app


@celery_app.task(name="backup.create")
def create_backup() -> dict[str, str]:
    return {
        "status": "pending-implementation",
    }

