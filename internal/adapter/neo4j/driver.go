package neo4jadapter

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// New creates a new Neo4j driver connection.
func New(ctx context.Context, uri, user, password string) (neo4j.DriverWithContext, error) {
	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(user, password, ""))
	if err != nil {
		return nil, fmt.Errorf("creating neo4j driver: %w", err)
	}

	if err := driver.VerifyConnectivity(ctx); err != nil {
		driver.Close(ctx)
		return nil, fmt.Errorf("verifying neo4j connectivity: %w", err)
	}

	return driver, nil
}

// SetupConstraints creates required indexes and uniqueness constraints.
func SetupConstraints(ctx context.Context, driver neo4j.DriverWithContext) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	constraints := []string{
		"CREATE CONSTRAINT user_uid IF NOT EXISTS FOR (u:User) REQUIRE u.uid IS UNIQUE",
		"CREATE INDEX user_trust_score IF NOT EXISTS FOR (u:User) ON (u.trust_score)",
	}

	for _, cypher := range constraints {
		_, err := session.Run(ctx, cypher, nil)
		if err != nil {
			return fmt.Errorf("executing constraint %q: %w", cypher, err)
		}
	}

	return nil
}
