# IIS Terraform Provider Examples

This directory contains examples for using the IIS Terraform provider with NTLM authentication.

## Security Configuration

### Method 1: Using terraform.tfvars (Recommended)

1. Copy the example variables file:
   ```bash
   cp terraform.tfvars.example terraform.tfvars
   ```

2. Edit `terraform.tfvars` with your actual values:
   ```hcl
   iis_host          = "https://your-iis-server:55539"
   iis_access_key    = "your-actual-access-token"
   iis_ntlm_username = "your-username"
   iis_ntlm_password = "your-secure-password"
   iis_ntlm_domain   = "your-domain"  # Optional
   iis_insecure      = true           # Set to false in production
   ```

3. Run Terraform:
   ```bash
   terraform init
   terraform plan
   terraform apply
   ```

### Method 2: Using Environment Variables

Set the following environment variables and remove the `variable` blocks from the `.tf` files:

```bash
export TF_VAR_iis_host="https://your-iis-server:55539"
export TF_VAR_iis_access_key="your-actual-access-token"
export TF_VAR_iis_ntlm_username="your-username"
export TF_VAR_iis_ntlm_password="your-secure-password"
export TF_VAR_iis_ntlm_domain="your-domain"
export TF_VAR_iis_insecure="true"
```

Or on Windows PowerShell:
```powershell
$env:TF_VAR_iis_host = "https://your-iis-server:55539"
$env:TF_VAR_iis_access_key = "your-actual-access-token"
$env:TF_VAR_iis_ntlm_username = "your-username"
$env:TF_VAR_iis_ntlm_password = "your-secure-password"
$env:TF_VAR_iis_ntlm_domain = "your-domain"
$env:TF_VAR_iis_insecure = "true"
```

### Method 3: Using Provider Environment Variables

The provider also supports direct environment variables:

```bash
export IIS_HOST="https://your-iis-server:55539"
export IIS_ACCESS_KEY="your-actual-access-token"
export IIS_NTLM_USERNAME="your-username"
export IIS_NTLM_PASSWORD="your-secure-password"
export IIS_NTLM_DOMAIN="your-domain"
export IIS_INSECURE="true"
```

Then use the provider block without explicit configuration:

```hcl
provider "iis" {
  # All configuration will be read from environment variables
}
```

## Security Best Practices

1. **Never commit sensitive values** to version control
2. **Use terraform.tfvars** for local development (already in .gitignore)
3. **Use environment variables** for CI/CD pipelines
4. **Use secure secret management** systems in production (HashiCorp Vault, Azure Key Vault, etc.)
5. **Set `iis_insecure = false`** in production with valid TLS certificates

## Example Resources

The configuration creates:
- An IIS Application Pool with .NET Framework v4.0
- An IIS Website using the application pool
- Outputs with resource information

## Troubleshooting

- Ensure the IIS Administration API is enabled and accessible
- Verify NTLM authentication credentials are correct
- Check network connectivity and firewall rules
- For certificate issues, temporarily set `iis_insecure = true` for testing