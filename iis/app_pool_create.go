package iis

import (
	"context"
	"encoding/json"
)

func (client Client) CreateAppPool(ctx context.Context, name string, managedRuntimeVersion string) (*ApplicationPool, error) {
	reqBody := CreateApplicationPoolRequest{
		Name:                  name,
		ManagedRuntimeVersion: managedRuntimeVersion,
	}
	res, err := httpPost(ctx, client, "/api/webserver/application-pools", reqBody)
	if err != nil {
		// If we get a 409 Conflict, the app pool already exists
		// Try to retrieve it by name instead
		if IsConflictError(err) {
			pool, getErr := client.GetAppPoolByName(ctx, name)
			if getErr == nil {
				// Found the existing pool, return it
				return pool, nil
			}
			// If we can't find it, return the original conflict error
		}
		return nil, err
	}
	var pool ApplicationPool
	err = json.Unmarshal(res, &pool)
	if err != nil {
		return nil, err
	}
	return &pool, nil
}

type CreateApplicationPoolRequest struct {
	Name                  string `json:"name"`
	ManagedRuntimeVersion string `json:"managed_runtime_version,omitempty"`
}
