package iis

import (
	"context"
	"encoding/json"
)

func (client Client) CreateFile(ctx context.Context, req CreateFileRequest) (*File, error) {
	res, err := httpPost(ctx, client, "/api/files", req)
	if err != nil {
		// If we get a 409 Conflict, the file/directory already exists
		// Try to retrieve it by name from the parent
		if IsConflictError(err) && req.Parent != nil {
			file, getErr := client.GetFileByName(ctx, req.Name, req.Parent.ID)
			if getErr == nil && file != nil {
				// Found the existing file/directory, return it
				return file, nil
			}
			// If we can't find it, return the original conflict error
		}
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
