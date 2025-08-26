package provider

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxjoehnk/terraform-provider-iis/iis"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"access_key": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("IIS_ACCESS_KEY", nil),
				Description: "The access key for IIS authentication. Can also be sourced from the IIS_ACCESS_KEY environment variable.",
			},
			"host": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("IIS_HOST", nil),
				Description: "The IIS host URL. Can also be sourced from the IIS_HOST environment variable.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"iis_application_pool": resourceApplicationPool(),
			"iis_application":      resourceApplication(),
			"iis_authentication":   resourceAuthentication(),
			"iis_website":          resourceWebsite(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"iis_website":      dataSourceIisWebsite(),
			"iis_certificates": dataSourceIisCertificates(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	host := d.Get("host").(string)
	if host == "" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Missing IIS Host",
			Detail:   "The IIS host must be configured via the 'host' argument or IIS_HOST environment variable.",
		})
	}

	accessKey := d.Get("access_key").(string)
	if accessKey == "" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Missing IIS Access Key",
			Detail:   "The IIS access key must be configured via the 'access_key' argument or IIS_ACCESS_KEY environment variable.",
		})
	}

	if diags.HasError() {
		return nil, diags
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	loggingTransport := logging.NewLoggingHTTPTransport(transport)
	client := &iis.Client{
		HttpClient: http.Client{
			Transport: loggingTransport,
		},
		Host:      host,
		AccessKey: accessKey,
	}

	return client, nil
}
