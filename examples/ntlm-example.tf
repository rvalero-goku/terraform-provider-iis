# Example configuration using NTLM authentication
# This demonstrates Windows domain authentication with IIS Administration API

terraform {
  required_providers {
    iis = {
      source = "terraform.local/maxjoehnk/iis"
      # No version constraint needed with dev_overrides
    }
  }
}

# Variables for sensitive configuration
variable "iis_host" {
  description = "IIS server URL"
  type        = string
  default     = "https://your-iis-server:55539"
}

variable "iis_access_key" {
  description = "API Access Token for IIS Administration API"
  type        = string
  sensitive   = true
}

variable "iis_ntlm_username" {
  description = "Username for NTLM authentication"
  type        = string
  sensitive   = true
}

variable "iis_ntlm_password" {
  description = "Password for NTLM authentication"
  type        = string
  sensitive   = true
}

variable "iis_ntlm_domain" {
  description = "Domain for NTLM authentication (optional)"
  type        = string
  default     = ""
}

variable "iis_insecure" {
  description = "Skip TLS certificate verification"
  type        = bool
  default     = false
}

# Provider configuration with NTLM authentication AND API access token
provider "iis" {
  # IIS server configuration
  host = var.iis_host

  # API Access Token (for authorization)
  access_key = var.iis_access_key

  # NTLM Authentication (for HTTP authentication)
  ntlm_username = var.iis_ntlm_username
  ntlm_password = var.iis_ntlm_password
  ntlm_domain   = var.iis_ntlm_domain

  # Proxy configuration (optional)
  #proxy_url = var.iis_proxy_url

  # TLS configuration
  insecure = var.iis_insecure
}

# Alternative: Using provider environment variables (no variables needed)
# Set these environment variables and remove the variable blocks above:
# IIS_HOST=https://your-iis-server:55539
# IIS_ACCESS_KEY=your-access-token
# IIS_NTLM_USERNAME=your-username
# IIS_NTLM_PASSWORD=your-password
# IIS_NTLM_DOMAIN=your-domain
# IIS_INSECURE=true
#
# Then use: provider "iis" {}

# Test resources
resource "iis_application_pool" "ntlm_test" {
  name                    = "NTLMTestAppPool"
  managed_runtime_version = "v4.0"
}

resource "iis_website" "ntlm_test" {
  name             = "NTLM Test Website"
  physical_path    = "C:\\inetpub\\wwwroot"
  application_pool = iis_application_pool.ntlm_test.id

  binding {
    protocol   = "http"
    port       = 8080
    ip_address = "*"
    hostname   = ""
  }
}

# Outputs
output "app_pool_info" {
  value = {
    id   = iis_application_pool.ntlm_test.id
    name = iis_application_pool.ntlm_test.name
  }
  description = "Information about the created application pool"
}

output "website_info" {
  value = {
    id   = iis_website.ntlm_test.id
    name = iis_website.ntlm_test.name
  }
  description = "Information about the created website"
}
