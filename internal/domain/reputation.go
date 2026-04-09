package domain

import "github.com/google/uuid"

// TrustScore represents a user's computed reputation score (0-100).
type TrustScore struct {
	UserID      uuid.UUID
	Score       int
	RatingCount int
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
