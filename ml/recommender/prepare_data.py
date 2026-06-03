"""Build the recommender training table from swipes + profiles.

Inputs (in data/raw/):
  * swipes.csv    columns: viewer_id, target_id, label   (1 = liked, 0 = passed)
  * profiles.csv  columns: user_id, age, lat, lon, city, niyyah, madhab,
                           languages, trust_score, is_kyc
  (See README for the exact COPY ... SQL to export these from Postgres.)

Output (data/processed/):
  * train.csv / test.csv — viewer_id, target_id, <FEATURES...>, label

Or generate a synthetic dataset to run the whole pipeline without real data:
  python prepare_data.py --synthesize 4000
"""
import argparse
import os
import random

import pandas as pd
from sklearn.model_selection import train_test_split

from features import FEATURES, compute

HERE = os.path.dirname(os.path.abspath(__file__))
RAW = os.path.join(HERE, "data", "raw")
OUT = os.path.join(HERE, "data", "processed")

CITIES = ["Almaty", "Astana", "Shymkent", "Karaganda"]
NIYYAHS = ["nikah_year", "serious_marriage", "friendship"]
MADHABS = ["hanafi", "shafii", "maliki", "hanbali", "none"]
LANGS = ["kk", "ru", "en", "tr", "ar"]


def parse_langs(v):
    if isinstance(v, list):
        return v
    s = str(v or "").strip().strip("{}[]")
    return [x.strip().strip('"') for x in s.split(",") if x.strip()]


def profiles_to_dict(df: pd.DataFrame) -> dict:
    out = {}
    for _, r in df.iterrows():
        out[str(r["user_id"])] = {
            "age": int(r["age"]) if pd.notna(r.get("age")) else None,
            "lat": float(r["lat"]) if pd.notna(r.get("lat")) else None,
            "lon": float(r["lon"]) if pd.notna(r.get("lon")) else None,
            "city": r.get("city") or "",
            "niyyah": r.get("niyyah") or "",
            "madhab": r.get("madhab") or "",
            "languages": parse_langs(r.get("languages")),
            "trust": int(r["trust_score"]) if pd.notna(r.get("trust_score")) else 0,
            "kyc": bool(r.get("is_kyc")),
        }
    return out


def build_rows(swipes: pd.DataFrame, profiles: dict) -> pd.DataFrame:
    records = []
    for _, s in swipes.iterrows():
        v, c = profiles.get(str(s["viewer_id"])), profiles.get(str(s["target_id"]))
        if not v or not c:
            continue
        feat = compute(v, c)
        feat.update({"viewer_id": s["viewer_id"], "target_id": s["target_id"],
                     "label": int(s["label"])})
        records.append(feat)
    cols = ["viewer_id", "target_id"] + FEATURES + ["label"]
    return pd.DataFrame.from_records(records)[cols]


def synthesize(n: int) -> pd.DataFrame:
    """Generate synthetic profiles + swipes with a latent preference function,
    so the learned model has real signal to recover (for pipeline demos)."""
    random.seed(42)
    n_users = max(40, n // 20)
    profiles = {}
    for i in range(n_users):
        uid = f"u{i}"
        profiles[uid] = {
            "age": random.randint(18, 45),
            "lat": 43 + random.uniform(-3, 8), "lon": 71 + random.uniform(-2, 7),
            "city": random.choice(CITIES),
            "niyyah": random.choice(NIYYAHS),
            "madhab": random.choice(MADHABS),
            "languages": random.sample(LANGS, k=random.randint(1, 3)),
            "trust": random.randint(20, 95),
            "kyc": random.random() < 0.4,
        }
    ids = list(profiles)

    def like_prob(f):  # latent "truth" the model should learn
        z = (-0.08 * f["age_gap"] - 0.004 * f["distance_km"]
             + 0.7 * f["same_city"] + 1.1 * f["niyyah_match"]
             + 0.5 * f["madhab_match"] + 0.3 * f["language_overlap"]
             + 0.02 * (f["cand_trust"] - 50) + 0.4 * f["cand_kyc"] - 0.3)
        return 1 / (1 + pow(2.718281828, -z))

    rows = []
    for _ in range(n):
        v, c = random.choice(ids), random.choice(ids)
        if v == c:
            continue
        f = compute(profiles[v], profiles[c])
        label = 1 if random.random() < like_prob(f) else 0
        rows.append({"viewer_id": v, "target_id": c, "label": label})
    return build_rows(pd.DataFrame(rows), profiles)


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--synthesize", type=int, default=0)
    args = ap.parse_args()
    os.makedirs(OUT, exist_ok=True)

    if args.synthesize > 0:
        df = synthesize(args.synthesize)
    else:
        swipes = pd.read_csv(os.path.join(RAW, "swipes.csv"))
        profiles = profiles_to_dict(pd.read_csv(os.path.join(RAW, "profiles.csv")))
        df = build_rows(swipes, profiles)

    df = df.dropna(subset=["label"])
    if df.empty:
        raise SystemExit("No rows. Provide data/raw/{swipes,profiles}.csv or --synthesize N.")

    print("Label balance:\n", df["label"].value_counts())
    strat = df["label"] if df["label"].nunique() > 1 else None
    train, test = train_test_split(df, test_size=0.2, random_state=42, stratify=strat)
    train.to_csv(os.path.join(OUT, "train.csv"), index=False)
    test.to_csv(os.path.join(OUT, "test.csv"), index=False)
    print(f"wrote train={len(train)} test={len(test)} → {OUT}")


if __name__ == "__main__":
    main()
