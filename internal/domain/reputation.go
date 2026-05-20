package domain

import "github.com/google/uuid"

// TrustScore represents a user's computed reputation score (0-100).
type TrustScore struct {
	UserID      uuid.UUID
	Score       int
	RatingCount int
}

// TrustScoreBreakdown explains how a user's trust score was derived.
// Self-only: it exposes moderation signals (report count/penalty).
type TrustScoreBreakdown struct {
	UserID         uuid.UUID `json:"user_id"`
	Score          int       `json:"score"`           // final clamped 0-100
	Badge          string    `json:"badge"`           // tier from GetBadge()
	RatingCount    int       `json:"rating_count"`    // # verified ratings
	SmoothedRating float64   `json:"smoothed_rating"` // 0-5 (Bayesian)
	BaseScore      float64   `json:"base_score"`      // smoothed * 20
	KYCBonus       float64   `json:"kyc_bonus"`       // 0 / 5 / 10
	ReportCount    int       `json:"report_count"`
	ReportPenalty  float64   `json:"report_penalty"` // report_count * 15
	RawScore       float64   `json:"raw_score"`      // pre-clamp
}

// LeaderboardEntry represents a single user in the reputation leaderboard.
type LeaderboardEntry struct {
	UserID      uuid.UUID `json:"user_id"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url"`
	Score       int       `json:"score"`
	Badge       string    `json:"badge"`
}

// GetBadge returns a string representation of the user's reputation tier.
func (t *TrustScore) GetBadge() string {
	switch {
	case t.Score >= 90:
		return "Platinum / Trusted"
	case t.Score >= 70:
		return "Gold / Member"
	case t.Score >= 30:
		return "Silver / Newbie"
	default:
		return "Bronze / Unverified"
	}
}
