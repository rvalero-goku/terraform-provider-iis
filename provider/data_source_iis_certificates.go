package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxjoehnk/terraform-provider-iis/iis"
)

const AliasKey = "alias"
const IdKey = "id"
const IssuedByKey = "issued_by"
const SubjectKey = "subject"
const ThumbprintKey = "thumbprint"

func dataSourceIisCertificates() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceIisCertificatesRead,
		Schema: map[string]*schema.Schema{
			"certificates": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: map[string]*schema.Schema{
					AliasKey: {
						Type:     schema.TypeString,
						Required: true,
					},
					IdKey: {
						Type:     schema.TypeString,
						Required: true,
					},
					IssuedByKey: {
						Type:     schema.TypeString,
						Required: true,
					},
					SubjectKey: {
						Type:     schema.TypeString,
						Optional: true,
					},
					ThumbprintKey: {
						Type:     schema.TypeString,
						Computed: true,
					},
				},
			},
		},
	}
}

func dataSourceIisCertificatesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*iis.Client)

	certificates, err := client.ListCertificates(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	certificateSet := mapCertificatesToSet(certificates)

	d.SetId(resource.UniqueId())
	if err := d.Set("certificates", certificateSet); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func mapCertificatesToSet(certificates []iis.Certificate) *schema.Set {
	result := schema.NewSet(schema.HashString, []interface{}{})
	for _, certificate := range certificates {
		mapped := map[string]interface{}{
			AliasKey:      certificate.Alias,
			IdKey:         certificate.ID,
			IssuedByKey:   certificate.IssuedBy,
			SubjectKey:    certificate.Subject,
			ThumbprintKey: certificate.Thumbprint,
		}
		result.Add(mapped)
	}
	return result
}
