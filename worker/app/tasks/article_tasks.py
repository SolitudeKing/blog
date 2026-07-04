from app.celery_app import celery_app


@celery_app.task(name="article.extract_metadata")
def extract_metadata(article_id: int) -> dict[str, int | str]:
    return {
        "article_id": article_id,
        "status": "pending-implementation",
    }

