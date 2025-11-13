package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxjoehnk/terraform-provider-iis/iis"
)

func dataSourceIisWebsite() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceIisWebsiteRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter websites by name. If not specified, all websites are returned.",
			},
			"websites": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"physical_path": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"application_pool_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"application_pool_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			// Deprecated: Use websites instead
			"ids": {
				Type:       schema.TypeList,
				Computed:   true,
				Elem:       &schema.Schema{Type: schema.TypeString},
				Deprecated: "Use websites attribute instead",
			},
		},
	}
}

func dataSourceIisWebsiteRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*iis.Client)

	sites, err := client.ListWebsites(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	nameFilter := d.Get("name").(string)
	siteIds := make([]string, 0)
	websiteList := make([]map[string]interface{}, 0)

	for _, site := range sites {
		// Apply name filter if specified
		if nameFilter != "" && site.Name != nameFilter {
			continue
		}

		siteIds = append(siteIds, site.ID)

		websiteMap := map[string]interface{}{
			"id":                    site.ID,
			"name":                  site.Name,
			"status":                site.Status,
			"physical_path":         site.PhysicalPath,
			"application_pool_id":   site.ApplicationPool.ID,
			"application_pool_name": site.ApplicationPool.Name,
		}
		websiteList = append(websiteList, websiteMap)
	}

	d.SetId(resource.UniqueId())
	if err := d.Set("ids", siteIds); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("websites", websiteList); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
