"""Canonical feature computation for the match recommender.

⚠️ MUST stay in lockstep with the Go implementation in
internal/pkg/recommender/recommender.go (featureValue / distanceKm).
Same feature names, same formulas, same imputation — otherwise the offline model
and the online scorer disagree.

Each example is a (viewer, candidate) pair. ``viewer`` and ``candidate`` are
dicts with keys: age (int|None), lat/lon (float|None), city, niyyah, madhab,
languages (list[str]); candidate additionally has trust (int) and kyc (bool).
"""
import math

# Canonical order. train.py exports model.json with feature_names in THIS order.
FEATURES = [
    "age_gap",
    "age_known",
    "distance_km",
    "distance_known",
    "same_city",
    "niyyah_match",
    "niyyah_compatible",
    "madhab_match",
    "language_overlap",
    "cand_trust",
    "cand_kyc",
]


def _haversine_km(lat1, lon1, lat2, lon2):
    r = 6371.0
    p1, p2 = math.radians(lat1), math.radians(lat2)
    dlat = math.radians(lat2 - lat1)
    dlon = math.radians(lon2 - lon1)
    h = math.sin(dlat / 2) ** 2 + math.cos(p1) * math.cos(p2) * math.sin(dlon / 2) ** 2
    return 2 * r * math.asin(min(1.0, math.sqrt(h)))


def _distance(viewer, cand):
    for v in (viewer.get("lat"), viewer.get("lon"), cand.get("lat"), cand.get("lon")):
        if v is None:
            return None
    return _haversine_km(viewer["lat"], viewer["lon"], cand["lat"], cand["lon"])


def _is_madhab(m):
    return bool(m) and m != "none"


def _niyyah_compatible(viewer_n, cand_n):
    if viewer_n == "nikah_year":
        return cand_n in ("nikah_year", "serious_marriage")
    return True


def _lang_overlap(a, b):
    sa = {str(x).strip().lower() for x in (a or [])}
    return sum(1 for y in (b or []) if str(y).strip().lower() in sa)


def compute(viewer: dict, cand: dict) -> dict:
    va, ca = viewer.get("age"), cand.get("age")
    age_known = va is not None and ca is not None
    dist = _distance(viewer, cand)
    vcity, ccity = (viewer.get("city") or ""), (cand.get("city") or "")
    vn, cn = (viewer.get("niyyah") or ""), (cand.get("niyyah") or "")
    vm, cm = (viewer.get("madhab") or ""), (cand.get("madhab") or "")

    return {
        "age_gap": abs(va - ca) if age_known else 0.0,
        "age_known": 1.0 if age_known else 0.0,
        "distance_km": dist if dist is not None else 0.0,
        "distance_known": 1.0 if dist is not None else 0.0,
        "same_city": 1.0 if vcity and vcity.lower() == ccity.lower() else 0.0,
        "niyyah_match": 1.0 if vn and vn == cn else 0.0,
        "niyyah_compatible": 1.0 if _niyyah_compatible(vn, cn) else 0.0,
        "madhab_match": 1.0 if _is_madhab(vm) and vm == cm else 0.0,
        "language_overlap": float(_lang_overlap(viewer.get("languages"), cand.get("languages"))),
        "cand_trust": float(cand.get("trust", 0) or 0),
        "cand_kyc": 1.0 if cand.get("kyc") else 0.0,
    }


def vector(viewer: dict, cand: dict) -> list:
    f = compute(viewer, cand)
    return [f[name] for name in FEATURES]
