# Artifacts Folder

This folder contains sample web content for deployment to IIS via the Terraform IIS Provider.

## Files

### Default.aspx
A dynamic ASP.NET page that displays:
- Server time
- Machine name
- Application pool information
- Authentication details
- User identity (for NTLM/Windows auth testing)
- Physical path

Perfect for testing:
- ✅ ASPX execution
- ✅ Application pool configuration
- ✅ NTLM/Windows authentication
- ✅ Server-side processing

### index.html
A static HTML landing page with:
- Clean, modern design
- Links to other pages
- Information about the deployment

Perfect for testing:
- ✅ Static file serving
- ✅ Basic HTTP/HTTPS access
- ✅ Website availability

## Deployment Methods

### Method 1: Manual Copy
Copy files directly to the IIS physical path:
```powershell
Copy-Item -Path .\artifacts\* -Destination "C:\inetpub\ProjectPulse" -Recurse -Force
```

### Method 2: Network Share
If IIS server is remote:
```powershell
Copy-Item -Path .\artifacts\* `
          -Destination "\\IIS-SERVER\c$\inetpub\ProjectPulse" `
          -Recurse -Force
```

### Method 3: Terraform Provisioner
Uncomment the `null_resource` in `ntlm-example.tf` to automatically deploy files when the directory is created.

### Method 4: CI/CD Pipeline
Integrate file deployment into your CI/CD pipeline:
- Azure DevOps: Use "Copy Files" task
- GitHub Actions: Use SCP or WinRM
- Jenkins: Use "Publish Over SSH" or PowerShell

## Testing the Deployment

After deployment, access:
- `http://projectpulse.com/index.html` - Static page
- `http://projectpulse.com/Default.aspx` - Dynamic ASPX page
- `https://projectpulse.com/` - HTTPS access (if certificate configured)

## Notes

The IIS Administration API manages IIS configuration (sites, app pools, bindings) but does not provide file content upload endpoints. File deployment must be handled separately from IIS configuration management.
