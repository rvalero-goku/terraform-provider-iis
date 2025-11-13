# Automatic API Token Generation

The IIS provider automatically generates API tokens when NTLM credentials are provided without an `access_key`.

## How It Works

The IIS Administration API requires **both NTLM authentication AND an access token** for most operations. When you configure the provider with only NTLM credentials, the provider will:

1. Use NTLM credentials to authenticate
2. Automatically generate an API token using the `GenerateApiToken` method
3. Use both NTLM auth + the generated token for all subsequent operations

## Configuration

Simply provide NTLM credentials without an `access_key`:

```hcl
provider "iis" {
  host          = "https://iis-server:55539"
  ntlm_username = "admin"
  ntlm_password = var.admin_password
  ntlm_domain   = "DOMAIN"  # Optional
  insecure      = true
}

# Resources will automatically use both NTLM + auto-generated token
resource "iis_application_pool" "myapp" {
  name = "MyAppPool"
}
```

## Multi-Server Example

```hcl
locals {
  config = yamldecode(file("${path.module}/iis-config.yaml"))
}

# Default provider auto-generates token
provider "iis" {
  host          = local.config.servers.server1.host
  ntlm_username = local.config.servers.server1.ntlm_username
  ntlm_password = local.config.servers.server1.ntlm_password
  insecure      = true
}

# Aliased providers also auto-generate tokens
provider "iis" {
  alias         = "server2"
  host          = local.config.servers.server2.host
  ntlm_username = local.config.servers.server2.ntlm_username
  ntlm_password = local.config.servers.server2.ntlm_password
  insecure      = true
}

resource "iis_application_pool" "server1_pool" {
  name = "Server1Pool"
}

resource "iis_application_pool" "server2_pool" {
  provider = iis.server2
  name     = "Server2Pool"
}
```

## Logging

The provider logs token generation activity:

```
INFO: No access_key provided, auto-generating API token using NTLM credentials
INFO: Successfully auto-generated API token (token_length: 54)
```

If token generation fails, operations will be attempted with NTLM only:

```
WARN: Failed to auto-generate API token, will attempt operations with NTLM only (error: ...)
```

## When to Use `iis_api_token` Resource

The `iis_api_token` resource is still useful for:

1. **Exporting tokens** for external tools/scripts
2. **CI/CD integration** - generate tokens to use outside Terraform
3. **Explicit token management** - control token lifecycle separately

```hcl
# Generate token for external use
resource "iis_api_token" "external" {
  host          = "https://iis-server:55539"
  ntlm_username = "admin"
  ntlm_password = var.admin_password
  insecure      = true
}

# Export for external scripts
output "api_token" {
  value     = iis_api_token.external.access_token
  sensitive = true
}
```

## Benefits

1. **Simplified Configuration** - No need to manually generate tokens
2. **Automatic Token Rotation** - New token generated on each provider initialization
3. **No State Dependencies** - Provider works immediately without resource creation
4. **Better Security** - Tokens are ephemeral, regenerated as needed
5. **Backward Compatible** - Existing configurations with `access_key` continue to work

## Implementation Details

- Token generation happens in `provider.go` during provider configuration
- Uses the same `GenerateApiToken` method as the `iis_api_token` resource
- Tokens are stored in the provider client instance (not in Terraform state)
- Each provider instance (including aliased providers) generates its own token
- Token generation is retried automatically on failure (with fallback to NTLM-only)

## Troubleshooting

### 403 Forbidden Errors

If you encounter 403 errors after the provider successfully initializes:
- Check IIS Administration API is running
- Verify NTLM credentials have proper permissions
- Review IIS Administration API access control settings

### Token Generation Failures

If auto-generation fails:
- Check network connectivity to IIS server
- Verify NTLM credentials are correct
- Check IIS Administration API XSRF protection is enabled
- Review provider logs for detailed error messages

## See Also

- [NTLM Authentication Guide](NTLM-AUTH.md)
- [API Token Resource Documentation](docs/resources/api_token.md)
- [Provider Configuration](README.md#provider-configuration)
