# IIS Multi-Server Configuration using YAML
# This example shows how to manage multiple IIS servers with a single configuration

terraform {
  required_providers {
    iis = {
      source = "terraform.local/maxjoehnk/iis"
    }
  }
}

# Variables for authentication
variable "iis_access_key" {
  description = "API Access Token for IIS Administration API"
  type        = string
  sensitive   = true
  default     = "" # Will be overridden by YAML if set
}

variable "iis_ntlm_username" {
  description = "Username for NTLM authentication"
  type        = string
  sensitive   = true
  default     = "" # Will be overridden by YAML if set
}

variable "iis_ntlm_password" {
  description = "Password for NTLM authentication"
  type        = string
  sensitive   = true
  default     = "" # Will be overridden by YAML if set
}

variable "iis_ntlm_domain" {
  description = "Domain for NTLM authentication"
  type        = string
  default     = ""
}

variable "iis_insecure" {
  description = "Skip TLS certificate verification"
  type        = bool
  default     = false
}

# Load configuration from YAML
locals {
  config = yamldecode(file("${path.module}/iis-config.yaml"))

  # Helper function to get server config value or fallback to variable
  server1_access_key    = try(local.config.servers.server1.access_key, "") != "" ? local.config.servers.server1.access_key : var.iis_access_key
  server1_ntlm_username = try(local.config.servers.server1.ntlm_username, "") != "" ? local.config.servers.server1.ntlm_username : var.iis_ntlm_username
  server1_ntlm_password = try(local.config.servers.server1.ntlm_password, "") != "" ? local.config.servers.server1.ntlm_password : var.iis_ntlm_password
  server1_ntlm_domain   = try(local.config.servers.server1.ntlm_domain, "") != "" ? local.config.servers.server1.ntlm_domain : var.iis_ntlm_domain
  server1_insecure      = try(local.config.servers.server1.insecure, var.iis_insecure)

  server2_access_key    = try(local.config.servers.server2.access_key, "") != "" ? local.config.servers.server2.access_key : var.iis_access_key
  server2_ntlm_username = try(local.config.servers.server2.ntlm_username, "") != "" ? local.config.servers.server2.ntlm_username : var.iis_ntlm_username
  server2_ntlm_password = try(local.config.servers.server2.ntlm_password, "") != "" ? local.config.servers.server2.ntlm_password : var.iis_ntlm_password
  server2_ntlm_domain   = try(local.config.servers.server2.ntlm_domain, "") != "" ? local.config.servers.server2.ntlm_domain : var.iis_ntlm_domain
  server2_insecure      = try(local.config.servers.server2.insecure, var.iis_insecure)

  # Certificate lookup helpers - finds certificate ID by CN (requires data sources to be loaded)
  server1_wildcard_cert = try(
    [for cert in data.iis_certificates.server1[0].certificates : cert.id if strcontains(cert.subject, "loved-quagga.projectpulse.me")][0],
    null
  )

  server2_wildcard_cert = try(
    [for cert in data.iis_certificates.server2[0].certificates : cert.id if strcontains(cert.subject, "loved-quagga.projectpulse.me")][0],
    null
  )
} # Provider configurations - must be defined statically
# Default provider (required even when using only aliased providers)
provider "iis" {
  host          = local.config.servers.server1.host
  access_key    = local.server1_access_key
  ntlm_username = local.server1_ntlm_username
  ntlm_password = local.server1_ntlm_password
  ntlm_domain   = local.server1_ntlm_domain
  insecure      = local.server1_insecure
}

# Provider for server1
provider "iis" {
  alias         = "server1"
  host          = local.config.servers.server1.host
  access_key    = local.server1_access_key
  ntlm_username = local.server1_ntlm_username
  ntlm_password = local.server1_ntlm_password
  ntlm_domain   = local.server1_ntlm_domain
  insecure      = local.server1_insecure
}

# Provider for server2
provider "iis" {
  alias         = "server2"
  host          = local.config.servers.server2.host
  access_key    = local.server2_access_key
  ntlm_username = local.server2_ntlm_username
  ntlm_password = local.server2_ntlm_password
  ntlm_domain   = local.server2_ntlm_domain
  insecure      = local.server2_insecure
}

# ============================================================================
# SERVER 1 Resources
# ============================================================================

# Data sources for server1
data "iis_certificates" "server1" {
  count = local.config.servers.server1.enabled ? 1 : 0
}

data "iis_file" "server1" {
  count = local.config.servers.server1.enabled ? 1 : 0
}

data "iis_website" "server1_existing" {
  count = local.config.servers.server1.enabled ? 1 : 0
}

# Import blocks removed - Default Web Site and DefaultAppPool were destroyed

# Application pools for server1
# NOTE: DefaultAppPool is a system resource - avoid destroying it in production
resource "iis_application_pool" "server1" {
  for_each = local.config.servers.server1.enabled ? local.config.application_pools : {}

  name                    = each.key
  managed_runtime_version = each.value.managed_runtime_version
  status                  = each.value.status
}

# Directories for server1
resource "iis_directory" "server1" {
  for_each = local.config.servers.server1.enabled ? local.config.directories : {}

  name = each.key
  parent_id = [
    for file in data.iis_file.server1[0].files :
    file.id if file.name == each.value.parent
  ][0]
}

# Websites for server1
# NOTE: "Default Web Site" is a system resource - avoid destroying it in production
resource "iis_website" "server1" {
  for_each = local.config.servers.server1.enabled ? local.config.websites : {}

  name             = each.key
  physical_path    = each.value.physical_path
  application_pool = iis_application_pool.server1[each.value.application_pool].id
  status           = each.value.status

  dynamic "binding" {
    for_each = each.value.bindings
    content {
      protocol   = binding.value.protocol
      port       = binding.value.port
      ip_address = try(binding.value.ip_address, "*")
      hostname   = try(binding.value.hostname, "")

      # Use certificate if specified by CN
      certificate = try(binding.value.certificate_cn, null) != null ? local.server1_wildcard_cert : (
        try(binding.value.use_certificate, false) && length(data.iis_certificates.server1) > 0 ? tolist(
          data.iis_certificates.server1[0].certificates
        )[0].id : null
      )
    }
  }

  depends_on = [
    iis_application_pool.server1,
    iis_directory.server1
  ]
}

# File operations for server1 - now using paths directly (provider will resolve IDs)
resource "iis_file_copy" "server1" {
  for_each = local.config.servers.server1.enabled && try(local.config.file_operations, null) != null ? local.config.file_operations : {}
  provider = iis.server1

  source_path      = each.value.source_path
  destination_path = each.value.destination_path
  move             = try(each.value.move, false)

  depends_on = [
    iis_directory.server1
  ]
}

# ============================================================================
# SERVER 2 Resources
# ============================================================================

# Data sources for server2
data "iis_certificates" "server2" {
  count    = local.config.servers.server2.enabled ? 1 : 0
  provider = iis.server2
}

data "iis_file" "server2" {
  count    = local.config.servers.server2.enabled ? 1 : 0
  provider = iis.server2
}

data "iis_website" "server2_existing" {
  count    = local.config.servers.server2.enabled ? 1 : 0
  provider = iis.server2
}

# Import blocks removed - Default Web Site and DefaultAppPool were destroyed

# Application pools for server2
resource "iis_application_pool" "server2" {
  for_each = local.config.servers.server2.enabled ? local.config.application_pools : {}
  provider = iis.server2

  name                    = each.key
  managed_runtime_version = each.value.managed_runtime_version
  status                  = each.value.status
}

# Directories for server2
resource "iis_directory" "server2" {
  for_each = local.config.servers.server2.enabled ? local.config.directories : {}
  provider = iis.server2

  name = each.key
  parent_id = [
    for file in data.iis_file.server2[0].files :
    file.id if file.name == each.value.parent
  ][0]
}

# Websites for server2
resource "iis_website" "server2" {
  for_each = local.config.servers.server2.enabled ? local.config.websites : {}
  provider = iis.server2

  name             = each.key
  physical_path    = each.value.physical_path
  application_pool = iis_application_pool.server2[each.value.application_pool].id
  status           = each.value.status

  dynamic "binding" {
    for_each = each.value.bindings
    content {
      protocol   = binding.value.protocol
      port       = binding.value.port
      ip_address = try(binding.value.ip_address, "*")
      hostname   = try(binding.value.hostname, "")

      # Use certificate if specified by CN
      certificate = try(binding.value.certificate_cn, null) != null ? local.server2_wildcard_cert : (
        try(binding.value.use_certificate, false) && length(data.iis_certificates.server2) > 0 ? tolist(
          data.iis_certificates.server2[0].certificates
        )[0].id : null
      )
    }
  }

  depends_on = [
    iis_application_pool.server2,
    iis_directory.server2
  ]
}

# File operations for server2 - now using paths directly (provider will resolve IDs)
resource "iis_file_copy" "server2" {
  for_each = local.config.servers.server2.enabled && try(local.config.file_operations, null) != null ? local.config.file_operations : {}
  provider = iis.server2

  source_path      = each.value.source_path
  destination_path = each.value.destination_path
  move             = try(each.value.move, false)

  depends_on = [
    iis_directory.server2
  ]
}

# ============================================================================
# Outputs
# ============================================================================

output "server1_resources" {
  value = local.config.servers.server1.enabled ? {
    host              = local.config.servers.server1.host
    application_pools = keys(iis_application_pool.server1)
    websites          = keys(iis_website.server1)
    directories       = keys(iis_directory.server1)
  } : null
  description = "Resources deployed to server1"
}

output "server2_resources" {
  value = local.config.servers.server2.enabled ? {
    host              = local.config.servers.server2.host
    application_pools = keys(iis_application_pool.server2)
    websites          = keys(iis_website.server2)
    directories       = keys(iis_directory.server2)
  } : null
  description = "Resources deployed to server2"
}

# Debug outputs for certificates
output "server1_certificates" {
  value = local.config.servers.server1.enabled && length(data.iis_certificates.server1) > 0 ? [
    for cert in data.iis_certificates.server1[0].certificates : {
      id      = cert.id
      subject = cert.subject
      alias   = cert.alias
    }
  ] : []
  description = "Available certificates on server1"
}

output "server2_certificates" {
  value = local.config.servers.server2.enabled && length(data.iis_certificates.server2) > 0 ? [
    for cert in data.iis_certificates.server2[0].certificates : {
      id      = cert.id
      subject = cert.subject
      alias   = cert.alias
    }
  ] : []
  description = "Available certificates on server2"
}

output "selected_certificate_ids" {
  value = {
    server1 = local.server1_wildcard_cert
    server2 = local.server2_wildcard_cert
  }
  description = "Selected wildcard certificate IDs"
}

output "server1_existing_websites" {
  value       = local.config.servers.server1.enabled && length(data.iis_website.server1_existing) > 0 ? data.iis_website.server1_existing[0].websites : []
  description = "Existing websites on server1"
}

output "server2_existing_websites" {
  value       = local.config.servers.server2.enabled && length(data.iis_website.server2_existing) > 0 ? data.iis_website.server2_existing[0].websites : []
  description = "Existing websites on server2"
}
