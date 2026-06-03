# Model Card — TrueConnect Photo Verification

## Purpose
Screen uploaded **profile photos** for (a) a visible human face and (b) absence
of explicit/NSFW content, so the public photo feed stays real and halal. This is
a *content-safety* layer, distinct from KYC identity verification (Sumsub).

## Components
| check | model | source | offline? |
|---|---|---|---|
| face presence | Haar cascade `haarcascade_frontalface_default` | bundled with OpenCV | yes |
| NSFW | ViT `Falconsai/nsfw_image_detection` | HuggingFace | downloaded once |

## Decision
`reject` if `face_count == 0` (when `REQUIRE_FACE=1`) **or**
`nsfw_score >= NSFW_THRESHOLD` (default 0.7); else `approve`.
Capabilities degrade independently — if the NSFW model can't load, only face
presence is enforced (the service never hard-fails).

## Evaluation
Binary task with **positive = "reject"** (catching a bad photo), measured on a
labeled set (`data/approve`, `data/reject`) via `evaluate.py`.

> Fill after running `evaluate.py`:
>
> | metric | value |
> |---|---|
> | precision | _ |
> | recall | _ |
> | F1 | _ |
> | accuracy | _ |

Tune `NSFW_THRESHOLD` on the precision/recall trade-off: lower = stricter (more
NSFW caught, more false rejects); higher = more permissive.

## Limitations & ethics
- **Haar cascade** is frontal-face only; profile angles, occlusion (niqab/hijab
  framing), low light → false "no_face". Consider a DNN face detector
  (MediaPipe / OpenCV SSD / RetinaFace) for production; keep a manual-review
  appeal path so legitimate users are never permanently blocked.
- **NSFW model** has its own biases/false positives; threshold is a policy knob.
- **Fail-open** by design: a verifier outage allows uploads (availability over
  strictness) — acceptable because explicit content is also caught by reporting
  + trust penalties downstream.
- No image is stored by the verifier; it only returns a verdict.
