# Secure Variable Management for GitHub Actions

This guide covers best practices for securely passing sensitive data to the IIS deployment workflows.

## Table of Contents

1. [GitHub Secrets](#github-secrets)
2. [Environment Variables](#environment-variables)
3. [Encrypted Files](#encrypted-files)
4. [HashiCorp Vault Integration](#hashicorp-vault-integration)
5. [Azure Key Vault Integration](#azure-key-vault-integration)
6. [Terraform Integration](#terraform-integration)

## GitHub Secrets

### Repository Secrets

Store sensitive configuration at the repository level:

**Setting up:**
1. Go to repository Settings → Secrets and variables → Actions
2. Click "New repository secret"
3. Add your configuration

**Example:**
```yaml
# .github/workflows/deploy-iis.yml
jobs:
  deploy:
    uses: owner/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
    with:
      config_yaml: ${{ secrets.IIS_CONFIG }}
      auto_approve: true
```

**Pros:**
- ✅ Simple to set up
- ✅ Encrypted at rest
- ✅ Masked in logs
- ✅ Scoped to repository

**Cons:**
- ⚠️ All collaborators with write access can use secrets
- ⚠️ No audit trail for secret usage

### Environment Secrets

More secure option with approval gates:

**Setting up:**
1. Go to Settings → Environments
2. Create environment (e.g., "production")
3. Add environment protection rules:
   - Required reviewers
   - Wait timer
   - Deployment branches
4. Add secrets to the environment

**Example:**
```yaml
jobs:
  deploy-production:
    environment: production  # Requires approval
    uses: owner/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
    with:
      config_yaml: ${{ secrets.IIS_CONFIG }}  # From environment
```

**Pros:**
- ✅ Requires manual approval
- ✅ Audit trail of deployments
- ✅ Can restrict to specific branches
- ✅ Multiple reviewers supported

**Cons:**
- ⚠️ More setup overhead
- ⚠️ Manual approval slows down deployments

### Organization Secrets

Share secrets across multiple repositories:

**Setting up:**
1. Go to Organization Settings → Secrets and variables → Actions
2. Create organization secret
3. Select repository access

**Example:**
```yaml
jobs:
  deploy:
    uses: owner/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
    with:
      config_yaml: ${{ secrets.IIS_CONFIG }}  # From organization
```

**Pros:**
- ✅ Central management
- ✅ Reusable across repos
- ✅ Consistent configuration

**Cons:**
- ⚠️ Requires organization admin access
- ⚠️ Broader access scope

## Environment Variables

### Separate Credentials from Configuration

Store only sensitive values as secrets:

```yaml
# Store in GitHub Secrets:
# - IIS_SERVER1_PASSWORD
# - IIS_SERVER2_PASSWORD

jobs:
  prepare-config:
    runs-on: ubuntu-latest
    outputs:
      config: ${{ steps.build.outputs.yaml }}
    steps:
      - name: Build Configuration
        id: build
        run: |
          cat > iis-config.yaml << 'EOF'
          servers:
            server1:
              host: "https://server1.example.com:55539"
              enabled: true
              ntlm_username: "admin"
              ntlm_password: "${{ secrets.IIS_SERVER1_PASSWORD }}"
              insecure: true
            
            server2:
              host: "https://server2.example.com:55539"
              enabled: true
              ntlm_username: "admin"
              ntlm_password: "${{ secrets.IIS_SERVER2_PASSWORD }}"
              insecure: true
          
          application_pools:
            MyAppPool:
              managed_runtime_version: "v4.0"
              status: "started"
          EOF
          
          echo "yaml<<EOFYAML" >> $GITHUB_OUTPUT
          cat iis-config.yaml >> $GITHUB_OUTPUT
          echo "EOFYAML" >> $GITHUB_OUTPUT
  
  deploy:
    needs: prepare-config
    uses: owner/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
    with:
      config_yaml: ${{ needs.prepare-config.outputs.config }}
```

### Using GitHub Actions Variables

Store non-sensitive configuration as variables:

**Setting up:**
1. Go to Settings → Secrets and variables → Actions → Variables tab
2. Add variables (e.g., `IIS_SERVER1_HOST`, `IIS_POOL_VERSION`)

**Example:**
```yaml
jobs:
  prepare-config:
    runs-on: ubuntu-latest
    outputs:
      config: ${{ steps.build.outputs.yaml }}
    steps:
      - name: Build Configuration
        id: build
        env:
          SERVER1_HOST: ${{ vars.IIS_SERVER1_HOST }}
          SERVER2_HOST: ${{ vars.IIS_SERVER2_HOST }}
          SERVER1_PASS: ${{ secrets.IIS_SERVER1_PASSWORD }}
          SERVER2_PASS: ${{ secrets.IIS_SERVER2_PASSWORD }}
          POOL_VERSION: ${{ vars.IIS_POOL_VERSION }}
        run: |
          cat > iis-config.yaml << EOF
          servers:
            server1:
              host: "$SERVER1_HOST"
              enabled: true
              ntlm_username: "admin"
              ntlm_password: "$SERVER1_PASS"
              insecure: true
            server2:
              host: "$SERVER2_HOST"
              enabled: true
              ntlm_username: "admin"
              ntlm_password: "$SERVER2_PASS"
              insecure: true
          
          application_pools:
            MyAppPool:
              managed_runtime_version: "$POOL_VERSION"
              status: "started"
          EOF
          
          echo "yaml<<EOFYAML" >> $GITHUB_OUTPUT
          cat iis-config.yaml >> $GITHUB_OUTPUT
          echo "EOFYAML" >> $GITHUB_OUTPUT
  
  deploy:
    needs: prepare-config
    uses: owner/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
    with:
      config_yaml: ${{ needs.prepare-config.outputs.config }}
```

## Encrypted Files

### Using git-crypt

Encrypt sensitive files in your repository:

**Setup:**
```bash
# Install git-crypt
brew install git-crypt  # macOS
apt-get install git-crypt  # Ubuntu

# Initialize in your repo
cd your-repo
git-crypt init

# Add GPG key
git-crypt add-gpg-user user@example.com

# Specify files to encrypt (.gitattributes)
echo "infrastructure/iis-config.yaml filter=git-crypt diff=git-crypt" >> .gitattributes

# Commit and push
git add .gitattributes infrastructure/iis-config.yaml
git commit -m "Encrypt IIS config"
git push
```

**GitHub Actions:**
```yaml
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Unlock git-crypt
        run: |
          echo "${{ secrets.GIT_CRYPT_KEY }}" | base64 -d > git-crypt-key
          git-crypt unlock git-crypt-key
          rm git-crypt-key
      
      - name: Read Config
        id: config
        run: |
          echo "yaml<<EOF" >> $GITHUB_OUTPUT
          cat infrastructure/iis-config.yaml >> $GITHUB_OUTPUT
          echo "EOF" >> $GITHUB_OUTPUT
      
      - name: Deploy
        uses: owner/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
        with:
          config_yaml: ${{ steps.config.outputs.yaml }}
```

### Using SOPS (Secrets OPerationS)

Encrypt specific values within YAML files:

**Setup:**
```bash
# Install SOPS
brew install sops  # macOS

# Encrypt file
sops --encrypt --age <public-key> infrastructure/iis-config.yaml > infrastructure/iis-config.enc.yaml
```

**GitHub Actions:**
```yaml
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Install SOPS
        run: |
          curl -LO https://github.com/mozilla/sops/releases/download/v3.7.3/sops-v3.7.3.linux
          chmod +x sops-v3.7.3.linux
          sudo mv sops-v3.7.3.linux /usr/local/bin/sops
      
      - name: Decrypt Config
        env:
          SOPS_AGE_KEY: ${{ secrets.SOPS_AGE_KEY }}
        run: |
          sops --decrypt infrastructure/iis-config.enc.yaml > iis-config.yaml
      
      - name: Read Config
        id: config
        run: |
          echo "yaml<<EOF" >> $GITHUB_OUTPUT
          cat iis-config.yaml >> $GITHUB_OUTPUT
          echo "EOF" >> $GITHUB_OUTPUT
      
      - name: Deploy
        uses: owner/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
        with:
          config_yaml: ${{ steps.config.outputs.yaml }}
```

## HashiCorp Vault Integration

### Using Vault for Secret Management

**Setup Vault:**
```bash
# Enable KV secrets engine
vault secrets enable -path=iis kv-v2

# Store secrets
vault kv put iis/server1 \
  host="https://server1.example.com:55539" \
  username="admin" \
  password="secret123"

vault kv put iis/server2 \
  host="https://server2.example.com:55539" \
  username="admin" \
  password="secret456"
```

**GitHub Actions with Vault:**
```yaml
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Import Secrets from Vault
        uses: hashicorp/vault-action@v2
        with:
          url: https://vault.example.com
          token: ${{ secrets.VAULT_TOKEN }}
          secrets: |
            iis/data/server1 host | IIS_SERVER1_HOST ;
            iis/data/server1 username | IIS_SERVER1_USER ;
            iis/data/server1 password | IIS_SERVER1_PASS ;
            iis/data/server2 host | IIS_SERVER2_HOST ;
            iis/data/server2 username | IIS_SERVER2_USER ;
            iis/data/server2 password | IIS_SERVER2_PASS
      
      - name: Build Configuration
        id: config
        run: |
          cat > iis-config.yaml << EOF
          servers:
            server1:
              host: "$IIS_SERVER1_HOST"
              enabled: true
              ntlm_username: "$IIS_SERVER1_USER"
              ntlm_password: "$IIS_SERVER1_PASS"
              insecure: true
            server2:
              host: "$IIS_SERVER2_HOST"
              enabled: true
              ntlm_username: "$IIS_SERVER2_USER"
              ntlm_password: "$IIS_SERVER2_PASS"
              insecure: true
          
          application_pools:
            MyAppPool:
              managed_runtime_version: "v4.0"
              status: "started"
          EOF
          
          echo "yaml<<EOFYAML" >> $GITHUB_OUTPUT
          cat iis-config.yaml >> $GITHUB_OUTPUT
          echo "EOFYAML" >> $GITHUB_OUTPUT
      
      - name: Deploy
        uses: owner/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
        with:
          config_yaml: ${{ steps.config.outputs.yaml }}
```

### Using Vault with OIDC (No Static Tokens)

More secure - uses GitHub's OIDC provider:

```yaml
jobs:
  deploy:
    runs-on: ubuntu-latest
    permissions:
      id-token: write
      contents: read
    
    steps:
      - name: Import Secrets from Vault
        uses: hashicorp/vault-action@v2
        with:
          url: https://vault.example.com
          method: jwt
          role: github-actions-role
          secrets: |
            iis/data/server1 * | IIS_SERVER1_ ;
            iis/data/server2 * | IIS_SERVER2_
      
      - name: Build Configuration
        id: config
        run: |
          cat > iis-config.yaml << EOF
          servers:
            server1:
              host: "$IIS_SERVER1_HOST"
              enabled: true
              ntlm_username: "$IIS_SERVER1_USERNAME"
              ntlm_password: "$IIS_SERVER1_PASSWORD"
              insecure: true
            server2:
              host: "$IIS_SERVER2_HOST"
              enabled: true
              ntlm_username: "$IIS_SERVER2_USERNAME"
              ntlm_password: "$IIS_SERVER2_PASSWORD"
              insecure: true
          EOF
          
          echo "yaml<<EOFYAML" >> $GITHUB_OUTPUT
          cat iis-config.yaml >> $GITHUB_OUTPUT
          echo "EOFYAML" >> $GITHUB_OUTPUT
```

## Azure Key Vault Integration

For Azure-hosted secrets:

```yaml
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Azure Login
        uses: azure/login@v1
        with:
          creds: ${{ secrets.AZURE_CREDENTIALS }}
      
      - name: Get Secrets from Key Vault
        uses: azure/CLI@v1
        with:
          inlineScript: |
            echo "IIS_SERVER1_HOST=$(az keyvault secret show --name iis-server1-host --vault-name myvault --query value -o tsv)" >> $GITHUB_ENV
            echo "IIS_SERVER1_USER=$(az keyvault secret show --name iis-server1-user --vault-name myvault --query value -o tsv)" >> $GITHUB_ENV
            echo "IIS_SERVER1_PASS=$(az keyvault secret show --name iis-server1-pass --vault-name myvault --query value -o tsv)" >> $GITHUB_ENV
      
      - name: Build Configuration
        id: config
        run: |
          cat > iis-config.yaml << EOF
          servers:
            server1:
              host: "$IIS_SERVER1_HOST"
              enabled: true
              ntlm_username: "$IIS_SERVER1_USER"
              ntlm_password: "$IIS_SERVER1_PASS"
              insecure: true
          
          application_pools:
            MyAppPool:
              managed_runtime_version: "v4.0"
              status: "started"
          EOF
          
          echo "yaml<<EOFYAML" >> $GITHUB_OUTPUT
          cat iis-config.yaml >> $GITHUB_OUTPUT
          echo "EOFYAML" >> $GITHUB_OUTPUT
      
      - name: Deploy
        uses: owner/terraform-provider-iis/.github/workflows/apply-iis-config.yml@main
        with:
          config_yaml: ${{ steps.config.outputs.config }}
```

## Best Practices

### 1. Principle of Least Privilege

Only give access to secrets that are needed:

```yaml
# BAD: Storing entire prod config in one secret
# If compromised, all servers are at risk
secrets.PROD_IIS_CONFIG

# GOOD: Separate secrets per server
secrets.IIS_SERVER1_PASSWORD
secrets.IIS_SERVER2_PASSWORD
```

### 2. Rotate Secrets Regularly

```yaml
jobs:
  rotate-passwords:
    runs-on: ubuntu-latest
    steps:
      - name: Generate New Password
        id: newpass
        run: |
          NEW_PASS=$(openssl rand -base64 32)
          echo "::add-mask::$NEW_PASS"
          echo "password=$NEW_PASS" >> $GITHUB_OUTPUT
      
      - name: Update IIS Password
        run: |
          # Update password on IIS server
          # Update GitHub secret via API
          gh secret set IIS_SERVER1_PASSWORD --body "${{ steps.newpass.outputs.password }}"
```

### 3. Audit Secret Usage

Enable audit logging:

```yaml
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Log Deployment
        run: |
          echo "Deployment started by ${{ github.actor }}"
          echo "Repository: ${{ github.repository }}"
          echo "Ref: ${{ github.ref }}"
          echo "SHA: ${{ github.sha }}"
          # Send to logging service
```

### 4. Use Short-Lived Tokens

When possible, generate temporary credentials:

```yaml
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Generate Temporary Credentials
        id: creds
        run: |
          # Generate temporary credentials that expire in 1 hour
          TEMP_TOKEN=$(generate-temp-token --expires 1h)
          echo "::add-mask::$TEMP_TOKEN"
          echo "token=$TEMP_TOKEN" >> $GITHUB_OUTPUT
      
      - name: Deploy with Temp Creds
        # Use temporary credentials
```

### 5. Never Log Secrets

```yaml
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Build Config (SAFE)
        run: |
          # Secrets are masked automatically
          echo "Password is: ${{ secrets.IIS_PASSWORD }}"  # Shows: Password is: ***
      
      - name: Build Config (UNSAFE - DON'T DO THIS)
        run: |
          # Base64 encoding can bypass masking
          echo "${{ secrets.IIS_PASSWORD }}" | base64  # DANGEROUS!
```

## Security Checklist

- [ ] All passwords stored as GitHub Secrets
- [ ] Secrets are not hardcoded in workflow files
- [ ] Production deployments require manual approval
- [ ] Environment protection rules enabled
- [ ] Branch protection rules configured
- [ ] Audit logging enabled
- [ ] Secrets rotated regularly
- [ ] Principle of least privilege applied
- [ ] Short-lived credentials used when possible
- [ ] Sensitive data never logged

## Related Documentation

- [GitHub Actions](GITHUB_ACTIONS.md)
- [Terraform Integration](TERRAFORM_INTEGRATION.md)
- [Quick Start](QUICK_START_GITHUB_ACTIONS.md)
