package iis

import (
	"context"
	"fmt"
)

func (client Client) ReadFile(ctx context.Context, id string) (*File, error) {
	url := fmt.Sprintf("/api/files/%s", id)
	var file File
	if err := getJson(ctx, client, url, &file); err != nil {
		return nil, err
	}
	return &file, nil
}

func (client Client) ReadWebServerFile(ctx context.Context, id string) (*File, error) {
	url := fmt.Sprintf("/api/webserver/files/%s", id)
	var file File
	if err := getJson(ctx, client, url, &file); err != nil {
		return nil, err
	}
	return &file, nil
}
