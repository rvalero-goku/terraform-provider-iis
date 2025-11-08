# Quick Start Guide for Local Development

This guide helps you quickly set up and test the Terraform IIS provider with proxy support.

## Problem: "Could not connect to terraform.local"

This error occurs because Terraform is trying to download the provider from a remote registry, but we're developing locally.

## Solution: Development Overrides

Use Terraform's development overrides feature to point to your local binary.

### Step 1: Build and Setup for Development

```powershell
# Navigate to the project directory
cd "c:\Code\terraform-provider-iis"

# Build the provider and set up development environment
.\build.ps1 -DevSetup
```

This will:
1. Build the provider binary as `bin/terraform-provider-iis.exe`
2. Create/update your `%APPDATA%\terraform.rc` file with dev overrides
3. Set up Terraform to use your local binary

### Step 2: Test the Configuration

```powershell
# Navigate to examples directory
cd examples

# Use the dev example (no version constraints)
# Edit dev-example.tf with your IIS server details

# Initialize Terraform
terraform init

# Plan the deployment
terraform plan
```

### Step 3: Alternative Methods

If you prefer not to modify your global terraform.rc:

#### Method A: Project-level terraform.rc

Create `examples/.terraformrc`:
```hcl
provider_installation {
  dev_overrides {
    "terraform.local/maxjoehnk/iis" = "../bin"
  }
  direct {}
}
```

Then run:
```powershell
$env:TF_CLI_CONFIG_FILE = ".\.terraformrc"
terraform init
```

#### Method B: Use local filesystem mirror

1. Build and install to plugins directory:
```powershell
.\build.ps1 -Install
```

2. Use the main.tf example (with version constraints)

### Configuration Files

- **dev-example.tf** - For development with dev overrides (no version)
- **main.tf** - For registry installation (with version constraints)
- **test-proxy.tf** - Various proxy configuration examples

### Environment Variables

Set these for testing:
```powershell
$env:IIS_HOST = "https://your-iis-server:55539"
$env:IIS_ACCESS_KEY = "your-access-key"
$env:IIS_PROXY_URL = "http://proxy.company.com:8080"
$env:IIS_INSECURE = "true"
```

### Troubleshooting

1. **Provider not found after dev setup**: Check that terraform.rc exists and points to correct path
2. **Build fails**: Ensure Go is installed and in PATH
3. **Permission errors**: Run PowerShell as Administrator

### Clean Up

To restore normal Terraform behavior:
```powershell
# Remove or rename the terraform.rc file
Rename-Item "$env:APPDATA\terraform.rc" "$env:APPDATA\terraform.rc.bak"
```

### Quick Test

```powershell
# Build and setup
.\build.ps1 -DevSetup

# Test
cd examples
terraform init
terraform validate

# If successful, you should see:
# Success! The configuration is valid.
```