package iis

import (
	"context"
	"encoding/json"
)

func (client Client) CreateFile(ctx context.Context, req CreateFileRequest) (*File, error) {
	res, err := httpPost(ctx, client, "/api/files", req)
	if err != nil {
		return nil, err
	}
	var file File
	err = json.Unmarshal(res, &file)
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (client Client) CreateDirectory(ctx context.Context, name string, parent *FileRef) (*File, error) {
	req := CreateFileRequest{
		Name:   name,
		Parent: parent,
		Type:   "directory",
	}
	return client.CreateFile(ctx, req)
}
