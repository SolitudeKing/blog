from datetime import datetime, timezone

from app.celery_app import celery_app


@celery_app.task(name="system.health")
def health() -> dict[str, str]:
    return {
        "status": "ok",
        "time": datetime.now(timezone.utc).isoformat(),
    }

