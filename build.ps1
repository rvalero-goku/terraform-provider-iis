# PowerShell script to build and install the Terraform IIS provider locally

param(
    [string]$Version = "0.1.0",
    [string]$GoOS = "windows",
    [string]$GoArch = "amd64",
    [switch]$Install,
    [switch]$DevSetup,
    [switch]$Clean,
    [switch]$Help
)

$ProviderName = "terraform-provider-iis"
$TerraformPluginsDir = "$env:APPDATA\terraform.d\plugins"
$LocalProviderPath = "terraform.local\maxjoehnk\iis\$Version\${GoOS}_$GoArch"

function Show-Help {
    Write-Host @"
Terraform IIS Provider Build Script

Usage: .\build.ps1 [OPTIONS]

Options:
  -Version <string>    Provider version (default: 0.1.0)
  -GoOS <string>       Target OS (default: windows)
  -GoArch <string>     Target architecture (default: amd64)
  -Install             Install provider locally after building
  -DevSetup            Build and setup for local development (recommended)
  -Clean               Clean build artifacts
  -Help               Show this help message

Examples:
  .\build.ps1                          # Build provider
  .\build.ps1 -DevSetup                # Build and setup for development (recommended)
  .\build.ps1 -Install                 # Build and install provider locally
  .\build.ps1 -Version "0.2.0"         # Build specific version
  .\build.ps1 -Clean                   # Clean build artifacts
  .\build.ps1 -GoOS linux -GoArch amd64 # Build for Linux

Environment Variables (for proxy testing):
  IIS_HOST         - IIS server URL
  IIS_ACCESS_KEY   - Access key for authentication
  IIS_PROXY_URL    - Proxy URL (e.g., http://proxy.company.com:8080)
  IIS_INSECURE     - Skip TLS verification (true/false)
"@
}

function Build-Provider {
    Write-Host "Building $ProviderName for $GoOS/$GoArch..." -ForegroundColor Green

    # Create bin directory if it doesn't exist
    if (!(Test-Path "bin")) {
        New-Item -ItemType Directory -Name "bin" | Out-Null
    }

    $env:GOOS = $GoOS
    $env:GOARCH = $GoArch

    # For dev_overrides, we need the binary named exactly "terraform-provider-iis"
    $outputName = if ($GoOS -eq "windows") {
        "bin\terraform-provider-iis.exe"
    }
    else {
        "bin/terraform-provider-iis"
    }

    $buildCmd = "go build -ldflags `"-w -s`" -o `"$outputName`" ."
    Write-Host "Running: $buildCmd" -ForegroundColor Yellow

    Invoke-Expression $buildCmd

    if ($LASTEXITCODE -eq 0) {
        Write-Host "Build completed successfully: $outputName" -ForegroundColor Green

        # Also create versioned binary for registry installation
        $versionedName = if ($GoOS -eq "windows") {
            "bin\${ProviderName}_v$Version.exe"
        }
        else {
            "bin/${ProviderName}_v$Version"
        }
        Copy-Item $outputName $versionedName -Force
        Write-Host "Also created versioned binary: $versionedName" -ForegroundColor Green

        return $outputName
    }
    else {
        Write-Host "Build failed!" -ForegroundColor Red
        exit 1
    }
}

function Install-ProviderLocally {
    param([string]$BinaryPath)

    Write-Host "Installing provider locally..." -ForegroundColor Green

    $fullInstallPath = "$TerraformPluginsDir\$LocalProviderPath"

    # Create directory structure
    if (!(Test-Path $fullInstallPath)) {
        New-Item -ItemType Directory -Path $fullInstallPath -Force | Out-Null
        Write-Host "Created directory: $fullInstallPath" -ForegroundColor Yellow
    }

    # Copy binary
    $destinationFile = "$fullInstallPath\${ProviderName}_v$Version.exe"
    Copy-Item $BinaryPath $destinationFile -Force

    Write-Host "Provider installed successfully!" -ForegroundColor Green
    Write-Host "Location: $destinationFile" -ForegroundColor Yellow

    # Show terraform configuration hint
    Write-Host @"

To use this provider in your Terraform configuration, add:

terraform {
  required_providers {
    iis = {
      source  = "terraform.local/maxjoehnk/iis"
      version = "~> $Version"
    }
  }
}

provider "iis" {
  host       = "https://your-iis-server.com:8443"
  access_key = "your-access-key"

  # Optional proxy configuration
  proxy_url = "http://proxy.company.com:8080"
  insecure  = false
}
"@ -ForegroundColor Cyan
}

function Setup-DevEnvironment {
    Write-Host "Setting up development environment..." -ForegroundColor Green

    # Get current directory and convert backslashes to forward slashes for HCL
    $currentDir = Get-Location
    $binPath = Join-Path $currentDir "bin"
    $binPath = $binPath -replace '\\', '/'

    # Create terraform.rc content for dev overrides
    $terraformRcContent = @"
provider_installation {
  dev_overrides {
    "terraform.local/maxjoehnk/iis" = "$binPath"
  }

  # For all other providers, install them directly as normal.
  direct {}
}
"@

    # Determine terraform.rc location
    $terraformRcPath = "$env:APPDATA\terraform.rc"

    # Backup existing terraform.rc if it exists
    if (Test-Path $terraformRcPath) {
        $backupPath = "$terraformRcPath.backup.$(Get-Date -Format 'yyyyMMdd-HHmmss')"
        Copy-Item $terraformRcPath $backupPath
        Write-Host "Backed up existing terraform.rc to: $backupPath" -ForegroundColor Yellow
    }

    # Write new terraform.rc
    Set-Content -Path $terraformRcPath -Value $terraformRcContent
    Write-Host "Created terraform.rc with dev overrides: $terraformRcPath" -ForegroundColor Green

    Write-Host @"

Development environment configured!

The terraform.rc file has been set up to use your local provider binary.
You can now run Terraform commands in the examples directory without needing
to install the provider to the plugins directory.

Next steps:
1. cd examples
2. terraform init
3. terraform plan

To restore normal provider behavior, delete or rename: $terraformRcPath
"@ -ForegroundColor Cyan
}

function Clean-BuildArtifacts {
    Write-Host "Cleaning build artifacts..." -ForegroundColor Green

    if (Test-Path "bin") {
        Remove-Item "bin" -Recurse -Force
        Write-Host "Removed bin directory" -ForegroundColor Yellow
    }

    Write-Host "Clean completed!" -ForegroundColor Green
}

function Test-Environment {
    Write-Host "Checking environment..." -ForegroundColor Green

    # Check if Go is installed
    try {
        $goVersion = go version
        Write-Host "Go: $goVersion" -ForegroundColor Yellow
    }
    catch {
        Write-Host "Error: Go is not installed or not in PATH" -ForegroundColor Red
        exit 1
    }

    # Check if Terraform is installed
    try {
        $tfVersion = terraform version
        Write-Host "Terraform: $($tfVersion.Split("`n")[0])" -ForegroundColor Yellow
    }
    catch {
        Write-Host "Warning: Terraform is not installed or not in PATH" -ForegroundColor Yellow
    }

    Write-Host "Environment check completed!" -ForegroundColor Green
}

# Main script logic
if ($Help) {
    Show-Help
    exit 0
}

if ($Clean) {
    Clean-BuildArtifacts
    exit 0
}

Test-Environment

$binaryPath = Build-Provider

if ($DevSetup) {
    Setup-DevEnvironment
}
elseif ($Install) {
    Install-ProviderLocally -BinaryPath $binaryPath
}

Write-Host "Done!" -ForegroundColor Green
