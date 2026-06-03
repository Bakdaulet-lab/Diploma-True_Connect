# Model Card — TrueConnect Match Recommender

## Overview
- **Task**: learning-to-rank swipe candidates by predicted **P(mutual like)**.
- **Model**: logistic regression on standardized pairwise features (interpretable,
  exports to `model.json`, scored inline in Go — no ML runtime in production).
- **Replaces**: static `created_at / last_login / trust_score` ordering.

## Features (viewer × candidate)
`age_gap`, `age_known`, `distance_km`, `distance_known`, `same_city`,
`niyyah_match`, `niyyah_compatible`, `madhab_match`, `language_overlap`,
`cand_trust`, `cand_kyc`. Single source of truth: `features.py` ↔
`internal/pkg/recommender/recommender.go` (kept in lockstep).

## Training data
- **Positives**: likes / mutual matches (`social.matches`).
- **Negatives**: passes (`social.swipe_rejections`).
- Features computed by joining the two users' profiles. Export SQL in README.
- A synthetic generator (`prepare_data.py --synthesize`) reproduces a known
  latent preference function to validate the pipeline before real data exists.

## Evaluation (`reports/report.md`, from `evaluate.py`)
- **Classification**: ROC-AUC, accuracy of P(like).
- **Ranking, per viewer**: precision@k, NDCG@k, MAP — **model vs the trust-sort
  baseline** (the current heuristic). This is the headline comparison.

> Fill after running `evaluate.py`:
>
> | metric | baseline (trust sort) | model |
> |---|---|---|
> | precision@5 | _ | _ |
> | NDCG@5 | _ | _ |
> | MAP | _ | _ |
> | ROC-AUC | — | _ |

## Limitations & ethics
- **Cold start**: little real swipe data → weak empirical numbers initially;
  synthetic results show methodology, not production performance.
- **Feedback loop**: ranking on past likes can amplify popularity/homophily.
  Keep hard filters (niyyah/halal/gender, distance, age) authoritative; the model
  only re-orders within the already-eligible pool.
- **Fairness**: monitor that `trust`/`kyc` weighting doesn't unfairly bury new
  users; consider exploration (occasionally surface unranked candidates).
- Interpretability: LR weights are reported by `train.py` for inspection.

## Serving
- Exported `model.json` → `RECOMMENDER_MODEL_PATH`. The Go API loads it at
  startup; absent/invalid → falls back to default ordering (no outage).
