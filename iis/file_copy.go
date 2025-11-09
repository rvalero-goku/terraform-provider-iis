package iis

import (
	"context"
	"encoding/json"
)

func (client Client) CopyFile(ctx context.Context, req CopyMoveFileRequest) (*File, error) {
	res, err := httpPost(ctx, client, "/api/files/copy", req)
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

func (client Client) MoveFile(ctx context.Context, req CopyMoveFileRequest) (*File, error) {
	res, err := httpPost(ctx, client, "/api/files/move", req)
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
