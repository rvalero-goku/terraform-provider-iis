# NTLM Authentication Documentation

This document explains how to use NTLM authentication with the Terraform IIS provider.

## Overview

The IIS provider now supports both access key authentication and NTLM (Windows domain) authentication:

- **Access Key Authentication**: Uses API tokens/keys (original method)
- **NTLM Authentication**: Uses Windows domain or local user credentials

## Configuration Options

### NTLM Authentication Parameters

| Parameter | Description | Environment Variable | Required |
|-----------|-------------|---------------------|----------|
| `ntlm_username` | Windows username | `IIS_NTLM_USERNAME` | Yes (for NTLM) |
| `ntlm_password` | Windows password | `IIS_NTLM_PASSWORD` | Yes (for NTLM) |
| `ntlm_domain` | Windows domain | `IIS_NTLM_DOMAIN` | No (for local accounts) |

### Authentication Method Rules

1. **Either** `access_key` **OR** NTLM credentials must be provided - not both
2. For NTLM: `ntlm_username` and `ntlm_password` are required
3. `ntlm_domain` is optional (leave empty for local accounts)

## Usage Examples

### 1. Domain User Authentication

```hcl
provider "iis" {
  host = "https://iis-server.company.com:55539"
  
  # Domain authentication
  ntlm_username = "serviceaccount"
  ntlm_password = "SecurePassword123!"
  ntlm_domain   = "COMPANY"
  
  # Optional settings
  proxy_url = "http://proxy.company.com:8080"
  insecure  = true
}
```

### 2. Local User Authentication

```hcl
provider "iis" {
  host = "https://iis-server.local:55539"
  
  # Local user authentication (no domain)
  ntlm_username = "iisadmin"
  ntlm_password = "LocalPassword123!"
  # ntlm_domain is omitted for local accounts
  
  insecure = true
}
```

### 3. Environment Variables

```powershell
# Set environment variables
$env:IIS_HOST = "https://iis-server.company.com:55539"
$env:IIS_NTLM_USERNAME = "serviceaccount"
$env:IIS_NTLM_PASSWORD = "SecurePassword123!"
$env:IIS_NTLM_DOMAIN = "COMPANY"
$env:IIS_PROXY_URL = "http://proxy.company.com:8080"
$env:IIS_INSECURE = "true"
```

```hcl
# Provider will use environment variables
provider "iis" {
  # All configuration from environment variables
}
```

### 4. Mixed Environments

For environments with multiple IIS servers using different authentication methods:

```hcl
# Provider for server with access key
provider "iis" {
  alias = "api_server"
  
  host       = "https://api-iis.company.com:55539"
  access_key = "api-key-here"
}

# Provider for server with NTLM
provider "iis" {
  alias = "domain_server"
  
  host          = "https://domain-iis.company.com:55539"
  ntlm_username = "serviceaccount"
  ntlm_password = "password"
  ntlm_domain   = "COMPANY"
}
```

## Authentication Flow

### NTLM Authentication Process

1. Provider validates that NTLM credentials are provided
2. HTTP transport is configured with NTLM support
3. For each API request:
   - Username is formatted as `DOMAIN\Username` (if domain provided)
   - Basic Auth header is set with credentials
   - NTLM negotiation occurs automatically via HTTP transport

### Security Considerations

1. **Credentials Storage**: Use environment variables in production
2. **Least Privilege**: Use service accounts with minimal required permissions
3. **TLS**: Always use HTTPS in production (`insecure = false`)
4. **Proxy**: Ensure proxy supports NTLM pass-through if needed

## Troubleshooting

### Common Issues

1. **Authentication Failed (401)**
   - Check username/password accuracy
   - Verify domain name (try with and without domain)
   - Ensure account has IIS administration permissions

2. **Domain Authentication Issues**
   - Try `DOMAIN\username` format manually
   - Check if account is locked or expired
   - Verify domain trust relationships

3. **Local Account Issues**
   - Omit `ntlm_domain` parameter
   - Ensure local account has "Log on as a service" rights
   - Check Windows local security policies

4. **Proxy Authentication**
   - Ensure proxy supports NTLM pass-through
   - Test direct connection first (remove proxy_url)
   - Check proxy logs for authentication failures

### Debug Mode

Enable Terraform debug logging:

```powershell
$env:TF_LOG = "DEBUG"
terraform plan
```

### Testing Credentials

Test NTLM credentials manually using PowerShell:

```powershell
# Test basic connectivity and auth
$cred = Get-Credential
$response = Invoke-RestMethod -Uri "https://iis-server:55539/api/webserver" -Credential $cred
```

## Migration from Access Key

To migrate from access key to NTLM authentication:

1. **Backup your terraform state**
2. **Update provider configuration**:
   ```hcl
   # Before
   provider "iis" {
     host       = "https://iis-server:55539"
     access_key = "old-api-key"
   }
   
   # After
   provider "iis" {
     host          = "https://iis-server:55539"
     ntlm_username = "serviceaccount"
     ntlm_password = "password"
     ntlm_domain   = "COMPANY"
   }
   ```
3. **Test with terraform plan**
4. **Apply changes**

## Best Practices

1. **Service Accounts**: Use dedicated service accounts for automation
2. **Environment Variables**: Store credentials in environment variables, not HCL files
3. **Rotation**: Implement regular password rotation
4. **Monitoring**: Monitor authentication failures in IIS logs
5. **Testing**: Test authentication changes in non-production first

## Example Complete Configuration

```hcl
terraform {
  required_providers {
    iis = {
      source = "terraform.local/maxjoehnk/iis"
    }
  }
}

provider "iis" {
  host = "https://iis-server.company.com:55539"
  
  # NTLM Authentication
  ntlm_username = var.iis_username
  ntlm_password = var.iis_password  
  ntlm_domain   = var.iis_domain
  
  # Network Configuration
  proxy_url = var.proxy_url
  insecure  = var.skip_tls_verify
}

variable "iis_username" {
  description = "IIS service account username"
  type        = string
  sensitive   = true
}

variable "iis_password" {
  description = "IIS service account password"
  type        = string
  sensitive   = true
}

variable "iis_domain" {
  description = "Windows domain name"
  type        = string
  default     = ""
}

variable "proxy_url" {
  description = "Corporate proxy URL"
  type        = string
  default     = ""
}

variable "skip_tls_verify" {
  description = "Skip TLS certificate verification"
  type        = bool
  default     = false
}
```