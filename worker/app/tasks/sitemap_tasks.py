from app.celery_app import celery_app


@celery_app.task(name="sitemap.generate")
def generate_sitemap() -> dict[str, str]:
    return {
        "status": "pending-implementation",
    }

