package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxjoehnk/terraform-provider-iis/iis"
)

const directoryNameKey = "name"
const directoryParentIDKey = "parent_id"
const directoryPhysicalPathKey = "physical_path"

func resourceDirectory() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDirectoryCreate,
		ReadContext:   resourceDirectoryRead,
		DeleteContext: resourceDirectoryDelete,

		Schema: map[string]*schema.Schema{
			directoryNameKey: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the directory to create",
			},
			directoryParentIDKey: {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Parent directory ID. If not specified, creates in root location.",
			},
			directoryPhysicalPathKey: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Physical path of the created directory",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Type of the resource (always 'directory')",
			},
			"exists": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the directory exists",
			},
		},
	}
}

func resourceDirectoryCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*iis.Client)
	name := d.Get(directoryNameKey).(string)
	parentID := d.Get(directoryParentIDKey).(string)

	var parent *iis.FileRef
	if parentID != "" {
		// Parent specified by ID
		parent = &iis.FileRef{ID: parentID}
	}
	// If parent is nil, the API will use a default root location

	tflog.Debug(ctx, "Creating directory: "+name+" with parent: "+parentID)
	dir, err := client.CreateDirectory(ctx, name, parent)
	if err != nil {
		return diag.FromErr(err)
	}

	tflog.Debug(ctx, "Created directory: "+toJSON(dir))
	d.SetId(dir.ID)
	
	// Set computed attributes
	if err := d.Set(directoryPhysicalPathKey, dir.PhysicalPath); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("type", dir.Type); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("exists", dir.Exists); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceDirectoryRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*iis.Client)
	
	dir, err := client.ReadFile(ctx, d.Id())
	if err != nil {
		d.SetId("")
		return diag.FromErr(err)
	}

	tflog.Debug(ctx, "Read directory: "+toJSON(dir))
	
	if err := d.Set(directoryNameKey, dir.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set(directoryPhysicalPathKey, dir.PhysicalPath); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("type", dir.Type); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("exists", dir.Exists); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceDirectoryDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*iis.Client)
	id := d.Id()
	
	tflog.Debug(ctx, "Deleting directory: "+id)
	err := client.DeleteFile(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}
	
	tflog.Debug(ctx, "Deleted directory: "+id)
	return nil
}
