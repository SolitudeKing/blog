from app.celery_app import celery_app


@celery_app.task(name="search.rebuild_index")
def rebuild_index() -> dict[str, str]:
    return {
        "status": "pending-implementation",
    }

