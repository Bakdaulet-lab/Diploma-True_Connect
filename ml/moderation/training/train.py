"""Fine-tune XLM-RoBERTa for 3-class content moderation (clean/warn/block).

XLM-R is multilingual (100+ languages incl. Kazakh & Russian), which is what
enables the cross-lingual story: train on RU+EN, generalize to Kazakh.

Usage:
    python train.py                       # uses data/processed/{train,val}.csv
    python train.py --epochs 3 --model xlm-roberta-base --batch 16

Outputs the fine-tuned model + tokenizer to ../model/ (loaded by the service).
"""
import argparse
import os

import numpy as np
import pandas as pd
from datasets import Dataset
from sklearn.metrics import f1_score, accuracy_score
from transformers import (
    AutoModelForSequenceClassification,
    AutoTokenizer,
    DataCollatorWithPadding,
    Trainer,
    TrainingArguments,
)

from labels import LABELS, LABEL2ID, ID2LABEL

HERE = os.path.dirname(os.path.abspath(__file__))
PROC = os.path.join(HERE, "..", "data", "processed")
MODEL_OUT = os.path.join(HERE, "..", "model")


def load_split(name: str) -> Dataset:
    df = pd.read_csv(os.path.join(PROC, f"{name}.csv"))
    df = df.dropna(subset=["text", "label"])
    df["labels"] = df["label"].map(LABEL2ID)
    return Dataset.from_pandas(df[["text", "labels"]], preserve_index=False)


def compute_metrics(eval_pred):
    logits, labels = eval_pred
    preds = np.argmax(logits, axis=-1)
    return {
        "accuracy": accuracy_score(labels, preds),
        "macro_f1": f1_score(labels, preds, average="macro"),
    }


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--model", default="xlm-roberta-base")
    ap.add_argument("--epochs", type=float, default=3)
    ap.add_argument("--batch", type=int, default=16)
    ap.add_argument("--lr", type=float, default=2e-5)
    ap.add_argument("--max-len", type=int, default=128)
    args = ap.parse_args()

    tokenizer = AutoTokenizer.from_pretrained(args.model)
    model = AutoModelForSequenceClassification.from_pretrained(
        args.model,
        num_labels=len(LABELS),
        id2label=ID2LABEL,
        label2id=LABEL2ID,
    )

    def tok(batch):
        return tokenizer(batch["text"], truncation=True, max_length=args.max_len)

    train_ds = load_split("train").map(tok, batched=True)
    val_ds = load_split("val").map(tok, batched=True)

    targs = TrainingArguments(
        output_dir=os.path.join(HERE, "..", "model", "_checkpoints"),
        num_train_epochs=args.epochs,
        per_device_train_batch_size=args.batch,
        per_device_eval_batch_size=args.batch,
        learning_rate=args.lr,
        eval_strategy="epoch",
        save_strategy="epoch",
        load_best_model_at_end=True,
        metric_for_best_model="macro_f1",
        logging_steps=50,
        report_to=[],
    )

    trainer = Trainer(
        model=model,
        args=targs,
        train_dataset=train_ds,
        eval_dataset=val_ds,
        tokenizer=tokenizer,
        data_collator=DataCollatorWithPadding(tokenizer),
        compute_metrics=compute_metrics,
    )

    trainer.train()
    print("Validation:", trainer.evaluate())

    os.makedirs(MODEL_OUT, exist_ok=True)
    trainer.save_model(MODEL_OUT)
    tokenizer.save_pretrained(MODEL_OUT)
    print(f"\nSaved model + tokenizer to {MODEL_OUT}")


if __name__ == "__main__":
    main()
