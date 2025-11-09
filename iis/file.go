package iis

import "time"

type File struct {
	Name         string    `json:"name"`
	ID           string    `json:"id"`
	Type         string    `json:"type"` // "file" or "directory"
	PhysicalPath string    `json:"physical_path"`
	Exists       bool      `json:"exists"`
	Size         int64     `json:"size,omitempty"`
	Created      time.Time `json:"created"`
	LastModified time.Time `json:"last_modified"`
	LastAccess   time.Time `json:"last_access"`
	ETag         string    `json:"e_tag,omitempty"`
	Parent       *FileRef  `json:"parent,omitempty"`
	TotalFiles   int       `json:"total_files,omitempty"` // For directories
	Claims       []string  `json:"claims"`
}

type FileRef struct {
	Name         string `json:"name"`
	ID           string `json:"id"`
	Type         string `json:"type"`
	PhysicalPath string `json:"physical_path"`
}

type FileListResponse struct {
	Files []File `json:"files"`
}

type CreateFileRequest struct {
	Name   string   `json:"name"`
	Parent *FileRef `json:"parent,omitempty"`
	Type   string   `json:"type"` // "file" or "directory"
}

type CopyMoveFileRequest struct {
	Name   string   `json:"name,omitempty"`
	File   *FileRef `json:"file"`
	Parent *FileRef `json:"parent"`
}
