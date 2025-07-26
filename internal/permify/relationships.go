package permify

import (
	"context"

	base "buf.build/gen/go/permifyco/permify/protocolbuffers/go/base/v1"
)

func AssignUserToDomain(userID, domainID, role, schemaVersion string) error {
	_, err := Client.Data.WriteRelationships(context.Background(), &base.RelationshipWriteRequest{
		TenantId: "t1",
		Metadata: &base.RelationshipWriteRequestMetadata{
			SchemaVersion: schemaVersion,
		},
		Tuples: []*base.Tuple{
			{
				Entity: &base.Entity{
					Type: "domain",
					Id:   domainID,
				},
				Relation: role, // "member" or "admin"
				Subject: &base.Subject{
					Type: "user",
					Id:   userID,
				},
			},
		},
	})
	return err
}
