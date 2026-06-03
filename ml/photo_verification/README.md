# TrueConnect — Photo Verification

A CV safety layer for **profile photos**: every avatar upload is checked for a
**visible face** and screened for **explicit (NSFW) content** before it is stored.
This complements KYC (which proves identity, outsourced to Sumsub) by keeping the
public photo feed real and halal.

- **Face presence** — OpenCV Haar cascade (bundled with opencv, no download).
- **NSFW** — HuggingFace ViT `Falconsai/nsfw_image_detection` (downloads on first run).
- Served as a Python/FastAPI microservice; the Go API calls `POST /verify` from
  `ProfileService.UploadPhoto`. If the service is down the API **fails open**
  (upload allowed) so verification never blocks the product.

## API
```
POST /verify   (multipart "image")
  → {has_face, face_count, nsfw_score, nsfw_label, decision: approve|reject, reasons}
GET  /health
```
`decision = reject` when `face_count == 0` (and REQUIRE_FACE) or `nsfw_score >= NSFW_THRESHOLD`.

## Run locally
```bash
cd ml/photo_verification/service
pip install -r requirements.txt
uvicorn app:app --port 8000
# contract smoke-test without models:
PHOTO_VERIFIER_STUB=1 uvicorn app:app --port 8000
```

Enable it in the API: set `PHOTO_VERIFIER_URL=http://photo-verifier:8000`
(docker) or `http://localhost:8000` (local), then restart the API.

## Config (env)
| var | default | meaning |
|---|---|---|
| `NSFW_THRESHOLD` | `0.7` | reject at/above this NSFW probability |
| `REQUIRE_FACE` | `1` | reject photos with no detected face |
| `PHOTO_VERIFIER_STUB` | `0` | `1` = always approve (no models, for tests) |

## Evaluation (diploma)
Put labeled images in `data/approve/` (real face + SFW) and `data/reject/`
(no face **or** NSFW), then:
```bash
pip install -r requirements.txt   # requests
python evaluate.py --url http://localhost:8000
# → reports/ : precision, recall, F1, accuracy, confusion matrix
```
See `reports/MODEL_CARD.md`.
