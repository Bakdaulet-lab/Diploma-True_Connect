"""Evaluate the ranking model vs the current heuristic baseline.

Loads ../recommender/model.json and scores test.csv with the SAME math the Go
service uses (standardize → logistic), then reports:
  * Classification: ROC-AUC, accuracy of P(like).
  * Ranking (grouped per viewer): precision@k, NDCG@k, MAP —
    MODEL (P(like)) vs BASELINE (sort by candidate trust, the current heuristic).

Writes reports/metrics.json and reports/report.md.

Usage:  python evaluate.py [--k 5]
"""
import argparse
import json
import math
import os

import numpy as np
import pandas as pd

from features import FEATURES

HERE = os.path.dirname(os.path.abspath(__file__))
PROC = os.path.join(HERE, "data", "processed")
MODEL = os.path.join(HERE, "model.json")
REPORTS = os.path.join(HERE, "reports")


def load_model():
    with open(MODEL, encoding="utf-8") as f:
        return json.load(f)


def score(df, m):
    names = m["feature_names"]
    X = df[names].astype(float).values
    mean = np.array(m["mean"]); scale = np.array(m["scale"]); scale[scale == 0] = 1
    z = (X - mean) / scale @ np.array(m["coef"]) + m["intercept"]
    return 1.0 / (1.0 + np.exp(-z))


def dcg(rels):
    return sum((2 ** r - 1) / math.log2(i + 2) for i, r in enumerate(rels))


def ndcg_at_k(order_labels, k):
    ideal = sorted(order_labels, reverse=True)[:k]
    idcg = dcg(ideal)
    return dcg(order_labels[:k]) / idcg if idcg > 0 else 0.0


def average_precision(labels):
    hits, ap = 0, 0.0
    for i, rel in enumerate(labels):
        if rel:
            hits += 1
            ap += hits / (i + 1)
    return ap / hits if hits else 0.0


def ranking_metrics(df, score_col, k):
    p_at_k, ndcgs, aps, n = [], [], [], 0
    for _, g in df.groupby("viewer_id"):
        if len(g) < 2 or g["label"].sum() == 0:
            continue  # need ranking choice + at least one positive
        ordered = g.sort_values(score_col, ascending=False)["label"].tolist()
        kk = min(k, len(ordered))
        p_at_k.append(sum(ordered[:kk]) / kk)
        ndcgs.append(ndcg_at_k(ordered, kk))
        aps.append(average_precision(ordered))
        n += 1
    return {
        "viewers_evaluated": n,
        f"precision@{k}": float(np.mean(p_at_k)) if p_at_k else None,
        f"ndcg@{k}": float(np.mean(ndcgs)) if ndcgs else None,
        "map": float(np.mean(aps)) if aps else None,
    }


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--k", type=int, default=5)
    args = ap.parse_args()
    os.makedirs(REPORTS, exist_ok=True)

    m = load_model()
    df = pd.read_csv(os.path.join(PROC, "test.csv"))
    df["model_score"] = score(df, m)

    report = {"classification": {}, "ranking_model": {}, "ranking_baseline": {}}

    # Classification metrics.
    try:
        from sklearn.metrics import roc_auc_score, accuracy_score
        report["classification"]["roc_auc"] = float(roc_auc_score(df["label"], df["model_score"]))
        report["classification"]["accuracy"] = float(
            accuracy_score(df["label"], (df["model_score"] >= 0.5).astype(int)))
    except Exception as e:
        report["classification"]["error"] = str(e)

    # Ranking: model vs baseline (candidate trust = current heuristic proxy).
    report["ranking_model"] = ranking_metrics(df, "model_score", args.k)
    report["ranking_baseline"] = ranking_metrics(df, "cand_trust", args.k)

    with open(os.path.join(REPORTS, "metrics.json"), "w", encoding="utf-8") as f:
        json.dump(report, f, ensure_ascii=False, indent=2)

    k = args.k
    rm, rb = report["ranking_model"], report["ranking_baseline"]
    lines = [
        "# Recommender evaluation\n",
        f"Classification: ROC-AUC = **{report['classification'].get('roc_auc', float('nan')):.3f}**, "
        f"accuracy = {report['classification'].get('accuracy', float('nan')):.3f}\n",
        f"Ranking (per viewer, {rm['viewers_evaluated']} viewers evaluated):\n",
        "| metric | baseline (trust sort) | **model** |",
        "|---|---|---|",
        f"| precision@{k} | {rb[f'precision@{k}']:.3f} | **{rm[f'precision@{k}']:.3f}** |",
        f"| ndcg@{k} | {rb[f'ndcg@{k}']:.3f} | **{rm[f'ndcg@{k}']:.3f}** |",
        f"| MAP | {rb['map']:.3f} | **{rm['map']:.3f}** |",
    ]
    with open(os.path.join(REPORTS, "report.md"), "w", encoding="utf-8") as f:
        f.write("\n".join(lines) + "\n")

    print("Wrote reports/metrics.json and reports/report.md")
    print(f"AUC={report['classification'].get('roc_auc')}  "
          f"model P@{k}={rm[f'precision@{k}']}  baseline P@{k}={rb[f'precision@{k}']}")


if __name__ == "__main__":
    main()
