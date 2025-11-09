# Deploying Artifacts to IIS

This guide explains how to deploy web content (HTML, ASPX, etc.) to your IIS website managed by Terraform.

## Quick Start

### Local Deployment
If Terraform is running on the same machine as IIS:
```powershell
.\deploy-artifacts.ps1
```

### Remote Deployment
If IIS is on a remote server:
```powershell
.\deploy-artifacts.ps1 -RemoteServer "192.168.99.111" -DestinationPath "C:\inetpub\ProjectPulse"
```

## What Gets Deployed

The `artifacts/` folder contains:
- **Default.aspx** - ASP.NET test page with server info and authentication details
- **index.html** - Static HTML landing page

## Manual Deployment Options

### Option 1: Local Copy
```powershell
Copy-Item -Path .\artifacts\* -Destination "C:\inetpub\ProjectPulse" -Recurse -Force
```

### Option 2: Remote Copy (UNC Path)
```powershell
Copy-Item -Path .\artifacts\* `
          -Destination "\\192.168.99.111\c$\inetpub\ProjectPulse" `
          -Recurse -Force
```

### Option 3: Use the Deployment Script
```powershell
# Local
.\deploy-artifacts.ps1

# Remote
.\deploy-artifacts.ps1 -RemoteServer "IIS-SERVER-NAME"

# Custom path
.\deploy-artifacts.ps1 -DestinationPath "C:\MyWebsite"

# Force overwrite
.\deploy-artifacts.ps1 -Force
```

## Terraform Integration

### Using null_resource with Provisioner
Add this to your `ntlm-example.tf`:

```hcl
resource "null_resource" "deploy_artifacts" {
  depends_on = [iis_directory.my_app_dir]
  
  triggers = {
    # Re-deploy when artifacts change
    artifacts_hash = md5(join("", [
      for f in fileset(path.module, "artifacts/*") : 
      filemd5("${path.module}/${f}")
    ]))
  }
  
  provisioner "local-exec" {
    command = "powershell.exe -File ${path.module}/deploy-artifacts.ps1 -RemoteServer ${var.iis_host}"
  }
}
```

## Testing Deployment

After deployment, test access:

### HTTP
```bash
curl http://projectpulse.com/index.html
curl http://projectpulse.com/Default.aspx
```

### HTTPS (if configured)
```bash
curl https://projectpulse.com/index.html
curl https://projectpulse.com/Default.aspx
```

### Browser
- http://projectpulse.com/
- https://projectpulse.com/

## Troubleshooting

### Access Denied
Ensure you have:
- Admin rights on the remote server
- File sharing enabled: `Enable-PSRemoting -Force`
- Firewall allows file sharing

### Path Not Found
1. Verify the directory exists: `Test-Path "C:\inetpub\ProjectPulse"`
2. Check Terraform output: `tofu output created_directory`
3. Create manually if needed: `New-Item -Path "C:\inetpub\ProjectPulse" -ItemType Directory`

### ASPX Not Executing
1. Ensure .NET Framework is installed
2. Verify app pool is using correct .NET version (v4.0)
3. Check application pool status is "Started"

## Why Separate File Deployment?

The IIS Administration API (used by this Terraform provider) manages:
- ✅ IIS configuration (sites, app pools, bindings)
- ✅ Directory structure metadata
- ✅ Authentication settings

It does NOT manage:
- ❌ File content upload/download
- ❌ File contents/binary data

This separation of concerns is intentional:
- **Configuration** → Terraform IIS Provider
- **Content** → Deployment pipeline, file copy, or provisioners

## CI/CD Integration

### Azure DevOps
```yaml
- task: CopyFiles@2
  inputs:
    SourceFolder: '$(Build.SourcesDirectory)/examples/artifacts'
    Contents: '**'
    TargetFolder: '\\\\$(IIS_SERVER)\\c$\\inetpub\\ProjectPulse'
```

### GitHub Actions
```yaml
- name: Deploy Artifacts
  run: |
    Copy-Item -Path ./examples/artifacts/* `
              -Destination "\\\\${{ secrets.IIS_SERVER }}\\c$\\inetpub\\ProjectPulse" `
              -Recurse -Force
  shell: powershell
```

### Jenkins
```groovy
bat '''
powershell -Command "Copy-Item -Path .\\examples\\artifacts\\* -Destination '\\\\IIS-SERVER\\c$\\inetpub\\ProjectPulse' -Recurse -Force"
'''
```
