# GitHub Actions Workflow Templates

Copy these templates to your repository's `.github/workflows/` directory.

## Template 1: Basic Deployment

**File:** `.github/workflows/deploy-iis.yml`

```yaml
name: Deploy IIS Infrastructure

on:
  workflow_dispatch:
    inputs:
      auto_approve:
        description: 'Auto approve deployment'
        type: boolean
        default: false

jobs:
  deploy:
    uses: rvalero-goku/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
    with:
      config_yaml: ${{ secrets.IIS_CONFIG }}
      auto_approve: ${{ inputs.auto_approve }}
```

**Required Secrets:**
- `IIS_CONFIG` - Full YAML configuration for your IIS infrastructure

---

## Template 2: Deploy on Push

**File:** `.github/workflows/deploy-iis-on-push.yml`

```yaml
name: Deploy IIS on Push

on:
  push:
    branches:
      - main
    paths:
      - 'infrastructure/iis/**'

jobs:
  deploy:
    uses: rvalero-goku/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
    with:
      config_yaml: ${{ secrets.IIS_CONFIG }}
      auto_approve: true
```

---

## Template 3: Multi-Environment

**File:** `.github/workflows/deploy-iis-multi-env.yml`

```yaml
name: Deploy IIS - Multi Environment

on:
  workflow_dispatch:
    inputs:
      environment:
        description: 'Target environment'
        type: choice
        required: true
        options:
          - dev
          - staging
          - production
      auto_approve:
        description: 'Auto approve deployment'
        type: boolean
        default: false

jobs:
  deploy:
    uses: rvalero-goku/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
    with:
      config_yaml: ${{ secrets[format('IIS_CONFIG_{0}', inputs.environment)] }}
      auto_approve: ${{ inputs.auto_approve }}
    environment: ${{ inputs.environment }}
```

**Required Secrets:**
- `IIS_CONFIG_dev` - Development environment configuration
- `IIS_CONFIG_staging` - Staging environment configuration
- `IIS_CONFIG_production` - Production environment configuration

---

## Template 4: Deploy from File

**File:** `.github/workflows/deploy-iis-from-file.yml`

```yaml
name: Deploy IIS from Config File

on:
  push:
    branches:
      - main
    paths:
      - 'infrastructure/iis-config.yaml'

jobs:
  load-config:
    runs-on: ubuntu-latest
    outputs:
      config: ${{ steps.read.outputs.yaml }}
    steps:
      - name: Checkout
        uses: actions/checkout@v4
      
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

**Required Files:**
- `infrastructure/iis-config.yaml` - IIS configuration in your repository

---

## Template 5: Destroy Infrastructure

**File:** `.github/workflows/destroy-iis.yml`

```yaml
name: Destroy IIS Infrastructure

on:
  workflow_dispatch:
    inputs:
      confirm:
        description: 'Type "destroy" to confirm'
        required: true
      environment:
        description: 'Environment to destroy'
        type: choice
        options:
          - dev
          - staging

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - name: Validate Confirmation
        if: inputs.confirm != 'destroy'
        run: |
          echo "❌ Destruction not confirmed"
          exit 1
  
  destroy:
    needs: validate
    uses: rvalero-goku/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
    with:
      config_yaml: ${{ secrets[format('IIS_CONFIG_{0}', inputs.environment)] }}
      destroy: true
      auto_approve: true
```

---

## Template 6: Scheduled Maintenance

**File:** `.github/workflows/scheduled-iis-maintenance.yml`

```yaml
name: Scheduled IIS Maintenance

on:
  schedule:
    # Run every Monday at 2 AM UTC
    - cron: '0 2 * * 1'
  workflow_dispatch:

jobs:
  deploy:
    uses: rvalero-goku/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
    with:
      config_yaml: ${{ secrets.IIS_CONFIG }}
      auto_approve: true
  
  notify:
    needs: deploy
    runs-on: ubuntu-latest
    if: always()
    steps:
      - name: Send Notification
        run: |
          echo "Deployment Status: ${{ needs.deploy.result }}"
          # Add your notification logic here (Slack, Teams, Email, etc.)
```

---

## Template 7: Pull Request Preview

**File:** `.github/workflows/pr-preview-iis.yml`

```yaml
name: IIS Configuration Preview

on:
  pull_request:
    paths:
      - 'infrastructure/iis-config.yaml'

jobs:
  preview:
    uses: rvalero-goku/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
    with:
      config_yaml: ${{ secrets.IIS_CONFIG_DEV }}
      auto_approve: false  # Only plan, don't apply
```

---

## Template 8: Matrix Deployment

**File:** `.github/workflows/deploy-iis-matrix.yml`

```yaml
name: Deploy IIS - Matrix

on:
  workflow_dispatch:

jobs:
  deploy:
    strategy:
      matrix:
        environment: [dev, staging]
        region: [east, west]
      fail-fast: false
    uses: rvalero-goku/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
    with:
      config_yaml: ${{ secrets[format('IIS_CONFIG_{0}_{1}', matrix.environment, matrix.region)] }}
      auto_approve: true
```

**Required Secrets:**
- `IIS_CONFIG_dev_east`
- `IIS_CONFIG_dev_west`
- `IIS_CONFIG_staging_east`
- `IIS_CONFIG_staging_west`

---

## Quick Setup Checklist

- [ ] Copy desired template to `.github/workflows/` in your repository
- [ ] Create required GitHub secrets
- [ ] Update workflow parameters as needed
- [ ] Commit and push workflow file
- [ ] Test workflow from Actions tab

## Example Secret Configuration

Example YAML to store in GitHub Secrets:

```yaml
servers:
  server1:
    host: "https://iis-server1.example.com:55539"
    enabled: true
    ntlm_username: "admin"
    ntlm_password: "your-secure-password"
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
```

## Additional Resources

- [Full Documentation](../docs/GITHUB_ACTIONS.md)
- [Main README](../README.md)
- [Examples](../examples/)
