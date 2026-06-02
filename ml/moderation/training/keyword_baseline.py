"""Keyword baseline — a faithful Python copy of the Go keyword filter
(internal/pkg/halalfilter/filter.go).

Used by evaluate.py as the BASELINE the ML model is compared against. The whole
point of the diploma experiment is to show this substring approach is brittle
(obfuscation/paraphrase bypass it) while the ML model generalizes.
"""

BLOCKED_WORDS = [
    # explicit sexual content
    "секс", "порно", "интим", "голая", "голый", "эротик",
    "sex", "porn", "nude", "naked", "xxx", "erotic",
    # solicitation
    "проституция", "эскорт", "escort",
]

WARNED_WORDS = [
    "scam", "мошенничество", "кидалово",
    "abuse", "harassment",
    "badword",
]


def predict(text: str) -> str:
    """Return 'clean' | 'warn' | 'block' using substring matching (like Go)."""
    lower = (text or "").lower()
    for w in BLOCKED_WORDS:
        if w in lower:
            return "block"
    for w in WARNED_WORDS:
        if w in lower:
            return "warn"
    return "clean"
