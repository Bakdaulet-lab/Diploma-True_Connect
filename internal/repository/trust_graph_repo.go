package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
)

// TrustGraphRepository abstracts Neo4j operations for the trust graph.
type TrustGraphRepository interface {
	// CreateUserNode creates a new user node in the graph.
	CreateUserNode(ctx context.Context, uid uuid.UUID, verificationLevel domain.VerificationLevel) error

	// AddRating creates a RATED edge between two user nodes.
	AddRating(ctx context.Context, raterUID, ratedUID uuid.UUID, score int, interactionContext string, verified bool) error

	// AddMeeting creates a MET_WITH edge between two user nodes.
	AddMeeting(ctx context.Context, uidA, uidB uuid.UUID, confirmedByBoth bool) error

	// AddReport creates a REPORTED edge between two user nodes.
	AddReport(ctx context.Context, reporterUID, reportedUID uuid.UUID, reason string) error

	// ComputeTrustScore runs the weighted trust score calculation and returns 0-100.
	ComputeTrustScore(ctx context.Context, uid uuid.UUID) (int, error)

	// DeleteUserNode completely removes a user and their edges from the graph.
	DeleteUserNode(ctx context.Context, uid uuid.UUID) error

	// UpdateTrustScore sets the trust_score property on a user node.
	UpdateTrustScore(ctx context.Context, uid uuid.UUID, score int) error

	// DetectSybilClusters runs community detection and returns suspicious clusters.
	DetectSybilClusters(ctx context.Context) ([]SybilCluster, error)

	// GetRecommendations returns user IDs based on graph properties like mutual connections.
	GetRecommendations(ctx context.Context, uid uuid.UUID, limit int) ([]uuid.UUID, error)

	// GetDirectInteractions gets all users who interacted with the target uid.
	GetDirectInteractions(ctx context.Context, uid uuid.UUID) ([]uuid.UUID, error)
}

// SybilCluster represents a group of suspicious accounts detected by graph analysis.
type SybilCluster struct {
	CommunityID         int
	Size                int
	ExternalConnections int
	SuspectUIDs         []uuid.UUID
}
