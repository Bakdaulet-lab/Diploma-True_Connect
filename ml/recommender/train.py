"""Train the match-ranking model and export model.json for the Go scorer.

Logistic regression on standardized pairwise features. LR is chosen because it
is interpretable (per-feature weights are defensible in a thesis) and exports to
a tiny JSON the Go service reads with no ML runtime.

Output: ../recommender/model.json  (feature_names, mean, scale, coef, intercept)
        — point the API at it via RECOMMENDER_MODEL_PATH.

Usage:  python train.py
"""
import json
import os

import pandas as pd
from sklearn.linear_model import LogisticRegression
from sklearn.preprocessing import StandardScaler

from features import FEATURES

HERE = os.path.dirname(os.path.abspath(__file__))
PROC = os.path.join(HERE, "data", "processed")
MODEL_OUT = os.path.join(HERE, "model.json")


def main():
    df = pd.read_csv(os.path.join(PROC, "train.csv"))
    X = df[FEATURES].astype(float).values
    y = df["label"].astype(int).values

    scaler = StandardScaler().fit(X)
    Xs = scaler.transform(X)

    clf = LogisticRegression(max_iter=1000, class_weight="balanced")
    clf.fit(Xs, y)

    model = {
        "feature_names": FEATURES,
        "mean": scaler.mean_.tolist(),
        "scale": scaler.scale_.tolist(),
        "coef": clf.coef_[0].tolist(),
        "intercept": float(clf.intercept_[0]),
    }
    with open(MODEL_OUT, "w", encoding="utf-8") as f:
        json.dump(model, f, indent=2)

    print(f"Saved {MODEL_OUT}")
    print("\nFeature weights (standardized — sign/magnitude = influence):")
    for name, w in sorted(zip(FEATURES, model["coef"]), key=lambda t: -abs(t[1])):
        print(f"  {name:<18} {w:+.3f}")


if __name__ == "__main__":
    main()
