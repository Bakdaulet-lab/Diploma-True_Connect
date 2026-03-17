package neo4jadapter

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/repository"
)

// TrustGraphRepo implements repository.TrustGraphRepository using Neo4j.
type TrustGraphRepo struct {
	driver neo4j.DriverWithContext
}

var _ repository.TrustGraphRepository = (*TrustGraphRepo)(nil)

// NewTrustGraphRepo creates a new Neo4j-backed trust graph repository.
func NewTrustGraphRepo(driver neo4j.DriverWithContext) *TrustGraphRepo {
	return &TrustGraphRepo{driver: driver}
}

// CreateUserNode creates a User node in the trust graph on registration.
func (r *TrustGraphRepo) CreateUserNode(ctx context.Context, uid uuid.UUID, verificationLevel domain.VerificationLevel) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	cypher := `
		MERGE (u:User {uid: $uid})
		ON CREATE SET
			u.verification_level = $level,
			u.trust_score        = 50,
			u.created_at         = datetime()
		ON MATCH SET
			u.verification_level = $level`

	_, err := session.Run(ctx, cypher, map[string]any{
		"uid":   uid.String(),
		"level": string(verificationLevel),
	})
	if err != nil {
		return fmt.Errorf("creating user node in neo4j: %w", err)
	}

	return nil
}

// AddRating creates a RATED edge from rater to rated. Implemented in Sprint 3.
func (r *TrustGraphRepo) AddRating(ctx context.Context, raterUID, ratedUID uuid.UUID, score int, interactionContext string, verified bool) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	cypher := `
		MATCH (a:User {uid: $rater_uid}), (b:User {uid: $rated_uid})
		CREATE (a)-[:RATED {
			score:    $score,
			context:  $context,
			verified: $verified,
			created_at: datetime()
		}]->(b)`

	_, err := session.Run(ctx, cypher, map[string]any{
		"rater_uid": raterUID.String(),
		"rated_uid": ratedUID.String(),
		"score":     score,
		"context":   interactionContext,
		"verified":  verified,
	})
	if err != nil {
		return fmt.Errorf("adding rating edge: %w", err)
	}

	return nil
}

// AddMeeting creates a MET_WITH edge between two users. Implemented in Sprint 3.
func (r *TrustGraphRepo) AddMeeting(ctx context.Context, uidA, uidB uuid.UUID, confirmedByBoth bool) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	cypher := `
		MATCH (a:User {uid: $uid_a}), (b:User {uid: $uid_b})
		MERGE (a)-[m:MET_WITH]->(b)
		SET m.confirmed_by_both = $confirmed,
		    m.created_at        = datetime()`

	_, err := session.Run(ctx, cypher, map[string]any{
		"uid_a":     uidA.String(),
		"uid_b":     uidB.String(),
		"confirmed": confirmedByBoth,
	})
	if err != nil {
		return fmt.Errorf("adding meeting edge: %w", err)
	}

	return nil
}

// AddReport creates a REPORTED edge. Implemented in Sprint 3.
func (r *TrustGraphRepo) AddReport(ctx context.Context, reporterUID, reportedUID uuid.UUID, reason string) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	cypher := `
		MATCH (a:User {uid: $reporter}), (b:User {uid: $reported})
		CREATE (a)-[:REPORTED {reason: $reason, created_at: datetime()}]->(b)`

	_, err := session.Run(ctx, cypher, map[string]any{
		"reporter": reporterUID.String(),
		"reported": reportedUID.String(),
		"reason":   reason,
	})
	if err != nil {
		return fmt.Errorf("adding report edge: %w", err)
	}

	return nil
}

// ComputeTrustScore runs the weighted Bayesian trust score calculation. Implemented in Sprint 3.
func (r *TrustGraphRepo) ComputeTrustScore(ctx context.Context, uid uuid.UUID) (int, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	// Weighted average with Bayesian smoothing toward 50 (neutral).
	// Verified interactions from high-trust, identity-verified raters count more.
	cypher := `
		MATCH (u:User {uid: $uid})<-[r:RATED]-(rater:User)
		WHERE r.verified = true
		WITH u, rater, r,
		     CASE
		       WHEN rater.verification_level IN ['id_verified', 'photo_verified'] THEN 1.5
		       ELSE 1.0
		     END AS id_weight,
		     rater.trust_score / 100.0 AS trust_weight
		WITH u,
		     AVG(r.score * id_weight * trust_weight) AS weighted_avg,
		     COUNT(r) AS rating_count
		WITH u,
		     (weighted_avg * rating_count + 2.5 * 5) / (rating_count + 5) AS smoothed
		RETURN toInteger(smoothed * 20) AS trust_score`

	result, err := session.Run(ctx, cypher, map[string]any{"uid": uid.String()})
	if err != nil {
		return 0, fmt.Errorf("computing trust score: %w", err)
	}

	record, err := result.Single(ctx)
	if err != nil {
		// No verified ratings yet — return neutral score.
		return 50, nil
	}

	scoreRaw, ok := record.Get("trust_score")
	if !ok {
		return 50, nil
	}

	score, ok := scoreRaw.(int64)
	if !ok {
		return 50, nil
	}

	// Clamp to 0-100.
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return int(score), nil
}

// UpdateTrustScore sets the trust_score property on a user node.
func (r *TrustGraphRepo) UpdateTrustScore(ctx context.Context, uid uuid.UUID, score int) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	cypher := `MATCH (u:User {uid: $uid}) SET u.trust_score = $score`

	_, err := session.Run(ctx, cypher, map[string]any{
		"uid":   uid.String(),
		"score": score,
	})
	if err != nil {
		return fmt.Errorf("updating trust score: %w", err)
	}

	return nil
}

// DetectSybilClusters runs Louvain community detection via Neo4j GDS to find suspicious clusters.
func (r *TrustGraphRepo) DetectSybilClusters(ctx context.Context) ([]repository.SybilCluster, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	// Drop any stale projection from a prior run.
	_, _ = session.Run(ctx, `CALL gds.graph.drop('trust-net', false)`, nil)

	// Step 1: Project the graph.
	projectCypher := `
		CALL gds.graph.project('trust-net', 'User', {
			RATED:     {orientation: 'UNDIRECTED'},
			MET_WITH:  {orientation: 'UNDIRECTED'}
		})`
	if _, err := session.Run(ctx, projectCypher, nil); err != nil {
		return nil, fmt.Errorf("projecting trust graph: %w", err)
	}

	// Step 2-3: Run Louvain and filter suspicious clusters.
	detectCypher := `
		CALL gds.louvain.stream('trust-net')
		YIELD nodeId, communityId
		WITH communityId, collect(gds.util.asNode(nodeId)) AS members, count(*) AS size
		WHERE size >= 3
		UNWIND members AS m
		OPTIONAL MATCH (m)-[:RATED|MET_WITH]-(external:User)
		WHERE NOT external IN members
		WITH communityId, size, count(external) AS external_connections, members
		WHERE external_connections < size * 0.3
		RETURN communityId, size, external_connections,
		       [m IN members | m.uid] AS suspect_uids`

	result, err := session.Run(ctx, detectCypher, nil)
	if err != nil {
		_, _ = session.Run(ctx, `CALL gds.graph.drop('trust-net', false)`, nil)
		return nil, fmt.Errorf("running sybil detection: %w", err)
	}

	var clusters []repository.SybilCluster
	for result.Next(ctx) {
		record := result.Record()

		communityID, _ := record.Get("communityId")
		size, _ := record.Get("size")
		extConns, _ := record.Get("external_connections")
		suspectRaw, _ := record.Get("suspect_uids")

		var uids []uuid.UUID
		if uidList, ok := suspectRaw.([]any); ok {
			for _, raw := range uidList {
				if s, ok := raw.(string); ok {
					if uid, err := uuid.Parse(s); err == nil {
						uids = append(uids, uid)
					}
				}
			}
		}

		clusters = append(clusters, repository.SybilCluster{
			CommunityID:         int(communityID.(int64)),
			Size:                int(size.(int64)),
			ExternalConnections: int(extConns.(int64)),
			SuspectUIDs:         uids,
		})
	}
	if err := result.Err(); err != nil {
		_, _ = session.Run(ctx, `CALL gds.graph.drop('trust-net', false)`, nil)
		return nil, fmt.Errorf("iterating sybil results: %w", err)
	}

	// Clean up the in-memory graph projection.
	_, _ = session.Run(ctx, `CALL gds.graph.drop('trust-net', false)`, nil)

	return clusters, nil
}
