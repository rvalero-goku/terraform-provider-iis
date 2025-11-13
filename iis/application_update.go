package iis

import (
	"context"
	"encoding/json"
	"fmt"
)

func (client Client) UpdateApplication(ctx context.Context, id string, application UpdateApplicationRequest) (*Application, error) {
	url := fmt.Sprintf("/api/webserver/webapps/%s", id)
	res, err := httpPatch(ctx, client, url, application)
	if err != nil {
		return nil, err
	}
	var app Application
	err = json.Unmarshal(res, &app)
	if err != nil {
		return nil, err
	}
	return &app, nil
}

type UpdateApplicationRequest struct {
	Path            string    `json:"path,omitempty"`
	PhysicalPath    string    `json:"physical_path,omitempty"`
	ApplicationPool Reference `json:"application_pool,omitempty"`
}
