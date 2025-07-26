package permify

import (
	"context"

	base "buf.build/gen/go/permifyco/permify/protocolbuffers/go/base/v1"
)

func CanUser(userID, domainID, permission, schemaVersion, snapToken string) (bool, error) {
	res, err := Client.Permission.Check(context.Background(), &base.PermissionCheckRequest{
		TenantId: "t1",
		Metadata: &base.PermissionCheckRequestMetadata{
			SchemaVersion: schemaVersion,
			SnapToken:     snapToken,
			Depth:         50,
		},
		Entity: &base.Entity{
			Type: "domain",
			Id:   domainID,
		},
		Permission: permission,
		Subject: &base.Subject{
			Type: "user",
			Id:   userID,
		},
	})
	if err != nil {
		return false, err
	}

	return res.Can == base.CheckResult_CHECK_RESULT_ALLOWED, nil
}
