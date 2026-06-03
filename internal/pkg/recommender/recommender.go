// Package recommender scores match candidates with a logistic-regression
// learning-to-rank model trained offline in Python (ml/recommender). The model
// is exported as a small model.json (feature names + standardizer + weights) and
// served inline here — no runtime ML service, sub-millisecond per candidate.
//
// CRITICAL: the feature computation below MUST stay in lockstep with
// ml/recommender/features.py. Features are addressed by name (driven by the
// model file), so the two sides agree as long as the names + formulas match.
package recommender

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
)

// Model is the JSON artifact exported by ml/recommender/train.py.
type Model struct {
	FeatureNames []string  `json:"feature_names"`
	Mean         []float64 `json:"mean"`      // StandardScaler.mean_
	Scale        []float64 `json:"scale"`     // StandardScaler.scale_
	Coef         []float64 `json:"coef"`      // LogisticRegression.coef_[0]
	Intercept    float64   `json:"intercept"` // LogisticRegression.intercept_[0]
}

// Ranker scores candidate pairs.
type Ranker struct {
	m Model
}

// Pair holds the raw viewer + candidate fields needed to compute features.
type Pair struct {
	ViewerAge       *int
	ViewerLat       *float64
	ViewerLon       *float64
	ViewerCity      string
	ViewerNiyyah    string
	ViewerMadhab    string
	ViewerLanguages []string

	CandAge       *int
	CandLat       *float64
	CandLon       *float64
	CandCity      string
	CandNiyyah    string
	CandMadhab    string
	CandLanguages []string
	CandTrust     int
	CandKYC       bool
}

// Load reads a model.json. Returns (nil, nil) if path is empty (feature
// disabled); returns an error only when a configured file is invalid.
func Load(path string) (*Ranker, error) {
	if strings.TrimSpace(path) == "" {
		return nil, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("recommender: read model: %w", err)
	}
	var m Model
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("recommender: parse model: %w", err)
	}
	n := len(m.FeatureNames)
	if n == 0 || len(m.Mean) != n || len(m.Scale) != n || len(m.Coef) != n {
		return nil, fmt.Errorf("recommender: inconsistent model dimensions")
	}
	return &Ranker{m: m}, nil
}

// Score returns P(mutual like) in [0,1] for the pair.
func (r *Ranker) Score(p Pair) float64 {
	z := r.m.Intercept
	for i, name := range r.m.FeatureNames {
		x := featureValue(name, p)
		scale := r.m.Scale[i]
		if scale == 0 {
			scale = 1
		}
		z += r.m.Coef[i] * ((x - r.m.Mean[i]) / scale)
	}
	return 1.0 / (1.0 + math.Exp(-z))
}

// RankIndices returns candidate indices sorted by descending score. The input
// slice is not mutated; ties preserve original order (stable).
func (r *Ranker) RankIndices(pairs []Pair) []int {
	idx := make([]int, len(pairs))
	scores := make([]float64, len(pairs))
	for i := range pairs {
		idx[i] = i
		scores[i] = r.Score(pairs[i])
	}
	sort.SliceStable(idx, func(a, b int) bool {
		return scores[idx[a]] > scores[idx[b]]
	})
	return idx
}

// ── Feature computation (must match ml/recommender/features.py) ──────────────

func featureValue(name string, p Pair) float64 {
	switch name {
	case "age_gap":
		if p.ViewerAge != nil && p.CandAge != nil {
			return math.Abs(float64(*p.ViewerAge - *p.CandAge))
		}
		return 0
	case "age_known":
		return b2f(p.ViewerAge != nil && p.CandAge != nil)
	case "distance_km":
		if d, ok := distanceKm(p); ok {
			return d
		}
		return 0
	case "distance_known":
		_, ok := distanceKm(p)
		return b2f(ok)
	case "same_city":
		return b2f(p.ViewerCity != "" && strings.EqualFold(p.ViewerCity, p.CandCity))
	case "niyyah_match":
		return b2f(p.ViewerNiyyah != "" && p.ViewerNiyyah == p.CandNiyyah)
	case "niyyah_compatible":
		return b2f(niyyahCompatible(p.ViewerNiyyah, p.CandNiyyah))
	case "madhab_match":
		return b2f(isMadhab(p.ViewerMadhab) && p.ViewerMadhab == p.CandMadhab)
	case "language_overlap":
		return float64(languageOverlap(p.ViewerLanguages, p.CandLanguages))
	case "cand_trust":
		return float64(p.CandTrust)
	case "cand_kyc":
		return b2f(p.CandKYC)
	default:
		return 0
	}
}

func b2f(v bool) float64 {
	if v {
		return 1
	}
	return 0
}

func isMadhab(m string) bool {
	return m != "" && m != "none"
}

// niyyahCompatible mirrors matching_service.niyyahCompatible: a nikah_year
// viewer is compatible only with nikah_year/serious_marriage candidates; any
// other viewer is open to all.
func niyyahCompatible(viewer, cand string) bool {
	if viewer == "nikah_year" {
		return cand == "nikah_year" || cand == "serious_marriage"
	}
	return true
}

func languageOverlap(a, b []string) int {
	set := make(map[string]struct{}, len(a))
	for _, x := range a {
		set[strings.ToLower(strings.TrimSpace(x))] = struct{}{}
	}
	count := 0
	for _, y := range b {
		if _, ok := set[strings.ToLower(strings.TrimSpace(y))]; ok {
			count++
		}
	}
	return count
}

// distanceKm returns the haversine distance in km, ok=false if any coord missing.
func distanceKm(p Pair) (float64, bool) {
	if p.ViewerLat == nil || p.ViewerLon == nil || p.CandLat == nil || p.CandLon == nil {
		return 0, false
	}
	const r = 6371.0 // earth radius km
	lat1 := *p.ViewerLat * math.Pi / 180
	lat2 := *p.CandLat * math.Pi / 180
	dLat := (*p.CandLat - *p.ViewerLat) * math.Pi / 180
	dLon := (*p.CandLon - *p.ViewerLon) * math.Pi / 180
	h := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(dLon/2)*math.Sin(dLon/2)
	return 2 * r * math.Asin(math.Min(1, math.Sqrt(h))), true
}
