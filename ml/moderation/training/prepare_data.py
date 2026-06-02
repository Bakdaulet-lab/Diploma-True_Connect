"""Build train/val/test splits for the moderation classifier.

Reads raw public datasets from ``data/raw/`` (whatever you downloaded), maps
each example to one of {clean, warn, block} using the rules in ``labels.py``,
combines, de-duplicates, and writes stratified splits to ``data/processed/``.

Supported raw formats (auto-detected by columns):
  * Jigsaw Toxic Comment (EN): columns ``comment_text`` + the 6 label columns.
  * Binary toxic (e.g. Russian Toxic Comments): a text column
    (``comment``/``text``/``comment_text``) + a ``toxic`` column.
  * Pre-labeled: any CSV with ``text`` and ``label`` (label already in
    {clean,warn,block}) — e.g. your hand-labeled Kazakh gold set.

Usage:
    python prepare_data.py                      # process everything in data/raw
    python prepare_data.py --synthesize 600     # no downloads → tiny synthetic set
                                                # (lets the whole pipeline run/smoke-test)
    python prepare_data.py --kk-gold kk_gold.csv  # also copy a Kazakh gold test set
"""
import argparse
import os
import random

import pandas as pd
from sklearn.model_selection import train_test_split

from labels import LABELS, map_jigsaw, map_binary_toxic

HERE = os.path.dirname(os.path.abspath(__file__))
RAW_DIR = os.path.join(HERE, "..", "data", "raw")
OUT_DIR = os.path.join(HERE, "..", "data", "processed")

JIGSAW_COLS = ["toxic", "severe_toxic", "obscene", "threat", "insult", "identity_hate"]
TEXT_COLS = ["text", "comment_text", "comment", "message"]


def _text_col(df: pd.DataFrame) -> str | None:
    for c in TEXT_COLS:
        if c in df.columns:
            return c
    return None


def load_raw() -> pd.DataFrame:
    """Load and label every CSV found in data/raw/."""
    frames = []
    if not os.path.isdir(RAW_DIR):
        return pd.DataFrame(columns=["text", "label"])

    for fname in sorted(os.listdir(RAW_DIR)):
        if not fname.endswith(".csv"):
            continue
        path = os.path.join(RAW_DIR, fname)
        df = pd.read_csv(path)
        tcol = _text_col(df)
        if tcol is None:
            print(f"  skip {fname}: no text column")
            continue

        if "label" in df.columns and set(df["label"].dropna().unique()) <= set(LABELS):
            out = df[[tcol, "label"]].rename(columns={tcol: "text"})
        elif all(c in df.columns for c in JIGSAW_COLS):
            out = pd.DataFrame({"text": df[tcol], "label": df.apply(map_jigsaw, axis=1)})
        elif "toxic" in df.columns:
            out = pd.DataFrame({"text": df[tcol], "label": df["toxic"].map(map_binary_toxic)})
        else:
            print(f"  skip {fname}: unrecognized schema {list(df.columns)[:6]}")
            continue

        frames.append(out)
        print(f"  loaded {fname}: {len(out)} rows")

    if not frames:
        return pd.DataFrame(columns=["text", "label"])
    return pd.concat(frames, ignore_index=True)


def synthesize(n: int) -> pd.DataFrame:
    """Generate a tiny balanced synthetic set so the pipeline runs without downloads."""
    clean = ["сәлеметсіз бе", "қалыңыз қалай", "никах туралы сөйлесейік",
             "let's talk about our families", "приятно познакомиться",
             "сіз қай қаладансыз", "мақсатым — отбасы құру"]
    warn = ["сен ақымақсың", "ты дурак и мошенник", "you are an idiot",
            "это какой-то scam", "harassment не прекращается"]
    block = ["давай про секс", "porn link here", "интим услуги escort",
             "хочешь nude фото", "проституция объявление"]
    rows = []
    per = max(1, n // 3)
    for label, pool in [("clean", clean), ("warn", warn), ("block", block)]:
        for _ in range(per):
            rows.append({"text": random.choice(pool), "label": label})
    random.shuffle(rows)
    return pd.DataFrame(rows)


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--synthesize", type=int, default=0,
                    help="generate N synthetic rows instead of reading data/raw")
    ap.add_argument("--kk-gold", type=str, default="",
                    help="path to a hand-labeled Kazakh CSV (text,label) → saved as kk_gold.csv")
    args = ap.parse_args()

    os.makedirs(OUT_DIR, exist_ok=True)
    random.seed(42)

    df = synthesize(args.synthesize) if args.synthesize > 0 else load_raw()
    df = df.dropna(subset=["text"]).copy()
    df["text"] = df["text"].astype(str).str.strip()
    df = df[df["text"].str.len() > 0]
    df = df[df["label"].isin(LABELS)].drop_duplicates(subset=["text"])

    if df.empty:
        raise SystemExit(
            "No data. Put CSVs in data/raw/ or run with --synthesize 600.")

    print("\nClass distribution:")
    print(df["label"].value_counts())

    train, tmp = train_test_split(df, test_size=0.2, stratify=df["label"], random_state=42)
    val, test = train_test_split(tmp, test_size=0.5, stratify=tmp["label"], random_state=42)

    for name, part in [("train", train), ("val", val), ("test", test)]:
        out = os.path.join(OUT_DIR, f"{name}.csv")
        part.to_csv(out, index=False)
        print(f"wrote {out}: {len(part)} rows")

    if args.kk_gold and os.path.exists(args.kk_gold):
        kk = pd.read_csv(args.kk_gold)
        kk.to_csv(os.path.join(OUT_DIR, "kk_gold.csv"), index=False)
        print(f"wrote kk_gold.csv: {len(kk)} rows (Kazakh zero-shot eval set)")


if __name__ == "__main__":
    main()
