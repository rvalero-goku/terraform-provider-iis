package iis

import (
	"context"
	"encoding/json"
)

type CreateWebsiteRequest struct {
	Name            string               `json:"name"`
	PhysicalPath    string               `json:"physical_path"`
	Bindings        []WebsiteBinding     `json:"bindings"`
	ApplicationPool ApplicationReference `json:"application_pool"`
}

func (client Client) CreateWebsite(ctx context.Context, req CreateWebsiteRequest) (*Website, error) {
	res, err := httpPost(ctx, client, "/api/webserver/websites", req)
	if err != nil {
		// If we get a 409 Conflict, the website already exists
		// Try to retrieve it by name instead
		if IsConflictError(err) {
			site, getErr := client.GetWebsiteByName(ctx, req.Name)
			if getErr == nil && site != nil {
				// Found the existing site, return it
				return site, nil
			}
			// If we can't find it, return the original conflict error
		}
		return nil, err
	}
	var site Website
	err = json.Unmarshal(res, &site)
	if err != nil {
		return nil, err
	}
	return &site, nil
}
