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
- An IIS Website with HTTP binding
- Example HTTPS binding configuration (commented out)
- File system browsing via Files API

### Files API

The provider supports the IIS Administration Files API for managing directories and browsing the file system:

#### List Root Locations

```hcl
# List configured root file locations
data "iis_file" "root_locations" {
  # No parent_id means list root locations
}

output "root_file_locations" {
  value = [
    for file in data.iis_file.root_locations.files : {
      name          = file.name
      type          = file.type
      physical_path = file.physical_path
    }
  ]
}
```

#### Browse a Directory

```hcl
# List files in a specific directory
data "iis_file" "directory_files" {
  parent_id = "YOUR_PARENT_DIRECTORY_ID"
}
```

#### Browse Web Server Files

```hcl
# List files for a specific website (virtual file structure)
data "iis_file" "website_files" {
  website_id = iis_website.ntlm_test.id
}
```

#### Create a Directory

```hcl
resource "iis_directory" "my_app" {
  name      = "myapp"
  parent_id = "PARENT_DIRECTORY_ID"
}
```

**Note:** File locations must be configured in the IIS Administration API `appsettings.json` file. By default, only specific root directories are accessible through the API.

### HTTP Binding

The basic example creates a website with an HTTP binding on port 8080:

```hcl
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
```

### HTTPS Binding with SSL Certificate

To create an HTTPS binding, you need to reference a certificate. The example includes a data source to list available certificates:

```hcl
# Fetch available certificates
data "iis_certificates" "available" {}

resource "iis_website" "https_example" {
  name             = "HTTPS Example Website"
  physical_path    = "C:\\inetpub\\wwwroot"
  application_pool = iis_application_pool.ntlm_test.id

  # HTTP binding
  binding {
    protocol   = "http"
    port       = 80
    ip_address = "*"
    hostname   = "example.com"
  }

  # HTTPS binding with certificate
  binding {
    protocol    = "https"
    port        = 443
    ip_address  = "*"
    hostname    = "example.com"
    certificate = "YOUR_CERTIFICATE_ID_HERE"
  }
}
```

The certificate ID can be obtained from:
1. The `available_certificates` output after running `terraform apply`
2. Directly from the IIS Administration API at `/api/certificates`
3. Using the certificate's thumbprint to find the ID

**Note:** Certificates must already exist on the IIS server in one of the certificate stores (My, WebHosting, or IIS Central Certificate Store). The provider currently supports referencing existing certificates, not creating new ones.

## Troubleshooting

- Ensure the IIS Administration API is enabled and accessible
- Verify NTLM authentication credentials are correct
- Check network connectivity and firewall rules
- For certificate issues, temporarily set `iis_insecure = true` for testing