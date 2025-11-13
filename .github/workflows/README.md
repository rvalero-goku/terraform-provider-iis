# GitHub Actions for IIS Terraform Provider

This directory contains GitHub Actions workflows for deploying IIS infrastructure using the custom Terraform provider.

## Workflows

### 1. `apply-iis-config.yml` - Reusable IIS Configuration Workflow

A reusable workflow that can be called from other repositories to apply IIS configurations using Terraform/OpenTofu.

**Features:**
- Builds the custom IIS Terraform provider from source
- Accepts YAML configuration as input
- Supports both apply and destroy operations
- Automatic plan generation with PR comments
- State file artifact upload
- Customizable Terraform/OpenTofu versions

**Inputs:**

| Input | Required | Default | Description |
|-------|----------|---------|-------------|
| `config_yaml` | Yes | - | IIS configuration YAML content |
| `provider_version` | No | `0.1.0` | IIS provider version to use |
| `terraform_version` | No | `latest` | Terraform/OpenTofu version |
| `working_directory` | No | `.` | Working directory for Terraform |
| `auto_approve` | No | `false` | Auto approve the Terraform apply |
| `destroy` | No | `false` | Destroy infrastructure instead of apply |

**Outputs:**

| Output | Description |
|--------|-------------|
| `terraform_output` | Terraform output JSON containing deployed resources |

**Secrets:**

| Secret | Required | Description |
|--------|----------|-------------|
| `ADDITIONAL_SECRETS` | No | Additional secrets if needed for IIS servers |

### 2. `example-caller.yml` - Example Usage

An example workflow demonstrating how to call the reusable workflow from another repository.

## Usage Examples

### Example 1: Manual Deployment with Inline YAML

```yaml
name: Deploy IIS

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
      config_yaml: |
        servers:
          server1:
            host: "https://iis-server1.example.com:55539"
            enabled: true
            ntlm_username: "admin"
            ntlm_password: "password123"
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
                hostname: "mywebsite.example.com"
      
      provider_version: "0.1.0"
      auto_approve: ${{ inputs.auto_approve }}
```

### Example 2: Deploy from Config File

```yaml
name: Deploy IIS from Config

on:
  push:
    branches: [main]
    paths:
      - 'infrastructure/iis-config.yaml'

jobs:
  load-config:
    runs-on: ubuntu-latest
    outputs:
      config: ${{ steps.read.outputs.yaml }}
    steps:
      - uses: actions/checkout@v4
      - name: Read Config
        id: read
        run: |
          echo "yaml<<EOF" >> $GITHUB_OUTPUT
          cat infrastructure/iis-config.yaml >> $GITHUB_OUTPUT
          echo "EOF" >> $GITHUB_OUTPUT
  
  deploy:
    needs: load-config
    uses: rvalero-goku/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
    with:
      config_yaml: ${{ needs.load-config.outputs.config }}
      auto_approve: true
```

### Example 3: Destroy Infrastructure

```yaml
name: Destroy IIS Infrastructure

on:
  workflow_dispatch:

jobs:
  destroy:
    uses: rvalero-goku/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
    with:
      config_yaml: |
        # Your IIS configuration here
        servers:
          server1:
            host: "https://iis-server.example.com:55539"
            enabled: true
            ntlm_username: "admin"
            ntlm_password: "password123"
        application_pools:
          TestPool:
            managed_runtime_version: "v4.0"
            status: "started"
      
      destroy: true
      auto_approve: true
```

### Example 4: Multi-Environment Deployment

```yaml
name: Deploy to Multiple Environments

on:
  workflow_dispatch:
    inputs:
      environment:
        type: choice
        options:
          - dev
          - staging
          - production

jobs:
  deploy-dev:
    if: inputs.environment == 'dev'
    uses: rvalero-goku/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
    with:
      config_yaml: ${{ secrets.DEV_IIS_CONFIG }}
      auto_approve: true
  
  deploy-staging:
    if: inputs.environment == 'staging'
    uses: rvalero-goku/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
    with:
      config_yaml: ${{ secrets.STAGING_IIS_CONFIG }}
      auto_approve: false  # Require manual approval
  
  deploy-production:
    if: inputs.environment == 'production'
    uses: rvalero-goku/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
    with:
      config_yaml: ${{ secrets.PROD_IIS_CONFIG }}
      auto_approve: false  # Require manual approval
    environment: production  # Use GitHub environment protection
```

## Configuration YAML Format

The `config_yaml` input should follow this structure:

```yaml
servers:
  server_name:
    host: "https://server.example.com:55539"
    enabled: true
    ntlm_username: "username"
    ntlm_password: "password"
    ntlm_domain: ""  # Optional
    insecure: true  # Skip TLS verification

application_pools:
  PoolName:
    managed_runtime_version: "v4.0"
    status: "started"

websites:
  WebsiteName:
    physical_path: "C:\\inetpub\\WebsiteName"
    application_pool: "PoolName"
    status: "started"
    bindings:
      - protocol: "http"
        port: 80
        ip_address: "*"
        hostname: "example.com"
      - protocol: "https"
        port: 443
        ip_address: "*"
        hostname: "example.com"
        certificate_cn: "example.com"

directories:
  DirectoryName:
    parent: "inetpub"

file_operations:
  operation_name:
    source_path: "C:\\source\\file.aspx"
    destination_path: "C:\\inetpub\\WebsiteName"
    move: false
```

## Security Best Practices

### 1. Store Sensitive Data in Secrets

**Never** commit passwords or sensitive data in your YAML configuration. Use GitHub Secrets:

```yaml
# Store the entire config in a secret
jobs:
  deploy:
    uses: rvalero-goku/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
    with:
      config_yaml: ${{ secrets.IIS_CONFIG_YAML }}
```

### 2. Use Environment Protection Rules

For production deployments, use GitHub environment protection:

```yaml
jobs:
  deploy:
    environment: production  # Requires approval
    uses: rvalero-goku/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
    with:
      config_yaml: ${{ secrets.PROD_IIS_CONFIG }}
```

### 3. Limit Auto-Approve

Only use `auto_approve: true` for non-production environments:

```yaml
jobs:
  deploy-dev:
    uses: rvalero-goku/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
    with:
      config_yaml: ${{ secrets.DEV_CONFIG }}
      auto_approve: true  # OK for dev
  
  deploy-prod:
    uses: rvalero-goku/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
    with:
      config_yaml: ${{ secrets.PROD_CONFIG }}
      auto_approve: false  # Require approval for prod
```

## Artifacts

The workflow automatically uploads the following artifacts:

- `terraform-state`: Terraform state file and plan
- Retention: 30 days

## Troubleshooting

### Provider Build Fails

If the provider build fails, check:
- Go version compatibility (requires Go 1.24+)
- Dependencies in `go.mod`
- Build errors in the action logs

### Terraform Init Fails

If `terraform init` fails:
- Verify the provider is built correctly
- Check the CLI config file is created properly
- Review the provider installation path

### Connection Errors

If IIS connection fails:
- Verify server URLs are correct
- Check NTLM credentials
- Ensure `insecure: true` is set for self-signed certificates
- Verify network connectivity from GitHub Actions runners

### State File Issues

The workflow uses local state by default. For production:
- Consider using remote state (S3, Azure Blob, etc.)
- Modify the workflow to configure backend
- Use state locking for concurrent deployments

## Advanced Configuration

### Custom Terraform Backend

To use a remote backend, modify your caller workflow:

```yaml
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Create Backend Config
        run: |
          cat > backend.tf << EOF
          terraform {
            backend "s3" {
              bucket = "my-terraform-state"
              key    = "iis/terraform.tfstate"
              region = "us-east-1"
            }
          }
          EOF
      
      - uses: rvalero-goku/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
        with:
          config_yaml: ${{ secrets.IIS_CONFIG }}
```

### Matrix Deployments

Deploy to multiple servers in parallel:

```yaml
jobs:
  deploy:
    strategy:
      matrix:
        server: [server1, server2, server3]
    uses: rvalero-goku/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
    with:
      config_yaml: ${{ secrets[format('IIS_CONFIG_{0}', matrix.server)] }}
```

## Contributing

When modifying these workflows:

1. Test changes in a fork first
2. Use proper versioning (tags or branches)
3. Update documentation
4. Follow GitHub Actions best practices

## Support

For issues or questions:
- Open an issue in the repository
- Check existing documentation in `/docs`
- Review example configurations in `/examples`
