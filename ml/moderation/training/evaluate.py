"""Evaluate the moderation model and produce the diploma report.

Generates, into ../reports/:
  * metrics.json         — per-class P/R/F1, macro-F1, ROC-AUC (ovr), accuracy
  * confusion_matrix.png — model confusion matrix on the test set
  * report.md            — human-readable summary incl. the BASELINE comparison,
                           the Kazakh zero-shot result, and the adversarial table

Compares the ML model against the keyword baseline (keyword_baseline.py) on:
  1. the held-out test set,
  2. the Kazakh gold set (zero-shot) if data/processed/kk_gold.csv exists,
  3. a curated adversarial/bypass set (obfuscation the keyword filter misses).

Usage:
    python evaluate.py
"""
import json
import os

import numpy as np
import pandas as pd

from labels import LABELS, LABEL2ID
import keyword_baseline

HERE = os.path.dirname(os.path.abspath(__file__))
PROC = os.path.join(HERE, "..", "data", "processed")
MODEL_DIR = os.path.join(HERE, "..", "model")
REPORTS = os.path.join(HERE, "..", "reports")

# Obfuscated / paraphrased examples the substring filter cannot catch but a
# language model should. (text, expected_label)
ADVERSARIAL = [
    ("давай про с е к с поговорим", "block"),     # spaced
    ("hey wanna see my nu_de pics", "block"),      # punctuation split
    ("ты полный придурок и неудачник", "warn"),    # insult, no keyword
    ("сен нағыз ақымақсың", "warn"),               # Kazakh insult, no keyword
    ("интимных услуг не предлагаю, ищу никах", "clean"),  # keyword present but clean
    ("сәлеметсіз бе, таныса аламыз ба?", "clean"),  # Kazakh clean
]


def _load_model():
    import torch
    from transformers import AutoModelForSequenceClassification, AutoTokenizer

    tok = AutoTokenizer.from_pretrained(MODEL_DIR)
    model = AutoModelForSequenceClassification.from_pretrained(MODEL_DIR).eval()

    def predict_proba(texts):
        probs = []
        with torch.no_grad():
            for i in range(0, len(texts), 32):
                batch = texts[i:i + 32]
                enc = tok(batch, truncation=True, max_length=128,
                          padding=True, return_tensors="pt")
                logits = model(**enc).logits
                probs.append(torch.softmax(logits, dim=-1).cpu().numpy())
        return np.vstack(probs)

    return predict_proba


def _metrics(y_true, y_pred, proba=None):
    from sklearn.metrics import classification_report, f1_score, accuracy_score
    out = {
        "accuracy": accuracy_score(y_true, y_pred),
        "macro_f1": f1_score(y_true, y_pred, average="macro",
                             labels=LABELS, zero_division=0),
        "per_class": classification_report(y_true, y_pred, labels=LABELS,
                                           output_dict=True, zero_division=0),
    }
    if proba is not None:
        try:
            from sklearn.metrics import roc_auc_score
            from sklearn.preprocessing import label_binarize
            yb = label_binarize([LABEL2ID[y] for y in y_true], classes=[0, 1, 2])
            out["roc_auc_ovr"] = roc_auc_score(yb, proba, average="macro", multi_class="ovr")
        except Exception as e:  # single-class test split, etc.
            out["roc_auc_ovr"] = None
            out["roc_auc_error"] = str(e)
    return out


def _confusion_png(y_true, y_pred, path):
    import matplotlib
    matplotlib.use("Agg")
    import matplotlib.pyplot as plt
    from sklearn.metrics import confusion_matrix

    cm = confusion_matrix(y_true, y_pred, labels=LABELS)
    fig, ax = plt.subplots(figsize=(4.5, 4))
    im = ax.imshow(cm, cmap="Blues")
    ax.set_xticks(range(len(LABELS)), LABELS)
    ax.set_yticks(range(len(LABELS)), LABELS)
    ax.set_xlabel("predicted"); ax.set_ylabel("true")
    ax.set_title("Moderation model — confusion matrix")
    for i in range(len(LABELS)):
        for j in range(len(LABELS)):
            ax.text(j, i, str(cm[i, j]), ha="center", va="center")
    fig.colorbar(im); fig.tight_layout(); fig.savefig(path, dpi=120)


def main():
    os.makedirs(REPORTS, exist_ok=True)
    predict_proba = _load_model()

    def model_pred(texts):
        proba = predict_proba(texts)
        labels = [LABELS[i] for i in proba.argmax(axis=1)]
        return labels, proba

    report = {}

    # 1. Held-out test set: model vs keyword baseline.
    test = pd.read_csv(os.path.join(PROC, "test.csv")).dropna(subset=["text", "label"])
    texts, y_true = test["text"].astype(str).tolist(), test["label"].tolist()
    y_model, proba = model_pred(texts)
    y_kw = [keyword_baseline.predict(t) for t in texts]

    report["test_model"] = _metrics(y_true, y_model, proba)
    report["test_keyword_baseline"] = _metrics(y_true, y_kw)
    _confusion_png(y_true, y_model, os.path.join(REPORTS, "confusion_matrix.png"))

    # 2. Kazakh zero-shot (the novelty), if a gold set exists.
    kk_path = os.path.join(PROC, "kk_gold.csv")
    if os.path.exists(kk_path):
        kk = pd.read_csv(kk_path).dropna(subset=["text", "label"])
        kt, ky = kk["text"].astype(str).tolist(), kk["label"].tolist()
        km, kp = model_pred(kt)
        report["kazakh_zero_shot_model"] = _metrics(ky, km, kp)
        report["kazakh_zero_shot_keyword"] = _metrics(ky, [keyword_baseline.predict(t) for t in kt])

    # 3. Adversarial / bypass set.
    at = [t for t, _ in ADVERSARIAL]
    ay = [y for _, y in ADVERSARIAL]
    am, _ = model_pred(at)
    ak = [keyword_baseline.predict(t) for t in at]
    report["adversarial"] = [
        {"text": t, "expected": e, "model": m, "keyword": k}
        for (t, e), m, k in zip(ADVERSARIAL, am, ak)
    ]

    with open(os.path.join(REPORTS, "metrics.json"), "w", encoding="utf-8") as f:
        json.dump(report, f, ensure_ascii=False, indent=2)

    # Markdown summary.
    lines = ["# Moderation evaluation\n"]
    lines.append("## Test set — ML model vs keyword baseline\n")
    lines.append("| metric | keyword baseline | **ML model** |")
    lines.append("|---|---|---|")
    lines.append(f"| accuracy | {report['test_keyword_baseline']['accuracy']:.3f} | "
                 f"**{report['test_model']['accuracy']:.3f}** |")
    lines.append(f"| macro-F1 | {report['test_keyword_baseline']['macro_f1']:.3f} | "
                 f"**{report['test_model']['macro_f1']:.3f}** |")
    auc = report["test_model"].get("roc_auc_ovr")
    lines.append(f"| ROC-AUC (ovr) | — | **{auc:.3f}** |" if auc else "| ROC-AUC (ovr) | — | n/a |")

    if "kazakh_zero_shot_model" in report:
        lines.append("\n## Kazakh (zero-shot) — trained on RU+EN, evaluated on Kazakh\n")
        lines.append(f"- ML model macro-F1: **{report['kazakh_zero_shot_model']['macro_f1']:.3f}**")
        lines.append(f"- keyword baseline macro-F1: {report['kazakh_zero_shot_keyword']['macro_f1']:.3f}")

    lines.append("\n## Adversarial / bypass set\n")
    lines.append("| text | expected | keyword | model |")
    lines.append("|---|---|---|---|")
    for r in report["adversarial"]:
        lines.append(f"| {r['text']} | {r['expected']} | {r['keyword']} | {r['model']} |")

    with open(os.path.join(REPORTS, "report.md"), "w", encoding="utf-8") as f:
        f.write("\n".join(lines) + "\n")

    print("Wrote reports/metrics.json, reports/report.md, reports/confusion_matrix.png")
    print(f"Test macro-F1: model={report['test_model']['macro_f1']:.3f} "
          f"keyword={report['test_keyword_baseline']['macro_f1']:.3f}")


if __name__ == "__main__":
    main()
