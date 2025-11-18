# State Drift Handling

This document explains how the IIS Terraform provider handles state drift and manual changes to IIS resources.

## Problem

When resources are manually deleted or modified outside of Terraform (directly in IIS), Terraform can encounter errors when trying to recreate them because the resources still exist on the server but not in Terraform's state file.

Common error scenarios:
- **409 Conflict**: Resource already exists on the server
- **State mismatch**: Terraform thinks resource doesn't exist but it does
- **Manual changes**: Administrator made direct changes in IIS Manager

## Solution

The provider now implements **graceful conflict resolution** by:

1. **Detecting 409 Conflict errors** when creating resources
2. **Automatically retrieving existing resources** by name
3. **Importing them into state** instead of failing

## Implementation

### Application Pools

When creating an application pool that already exists:

```go
// In app_pool_create.go
func (client Client) CreateAppPool(ctx context.Context, name string, managedRuntimeVersion string) (*ApplicationPool, error) {
    // Try to create
    res, err := httpPost(ctx, client, "/api/webserver/application-pools", reqBody)
    if err != nil {
        // If 409 Conflict, retrieve existing pool by name
        if IsConflictError(err) {
            pool, getErr := client.GetAppPoolByName(ctx, name)
            if getErr == nil {
                return pool, nil  // Return existing pool
            }
        }
        return nil, err
    }
    // Normal creation succeeded
    return &pool, nil
}
```

### Websites

When creating a website that already exists:

```go
// In website_create.go
func (client Client) CreateWebsite(ctx context.Context, req CreateWebsiteRequest) (*Website, error) {
    res, err := httpPost(ctx, client, "/api/webserver/websites", req)
    if err != nil {
        // If 409 Conflict, retrieve existing website by name
        if IsConflictError(err) {
            site, getErr := client.GetWebsiteByName(ctx, req.Name)
            if getErr == nil && site != nil {
                return site, nil  // Return existing site
            }
        }
        return nil, err
    }
    return &site, nil
}
```

### Directories & Files

When creating a directory or file that already exists:

```go
// In file_create.go
func (client Client) CreateFile(ctx context.Context, req CreateFileRequest) (*File, error) {
    res, err := httpPost(ctx, client, "/api/files", req)
    if err != nil {
        // If 409 Conflict, retrieve existing file/directory
        if IsConflictError(err) && req.Parent != nil {
            file, getErr := client.GetFileByName(ctx, req.Name, req.Parent.ID)
            if getErr == nil && file != nil {
                return file, nil  // Return existing file/directory
            }
        }
        return nil, err
    }
    return &file, nil
}
```

## Helper Functions

### IsConflictError

Detects 409 Conflict HTTP errors:

```go
// In util.go
func IsConflictError(err error) bool {
    if err == nil {
        return false
    }
    return bytes.Contains([]byte(err.Error()), []byte("409 Conflict"))
}
```

### GetAppPoolByName

Retrieves an application pool by name:

```go
// In app_pool.go
func (client Client) GetAppPoolByName(ctx context.Context, name string) (*ApplicationPool, error) {
    var response struct {
        AppPools []ApplicationPool `json:"app_pools"`
    }
    if err := getJson(ctx, client, "/api/webserver/application-pools", &response); err != nil {
        return nil, err
    }
    
    for _, pool := range response.AppPools {
        if pool.Name == name {
            return client.ReadAppPool(ctx, pool.ID)
        }
    }
    
    return nil, fmt.Errorf("application pool '%s' not found", name)
}
```

### GetWebsiteByName

Retrieves a website by name:

```go
// In website_list.go
func (client Client) GetWebsiteByName(ctx context.Context, name string) (*Website, error) {
    websites, err := client.ListWebsites(ctx)
    if err != nil {
        return nil, err
    }
    
    for _, site := range websites {
        if site.Name == name {
            return client.ReadWebsite(ctx, site.ID)
        }
    }
    
    return nil, nil
}
```

### GetFileByName

Retrieves a file/directory by name:

```go
// In file_list.go
func (client Client) GetFileByName(ctx context.Context, name string, parentID string) (*File, error) {
    files, err := client.ListFiles(ctx, parentID)
    if err != nil {
        return nil, err
    }
    
    for _, file := range files {
        if file.Name == name {
            return &file, nil
        }
    }
    
    return nil, nil
}
```

## Usage Example

### Scenario: Manual Deletion Recovery

1. **Resources exist** in Terraform state and on IIS server
2. **Administrator manually deletes** app pool in IIS Manager
3. **Terraform state** still shows the app pool exists
4. **Run `terraform apply`**:
   - Terraform detects resource is missing from server
   - Tries to create it
   - Gets 409 Conflict (resource actually still exists due to timing/cache)
   - Provider retrieves existing resource by name
   - Imports it back into state
   - **Success!** No error

### Example Output

Before the fix:
```
Error: POST https://server:55539/api/webserver/application-pools returned invalid status code: 409 Conflict
{"title":"Conflict","detail":"Already exists","name":"name","status":409}
```

After the fix:
```
iis_application_pool.server1["ProjectPulseAppPool"]: Creating...
iis_application_pool.server1["ProjectPulseAppPool"]: Creation complete after 0s [id=8tzq8yTGzSvKg3JFIRfkrMibqovQ9oVTWtTgfy0x1ps]

Apply complete! Resources: 1 added, 0 changed, 0 destroyed.
```

## When to Use terraform import

You still need `terraform import` when:

1. **Resource was created manually** and never existed in Terraform state
2. **Multiple resources** with same name exist (rare but possible)
3. **Complex configurations** that need manual mapping

### Manual Import Example

```bash
# Import existing app pool
terraform import 'iis_application_pool.server1["MyPool"]' <pool-id>

# Import existing website
terraform import 'iis_website.server1["MySite"]' <site-id>

# Import existing directory
terraform import 'iis_directory.server1["MyDir"]' <dir-id>
```

## Best Practices

### 1. Let Terraform Manage Everything

Avoid manual changes when possible:
- ✅ Use Terraform for all IIS configuration
- ❌ Don't mix Terraform + manual IIS Manager changes

### 2. Use terraform plan Before Apply

Always check what will change:
```bash
terraform plan
```

### 3. State File Management

Keep state file safe:
```hcl
terraform {
  backend "azurerm" {
    # Store state remotely
  }
}
```

### 4. Handle Drift Detection

Use `terraform plan` to detect manual changes:
```bash
# Check for drift
terraform plan

# Refresh state to match reality
terraform apply -refresh-only
```

### 5. Team Communication

When manual changes are necessary:
1. Document the change
2. Run `terraform apply -refresh-only` to update state
3. Commit updated state file
4. Notify team

## Troubleshooting

### Issue: Resource Still Shows as "to be created"

**Cause**: State file doesn't have the resource

**Solution**:
```bash
# Let provider auto-import on apply
terraform apply

# Or manually import
terraform import 'resource.name["key"]' <id>
```

### Issue: 409 Conflict Despite No Manual Changes

**Cause**: IIS API caching or timing issues

**Solution**: Provider now handles this automatically by:
1. Catching 409 error
2. Listing resources to find existing one
3. Importing it into state

### Issue: Resource Properties Don't Match

**Cause**: Manual changes to resource properties

**Solution**:
```bash
# Show what changed
terraform plan

# Apply to fix drift
terraform apply
```

### Issue: Can't Find Resource by Name

**Cause**: Resource was renamed manually

**Solution**:
```bash
# Remove from state
terraform state rm 'resource.name["key"]'

# Re-import with new name
terraform import 'resource.name["new_key"]' <id>
```

## Testing Conflict Resolution

To test the conflict handling:

1. **Create resource with Terraform**:
   ```bash
   terraform apply
   ```

2. **Remove from state** (simulating drift):
   ```bash
   terraform state rm 'iis_application_pool.server1["TestPool"]'
   ```

3. **Re-apply** (resource exists on server, not in state):
   ```bash
   terraform apply
   ```

4. **Result**: Provider detects conflict and imports existing resource ✅

## Related Documentation

- [Terraform Import](https://www.terraform.io/docs/cli/import/index.html)
- [State Management](https://www.terraform.io/docs/language/state/index.html)
- [Resource Drift](https://www.terraform.io/docs/cloud/workspaces/state.html#detecting-drift)

## Summary

The IIS Terraform provider now gracefully handles:

✅ **409 Conflict errors** - Auto-imports existing resources  
✅ **State drift** - Recovers from manual deletions  
✅ **Manual changes** - Re-imports modified resources  
✅ **Cache issues** - Handles IIS API timing problems  
✅ **Team workflows** - Supports multiple operators  

This makes the provider more resilient and easier to use in real-world scenarios where manual changes sometimes occur.
