package recommender

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"
)

func ptrInt(v int) *int             { return &v }
func ptrF(v float64) *float64       { return &v }

func TestLoadEmptyPathDisabled(t *testing.T) {
	r, err := Load("")
	if err != nil || r != nil {
		t.Fatalf("empty path should disable ranker, got r=%v err=%v", r, err)
	}
}

func TestDistanceKm(t *testing.T) {
	// Almaty → Astana ≈ 970 km.
	p := Pair{
		ViewerLat: ptrF(43.2389), ViewerLon: ptrF(76.8897),
		CandLat: ptrF(51.1694), CandLon: ptrF(71.4491),
	}
	d, ok := distanceKm(p)
	if !ok {
		t.Fatal("expected distance to be computable")
	}
	if math.Abs(d-970) > 60 {
		t.Fatalf("distance ~970km expected, got %.1f", d)
	}
	// Missing a coordinate → not computable.
	if _, ok := distanceKm(Pair{ViewerLat: ptrF(1)}); ok {
		t.Fatal("expected ok=false when coords missing")
	}
}

func TestFeatureValues(t *testing.T) {
	p := Pair{
		ViewerAge: ptrInt(30), CandAge: ptrInt(25),
		ViewerCity: "Almaty", CandCity: "almaty", // case-insensitive
		ViewerNiyyah: "nikah_year", CandNiyyah: "serious_marriage",
		ViewerMadhab: "hanafi", CandMadhab: "hanafi",
		ViewerLanguages: []string{"kk", "ru"}, CandLanguages: []string{"RU", "en"},
		CandTrust: 80, CandKYC: true,
	}
	cases := map[string]float64{
		"age_gap":           5,
		"age_known":         1,
		"same_city":         1,
		"niyyah_match":      0, // nikah_year != serious_marriage
		"niyyah_compatible": 1, // nikah_year viewer compatible with serious_marriage
		"madhab_match":      1,
		"language_overlap":  1, // "ru"
		"cand_trust":        80,
		"cand_kyc":          1,
	}
	for name, want := range cases {
		if got := featureValue(name, p); got != want {
			t.Errorf("feature %q = %v, want %v", name, got, want)
		}
	}

	// madhab "none" must not count as a match.
	none := Pair{ViewerMadhab: "none", CandMadhab: "none"}
	if featureValue("madhab_match", none) != 0 {
		t.Error("madhab_match should be 0 when both are 'none'")
	}
}

func TestScoreRankingOrder(t *testing.T) {
	model := Model{
		FeatureNames: []string{"cand_trust", "niyyah_match"},
		Mean:         []float64{50, 0.5},
		Scale:        []float64{20, 0.5},
		Coef:         []float64{1.0, 2.0}, // both positive → higher trust + match ⇒ higher score
		Intercept:    0,
	}
	path := filepath.Join(t.TempDir(), "model.json")
	b, _ := json.Marshal(model)
	if err := os.WriteFile(path, b, 0o600); err != nil {
		t.Fatal(err)
	}
	r, err := Load(path)
	if err != nil || r == nil {
		t.Fatalf("load failed: %v", err)
	}

	strong := Pair{CandTrust: 95, ViewerNiyyah: "serious_marriage", CandNiyyah: "serious_marriage"}
	weak := Pair{CandTrust: 10, ViewerNiyyah: "serious_marriage", CandNiyyah: "friendship"}

	if r.Score(strong) <= r.Score(weak) {
		t.Fatalf("strong candidate should score higher: strong=%.3f weak=%.3f",
			r.Score(strong), r.Score(weak))
	}

	// Input order [weak, strong] → ranked indices should put strong (index 1) first.
	order := r.RankIndices([]Pair{weak, strong})
	if order[0] != 1 {
		t.Fatalf("expected strong (idx 1) ranked first, got order=%v", order)
	}
}
