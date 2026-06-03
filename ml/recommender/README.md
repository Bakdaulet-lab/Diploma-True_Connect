# TrueConnect — Match Recommender (learning-to-rank)

Personalized candidate ranking for the swipe feed. Replaces the static
`ORDER BY created_at DESC, last_login DESC, trust_score DESC` in
`internal/service/matching_service.go` with a model that predicts **P(mutual
like)** from pairwise features and re-orders each batch.

**Served inline in Go** — training happens here (Python), the model is exported
as a tiny `model.json` (weights only), and `internal/pkg/recommender` scores
candidates with no runtime ML service (sub-ms per candidate). If the model is
absent the feed silently keeps the old ordering.

## Features (pairwise viewer × candidate)
Defined in `features.py` and mirrored byte-for-byte in
`internal/pkg/recommender/recommender.go`:

`age_gap`, `age_known`, `distance_km` (haversine), `distance_known`,
`same_city`, `niyyah_match`, `niyyah_compatible`, `madhab_match`,
`language_overlap`, `cand_trust`, `cand_kyc`.

> Changing a feature means editing **both** `features.py` and `recommender.go`.

## Quickstart

```bash
cd ml/recommender
python -m venv .venv && source .venv/bin/activate   # Windows: .venv\Scripts\activate
pip install -r requirements.txt

# Smoke-test the whole pipeline with synthetic data (no DB needed):
python prepare_data.py --synthesize 4000
python train.py        # → model.json (+ prints feature weights)
python evaluate.py     # → reports/ (model vs trust-sort baseline: P@k, NDCG, MAP, AUC)
```

Enable it in the API:
```bash
export RECOMMENDER_MODEL_PATH=$(pwd)/model.json   # then restart the API
```

## Using real data
Export from Postgres, then run `prepare_data.py` (no `--synthesize`):

```sql
-- data/raw/swipes.csv  (1 = liked, 0 = passed)
COPY (
  SELECT user_a_id AS viewer_id, user_b_id AS target_id, 1 AS label
    FROM social.matches WHERE user_a_liked
  UNION ALL
  SELECT user_b_id, user_a_id, 1 FROM social.matches WHERE user_b_liked
  UNION ALL
  SELECT user_id, target_id, 0 FROM social.swipe_rejections
) TO '/tmp/swipes.csv' CSV HEADER;

-- data/raw/profiles.csv
COPY (
  SELECT p.user_id,
         EXTRACT(year FROM AGE(p.birth_date))::int AS age,
         ST_Y(p.location::geometry) AS lat, ST_X(p.location::geometry) AS lon,
         p.city, p.niyyah::text AS niyyah, p.madhab::text AS madhab,
         p.languages, u.trust_score,
         (u.verification_level IN ('id_verified','photo_verified')) AS is_kyc
    FROM social.profiles p JOIN social.users u ON u.id = p.user_id
) TO '/tmp/profiles.csv' CSV HEADER;
```

> **Cold start (be honest in the defense):** with few real swipes the metrics are
> weak; the synthetic run demonstrates the *methodology* (the model recovers a
> known latent preference function and beats the trust-sort baseline). As real
> swipe volume grows, retrain to get genuine uplift.
