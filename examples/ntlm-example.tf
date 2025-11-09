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

# Data source to fetch available certificates
data "iis_certificates" "available" {}

# Data source: List root file locations
# These are the configured root directories in appsettings.json
data "iis_file" "root_locations" {
  # No parent_id means list root locations
}

# Data source: List files in a specific directory
# Uncomment and set parent_id to browse a specific directory
# data "iis_file" "directory_files" {
#   parent_id = "PARENT_DIRECTORY_ID_HERE"
# }

# Data source: List web server files for a website
# This uses the /api/webserver/files endpoint
# data "iis_file" "website_files" {
#   website_id = iis_website.ntlm_test.id
# }

# Test resources
resource "iis_application_pool" "ntlm_test" {
  name                    = "NTLMTestAppPool"
  managed_runtime_version = "v4.0"
}

# Example 1: HTTP-only website
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

# Example 2: HTTPS website with certificate
# To use this, uncomment and update the certificate reference
resource "iis_website" "https_example" {
  name             = "HTTPS Example Website"
  physical_path    = "C:\\inetpub\\wwwroot"
  application_pool = iis_application_pool.ntlm_test.id

  # HTTP binding (redirect to HTTPS in production)
  binding {
    protocol   = "http"
    port       = 80
    ip_address = "*"
    hostname   = "example.com"
  }

  # HTTPS binding with certificate
  binding {
    protocol   = "https"
    port       = 443
    ip_address = "*"
    hostname   = "example.com"
    # Use certificate ID from the data source
    # Find the desired certificate from: data.iis_certificates.available.certificates
    certificate = tolist(data.iis_certificates.available.certificates)[0].id
    #certificate = "YOUR_CERTIFICATE_ID_HERE"
  }
}

# Resource: Create a new directory
# This example shows how to use a data source to find the parent directory
# and then create a subdirectory within it
#
# Uncomment the following to create a directory:
#
locals {
  # Convert set to list to access the first element
  root_locations_list = tolist(data.iis_file.root_locations.files)
  # Use the first root location as parent
  first_root_location = length(local.root_locations_list) > 0 ? local.root_locations_list[0] : null
}

resource "iis_directory" "my_app_dir" {
  name      = "terraform-test-dir"
  parent_id = local.first_root_location.id
}

output "created_directory" {
  value = {
    id            = iis_directory.my_app_dir.id
    name          = iis_directory.my_app_dir.name
    physical_path = iis_directory.my_app_dir.physical_path
  }
  description = "Information about the created directory"
}

# Outputs
output "available_certificates" {
  value = [
    for cert in data.iis_certificates.available.certificates : {
      id         = cert.id
      alias      = cert.alias
      subject    = cert.subject
      issued_by  = cert.issued_by
      thumbprint = cert.thumbprint
    }
  ]
  description = "List of available certificates on the IIS server"
}

output "root_file_locations" {
  value = [
    for file in data.iis_file.root_locations.files : {
      name          = file.name
      id            = file.id
      type          = file.type
      physical_path = file.physical_path
      exists        = file.exists
      total_files   = file.total_files
    }
  ]
  description = "List of root file locations configured in IIS"
}

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
