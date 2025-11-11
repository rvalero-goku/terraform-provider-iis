package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxjoehnk/terraform-provider-iis/iis"
)

func resourceFileCopy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFileCopyCreate,
		ReadContext:   resourceFileCopyRead,
		DeleteContext: resourceFileCopyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"source_path": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Full path to the source file",
			},
			"destination_path": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Full path to the destination (directory or file)",
			},
			"destination_name": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Name of the destination file (optional, defaults to source filename)",
			},
			"move": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
				Description: "If true, moves the file instead of copying it",
			},
			"file_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the resulting file",
			},
		},
	}
}

func resourceFileCopyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*iis.Client)
	sourcePath := d.Get("source_path").(string)
	destPath := d.Get("destination_path").(string)
	destName := d.Get("destination_name").(string)
	move := d.Get("move").(bool)

	tflog.Debug(ctx, "Looking up source file: "+sourcePath)
	
	// First, we need to get the file IDs by reading the file system
	// This is a simplified approach - you'll need to enhance this to handle paths properly
	sourceFile, err := findFileByPath(ctx, client, sourcePath)
	if err != nil {
		return diag.Errorf("failed to find source file %s: %v", sourcePath, err)
	}

	destParent, err := findFileByPath(ctx, client, destPath)
	if err != nil {
		return diag.Errorf("failed to find destination directory %s: %v", destPath, err)
	}

	req := iis.CopyMoveFileRequest{
		File: &iis.FileRef{
			ID: sourceFile.ID,
		},
		Parent: &iis.FileRef{
			ID: destParent.ID,
		},
	}

	if destName != "" {
		req.Name = destName
	}

	var resultFile *iis.File
	if move {
		tflog.Info(ctx, "Moving file from "+sourcePath+" to "+destPath)
		resultFile, err = client.MoveFile(ctx, req)
	} else {
		tflog.Info(ctx, "Copying file from "+sourcePath+" to "+destPath)
		resultFile, err = client.CopyFile(ctx, req)
	}

	if err != nil {
		return diag.FromErr(err)
	}

	tflog.Info(ctx, "File copy/move completed successfully, result file ID: "+resultFile.ID)
	d.SetId(resultFile.ID)
	d.Set("file_id", resultFile.ID)
	
	// Set the original input values to maintain state consistency
	d.Set("source_path", sourcePath)
	d.Set("destination_path", destPath)
	d.Set("move", move)
	if destName != "" {
		d.Set("destination_name", destName)
	}

	return nil
}

func resourceFileCopyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*iis.Client)
	
	tflog.Debug(ctx, "Reading file copy resource with ID: "+d.Id())

	file, err := client.ReadFile(ctx, d.Id())
	if err != nil {
		tflog.Warn(ctx, "File not found during read, marking as deleted: "+err.Error())
		d.SetId("")
		return nil
	}

	tflog.Debug(ctx, "Successfully read file: "+file.Name)
	d.Set("file_id", file.ID)
	d.Set("source_path", d.Get("source_path").(string))
	d.Set("destination_path", d.Get("destination_path").(string))
	d.Set("move", d.Get("move").(bool))

	return nil
}

func resourceFileCopyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*iis.Client)

	tflog.Info(ctx, "Deleting copied/moved file: "+d.Id())
	err := client.DeleteFile(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

// Helper function to find a file by its full path
func findFileByPath(ctx context.Context, client *iis.Client, path string) (*iis.File, error) {
	// Try to read directly if it's an ID first
	file, err := client.ReadFile(ctx, path)
	if err == nil {
		return file, nil
	}

	// If not an ID, search by physical path
	return findFileByPhysicalPath(ctx, client, path, "")
}

// Recursive function to search for a file by physical path
func findFileByPhysicalPath(ctx context.Context, client *iis.Client, targetPath string, parentID string) (*iis.File, error) {
	files, err := client.ListFiles(ctx, parentID)
	if err != nil {
		return nil, err
	}

	// Normalize paths for comparison (case-insensitive)
	normalizedTarget := strings.ToLower(strings.ReplaceAll(targetPath, "/", "\\"))

	for _, file := range files {
		normalizedFilePath := strings.ToLower(strings.ReplaceAll(file.PhysicalPath, "/", "\\"))
		
		// Check if this file matches our target path
		if normalizedFilePath == normalizedTarget {
			return &file, nil
		}

		// If this is a directory, recursively search within it
		if file.Type == "directory" {
			result, err := findFileByPhysicalPath(ctx, client, targetPath, file.ID)
			if err == nil {
				return result, nil
			}
			// Continue searching other directories if not found in this one
		}
	}

	return nil, fmt.Errorf("file not found: %s", targetPath)
}
