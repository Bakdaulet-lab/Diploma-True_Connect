"""FastAPI photo-verification service: face presence + NSFW screening.

Endpoint:
  POST /verify   (multipart field "image") ->
      {has_face, face_count, nsfw_score, nsfw_label, decision, reasons}
  GET  /health

The Go backend (internal/adapter/photoverify) calls /verify before storing a
profile photo. If this service is down the Go side fails open (allows upload).

Capabilities degrade independently and gracefully:
  * Face detection — OpenCV Haar cascade (ships with opencv, no download).
  * NSFW detection — HuggingFace ViT 'Falconsai/nsfw_image_detection'
    (downloaded on first run). If transformers/torch/model are unavailable,
    NSFW scoring is skipped (score 0) instead of failing.
  * PHOTO_VERIFIER_STUB=1 disables both (always approve) for contract tests.
"""
import io
import os

from fastapi import FastAPI, File, UploadFile

NSFW_THRESHOLD = float(os.getenv("NSFW_THRESHOLD", "0.7"))
REQUIRE_FACE = os.getenv("REQUIRE_FACE", "1") == "1"
STUB = os.getenv("PHOTO_VERIFIER_STUB", "0") == "1"

app = FastAPI(title="TrueConnect Photo Verifier", version="1.0")

_face_cascade = None
_nsfw_pipe = None


@app.on_event("startup")
def _load():
    global _face_cascade, _nsfw_pipe
    if STUB:
        return
    # Face detector (OpenCV Haar — bundled, offline).
    try:
        import cv2
        path = os.path.join(cv2.data.haarcascades, "haarcascade_frontalface_default.xml")
        _face_cascade = cv2.CascadeClassifier(path)
    except Exception as e:  # pragma: no cover
        print(f"[photo-verifier] face detector unavailable: {e}")
    # NSFW classifier (optional; downloads on first run).
    try:
        from transformers import pipeline
        _nsfw_pipe = pipeline("image-classification",
                              model="Falconsai/nsfw_image_detection")
    except Exception as e:  # pragma: no cover
        print(f"[photo-verifier] NSFW model unavailable, scoring disabled: {e}")


def _count_faces(image_bytes: bytes) -> int | None:
    if _face_cascade is None:
        return None
    import cv2
    import numpy as np
    arr = np.frombuffer(image_bytes, np.uint8)
    img = cv2.imdecode(arr, cv2.IMREAD_COLOR)
    if img is None:
        return 0
    gray = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)
    faces = _face_cascade.detectMultiScale(gray, scaleFactor=1.1,
                                           minNeighbors=5, minSize=(40, 40))
    return len(faces)


def _nsfw_score(image_bytes: bytes) -> float | None:
    if _nsfw_pipe is None:
        return None
    from PIL import Image
    img = Image.open(io.BytesIO(image_bytes)).convert("RGB")
    preds = _nsfw_pipe(img)
    for p in preds:
        if str(p["label"]).lower() == "nsfw":
            return float(p["score"])
    return 0.0


@app.get("/health")
def health():
    mode = "stub" if STUB else "active"
    return {
        "status": "ok",
        "mode": mode,
        "face_detection": _face_cascade is not None,
        "nsfw_detection": _nsfw_pipe is not None,
    }


@app.post("/verify")
async def verify(image: UploadFile = File(...)):
    data = await image.read()

    if STUB:
        return {"has_face": True, "face_count": 1, "nsfw_score": 0.0,
                "nsfw_label": "normal", "decision": "approve", "reasons": []}

    face_count = _count_faces(data)
    nsfw = _nsfw_score(data)

    reasons = []
    if REQUIRE_FACE and face_count is not None and face_count == 0:
        reasons.append("no_face")
    if nsfw is not None and nsfw >= NSFW_THRESHOLD:
        reasons.append("nsfw")

    return {
        "has_face": (face_count or 0) > 0 if face_count is not None else True,
        "face_count": face_count if face_count is not None else -1,
        "nsfw_score": nsfw if nsfw is not None else 0.0,
        "nsfw_label": "nsfw" if (nsfw is not None and nsfw >= NSFW_THRESHOLD) else "normal",
        "decision": "reject" if reasons else "approve",
        "reasons": reasons,
    }
