package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxjoehnk/terraform-provider-iis/iis"
)

const NameKey = "name"
const StatusKey = "status"

func resourceApplicationPool() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceApplicationPoolCreate,
		ReadContext:   resourceApplicationPoolRead,
		UpdateContext: resourceApplicationPoolUpdate,
		DeleteContext: resourceApplicationPoolDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			NameKey: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true, // Renaming requires recreate
			},
			StatusKey: {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "started",
			},
			"managed_runtime_version": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "v4.0",
				Description: ".NET CLR version for the app pool (e.g., v4.0, v2.0)",
			},
		},
	}
}

func resourceApplicationPoolCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*iis.Client)
	name := d.Get(NameKey).(string)
	runtimeVersion := d.Get("managed_runtime_version").(string)
	tflog.Debug(ctx, "Creating application pool: "+toJSON(name)+", runtime: "+runtimeVersion)
	pool, err := client.CreateAppPool(ctx, name, runtimeVersion)
	if err != nil {
		return diag.FromErr(err)
	}
	tflog.Debug(ctx, "Created application pool: "+toJSON(pool))
	d.SetId(pool.ID)
	
	// If user specified a status different from the created status, apply it
	if status, ok := d.GetOk(StatusKey); ok && status.(string) != "" && status.(string) != pool.Status {
		desiredStatus := status.(string)
		tflog.Debug(ctx, "Setting app pool status to: "+desiredStatus)
		_, err = client.UpdateAppPool(ctx, pool.ID, runtimeVersion, desiredStatus)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	
	return resourceApplicationPoolRead(ctx, d, m)
}

func resourceApplicationPoolRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*iis.Client)
	id := d.Id()
	appPool, err := client.ReadAppPool(ctx, id)
	if err != nil {
		d.SetId("")
		return diag.FromErr(err)
	}
	tflog.Debug(ctx, "Read application pool: "+toJSON(appPool))

       if err = d.Set(NameKey, appPool.Name); err != nil {
	       return diag.FromErr(err)
       }
       if err = d.Set(StatusKey, appPool.Status); err != nil {
	       return diag.FromErr(err)
       }
       if err = d.Set("managed_runtime_version", appPool.ManagedRuntimeVersion); err != nil {
	       return diag.FromErr(err)
       }
	return nil
}

func resourceApplicationPoolUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*iis.Client)
	id := d.Id()
	
	if d.HasChange("managed_runtime_version") || d.HasChange(StatusKey) {
		runtimeVersion := d.Get("managed_runtime_version").(string)
		status := d.Get(StatusKey).(string)
		tflog.Debug(ctx, "Updating application pool: "+toJSON(id)+", runtime: "+runtimeVersion+", status: "+status)
		
		applicationPool, err := client.UpdateAppPool(ctx, id, runtimeVersion, status)
		if err != nil {
			return diag.FromErr(err)
		}
		
		tflog.Debug(ctx, "Updated application pool: "+toJSON(applicationPool))
		
		// Re-read to update state
		return resourceApplicationPoolRead(ctx, d, m)
	}
	
	return nil
}

func resourceApplicationPoolDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*iis.Client)
	id := d.Id()
	tflog.Debug(ctx, "Deleting application pool: "+toJSON(id))
	err := client.DeleteAppPool(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}
	tflog.Debug(ctx, "Deleted application pool: "+toJSON(id))
	return nil
}
