package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxjoehnk/terraform-provider-iis/iis"
)

const fileNameKey = "name"
const fileIDKey = "id"
const fileTypeKey = "type"
const filePhysicalPathKey = "physical_path"
const fileExistsKey = "exists"
const fileSizeKey = "size"
const fileTotalFilesKey = "total_files"

func dataSourceIisFile() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceIisFileRead,
		Schema: map[string]*schema.Schema{
			"parent_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Parent directory ID to filter files. If not specified, lists root locations.",
			},
			"website_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Website ID for web server files. If specified, uses /api/webserver/files endpoint.",
			},
			"files": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						fileNameKey: {
							Type:     schema.TypeString,
							Computed: true,
						},
						fileIDKey: {
							Type:     schema.TypeString,
							Computed: true,
						},
						fileTypeKey: {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of file: 'file' or 'directory'",
						},
						filePhysicalPathKey: {
							Type:     schema.TypeString,
							Computed: true,
						},
						fileExistsKey: {
							Type:     schema.TypeBool,
							Computed: true,
						},
						fileSizeKey: {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "File size in bytes (only for files, not directories)",
						},
						fileTotalFilesKey: {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Total number of files in directory (only for directories)",
						},
					},
				},
			},
		},
	}
}

func dataSourceIisFileRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*iis.Client)

	parentID := d.Get("parent_id").(string)
	websiteID := d.Get("website_id").(string)

	var files []iis.File
	var err error

	// Choose endpoint based on website_id parameter
	if websiteID != "" {
		files, err = client.ListWebServerFiles(ctx, websiteID)
	} else {
		files, err = client.ListFiles(ctx, parentID)
	}

	if err != nil {
		return diag.FromErr(err)
	}

	fileSet := mapFilesToSet(files)

	d.SetId(resource.UniqueId())
	if err := d.Set("files", fileSet); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func mapFilesToSet(files []iis.File) *schema.Set {
	var set []interface{}
	for _, file := range files {
		fileMap := map[string]interface{}{
			fileNameKey:         file.Name,
			fileIDKey:           file.ID,
			fileTypeKey:         file.Type,
			filePhysicalPathKey: file.PhysicalPath,
			fileExistsKey:       file.Exists,
		}
		
		// Only set size for files (not directories)
		if file.Type == "file" {
			fileMap[fileSizeKey] = int(file.Size)
		} else {
			fileMap[fileSizeKey] = 0
		}
		
		// Only set total_files for directories
		if file.Type == "directory" {
			fileMap[fileTotalFilesKey] = file.TotalFiles
		} else {
			fileMap[fileTotalFilesKey] = 0
		}
		
		set = append(set, fileMap)
	}
	return schema.NewSet(hashFile, set)
}

func hashFile(v interface{}) int {
	fileMap := v.(map[string]interface{})
	return schema.HashString(fileMap[fileIDKey].(string))
}
