# IIS Multi-Server Configuration using YAML
# This example shows how to manage multiple IIS servers with a single configuration
# The provider automatically generates API tokens when NTLM credentials are provided

terraform {
  required_providers {
    iis = {
      source = "terraform.local/maxjoehnk/iis"
    }
  }
}

# Load configuration from YAML
locals {
  config = yamldecode(file("${path.module}/iis-config-example.yaml"))
}

# Provider configurations
# The provider automatically generates API tokens when NTLM credentials are provided
# Default provider (required even when using only aliased providers)
provider "iis" {
  host          = local.config.servers.server1.host
  ntlm_username = local.config.servers.server1.ntlm_username
  ntlm_password = local.config.servers.server1.ntlm_password
  ntlm_domain   = try(local.config.servers.server1.ntlm_domain, "")
  insecure      = try(local.config.servers.server1.insecure, false)
}

# Provider for server1
provider "iis" {
  alias         = "server1"
  host          = local.config.servers.server1.host
  ntlm_username = local.config.servers.server1.ntlm_username
  ntlm_password = local.config.servers.server1.ntlm_password
  ntlm_domain   = try(local.config.servers.server1.ntlm_domain, "")
  insecure      = try(local.config.servers.server1.insecure, false)
}

# Provider for server2
provider "iis" {
  alias         = "server2"
  host          = local.config.servers.server2.host
  ntlm_username = local.config.servers.server2.ntlm_username
  ntlm_password = local.config.servers.server2.ntlm_password
  ntlm_domain   = try(local.config.servers.server2.ntlm_domain, "")
  insecure      = try(local.config.servers.server2.insecure, false)
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

      # Dynamically look up certificate by CN from YAML if specified
      certificate = try(binding.value.certificate_cn, null) != null ? try(
        [for cert in data.iis_certificates.server1[0].certificates : cert.id if strcontains(cert.subject, binding.value.certificate_cn)][0],
        null
      ) : null
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

      # Dynamically look up certificate by CN from YAML if specified
      certificate = try(binding.value.certificate_cn, null) != null ? try(
        [for cert in data.iis_certificates.server2[0].certificates : cert.id if strcontains(cert.subject, binding.value.certificate_cn)][0],
        null
      ) : null
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

output "server1_existing_websites" {
  value       = local.config.servers.server1.enabled && length(data.iis_website.server1_existing) > 0 ? data.iis_website.server1_existing[0].websites : []
  description = "Existing websites on server1"
}

output "server2_existing_websites" {
  value       = local.config.servers.server2.enabled && length(data.iis_website.server2_existing) > 0 ? data.iis_website.server2_existing[0].websites : []
  description = "Existing websites on server2"
}
