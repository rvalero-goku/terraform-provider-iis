# GitHub Actions Integration Guide

This guide explains how to use the IIS Terraform Provider GitHub Actions from other repositories to automate your IIS infrastructure deployments.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Workflow Files](#workflow-files)
3. [Setup Instructions](#setup-instructions)
4. [Usage Examples](#usage-examples)
5. [Configuration](#configuration)
6. [Security](#security)
7. [Troubleshooting](#troubleshooting)

## Quick Start

### In Your Repository

1. Create `.github/workflows/deploy-iis.yml`:

```yaml
name: Deploy IIS Infrastructure

on:
  workflow_dispatch:
    inputs:
      auto_approve:
        type: boolean
        default: false

jobs:
  deploy:
    uses: rvalero-goku/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
    with:
      config_yaml: ${{ secrets.IIS_CONFIG }}
      auto_approve: ${{ inputs.auto_approve }}
```

2. Create a GitHub secret named `IIS_CONFIG` with your YAML configuration

3. Run the workflow from Actions tab

## Workflow Files

### Reusable Workflow: `apply-iis-config.yml`

The main reusable workflow that can be called from other repositories.

**Location:** `.github/workflows/apply-iis-config.yml`

**Purpose:** 
- Build the custom IIS Terraform provider
- Apply IIS configuration using OpenTofu
- Support both apply and destroy operations
- Generate and store Terraform state

**Key Features:**
- ✅ Builds provider from source (no pre-built binaries needed)
- ✅ Accepts YAML configuration as workflow input
- ✅ Supports multiple IIS servers
- ✅ Auto-approve option for CI/CD
- ✅ Artifact upload for state files
- ✅ Pull request plan comments

### Test Workflow: `test-deployment.yml`

A standalone workflow for testing in this repository.

**Location:** `.github/workflows/test-deployment.yml`

**Purpose:**
- Test the provider and workflow locally
- Validate configuration changes
- Quick development feedback

### Example Caller: `example-caller.yml`

Example implementations showing various usage patterns.

**Location:** `.github/workflows/example-caller.yml`

**Purpose:**
- Reference implementation
- Copy-paste examples for your repository
- Best practices demonstration

## Setup Instructions

### For the Provider Repository (This Repo)

No additional setup needed. The workflows are ready to use.

### For Consuming Repositories

#### Step 1: Create Workflow File

Create `.github/workflows/deploy-iis.yml` in your repository:

```yaml
name: Deploy IIS Infrastructure

on:
  workflow_dispatch:
  push:
    branches: [main]

jobs:
  deploy:
    uses: rvalero-goku/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
    with:
      config_yaml: ${{ secrets.IIS_CONFIG }}
      auto_approve: true
```

#### Step 2: Configure Secrets

Go to your repository Settings → Secrets and variables → Actions

Add secret `IIS_CONFIG` with your YAML configuration:

```yaml
servers:
  server1:
    host: "https://your-server.example.com:55539"
    enabled: true
    ntlm_username: "admin"
    ntlm_password: "your-password"
    insecure: true

application_pools:
  MyAppPool:
    managed_runtime_version: "v4.0"
    status: "started"

websites:
  MyWebsite:
    physical_path: "C:\\inetpub\\MyWebsite"
    application_pool: "MyAppPool"
    status: "started"
    bindings:
      - protocol: "http"
        port: 80
        ip_address: "*"
        hostname: ""
```

#### Step 3: Run Workflow

- Go to Actions tab
- Select "Deploy IIS Infrastructure"
- Click "Run workflow"

## Usage Examples

### Example 1: Simple Deployment

```yaml
name: Deploy IIS

on:
  push:
    branches: [main]

jobs:
  deploy:
    uses: rvalero-goku/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
    with:
      config_yaml: ${{ secrets.IIS_CONFIG }}
      auto_approve: true
```

### Example 2: Manual Approval

```yaml
name: Deploy IIS (Manual Approval)

on:
  workflow_dispatch:

jobs:
  deploy:
    uses: rvalero-goku/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
    with:
      config_yaml: ${{ secrets.IIS_CONFIG }}
      auto_approve: false  # Requires manual approval in PR
```

### Example 3: Deploy from File

```yaml
name: Deploy from Config File

on:
  push:
    paths:
      - 'infrastructure/iis-config.yaml'

jobs:
  load-and-deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Load Config
        id: config
        run: |
          echo "yaml<<EOF" >> $GITHUB_OUTPUT
          cat infrastructure/iis-config.yaml >> $GITHUB_OUTPUT
          echo "EOF" >> $GITHUB_OUTPUT
      
      - name: Deploy
        uses: rvalero-goku/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
        with:
          config_yaml: ${{ steps.config.outputs.yaml }}
          auto_approve: true
```

### Example 4: Multi-Environment

```yaml
name: Deploy Multi-Environment

on:
  workflow_dispatch:
    inputs:
      environment:
        type: choice
        options: [dev, staging, prod]

jobs:
  deploy:
    uses: rvalero-goku/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
    with:
      config_yaml: ${{ secrets[format('IIS_CONFIG_{0}', inputs.environment)] }}
      auto_approve: ${{ inputs.environment != 'prod' }}
    environment: ${{ inputs.environment }}
```

### Example 5: Scheduled Deployment

```yaml
name: Scheduled IIS Deployment

on:
  schedule:
    - cron: '0 2 * * 1'  # Every Monday at 2 AM

jobs:
  deploy:
    uses: rvalero-goku/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
    with:
      config_yaml: ${{ secrets.IIS_CONFIG }}
      auto_approve: true
```

### Example 6: Destroy Infrastructure

```yaml
name: Destroy IIS Infrastructure

on:
  workflow_dispatch:

jobs:
  destroy:
    uses: rvalero-goku/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
    with:
      config_yaml: ${{ secrets.IIS_CONFIG }}
      destroy: true
      auto_approve: true
```

## Configuration

### Workflow Inputs

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `config_yaml` | string | Required | Full YAML configuration |
| `provider_version` | string | `0.1.0` | Provider version |
| `terraform_version` | string | `latest` | OpenTofu version |
| `working_directory` | string | `.` | Work directory |
| `auto_approve` | boolean | `false` | Auto-approve apply |
| `destroy` | boolean | `false` | Destroy mode |

### YAML Configuration Structure

```yaml
servers:
  <server_name>:
    host: "<url>"
    enabled: true|false
    ntlm_username: "<username>"
    ntlm_password: "<password>"
    ntlm_domain: "<domain>"  # Optional
    insecure: true|false

application_pools:
  <pool_name>:
    managed_runtime_version: "v4.0|v2.0|No Managed Code"
    status: "started|stopped"

websites:
  <website_name>:
    physical_path: "<path>"
    application_pool: "<pool_name>"
    status: "started|stopped"
    bindings:
      - protocol: "http|https"
        port: <number>
        ip_address: "*|<ip>"
        hostname: "<hostname>"
        certificate_cn: "<cn>"  # For HTTPS only

directories:
  <directory_name>:
    parent: "<parent_name>"

file_operations:
  <operation_name>:
    source_path: "<source>"
    destination_path: "<destination>"
    move: true|false
```

## Security

### Best Practices

1. **Never commit credentials** - Always use GitHub Secrets
2. **Use environment protection** for production deployments
3. **Enable required reviewers** for production changes
4. **Use branch protection rules**
5. **Rotate credentials regularly**
6. **Use separate configs** for each environment

### Setting Up Environment Protection

1. Go to Settings → Environments
2. Create environment (e.g., "production")
3. Add protection rules:
   - ✅ Required reviewers
   - ✅ Wait timer
   - ✅ Deployment branches (only main)

```yaml
jobs:
  deploy-production:
    environment: production  # Applies protection rules
    uses: rvalero-goku/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
    with:
      config_yaml: ${{ secrets.PROD_IIS_CONFIG }}
```

### Storing Credentials

**Option 1: Repository Secrets**
- Settings → Secrets and variables → Actions
- Add secret with full YAML config

**Option 2: Environment Secrets**
- Settings → Environments → [Environment Name] → Secrets
- More secure for production

**Option 3: GitHub Secrets Manager**
- Enterprise feature
- Centralized secret management

## Troubleshooting

### Common Issues

#### 1. Provider Build Fails

**Error:** `go: cannot find main module`

**Solution:** Ensure the workflow checks out the provider repository correctly

```yaml
- uses: actions/checkout@v4
  with:
    repository: rvalero-goku/terraform-provider-iis
```

#### 2. Terraform Init Fails

**Error:** `Could not retrieve providers`

**Solution:** Verify the CLI config file path is correct and the provider binary exists

```bash
# Check in workflow
- name: Debug
  run: |
    ls -la ~/.terraform.d/plugins/terraform.local/maxjoehnk/iis/0.1.0/linux_amd64/
    cat ~/.terraform.d/terraform.rc
```

#### 3. Connection to IIS Server Fails

**Error:** `connection refused` or `timeout`

**Solution:**
- Verify server URL is accessible from GitHub Actions runners
- Check if firewall allows GitHub Actions IPs
- Consider using self-hosted runners for private networks

#### 4. Authentication Fails

**Error:** `401 Unauthorized`

**Solution:**
- Verify NTLM credentials are correct
- Check if `ntlm_domain` is required
- Ensure user has IIS management permissions

#### 5. Certificate Issues

**Error:** `x509: certificate signed by unknown authority`

**Solution:** Set `insecure: true` in server configuration for self-signed certificates

```yaml
servers:
  server1:
    host: "https://server.example.com:55539"
    insecure: true  # Skip TLS verification
```

### Debugging

Enable debug output:

```yaml
jobs:
  deploy:
    uses: rvalero-goku/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
    with:
      config_yaml: ${{ secrets.IIS_CONFIG }}
    env:
      TF_LOG: DEBUG  # Enable Terraform debug logs
```

View artifacts:

1. Go to workflow run
2. Scroll to "Artifacts" section
3. Download `terraform-state`
4. Inspect state file and plan

### Getting Help

1. Check [README.md](../README.md) for provider documentation
2. Review [examples/](../examples/) for working configurations
3. Check GitHub Actions logs for detailed error messages
4. Open an issue with:
   - Workflow YAML
   - Error messages
   - Terraform version
   - Provider version

## Advanced Topics

### Using Self-Hosted Runners

For private networks, use self-hosted runners:

```yaml
jobs:
  deploy:
    runs-on: self-hosted  # Use your runner
    uses: rvalero-goku/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
    with:
      config_yaml: ${{ secrets.IIS_CONFIG }}
```

### Remote State Backend

Modify the workflow to use remote state:

```yaml
jobs:
  configure-backend:
    runs-on: ubuntu-latest
    steps:
      - name: Create Backend Config
        run: |
          cat > backend.tf << EOF
          terraform {
            backend "azurerm" {
              resource_group_name  = "terraform-state"
              storage_account_name = "tfstate"
              container_name       = "tfstate"
              key                  = "iis.tfstate"
            }
          }
          EOF
      
      - name: Deploy
        uses: rvalero-goku/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
        with:
          config_yaml: ${{ secrets.IIS_CONFIG }}
```

### Matrix Deployments

Deploy to multiple configurations in parallel:

```yaml
jobs:
  deploy:
    strategy:
      matrix:
        environment: [dev, staging, prod]
        region: [us-east, us-west]
    uses: rvalero-goku/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
    with:
      config_yaml: ${{ secrets[format('IIS_CONFIG_{0}_{1}', matrix.environment, matrix.region)] }}
```

## Migration Guide

### From Manual Deployment

**Before:**
```powershell
cd examples
terraform apply -auto-approve
```

**After:**
1. Store config in GitHub Secret
2. Create workflow file
3. Push to repository
4. Automatic deployment on push

### From Other CI/CD

**Jenkins:**
```groovy
stage('Deploy IIS') {
    steps {
        // Replace with GitHub Actions workflow trigger
        sh 'gh workflow run deploy-iis.yml'
    }
}
```

**Azure DevOps:**
```yaml
- task: GitHubRelease@1
  inputs:
    gitHubConnection: 'github-connection'
    repositoryName: 'your-org/your-repo'
    action: 'trigger-workflow'
    workflow: 'deploy-iis.yml'
```

## Related Documentation

- [Main README](../README.md) - Provider overview
- [SETUP.md](../SETUP.md) - Local setup guide
- [QUICKSTART.md](../QUICKSTART.md) - Quick start guide
- [NTLM-AUTH.md](../NTLM-AUTH.md) - NTLM authentication details
- [Examples](../examples/) - Working configuration examples

## Contributing

To improve these workflows:

1. Fork the repository
2. Make changes in a feature branch
3. Test thoroughly
4. Submit pull request
5. Update documentation

## License

Same license as the main project.
