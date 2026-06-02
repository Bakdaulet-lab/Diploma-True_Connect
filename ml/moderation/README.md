# TrueConnect — ML Content Moderation (Kazakh / Russian / English)

An XLM-RoBERTa classifier that replaces the brittle keyword filter
(`internal/pkg/halalfilter`) with a multilingual model. Served as a FastAPI
microservice and called from the Go backend over HTTP; the Go side falls back to
the keyword filter if this service is unavailable.

Three output classes:

| label | meaning | action in app |
|---|---|---|
| `clean` | allowed | delivered |
| `warn`  | toxic (insult/harassment/spam) | delivered + flagged, −trust |
| `block` | explicit/sexual, threats, solicitation | rejected |

## Layout

```
ml/moderation/
  training/   prepare_data.py · train.py · evaluate.py · labels.py · keyword_baseline.py
  service/    app.py · Dockerfile · requirements.txt        # FastAPI inference
  data/       raw/ (downloads)  processed/ (splits)         # git-ignored
  model/      fine-tuned XLM-R                               # git-ignored
  reports/    metrics.json · report.md · confusion_matrix.png · MODEL_CARD.md
```

## Quickstart (diploma workflow)

```bash
cd ml/moderation
python -m venv .venv && source .venv/bin/activate    # Windows: .venv\Scripts\activate
pip install -r requirements.txt

# 1. Data → splits
#    Put public datasets in data/raw/ (Jigsaw EN, Russian Toxic Comments, …),
#    or smoke-test the whole pipeline with synthetic data:
python training/prepare_data.py --synthesize 600
#    (add a hand-labeled Kazakh gold set for the zero-shot result:)
#    python training/prepare_data.py --kk-gold path/to/kk_gold.csv

# 2. Fine-tune XLM-RoBERTa  → ml/moderation/model/
python training/train.py --epochs 3

# 3. Evaluate → reports/ (model vs keyword baseline, Kazakh zero-shot, adversarial)
python training/evaluate.py
```

## Run the service

```bash
# With a trained model in ./model:
MODEL_DIR=./model uvicorn service.app:app --host 0.0.0.0 --port 8000
# Without a model (lexicon stub, for contract testing):
MODERATION_STUB=1 uvicorn service.app:app --port 8000

curl -s localhost:8000/health
curl -s -X POST localhost:8000/moderate -H 'Content-Type: application/json' \
     -d '{"text":"давай про секс"}'
```

In docker-compose the service is built and started automatically; the API reads
`MODERATION_SERVICE_URL=http://moderation:8000`.

## Datasets (suggested, public)
- **English**: Jigsaw Toxic Comment Classification (Wikipedia, multi-label).
- **Russian**: "Russian Language Toxic Comments" (Kaggle).
- **Kazakh**: machine-translate a balanced RU/EN subset (silver) for training +
  a small hand-labeled gold set for honest zero-shot evaluation.

The label mapping from each source to `{clean,warn,block}` lives in
`training/labels.py` and is documented in `reports/MODEL_CARD.md`.
