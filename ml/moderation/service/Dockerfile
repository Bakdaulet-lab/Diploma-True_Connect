# Build context is ml/moderation (so the trained ./model can be baked in if present).
FROM python:3.11-slim

WORKDIR /app

# Install serving deps first for better layer caching.
COPY service/requirements.txt /app/requirements.txt
RUN pip install --no-cache-dir -r /app/requirements.txt

# App code.
COPY service/app.py /app/app.py

# The trained model is provided at runtime via a bind mount at /app/model
# (see deployments/docker-compose.yml). If it is absent, the service
# auto-falls back to lexicon STUB mode, so the container always starts.
ENV MODEL_DIR=/app/model
EXPOSE 8000

CMD ["uvicorn", "app:app", "--host", "0.0.0.0", "--port", "8000"]
