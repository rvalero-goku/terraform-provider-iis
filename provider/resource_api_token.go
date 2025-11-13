package provider

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maxjoehnk/terraform-provider-iis/iis"
)

func resourceApiToken() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceApiTokenCreate,
		ReadContext:   resourceApiTokenRead,
		DeleteContext: resourceApiTokenDelete,
		
		Schema: map[string]*schema.Schema{
			"host": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "IIS Administration API host URL (e.g., https://server:55539)",
			},
			"ntlm_username": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Sensitive:   true,
				Description: "NTLM username for authentication",
			},
			"ntlm_password": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Sensitive:   true,
				Description: "NTLM password for authentication",
			},
			"ntlm_domain": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "",
				Description: "NTLM domain (optional)",
			},
			"insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Default:     false,
				Description: "Skip TLS certificate verification",
			},
			"expires_on": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "",
				Description: "Token expiration date (empty = never expires)",
			},
			"access_token": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "Generated API access token",
			},
		},
	}
}

func resourceApiTokenCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	host := d.Get("host").(string)
	username := d.Get("ntlm_username").(string)
	password := d.Get("ntlm_password").(string)
	domain := d.Get("ntlm_domain").(string)
	insecure := d.Get("insecure").(bool)
	
	tflog.Info(ctx, "Generating IIS API token", map[string]interface{}{
		"host":     host,
		"username": username,
		"insecure": insecure,
	})
	
	// Create HTTP client with appropriate TLS settings
	httpClient := http.Client{}
	if insecure {
		tflog.Warn(ctx, "TLS certificate verification disabled")
		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}
	
	// Create temporary IIS client for token generation
	// We don't need an access key yet since we're generating one
	client := iis.Client{
		HttpClient:   httpClient,
		Host:         host,
		AccessKey:    "",
		NTLMUsername: username,
		NTLMPassword: password,
		NTLMDomain:   domain,
	}
	
	// Generate API token
	tflog.Debug(ctx, "Calling GenerateApiToken")
	token, err := client.GenerateApiToken(ctx, username, password, domain)
	if err != nil {
		tflog.Error(ctx, "Failed to generate API token", map[string]interface{}{
			"error": err.Error(),
		})
		return diag.FromErr(err)
	}
	
	tflog.Info(ctx, "Successfully generated API token", map[string]interface{}{
		"token_length": len(token),
	})
	
	// Use host as ID since each token is tied to a specific server
	d.SetId(host)
	d.Set("access_token", token)
	
	return nil
}

func resourceApiTokenRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// API tokens are ephemeral - they exist if they're in state
	// We don't validate them on read to avoid unnecessary API calls
	tflog.Debug(ctx, "Reading API token resource", map[string]interface{}{
		"id": d.Id(),
	})
	return nil
}

func resourceApiTokenDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Optionally: we could revoke the token via the API here
	// For now, just remove from state
	tflog.Info(ctx, "Deleting API token resource", map[string]interface{}{
		"id": d.Id(),
	})
	d.SetId("")
	return nil
}
