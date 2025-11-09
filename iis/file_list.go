package iis

import (
	"context"
	"fmt"
)

func (client Client) ListFiles(ctx context.Context, parentID string) ([]File, error) {
	path := "/api/files"
	if parentID != "" {
		path = fmt.Sprintf("/api/files?parent.id=%s", parentID)
	}
	
	var res FileListResponse
	err := getJson(ctx, client, path, &res)
	if err != nil {
		return nil, err
	}
	return res.Files, nil
}

func (client Client) ListWebServerFiles(ctx context.Context, websiteID string) ([]File, error) {
	path := "/api/webserver/files"
	if websiteID != "" {
		path = fmt.Sprintf("/api/webserver/files?website.id=%s", websiteID)
	}
	
	var res FileListResponse
	err := getJson(ctx, client, path, &res)
	if err != nil {
		return nil, err
	}
	return res.Files, nil
}
