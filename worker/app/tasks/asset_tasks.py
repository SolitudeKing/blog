from app.celery_app import celery_app


@celery_app.task(name="asset.generate_thumbnail")
def generate_thumbnail(asset_id: int) -> dict[str, int | str]:
    return {
        "asset_id": asset_id,
        "status": "pending-implementation",
    }

