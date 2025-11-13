package provider

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/url"
	"time"

	"github.com/Azure/go-ntlmssp"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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
			"proxy_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("IIS_PROXY_URL", nil),
				Description: "The proxy URL for HTTP/HTTPS requests. Can also be sourced from the IIS_PROXY_URL environment variable. Format: http://[username:password@]proxy.example.com:port",
			},
			"insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("IIS_INSECURE", false),
				Description: "Skip TLS certificate verification. Can also be sourced from the IIS_INSECURE environment variable.",
			},
			"ntlm_username": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("IIS_NTLM_USERNAME", nil),
				Description: "Username for NTLM authentication. Can also be sourced from the IIS_NTLM_USERNAME environment variable. Use either access_key OR NTLM credentials.",
			},
			"ntlm_password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("IIS_NTLM_PASSWORD", nil),
				Description: "Password for NTLM authentication. Can also be sourced from the IIS_NTLM_PASSWORD environment variable. Use either access_key OR NTLM credentials.",
			},
			"ntlm_domain": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("IIS_NTLM_DOMAIN", nil),
				Description: "Domain for NTLM authentication. Can also be sourced from the IIS_NTLM_DOMAIN environment variable. Optional, can be empty for local accounts.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"iis_application_pool": resourceApplicationPool(),
			"iis_application":      resourceApplication(),
			"iis_authentication":   resourceAuthentication(),
			"iis_website":          resourceWebsite(),
			"iis_directory":        resourceDirectory(),
			"iis_file_copy":        resourceFileCopy(),
			"iis_api_token":        resourceApiToken(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"iis_website":      dataSourceIisWebsite(),
			"iis_certificates": dataSourceIisCertificates(),
			"iis_file":         dataSourceIisFile(),
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

	// Get authentication credentials
	accessKey := d.Get("access_key").(string)
	ntlmUsername := d.Get("ntlm_username").(string)
	ntlmPassword := d.Get("ntlm_password").(string)
	ntlmDomain := d.Get("ntlm_domain").(string)

	// Validate authentication method
	hasAccessKey := accessKey != ""
	hasNtlmCreds := ntlmUsername != "" && ntlmPassword != ""

	if !hasAccessKey && !hasNtlmCreds {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Missing Authentication Credentials",
			Detail:   "Either access_key OR NTLM credentials (username/password) must be provided. Both can be used together for IIS Administration API. Use IIS_ACCESS_KEY and/or IIS_NTLM_USERNAME/IIS_NTLM_PASSWORD environment variables.",
		})
	}

	// Note: Both access_key and NTLM credentials can be used together
	// NTLM for authentication, access_key for API authorization

	if diags.HasError() {
		return nil, diags
	}

	// Configure TLS settings
	insecure := d.Get("insecure").(bool)
	tlsConfig := &tls.Config{
		InsecureSkipVerify: insecure,
	}

	// Configure proxy if provided
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
		// Connection pool settings to improve NTLM performance
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
		// Enable keep-alive for better NTLM session persistence
		DisableKeepAlives: false,
		// Add timeouts for better reliability
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	proxyURL := d.Get("proxy_url").(string)
	if proxyURL != "" {
		parsedProxyURL, err := url.Parse(proxyURL)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Invalid Proxy URL",
				Detail:   "The proxy URL is not valid: " + err.Error(),
			})
			return nil, diags
		}
		transport.Proxy = http.ProxyURL(parsedProxyURL)
	}

	// Configure NTLM authentication if credentials are provided
	var finalTransport http.RoundTripper = transport
	if hasNtlmCreds {
		// Wrap transport with NTLM authentication
		finalTransport = &ntlmssp.Negotiator{
			RoundTripper: transport,
		}
	}

	loggingTransport := logging.NewLoggingHTTPTransport(finalTransport)
	client := &iis.Client{
		HttpClient: http.Client{
			Transport: loggingTransport,
			// Increased timeout to accommodate retries
			// Total time: 5 retries * max 16s backoff + 60s request time
			Timeout: 120 * time.Second,
		},
		Host:         host,
		AccessKey:    accessKey,
		NTLMUsername: ntlmUsername,
		NTLMPassword: ntlmPassword,
		NTLMDomain:   ntlmDomain,
	}

	// Auto-generate API token if only NTLM credentials are provided
	// IIS Administration API requires both NTLM auth + access token for most operations
	if hasNtlmCreds && !hasAccessKey {
		tflog.Info(context.Background(), "No access_key provided, auto-generating API token using NTLM credentials")
		
		token, err := client.GenerateApiToken(context.Background(), ntlmUsername, ntlmPassword, ntlmDomain)
		if err != nil {
			tflog.Warn(context.Background(), "Failed to auto-generate API token, will attempt operations with NTLM only", map[string]interface{}{
				"error": err.Error(),
			})
			// Don't fail - some operations might work with NTLM only
		} else {
			tflog.Info(context.Background(), "Successfully auto-generated API token", map[string]interface{}{
				"token_length": len(token),
			})
			client.AccessKey = token
		}
	}

	return client, nil
}
