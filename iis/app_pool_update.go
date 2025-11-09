package iis

import (
	"context"
	"encoding/json"
	"fmt"
)

func (client Client) UpdateAppPool(ctx context.Context, id, name, runtimeVersion, status string) (*ApplicationPool, error) {
	reqBody := struct {
		Name                  string `json:"name"`
		ManagedRuntimeVersion string `json:"managed_runtime_version"`
		Status                string `json:"status"`
	}{
		Name:                  name,
		ManagedRuntimeVersion: runtimeVersion,
		Status:                status,
	}
	url := fmt.Sprintf("/api/webserver/application-pools/%s", id)
	res, err := httpPatch(ctx, client, url, reqBody)
	if err != nil {
		return nil, err
	}
	var pool ApplicationPool
	err = json.Unmarshal(res, &pool)
	if err != nil {
		return nil, err
	}
	return &pool, nil
}
