# Terraform IIS Provider - Setup and Build Instructions

This document provides step-by-step instructions for adding proxy support and building the Terraform IIS provider.

## Prerequisites Setup

### 1. Install Go (if not installed)

Download and install Go from: https://golang.org/downloads/

For Windows:
- Download the Windows installer (go1.21.x.windows-amd64.msi)
- Run the installer
- Restart your PowerShell/Command Prompt

Verify installation:
```powershell
go version
```

### 2. Install Terraform (if not installed)

Download from: https://www.terraform.io/downloads

For Windows:
- Download the Windows binary
- Extract to a folder in your PATH (e.g., C:\terraform)
- Add the folder to your PATH environment variable

Verify installation:
```powershell
terraform version
```

## Proxy Support Implementation

The following changes have been made to add proxy support:

### 1. Provider Schema Updates (`provider/provider.go`)

Added two new configuration options:
- `proxy_url`: HTTP/HTTPS proxy URL with optional authentication
- `insecure`: Skip TLS certificate verification

### 2. HTTP Client Configuration

Updated the `providerConfigure` function to:
- Parse proxy URL and configure HTTP transport
- Handle proxy authentication
- Configure TLS settings

### 3. Environment Variable Support

The provider now supports these environment variables:
- `IIS_PROXY_URL`: Proxy URL
- `IIS_INSECURE`: Skip TLS verification (true/false)

## Building the Provider

### Method 1: Using PowerShell Script (Recommended)

```powershell
# Build only
.\build.ps1

# Build and install locally for Terraform
.\build.ps1 -Install

# Build specific version
.\build.ps1 -Version "0.2.0" -Install

# Clean build artifacts
.\build.ps1 -Clean

# Show help
.\build.ps1 -Help
```

### Method 2: Using Make

```bash
# Build the provider
make build

# Build and install locally
make install-local

# Clean build artifacts
make clean

# Run tests
make test
```

### Method 3: Manual Build

```powershell
# Create bin directory
New-Item -ItemType Directory -Name "bin" -Force

# Build for Windows
go build -ldflags "-w -s" -o "bin/terraform-provider-iis_v0.1.0.exe" .

# Build for Linux
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -ldflags "-w -s" -o "bin/terraform-provider-iis_v0.1.0_linux_amd64" .
```

## Installation for Local Development

### Automatic Installation (Recommended)

```powershell
.\build.ps1 -Install
```

### Manual Installation

1. Build the provider first
2. Create the Terraform plugins directory:
```powershell
$pluginDir = "$env:APPDATA\terraform.d\plugins\terraform.local\maxjoehnk\iis\0.1.0\windows_amd64"
New-Item -ItemType Directory -Path $pluginDir -Force
```

3. Copy the binary:
```powershell
Copy-Item "bin\terraform-provider-iis_v0.1.0.exe" "$pluginDir\terraform-provider-iis_v0.1.0.exe"
```

## Using the Provider with Proxy

### Basic Configuration

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
  host       = "https://iis-server.example.com:8443"
  access_key = "your-access-key"
  
  # Proxy configuration
  proxy_url = "http://proxy.company.com:8080"
  insecure  = true  # Skip TLS verification for internal servers
}
```

### Proxy with Authentication

```hcl
provider "iis" {
  host       = "https://iis-server.example.com:8443"
  access_key = "your-access-key"
  proxy_url  = "http://username:password@proxy.company.com:8080"
}
```

### Using Environment Variables

```powershell
# Set environment variables
$env:IIS_HOST = "https://iis-server.example.com:8443"
$env:IIS_ACCESS_KEY = "your-access-key"
$env:IIS_PROXY_URL = "http://proxy.company.com:8080"
$env:IIS_INSECURE = "true"

# Use in Terraform (no provider block configuration needed)
terraform plan
```

## Testing the Setup

### 1. Initialize Terraform

```powershell
cd examples
terraform init
```

### 2. Validate Configuration

```powershell
terraform validate
```

### 3. Plan Deployment

```powershell
terraform plan
```

## Troubleshooting

### Common Issues

1. **Provider not found**
   - Ensure the provider is built and installed in the correct directory
   - Check Terraform plugin directory: `%APPDATA%\terraform.d\plugins`

2. **Proxy authentication fails**
   - Verify proxy URL format: `http://username:password@proxy.com:port`
   - Check proxy credentials and network connectivity

3. **TLS certificate errors**
   - Set `insecure = true` for internal/self-signed certificates
   - Ensure proper CA certificates are installed

4. **Build fails**
   - Ensure Go is properly installed and in PATH
   - Check Go version compatibility (1.21+)

### Debug Mode

Enable debug logging:
```powershell
$env:TF_LOG = "DEBUG"
terraform plan
```

### Verify Installation

Check provider installation:
```powershell
terraform providers
```

## Next Steps

1. Test the provider with your IIS server
2. Create Terraform configurations for your IIS resources
3. Consider creating a proper provider registry for distribution
4. Add unit tests for proxy functionality
5. Update documentation with specific IIS server setup requirements

## Files Created/Modified

- `provider/provider.go` - Added proxy and TLS configuration
- `build.ps1` - PowerShell build script
- `Makefile` - Make-based build system
- `examples/` - Example Terraform configurations
- `README.md` - Updated documentation
- `.gitignore` - Updated to exclude build artifacts