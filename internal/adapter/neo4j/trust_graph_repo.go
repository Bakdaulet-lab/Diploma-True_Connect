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
		MERGE (a)-[r:RATED]->(b)
		ON CREATE SET
			r.score      = $score,
			r.context    = $context,
			r.verified   = $verified,
			r.created_at = datetime()
		ON MATCH SET
			r.verified   = $verified,
			r.updated_at = datetime()`

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

// ComputeTrustScore runs the weighted Bayesian trust score calculation.
// It incorporates Neo4j trust graph interactions (RATED) and platform behavior (KYC, REPORTED).
func (r *TrustGraphRepo) ComputeTrustScore(ctx context.Context, uid uuid.UUID) (int, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	// Weighted average with Bayesian smoothing toward 50 (neutral).
	// Verified interactions from high-trust, identity-verified raters count more.
	// Platform behavior: bonuses for own KYC, penalties for being reported.
	cypher := `
		MATCH (u:User {uid: $uid})
		
		// 1. Base KYC Platform Bonus
		WITH u,
		     CASE u.verification_level
		       WHEN 'id_verified' THEN 10.0
		       WHEN 'photo_verified' THEN 5.0
		       ELSE 0.0
		     END AS kyc_bonus

		// 2. Aggregate Verified Ratings
		OPTIONAL MATCH (u)<-[r:RATED]-(rater:User)
		WHERE r.verified = true
		WITH u, kyc_bonus, rater, r,
		     CASE
		       WHEN rater.verification_level IN ['id_verified', 'photo_verified'] THEN 1.5
		       ELSE 1.0
		     END AS id_weight,
		     (COALESCE(rater.trust_score, 50) / 100.0) AS trust_weight
		
		WITH u, kyc_bonus,
		     CASE WHEN r IS NOT NULL THEN (r.score * id_weight * trust_weight) ELSE null END as effective_rating
		
		WITH u, kyc_bonus,
		     SUM(effective_rating) AS sum_effective,
		     COUNT(effective_rating) AS rating_count
		
		// Bayesian Smoothing (default = 2.5 score over 5 virtual ratings)
		WITH u, kyc_bonus,
		     (sum_effective + (2.5 * 5.0)) / (rating_count + 5.0) AS smoothed

		// 3. Aggregate Reports (Platform Behavior Penalty)
		OPTIONAL MATCH (u)<-[rep:REPORTED]-(reporter:User)
		WITH u, kyc_bonus, smoothed, COUNT(DISTINCT reporter) AS report_count

		// 4. Final Calculation
		WITH (smoothed * 20.0) AS base_score,
		     kyc_bonus,
		     (report_count * 15.0) AS report_penalty

		WITH (base_score + kyc_bonus - report_penalty) AS raw_score
		
		RETURN toInteger(
		    CASE 
		      WHEN raw_score > 100.0 THEN 100.0
		      WHEN raw_score < 0.0 THEN 0.0 
		      ELSE raw_score 
		    END
		) AS trust_score`

	result, err := session.Run(ctx, cypher, map[string]any{"uid": uid.String()})
	if err != nil {
		return 0, fmt.Errorf("computing trust score: %w", err)
	}

	record, err := result.Single(ctx)
	if err != nil {
		// No node or error in calculation — return neutral score.
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

	return int(score), nil
}

// DeleteUserNode completely removes a user and their edges from the graph.
func (r *TrustGraphRepo) DeleteUserNode(ctx context.Context, uid uuid.UUID) error {
	query := `
		MATCH (u:User {id: $uid})
		DETACH DELETE u
	`
	params := map[string]any{"uid": uid.String()}

	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		return tx.Run(ctx, query, params)
	})

	if err != nil {
		return fmt.Errorf("neo4j delete user node: %w", err)
	}

	return nil
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

// GetRecommendations returns user IDs based on friends-of-friends connections.
func (r *TrustGraphRepo) GetRecommendations(ctx context.Context, uid uuid.UUID, limit int) ([]uuid.UUID, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	cypher := `MATCH (me:User {uid: $uid})-[:MET_WITH|RATED]-(friend:User)-[:MET_WITH|RATED]-(fof:User)
	WHERE NOT (me)-[:MET_WITH|RATED]-(fof) AND me <> fof
	WITH fof, count(friend) AS shared_connections
	ORDER BY shared_connections DESC, fof.trust_score DESC
	LIMIT $limit
	RETURN fof.uid AS recommendedId`

	result, err := session.Run(ctx, cypher, map[string]any{
		"uid":   uid.String(),
		"limit": limit,
	})
	if err != nil {
		return nil, fmt.Errorf("neo4j getting recommendations: %w", err)
	}

	var recommendedIDs []uuid.UUID
	for result.Next(ctx) {
		uidStr, _ := result.Record().Get("recommendedId")
		parsedUID, err := uuid.Parse(uidStr.(string))
		if err == nil {
			recommendedIDs = append(recommendedIDs, parsedUID)
		}
	}
	return recommendedIDs, result.Err()
}

// GetDirectInteractions returns user IDs based on RATED, MET_WITH, or REPORTED edges.
func (r *TrustGraphRepo) GetDirectInteractions(ctx context.Context, uid uuid.UUID) ([]uuid.UUID, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	// We want anyone who rated this user or who was rated by this user.
	cypher := `
MATCH (u:User {uid: $uid})-[]-(other:User)
RETURN DISTINCT other.uid AS targetUid
`

	res, err := session.Run(ctx, cypher, map[string]any{"uid": uid.String()})
	if err != nil {
		return nil, fmt.Errorf("neo4j GetDirectInteractions: %w", err)
	}

	var targets []uuid.UUID
	for res.Next(ctx) {
		record := res.Record()
		targetStr, _ := record.Get("targetUid")
		if t, err := uuid.Parse(targetStr.(string)); err == nil {
			targets = append(targets, t)
		}
	}

	if err = res.Err(); err != nil {
		return nil, fmt.Errorf("reading direct interactions: %w", err)
	}

	return targets, nil
}
