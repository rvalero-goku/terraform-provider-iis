# Triggering GitHub Actions from Terraform

This guide explains how to trigger the IIS deployment GitHub Actions workflow from a Terraform deployment.

## Table of Contents

1. [Overview](#overview)
2. [Methods](#methods)
3. [Using GitHub Provider](#using-github-provider)
4. [Using Null Resource with GitHub CLI](#using-null-resource-with-github-cli)
5. [Using HTTP Provider](#using-http-provider)
6. [Using External Data Source](#using-external-data-source)
7. [Complete Examples](#complete-examples)

## Overview

There are several ways to trigger GitHub Actions workflows from Terraform:

| Method | Pros | Cons | Best For |
|--------|------|------|----------|
| GitHub Provider | Native Terraform, managed state | Requires GitHub token | Production deployments |
| GitHub CLI + null_resource | Simple, flexible | External dependency | Quick automation |
| HTTP Provider | Direct API calls | Manual API handling | Custom integrations |
| External Data Source | Real-time trigger | Less state management | Event-driven flows |

## Methods

### 1. Using GitHub Provider (Recommended)

The most native Terraform approach using the official GitHub provider.

**Setup:**

```hcl
terraform {
  required_providers {
    github = {
      source  = "integrations/github"
      version = "~> 6.0"
    }
  }
}

provider "github" {
  token = var.github_token  # Or use GITHUB_TOKEN env var
  owner = "your-org"
}
```

**Trigger Workflow Dispatch:**

```hcl
# variables.tf
variable "github_token" {
  description = "GitHub Personal Access Token with repo and workflow permissions"
  type        = string
  sensitive   = true
}

variable "iis_config" {
  description = "IIS configuration YAML"
  type        = string
}

# main.tf
resource "github_repository_dispatch" "trigger_iis_deployment" {
  repository = "your-repo"
  event_type = "deploy-iis"
  
  client_payload = jsonencode({
    config_yaml = var.iis_config
    environment = "production"
    auto_approve = true
  })
}
```

**Corresponding GitHub Workflow:**

```yaml
# .github/workflows/terraform-triggered-deploy.yml
name: IIS Deploy (Terraform Triggered)

on:
  repository_dispatch:
    types: [deploy-iis]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Deploy IIS Configuration
        uses: owner/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
        with:
          config_yaml: ${{ github.event.client_payload.config_yaml }}
          auto_approve: ${{ github.event.client_payload.auto_approve }}
```

**Full Example with Outputs:**

```hcl
# Track deployment status
resource "github_repository_dispatch" "trigger_iis_deployment" {
  repository = "infrastructure-repo"
  event_type = "deploy-iis"
  
  client_payload = jsonencode({
    config_yaml = templatefile("${path.module}/iis-config.yaml.tpl", {
      app_pool_name    = var.app_pool_name
      website_name     = var.website_name
      physical_path    = var.physical_path
      server1_host     = var.server1_host
      server1_password = var.server1_password
    })
    environment     = var.environment
    auto_approve    = var.auto_approve
    triggered_by    = "terraform"
    deployment_id   = uuid()
  })
}

output "deployment_triggered" {
  value = {
    repository   = github_repository_dispatch.trigger_iis_deployment.repository
    event_type   = github_repository_dispatch.trigger_iis_deployment.event_type
    triggered_at = timestamp()
  }
}
```

### 2. Using Null Resource with GitHub CLI

Simple approach using local-exec provisioner with GitHub CLI.

**Prerequisites:**

```bash
# Install GitHub CLI
# macOS
brew install gh

# Windows
winget install GitHub.cli

# Authenticate
gh auth login
```

**Terraform Configuration:**

```hcl
# variables.tf
variable "iis_config_file" {
  description = "Path to IIS configuration YAML file"
  type        = string
  default     = "iis-config.yaml"
}

variable "github_repo" {
  description = "GitHub repository (owner/repo)"
  type        = string
  default     = "your-org/infrastructure-repo"
}

# main.tf
resource "null_resource" "trigger_iis_deployment" {
  triggers = {
    config_hash = filemd5(var.iis_config_file)
    timestamp   = timestamp()
  }
  
  provisioner "local-exec" {
    command = <<-EOT
      gh workflow run deploy-iis.yml \
        --repo ${var.github_repo} \
        --ref main \
        --field config_yaml="$(cat ${var.iis_config_file})" \
        --field auto_approve=true \
        --field environment=production
    EOT
  }
}

output "deployment_info" {
  value = {
    triggered_at = null_resource.trigger_iis_deployment.triggers.timestamp
    config_hash  = null_resource.trigger_iis_deployment.triggers.config_hash
  }
}
```

**With Workflow Run Status Check:**

```hcl
resource "null_resource" "trigger_and_wait" {
  provisioner "local-exec" {
    command = <<-EOT
      # Trigger workflow
      RUN_ID=$(gh workflow run deploy-iis.yml \
        --repo ${var.github_repo} \
        --ref main \
        --field config_yaml="$(cat ${var.iis_config_file})" \
        --json url,databaseId \
        --jq '.databaseId')
      
      # Wait for completion
      gh run watch $RUN_ID --repo ${var.github_repo}
      
      # Check status
      STATUS=$(gh run view $RUN_ID --repo ${var.github_repo} --json conclusion --jq '.conclusion')
      if [ "$STATUS" != "success" ]; then
        echo "Workflow failed with status: $STATUS"
        exit 1
      fi
    EOT
  }
}
```

### 3. Using HTTP Provider

Direct GitHub API integration for more control.

**Setup:**

```hcl
terraform {
  required_providers {
    http = {
      source  = "hashicorp/http"
      version = "~> 3.0"
    }
  }
}
```

**Trigger Workflow via API:**

```hcl
variable "github_token" {
  description = "GitHub PAT with workflow permissions"
  type        = string
  sensitive   = true
}

variable "github_owner" {
  type    = string
  default = "your-org"
}

variable "github_repo" {
  type    = string
  default = "infrastructure-repo"
}

locals {
  iis_config = file("${path.module}/iis-config.yaml")
}

# Trigger workflow_dispatch event
resource "terraform_data" "trigger_workflow" {
  triggers_replace = {
    config_hash = md5(local.iis_config)
  }
  
  provisioner "local-exec" {
    command = <<-EOT
      curl -X POST \
        -H "Accept: application/vnd.github+json" \
        -H "Authorization: Bearer ${var.github_token}" \
        -H "X-GitHub-Api-Version: 2022-11-28" \
        https://api.github.com/repos/${var.github_owner}/${var.github_repo}/actions/workflows/deploy-iis.yml/dispatches \
        -d '{"ref":"main","inputs":{"config_yaml":"${replace(local.iis_config, "\n", "\\n")}","auto_approve":"true"}}'
    EOT
  }
}
```

**With Status Polling:**

```hcl
# Trigger workflow
resource "terraform_data" "trigger_workflow" {
  provisioner "local-exec" {
    command = <<-EOT
      # Trigger workflow
      curl -X POST \
        -H "Accept: application/vnd.github+json" \
        -H "Authorization: Bearer ${var.github_token}" \
        https://api.github.com/repos/${var.github_owner}/${var.github_repo}/actions/workflows/deploy-iis.yml/dispatches \
        -d '{"ref":"main","inputs":{"config_yaml":"${replace(local.iis_config, "\n", "\\n")}"}}'
      
      # Wait a moment for run to start
      sleep 5
      
      # Get latest run ID
      RUN_ID=$(curl -s \
        -H "Accept: application/vnd.github+json" \
        -H "Authorization: Bearer ${var.github_token}" \
        https://api.github.com/repos/${var.github_owner}/${var.github_repo}/actions/workflows/deploy-iis.yml/runs \
        | jq -r '.workflow_runs[0].id')
      
      # Poll for completion
      while true; do
        STATUS=$(curl -s \
          -H "Accept: application/vnd.github+json" \
          -H "Authorization: Bearer ${var.github_token}" \
          https://api.github.com/repos/${var.github_owner}/${var.github_repo}/actions/runs/$RUN_ID \
          | jq -r '.status')
        
        if [ "$STATUS" = "completed" ]; then
          CONCLUSION=$(curl -s \
            -H "Accept: application/vnd.github+json" \
            -H "Authorization: Bearer ${var.github_token}" \
            https://api.github.com/repos/${var.github_owner}/${var.github_repo}/actions/runs/$RUN_ID \
            | jq -r '.conclusion')
          
          if [ "$CONCLUSION" != "success" ]; then
            echo "Workflow failed: $CONCLUSION"
            exit 1
          fi
          break
        fi
        
        echo "Waiting for workflow to complete... (status: $STATUS)"
        sleep 10
      done
    EOT
  }
}
```

### 4. Using External Data Source

Query GitHub API in real-time during Terraform plan/apply.

```hcl
# Trigger workflow (using any method above)
resource "terraform_data" "trigger" {
  provisioner "local-exec" {
    command = <<-EOT
      gh workflow run deploy-iis.yml \
        --repo ${var.github_repo} \
        --field config_yaml="$(cat iis-config.yaml)" > /tmp/workflow_id.txt
    EOT
  }
}

# Query workflow status
data "external" "workflow_status" {
  depends_on = [terraform_data.trigger]
  
  program = ["bash", "-c", <<-EOT
    RUN_ID=$(cat /tmp/workflow_id.txt)
    gh api repos/${var.github_repo}/actions/runs/$RUN_ID \
      --jq '{status: .status, conclusion: .conclusion, url: .html_url}'
  EOT
  ]
}

output "workflow_result" {
  value = {
    status     = data.external.workflow_status.result.status
    conclusion = data.external.workflow_status.result.conclusion
    url        = data.external.workflow_status.result.url
  }
}
```

## Complete Examples

### Example 1: Azure VM + IIS Deployment

Deploy Azure VM, then configure IIS via GitHub Actions:

```hcl
# main.tf
terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
    github = {
      source  = "integrations/github"
      version = "~> 6.0"
    }
  }
}

provider "azurerm" {
  features {}
}

provider "github" {
  token = var.github_token
}

# Deploy Azure VM
resource "azurerm_windows_virtual_machine" "iis_server" {
  name                = "iis-server-${var.environment}"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  size                = "Standard_D2s_v3"
  admin_username      = var.admin_username
  admin_password      = var.admin_password
  
  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "Premium_LRS"
  }
  
  source_image_reference {
    publisher = "MicrosoftWindowsServer"
    offer     = "WindowsServer"
    sku       = "2022-datacenter"
    version   = "latest"
  }
}

# Install IIS
resource "azurerm_virtual_machine_extension" "iis" {
  name                 = "install-iis"
  virtual_machine_id   = azurerm_windows_virtual_machine.iis_server.id
  publisher            = "Microsoft.Compute"
  type                 = "CustomScriptExtension"
  type_handler_version = "1.10"
  
  settings = jsonencode({
    commandToExecute = <<-EOT
      powershell -Command "
        Install-WindowsFeature -Name Web-Server -IncludeManagementTools;
        Install-Module -Name IISAdministration -Force;
      "
    EOT
  })
}

# Configure IIS via GitHub Actions
resource "github_repository_dispatch" "configure_iis" {
  depends_on = [azurerm_virtual_machine_extension.iis]
  
  repository = "your-org/infrastructure"
  event_type = "deploy-iis"
  
  client_payload = jsonencode({
    config_yaml = templatefile("${path.module}/iis-config.yaml.tpl", {
      server_host     = azurerm_windows_virtual_machine.iis_server.public_ip_address
      admin_username  = var.admin_username
      admin_password  = var.admin_password
      app_pool_name   = var.app_pool_name
      website_name    = var.website_name
      website_port    = 80
      physical_path   = "C:\\inetpub\\wwwroot\\${var.website_name}"
    })
    environment  = var.environment
    auto_approve = true
  })
}

output "server_details" {
  value = {
    ip_address      = azurerm_windows_virtual_machine.iis_server.public_ip_address
    vm_name         = azurerm_windows_virtual_machine.iis_server.name
    iis_configured  = github_repository_dispatch.configure_iis.id != null
  }
}
```

**iis-config.yaml.tpl:**
```yaml
servers:
  primary:
    host: "https://${server_host}:55539"
    enabled: true
    ntlm_username: "${admin_username}"
    ntlm_password: "${admin_password}"
    insecure: true

application_pools:
  ${app_pool_name}:
    managed_runtime_version: "v4.0"
    managed_pipeline_mode: "Integrated"
    enable_32bit_app_on_win64: false
    status: "started"

websites:
  ${website_name}:
    application_pool: "${app_pool_name}"
    physical_path: "${physical_path}"
    protocol: "http"
    binding_information: "*:${website_port}:"
```

### Example 2: Multi-Environment Deployment

Deploy to dev, then trigger staging and production sequentially:

```hcl
# Deploy to dev immediately
resource "github_repository_dispatch" "deploy_dev" {
  repository = "your-org/infrastructure"
  event_type = "deploy-iis"
  
  client_payload = jsonencode({
    config_yaml  = file("configs/dev-iis-config.yaml")
    environment  = "dev"
    auto_approve = true
  })
}

# Wait for dev, then deploy to staging
resource "null_resource" "deploy_staging" {
  depends_on = [github_repository_dispatch.deploy_dev]
  
  triggers = {
    dev_deployment = github_repository_dispatch.deploy_dev.id
  }
  
  provisioner "local-exec" {
    command = <<-EOT
      # Wait for dev deployment
      sleep 30
      
      # Trigger staging
      gh workflow run deploy-iis.yml \
        --repo your-org/infrastructure \
        --ref main \
        --field config_yaml="$(cat configs/staging-iis-config.yaml)" \
        --field environment=staging \
        --field auto_approve=false
    EOT
  }
}

# Manual approval required for production
output "production_deployment_command" {
  value = <<-EOT
    To deploy to production, run:
    gh workflow run deploy-iis.yml \
      --repo your-org/infrastructure \
      --ref main \
      --field config_yaml="$(cat configs/prod-iis-config.yaml)" \
      --field environment=production \
      --field auto_approve=false
  EOT
}
```

### Example 3: Dynamic Configuration Generation

Generate IIS config based on Terraform outputs:

```hcl
locals {
  # Generate server configuration from Terraform state
  servers = {
    for vm in azurerm_windows_virtual_machine.iis_servers : 
    vm.name => {
      host           = "https://${vm.public_ip_address}:55539"
      enabled        = true
      ntlm_username  = var.admin_username
      ntlm_password  = var.admin_password
      insecure       = true
    }
  }
  
  # Generate application pool config
  app_pools = {
    for app in var.applications :
    app.name => {
      managed_runtime_version   = app.dotnet_version
      managed_pipeline_mode     = "Integrated"
      enable_32bit_app_on_win64 = app.enable_32bit
      status                    = "started"
    }
  }
  
  # Generate website config
  websites = {
    for site in var.websites :
    site.name => {
      application_pool     = site.app_pool
      physical_path        = "C:\\inetpub\\wwwroot\\${site.name}"
      protocol             = "http"
      binding_information  = "*:${site.port}:"
    }
  }
  
  # Complete IIS configuration
  iis_config = yamlencode({
    servers           = local.servers
    application_pools = local.app_pools
    websites          = local.websites
  })
}

# Trigger deployment with generated config
resource "github_repository_dispatch" "deploy_iis" {
  repository = var.github_repo
  event_type = "deploy-iis"
  
  client_payload = jsonencode({
    config_yaml  = local.iis_config
    environment  = var.environment
    auto_approve = true
  })
}

output "generated_config" {
  value     = local.iis_config
  sensitive = true
}
```

### Example 4: Terraform Cloud Integration

Using Terraform Cloud with GitHub Actions:

```hcl
# terraform.tf
terraform {
  cloud {
    organization = "your-org"
    
    workspaces {
      name = "iis-infrastructure"
    }
  }
}

# main.tf
variable "tfc_github_token" {
  description = "GitHub token stored in Terraform Cloud"
  type        = string
  sensitive   = true
}

# Use terraform_data for triggers
resource "terraform_data" "trigger_github_action" {
  triggers_replace = {
    config_version = filemd5("iis-config.yaml")
  }
  
  provisioner "local-exec" {
    command = <<-EOT
      curl -X POST \
        -H "Authorization: Bearer ${var.tfc_github_token}" \
        -H "Accept: application/vnd.github+json" \
        https://api.github.com/repos/your-org/infrastructure/dispatches \
        -d '{"event_type":"deploy-iis","client_payload":{"config_yaml":"$(cat iis-config.yaml | base64)"}}'
    EOT
  }
}
```

## GitHub Actions Workflow Configuration

For all the above methods, your GitHub Actions workflow should accept the configuration:

```yaml
# .github/workflows/deploy-iis.yml
name: Deploy IIS Configuration

on:
  workflow_dispatch:
    inputs:
      config_yaml:
        description: 'IIS Configuration YAML'
        required: true
        type: string
      environment:
        description: 'Deployment environment'
        required: false
        default: 'production'
        type: string
      auto_approve:
        description: 'Auto-approve deployment'
        required: false
        default: 'false'
        type: string
  
  repository_dispatch:
    types: [deploy-iis]

jobs:
  deploy:
    runs-on: ubuntu-latest
    environment: ${{ github.event.inputs.environment || github.event.client_payload.environment }}
    
    steps:
      - name: Deploy IIS
        uses: your-org/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
        with:
          config_yaml: ${{ github.event.inputs.config_yaml || github.event.client_payload.config_yaml }}
          auto_approve: ${{ github.event.inputs.auto_approve == 'true' || github.event.client_payload.auto_approve == true }}
```

## Security Considerations

### 1. GitHub Token Permissions

Required token permissions:
- `repo` (full control)
- `workflow` (trigger workflows)

Create fine-grained PAT:
```bash
# Via GitHub CLI
gh auth refresh -s workflow

# Or create at: https://github.com/settings/tokens
```

### 2. Store Tokens Securely

```hcl
# In Terraform Cloud/Enterprise
# Store as sensitive variable

# For local development
export TF_VAR_github_token="ghp_xxx"

# In CI/CD
# Use secrets management (Vault, AWS Secrets Manager, etc.)
```

### 3. Audit Trail

```hcl
resource "github_repository_dispatch" "deploy" {
  client_payload = jsonencode({
    config_yaml   = var.iis_config
    triggered_by  = "terraform"
    terraform_run = terraform.workspace
    timestamp     = timestamp()
    operator      = var.operator_email
  })
}
```

## Troubleshooting

### Issue: Workflow Not Triggering

Check GitHub token permissions:
```bash
gh api user -H "Authorization: token YOUR_TOKEN"
```

### Issue: Configuration Too Large

GitHub API has payload limits. For large configs, use artifact storage:

```hcl
resource "null_resource" "upload_config" {
  provisioner "local-exec" {
    command = <<-EOT
      # Upload to artifact storage
      aws s3 cp iis-config.yaml s3://your-bucket/configs/iis-config-${timestamp()}.yaml
      
      # Trigger with URL
      gh workflow run deploy-iis.yml \
        --field config_url="s3://your-bucket/configs/iis-config-${timestamp()}.yaml"
    EOT
  }
}
```

### Issue: Race Conditions

Add explicit dependencies:
```hcl
resource "github_repository_dispatch" "deploy" {
  depends_on = [
    azurerm_virtual_machine_extension.iis,
    azurerm_network_security_rule.allow_https
  ]
  # ...
}
```

## Related Documentation

- [Secure Variables](SECURE_VARIABLES.md)
- [GitHub Actions](GITHUB_ACTIONS.md)
- [Quick Start](QUICK_START_GITHUB_ACTIONS.md)
