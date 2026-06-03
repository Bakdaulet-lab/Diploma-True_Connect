"""Evaluate the photo-verifier against a labeled image set.

Folder layout (you provide the images):
    data/approve/   real face, safe-for-work  → expected decision "approve"
    data/reject/    no face OR explicit (NSFW) → expected decision "reject"

Runs each image through the live service (POST /verify) and reports
precision / recall / accuracy treating "reject" (catching a bad photo) as the
positive class. Writes reports/metrics.json and reports/report.md.

Usage:
    # start the service first (docker compose up photo-verifier), then:
    python evaluate.py --url http://localhost:8000 --data ./data
"""
import argparse
import json
import os

import requests

HERE = os.path.dirname(os.path.abspath(__file__))
REPORTS = os.path.join(HERE, "reports")
EXTS = (".jpg", ".jpeg", ".png", ".webp")


def classify_dir(url, folder):
    decisions = []
    if not os.path.isdir(folder):
        return decisions
    for name in sorted(os.listdir(folder)):
        if not name.lower().endswith(EXTS):
            continue
        with open(os.path.join(folder, name), "rb") as f:
            resp = requests.post(f"{url}/verify",
                                 files={"image": (name, f, "application/octet-stream")},
                                 timeout=30)
        resp.raise_for_status()
        decisions.append((name, resp.json().get("decision")))
    return decisions


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--url", default=os.getenv("PHOTO_VERIFIER_URL", "http://localhost:8000"))
    ap.add_argument("--data", default=os.path.join(HERE, "data"))
    args = ap.parse_args()
    os.makedirs(REPORTS, exist_ok=True)

    approve = classify_dir(args.url, os.path.join(args.data, "approve"))
    reject = classify_dir(args.url, os.path.join(args.data, "reject"))
    if not approve and not reject:
        raise SystemExit("No images. Put files in data/approve and data/reject.")

    # Positive class = "reject" (correctly catching a bad photo).
    tp = sum(1 for _, d in reject if d == "reject")
    fn = sum(1 for _, d in reject if d != "reject")
    tn = sum(1 for _, d in approve if d == "approve")
    fp = sum(1 for _, d in approve if d != "approve")

    precision = tp / (tp + fp) if (tp + fp) else 0.0
    recall = tp / (tp + fn) if (tp + fn) else 0.0
    total = tp + fp + tn + fn
    accuracy = (tp + tn) / total if total else 0.0
    f1 = (2 * precision * recall / (precision + recall)) if (precision + recall) else 0.0

    metrics = {
        "counts": {"tp": tp, "fp": fp, "tn": tn, "fn": fn, "total": total},
        "precision": precision, "recall": recall, "f1": f1, "accuracy": accuracy,
    }
    with open(os.path.join(REPORTS, "metrics.json"), "w", encoding="utf-8") as f:
        json.dump(metrics, f, ensure_ascii=False, indent=2)

    lines = [
        "# Photo verification evaluation\n",
        f"- images: {total} (reject={tp+fn}, approve={tn+fp})",
        f"- **precision={precision:.3f}  recall={recall:.3f}  F1={f1:.3f}  accuracy={accuracy:.3f}**",
        "",
        "| | predicted reject | predicted approve |",
        "|---|---|---|",
        f"| actual reject | {tp} (TP) | {fn} (FN) |",
        f"| actual approve | {fp} (FP) | {tn} (TN) |",
    ]
    with open(os.path.join(REPORTS, "report.md"), "w", encoding="utf-8") as f:
        f.write("\n".join(lines) + "\n")

    print(f"precision={precision:.3f} recall={recall:.3f} F1={f1:.3f} accuracy={accuracy:.3f}")
    print("Wrote reports/metrics.json and reports/report.md")


if __name__ == "__main__":
    main()
