"""Shared label scheme and dataset→label mapping for content moderation.

Three collapsed classes consumed by the Go API and the FastAPI service:

    clean (0)  — allowed
    warn  (1)  — allowed but flagged toxic (insult/harassment/spam) → trust penalty
    block (2)  — rejected outright (sexual/explicit, threats, solicitation)

The mapping rules below are documented in reports/MODEL_CARD.md and are the
single source of truth used by prepare_data.py.
"""

LABELS = ["clean", "warn", "block"]
LABEL2ID = {name: i for i, name in enumerate(LABELS)}
ID2LABEL = {i: name for i, name in enumerate(LABELS)}


def map_jigsaw(row) -> str:
    """Map a Jigsaw multi-label row (English) to one collapsed class.

    Jigsaw columns: toxic, severe_toxic, obscene, threat, insult, identity_hate.
    """
    def f(col):
        try:
            return float(row.get(col, 0) or 0) >= 0.5
        except (TypeError, ValueError):
            return False

    if f("threat") or f("severe_toxic") or f("obscene"):
        return "block"
    if f("toxic") or f("insult") or f("identity_hate"):
        return "warn"
    return "clean"


def map_binary_toxic(is_toxic) -> str:
    """Map a binary toxic flag (e.g. the Russian toxic-comments dataset).

    Binary datasets cannot distinguish 'block' from 'warn', so toxic→warn.
    Explicit/'block' examples come from Jigsaw + the silver/keyword sources.
    """
    try:
        return "warn" if float(is_toxic) >= 0.5 else "clean"
    except (TypeError, ValueError):
        return "clean"
