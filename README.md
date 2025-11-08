# Terraform IIS Provider

Terraform Provider using the [Microsoft IIS Administration](https://docs.microsoft.com/en-us/IIS-Administration/) API with proxy support.

## Features

- ✅ Create and manage IIS Application Pools
- ✅ Create and manage IIS Applications
- ✅ Create and manage IIS Websites
- ✅ Configure Authentication settings
- ✅ **Proxy Support** - HTTP/HTTPS proxy with authentication
- ✅ **NTLM Authentication** - Windows domain and local user authentication
- ✅ **TLS Configuration** - Skip verification for internal servers
- ✅ Environment variable configuration

## Quick Start

### 1. Build and Install the Provider

Using PowerShell (Recommended for Windows):
```powershell
# Build and install locally
.\build.ps1 -Install

# Or build for a specific version
.\build.ps1 -Version "0.2.0" -Install
```

Using Make (Alternative):
```bash
# Build the provider
make build

# Install locally for Terraform
make install-local
```

### 2. Basic Usage

```hcl
terraform {
  required_providers {
    iis = {
      source  = "terraform.local/maxjoehnk/iis"
      version = "~> 0.1.0"
    }
  }
}

provider "iis" {
  access_key = "your-access-key"
  host       = "https://localhost:55539"
}

resource "iis_application_pool" "example" {
  name = "MyAppPool"
}

resource "iis_application" "example" {
  physical_path    = "%systemdrive%\\inetpub\\myapp"
  application_pool = iis_application_pool.example.id
  path            = "MyApp"
  website         = data.iis_website.default.ids[0]
}

data "iis_website" "default" {}
```

### 3. Using with Proxy

```hcl
provider "iis" {
  host       = "https://iis-server.company.com:8443"
  access_key = "your-access-key"
  
  # Proxy configuration
  proxy_url = "http://proxy.company.com:8080"
  
  # Skip TLS verification for internal servers
  insecure = true
}
```

### 4. Proxy with Authentication

```hcl
provider "iis" {
  host       = "https://iis-server.company.com:8443"
  access_key = "your-access-key"
  
  # Proxy with username/password
  proxy_url = "http://username:password@proxy.company.com:8080"
}
```

### 5. NTLM Authentication (Windows Domain)

```hcl
provider "iis" {
  host = "https://iis-server.company.com:8443"
  
  # NTLM Authentication (replaces access_key)
  ntlm_username = "serviceaccount"
  ntlm_password = "SecurePassword123!"
  ntlm_domain   = "COMPANY"  # Optional for local accounts
  
  # Optional proxy and TLS settings
  proxy_url = "http://proxy.company.com:8080"
  insecure  = true
}
```

### 6. Local User Authentication

```hcl
provider "iis" {
  host = "https://iis-server.local:8443"
  
  # Local user authentication (no domain)
  ntlm_username = "iisadmin"
  ntlm_password = "LocalPassword123!"
  # Omit ntlm_domain for local accounts
}
```

### 7. Dual Authentication (Enterprise)

For environments requiring both NTLM authentication AND API authorization:

```hcl
provider "iis" {
  host = "https://iis-server.company.com:8443"
  
  # API Authorization (IIS Administration API)
  access_key = "your-api-access-token"
  
  # HTTP Authentication (Windows/Domain)
  ntlm_username = "serviceaccount" 
  ntlm_password = "SecurePassword123!"
  ntlm_domain   = "COMPANY"
  
  # Network settings
  proxy_url = "http://proxy.company.com:8080"
  insecure  = true
}
```

## Configuration Options

| Parameter | Description | Environment Variable | Required |
|-----------|-------------|---------------------|----------|
| `host` | IIS Administration API host URL | `IIS_HOST` | Yes |
| `access_key` | Access key for authentication | `IIS_ACCESS_KEY` | Yes* |
| `ntlm_username` | Username for NTLM authentication | `IIS_NTLM_USERNAME` | Yes* |
| `ntlm_password` | Password for NTLM authentication | `IIS_NTLM_PASSWORD` | Yes* |
| `ntlm_domain` | Domain for NTLM authentication | `IIS_NTLM_DOMAIN` | No |
| `proxy_url` | HTTP/HTTPS proxy URL | `IIS_PROXY_URL` | No |
| `insecure` | Skip TLS certificate verification | `IIS_INSECURE` | No |

**\* Authentication**: Either `access_key` OR NTLM credentials must be provided. Both can be used together for dual authentication (NTLM + API token).

### Proxy URL Format

The proxy URL supports the following formats:
- `http://proxy.company.com:8080` - Basic proxy
- `http://username:password@proxy.company.com:8080` - Proxy with authentication
- `https://secure-proxy.company.com:443` - HTTPS proxy

## Environment Variables

You can configure the provider using environment variables:

```powershell
# Windows PowerShell - Access Key Authentication
$env:IIS_HOST = "https://iis-server.company.com:8443"
$env:IIS_ACCESS_KEY = "your-access-key"
$env:IIS_PROXY_URL = "http://proxy.company.com:8080"
$env:IIS_INSECURE = "true"

# OR - NTLM Authentication
$env:IIS_HOST = "https://iis-server.company.com:8443"
$env:IIS_NTLM_USERNAME = "serviceaccount"
$env:IIS_NTLM_PASSWORD = "SecurePassword123!"
$env:IIS_NTLM_DOMAIN = "COMPANY"
$env:IIS_PROXY_URL = "http://proxy.company.com:8080"
$env:IIS_INSECURE = "true"
```

```bash
# Linux/macOS
export IIS_HOST="https://iis-server.company.com:8443"
export IIS_ACCESS_KEY="your-access-key"
export IIS_PROXY_URL="http://proxy.company.com:8080"
export IIS_INSECURE="true"
```

## Building from Source

### Prerequisites

- Go 1.21 or later
- Git

### Build Steps

1. Clone the repository:
```bash
git clone https://github.com/maxjoehnk/terraform-provider-iis.git
cd terraform-provider-iis
```

2. Build the provider:
```powershell
# Using PowerShell script
.\build.ps1

# Using Make
make build

# Using Go directly
go build -o bin/terraform-provider-iis.exe .
```

3. Install locally:
```powershell
# Using PowerShell script
.\build.ps1 -Install

# Using Make
make install-local
```

### Cross-Platform Building

```powershell
# Build for multiple platforms
.\build.ps1 -GoOS linux -GoArch amd64
.\build.ps1 -GoOS darwin -GoArch arm64

# Or using Make
make build-all
```

## Development

### Running Tests

```bash
go test -v ./...
# or
make test
```

### Code Formatting

```bash
go fmt ./...
# or
make fmt
```

### Linting

```bash
golangci-lint run
# or
make lint
```

## Troubleshooting

### Common Issues

1. **Provider not found**: Ensure the provider is installed in the correct Terraform plugins directory
2. **Proxy authentication fails**: Check proxy URL format and credentials
3. **TLS verification errors**: Set `insecure = true` for internal/self-signed certificates

### Debug Mode

Enable Terraform debug logging:
```powershell
$env:TF_LOG = "DEBUG"
terraform plan
```

### Verify Installation

Check if the provider is correctly installed:
```bash
terraform providers
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.