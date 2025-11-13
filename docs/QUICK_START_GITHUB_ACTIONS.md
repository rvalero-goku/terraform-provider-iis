# Quick Start: GitHub Actions for IIS Provider

Get your IIS infrastructure automated in 5 minutes!

## Step 1: Create Workflow File (2 minutes)

In your repository, create `.github/workflows/deploy-iis.yml`:

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

## Step 2: Create Configuration Secret (2 minutes)

1. Go to your repository Settings → Secrets and variables → Actions
2. Click "New repository secret"
3. Name: `IIS_CONFIG`
4. Value: Your YAML configuration (example below)

```yaml
servers:
  server1:
    host: "https://your-iis-server.example.com:55539"
    enabled: true
    ntlm_username: "your-username"
    ntlm_password: "your-password"
    insecure: true  # For self-signed certificates

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

## Step 3: Run Workflow (1 minute)

1. Go to Actions tab in your repository
2. Select "Deploy IIS Infrastructure"
3. Click "Run workflow"
4. Check "auto approve deployment" (for first test)
5. Click "Run workflow"

## Step 4: Monitor Deployment

Watch the workflow run:
- ✅ Provider build
- ✅ Terraform init
- ✅ Terraform plan
- ✅ Terraform apply
- ✅ Outputs

## Step 5: Verify

Check your IIS server:
- Application pool created
- Website created
- Bindings configured

## That's It! 🎉

You now have automated IIS deployments via GitHub Actions!

## Next Steps

### Add More Servers

Update your secret to include multiple servers:

```yaml
servers:
  server1:
    host: "https://server1.example.com:55539"
    enabled: true
    ntlm_username: "admin"
    ntlm_password: "password1"
  
  server2:
    host: "https://server2.example.com:55539"
    enabled: true
    ntlm_username: "admin"
    ntlm_password: "password2"

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
```

### Automatic Deployment on Push

Update your workflow to deploy automatically:

```yaml
on:
  push:
    branches: [main]
  workflow_dispatch:

jobs:
  deploy:
    uses: rvalero-goku/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
    with:
      config_yaml: ${{ secrets.IIS_CONFIG }}
      auto_approve: true  # Auto-deploy on push
```

### Multiple Environments

Create separate secrets for each environment:
- `IIS_CONFIG_DEV`
- `IIS_CONFIG_STAGING`
- `IIS_CONFIG_PROD`

```yaml
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
```

### Add HTTPS with Certificates

```yaml
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
      - protocol: "https"
        port: 443
        ip_address: "*"
        hostname: "mywebsite.example.com"
        certificate_cn: "mywebsite.example.com"  # Certificate common name
```

### Add Directories and Files

```yaml
directories:
  MyWebsiteDir:
    parent: "inetpub"

file_operations:
  copy_default_page:
    source_path: "C:\\Source\\default.aspx"
    destination_path: "C:\\inetpub\\MyWebsite"
    move: false  # Copy instead of move
```

## Full Configuration Example

Complete YAML with all features:

```yaml
servers:
  server1:
    host: "https://iis-server1.example.com:55539"
    enabled: true
    ntlm_username: "admin"
    ntlm_password: "SecurePassword123!"
    ntlm_domain: ""  # Empty for local accounts
    insecure: true

application_pools:
  MyAppPool:
    managed_runtime_version: "v4.0"
    status: "started"
  
  AnotherPool:
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
      - protocol: "https"
        port: 443
        ip_address: "*"
        hostname: "mywebsite.example.com"
        certificate_cn: "mywebsite.example.com"
  
  AnotherSite:
    physical_path: "C:\\inetpub\\AnotherSite"
    application_pool: "AnotherPool"
    status: "started"
    bindings:
      - protocol: "http"
        port: 8080
        ip_address: "*"
        hostname: "another.example.com"

directories:
  MyWebsiteDir:
    parent: "inetpub"
  
  AnotherDir:
    parent: "inetpub"

file_operations:
  copy_default:
    source_path: "C:\\Source\\default.aspx"
    destination_path: "C:\\inetpub\\MyWebsite"
    move: false
  
  copy_web_config:
    source_path: "C:\\Source\\Web.config"
    destination_path: "C:\\inetpub\\MyWebsite"
    move: false
```

## Troubleshooting

### Connection Failed

**Error:** Cannot connect to IIS server

**Solutions:**
1. Verify server URL is correct
2. Check `insecure: true` for self-signed certs
3. Verify NTLM credentials
4. Ensure port is accessible from GitHub Actions

### Authentication Failed

**Error:** 401 Unauthorized

**Solutions:**
1. Check username/password
2. Verify user has IIS admin permissions
3. For domain accounts, set `ntlm_domain`
4. For local accounts, leave `ntlm_domain` empty

### Certificate Not Found

**Error:** Certificate with CN not found

**Solutions:**
1. List certificates on server first
2. Verify certificate common name
3. Ensure certificate is in correct store
4. Check certificate_cn matches exactly

### Provider Build Failed

**Error:** Go build failed

**Solutions:**
1. Wait for workflow to complete (can take 2-3 minutes)
2. Check Go version compatibility
3. Review workflow logs for specific error
4. Try re-running the workflow

## Getting Help

1. **Documentation:** [Full Guide](GITHUB_ACTIONS.md)
2. **Templates:** [Workflow Templates](WORKFLOW_TEMPLATES.md)
3. **Examples:** See `.github/workflows/example-caller.yml`
4. **Issues:** Open an issue in the repository

## Best Practices

✅ **Use Secrets** - Never commit passwords
✅ **Test in Dev** - Test with dev servers first
✅ **Manual Approval** - Use for production
✅ **Environment Protection** - Enable for prod
✅ **Monitor Logs** - Check workflow runs
✅ **State Artifacts** - Keep for audit trail

## Common Workflows

### Daily Deployment
```yaml
on:
  schedule:
    - cron: '0 2 * * *'  # 2 AM daily
```

### Weekly Maintenance
```yaml
on:
  schedule:
    - cron: '0 2 * * 1'  # Monday 2 AM
```

### On File Change
```yaml
on:
  push:
    paths:
      - 'infrastructure/**'
```

### Manual Only
```yaml
on:
  workflow_dispatch:
```

## Success! 🚀

You've successfully automated your IIS infrastructure deployment!

Key achievements:
- ✅ No manual Terraform setup needed
- ✅ Consistent deployments via CI/CD
- ✅ Multi-server support
- ✅ Secure credential management
- ✅ Audit trail via state artifacts

## What's Next?

1. Add more servers to your configuration
2. Set up multiple environments (dev, staging, prod)
3. Enable automatic deployments on push
4. Add notifications (Slack, Teams, Email)
5. Configure environment protection rules
6. Explore advanced features in the full documentation

Happy automating! 🎉
