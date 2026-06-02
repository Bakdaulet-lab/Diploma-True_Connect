# Model Card — TrueConnect Content Moderation

## Overview
- **Task**: single-label text classification into `clean` / `warn` / `block`.
- **Base model**: `xlm-roberta-base` (multilingual, 100+ languages incl. Kazakh & Russian).
- **Purpose**: moderate chat messages, posts, and comments in a halal matrimonial
  app; replaces a substring keyword filter that is trivially bypassed.
- **Languages**: Russian, English (training), Kazakh (zero-shot / silver-augmented).

## Intended use & integration
- Called by the Go backend via `POST /moderate`. `block` → reject; `warn` →
  deliver + flag toxic (trust penalty); `clean` → allow.
- **Fail-open by design**: if the service is unavailable the backend falls back
  to the keyword filter, so moderation degrades but never disappears.

## Label scheme (mapping from public datasets)
Defined in `training/labels.py`:
- **block** ← Jigsaw `threat | severe_toxic | obscene`; explicit/solicitation terms.
- **warn**  ← Jigsaw `toxic | insult | identity_hate`; binary-toxic = 1.
- **clean** ← everything else.

## Training data
- English: Jigsaw Toxic Comment Classification (Wikipedia comments).
- Russian: Russian Language Toxic Comments.
- Kazakh: machine-translated silver set (optional) + hand-labeled **gold eval** set.
- Splits: stratified 80/10/10 (`prepare_data.py`).

## Evaluation (see `report.md` / `metrics.json`)
Reported by `evaluate.py`:
- Per-class precision/recall/F1, macro-F1, accuracy, ROC-AUC (one-vs-rest).
- **Baseline comparison** vs the keyword filter on the same test set.
- **Kazakh zero-shot**: trained on RU+EN, evaluated on the Kazakh gold set
  (the project's novel contribution for an under-resourced language).
- **Adversarial/bypass set**: spaced/obfuscated/paraphrased inputs that defeat
  the keyword filter; demonstrates the model's robustness.

> Fill the numbers below after running `evaluate.py`:
>
> | metric | keyword baseline | ML model |
> |---|---|---|
> | test macro-F1 | _ | _ |
> | test ROC-AUC | — | _ |
> | Kazakh zero-shot macro-F1 | _ | _ |

## Limitations & ethics
- Trained largely on RU/EN; Kazakh performance depends on silver augmentation
  and the gold set size — report it honestly.
- Toxicity labels carry annotator/cultural bias; thresholds (`block` vs `warn`)
  are a product/safety choice, not ground truth.
- Not a substitute for human review of appeals; false positives can wrongly
  penalize trust score — keep the trust penalty modest and appealable.

## Latency
- XLM-R base on CPU: ~tens–hundreds of ms for short messages. Report measured
  ms/msg from `evaluate.py`. Optional: distill or export to ONNX for speedups.
