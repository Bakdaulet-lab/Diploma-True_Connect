"""FastAPI inference service for content moderation.

Endpoints:
  GET  /health            → {"status": "ok", "mode": "model"|"stub"}
  POST /moderate          → {text} -> {label, scores, lang_detected}
  POST /moderate/batch    → {texts: [...]} -> {results: [...]}

The Go backend (internal/adapter/moderation) calls /moderate. If this service is
down, the Go side falls back to its keyword filter — so this service can fail
without taking moderation offline.

Model loading:
  * Loads a fine-tuned XLM-RoBERTa from MODEL_DIR (default /app/model).
  * If MODERATION_STUB=1 (or the model dir is missing), runs a lightweight
    lexicon STUB so the HTTP contract works without torch/a trained model —
    handy for smoke tests and CI.
"""
import os
from typing import List

from fastapi import FastAPI
from pydantic import BaseModel

LABELS = ["clean", "warn", "block"]
MODEL_DIR = os.getenv("MODEL_DIR", "/app/model")
# Use the real model only when a trained checkpoint is actually present
# (config.json). An empty bind-mounted dir → fall back to lexicon STUB mode.
_HAS_MODEL = os.path.exists(os.path.join(MODEL_DIR, "config.json"))
STUB = os.getenv("MODERATION_STUB", "0") == "1" or not _HAS_MODEL

KAZAKH_CHARS = set("әғқңөұүһі")

app = FastAPI(title="TrueConnect Moderation", version="1.0")

_tokenizer = None
_model = None
_torch = None


class ModerateRequest(BaseModel):
    text: str
    lang: str | None = None


class BatchRequest(BaseModel):
    texts: List[str]


def detect_lang(text: str) -> str:
    low = text.lower()
    if any(ch in KAZAKH_CHARS for ch in low):
        return "kk"
    if any("Ѐ" <= ch <= "ӿ" for ch in low):
        return "ru"
    return "en"


# ── Stub classifier (no ML) ──────────────────────────────────────────────────
_STUB_BLOCK = ["секс", "порно", "интим", "эротик", "sex", "porn", "nude",
               "naked", "xxx", "escort", "проституция", "эскорт"]
_STUB_WARN = ["ақымақ", "дурак", "идиот", "придурок", "scam", "мошенник",
              "harassment", "idiot", "badword"]


def _stub_predict(text: str):
    low = text.lower()
    if any(w in low for w in _STUB_BLOCK):
        scores = {"clean": 0.05, "warn": 0.1, "block": 0.85}
    elif any(w in low for w in _STUB_WARN):
        scores = {"clean": 0.15, "warn": 0.75, "block": 0.1}
    else:
        scores = {"clean": 0.9, "warn": 0.07, "block": 0.03}
    label = max(scores, key=scores.get)
    return label, scores


# ── Real model ───────────────────────────────────────────────────────────────
@app.on_event("startup")
def _load():
    global _tokenizer, _model, _torch
    if STUB:
        return
    import torch
    from transformers import AutoModelForSequenceClassification, AutoTokenizer
    _torch = torch
    _tokenizer = AutoTokenizer.from_pretrained(MODEL_DIR)
    _model = AutoModelForSequenceClassification.from_pretrained(MODEL_DIR).eval()


def _model_predict_batch(texts: List[str]):
    enc = _tokenizer(texts, truncation=True, max_length=128,
                     padding=True, return_tensors="pt")
    with _torch.no_grad():
        logits = _model(**enc).logits
        probs = _torch.softmax(logits, dim=-1).cpu().tolist()
    # Respect the model's own id2label ordering.
    id2label = _model.config.id2label
    results = []
    for row in probs:
        scores = {id2label[i]: float(p) for i, p in enumerate(row)}
        label = max(scores, key=scores.get)
        results.append((label, scores))
    return results


def _predict_one(text: str):
    if STUB:
        return _stub_predict(text)
    return _model_predict_batch([text])[0]


@app.get("/health")
def health():
    return {"status": "ok", "mode": "stub" if STUB else "model"}


@app.post("/moderate")
def moderate(req: ModerateRequest):
    label, scores = _predict_one(req.text)
    return {
        "label": label,
        "scores": scores,
        "lang_detected": req.lang or detect_lang(req.text),
    }


@app.post("/moderate/batch")
def moderate_batch(req: BatchRequest):
    if STUB:
        preds = [_stub_predict(t) for t in req.texts]
    else:
        preds = _model_predict_batch(req.texts)
    return {
        "results": [
            {"label": label, "scores": scores, "lang_detected": detect_lang(t)}
            for t, (label, scores) in zip(req.texts, preds)
        ]
    }
