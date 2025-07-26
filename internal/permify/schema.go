package permify

import (
	"context"
	"fmt"
	"time"

	base "buf.build/gen/go/permifyco/permify/protocolbuffers/go/base/v1"
)

const (
	defaultTenantID = "t1"
	defaultTimeout  = 5 * time.Second
)

func UploadSchema(ctx context.Context, tenantID string) (string, error) {
	if tenantID == "" {
		tenantID = defaultTenantID
	}

	schema := `
entity user {}

entity domain {
  relation member @user
  relation admin @user

  action view = member or admin
  action edit = admin
}

entity post {
	relation member @user 
	relation admin @user 
	
	action view = member or admin 
	action edit = admin
}
`

	// Create a context with timeout if none provided
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), defaultTimeout)
		defer cancel()
	}

	res, err := Client.Schema.Write(ctx, &base.SchemaWriteRequest{
		TenantId: tenantID,
		Schema:   schema,
	})
	if err != nil {
		return "", fmt.Errorf("failed to write schema: %w", err)
	}

	return res.SchemaVersion, nil
}
