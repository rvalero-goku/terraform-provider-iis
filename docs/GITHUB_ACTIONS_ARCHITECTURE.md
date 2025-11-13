# GitHub Actions Workflow Architecture

## High-Level Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                     Your Repository                              │
│                                                                  │
│  ┌────────────────────────────────────────────────────────┐   │
│  │  .github/workflows/deploy-iis.yml                       │   │
│  │                                                          │   │
│  │  jobs:                                                   │   │
│  │    deploy:                                               │   │
│  │      uses: rvalero-goku/terraform-provider-iis/         │   │
│  │            .github/workflows/apply-iis-config.yml       │   │
│  │      with:                                               │   │
│  │        config_yaml: ${{ secrets.IIS_CONFIG }}           │   │
│  └────────────────────────────────────────────────────────┘   │
│                            │                                     │
│                            │ Trigger                             │
└────────────────────────────┼─────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│            IIS Provider Repository (Reusable Workflow)           │
│                                                                  │
│  Step 1: Checkout & Build Provider                              │
│  ┌──────────────────────────────────────────────────┐          │
│  │  • git clone provider repository                 │          │
│  │  • go build -o terraform-provider-iis            │          │
│  │  • Install to ~/.terraform.d/plugins/            │          │
│  └──────────────────────────────────────────────────┘          │
│                            │                                     │
│                            ▼                                     │
│  Step 2: Setup Terraform Environment                            │
│  ┌──────────────────────────────────────────────────┐          │
│  │  • Create terraform.rc with dev_overrides        │          │
│  │  • Install OpenTofu                              │          │
│  │  • Create working directory                      │          │
│  └──────────────────────────────────────────────────┘          │
│                            │                                     │
│                            ▼                                     │
│  Step 3: Prepare Configuration                                  │
│  ┌──────────────────────────────────────────────────┐          │
│  │  • Write iis-config.yaml                         │          │
│  │  • Copy main.tf from examples                    │          │
│  │  • Configure backend (optional)                  │          │
│  └──────────────────────────────────────────────────┘          │
│                            │                                     │
│                            ▼                                     │
│  Step 4: Terraform Operations                                   │
│  ┌──────────────────────────────────────────────────┐          │
│  │  • tofu init                                     │          │
│  │  • tofu validate                                 │          │
│  │  • tofu plan -out=tfplan                        │          │
│  │  • tofu apply tfplan (if auto_approve)          │          │
│  └──────────────────────────────────────────────────┘          │
│                            │                                     │
│                            ▼                                     │
│  Step 5: Outputs & Artifacts                                    │
│  ┌──────────────────────────────────────────────────┐          │
│  │  • Capture terraform outputs                     │          │
│  │  • Upload state files as artifacts               │          │
│  │  • Return outputs to caller                      │          │
│  └──────────────────────────────────────────────────┘          │
└─────────────────────────────┬───────────────────────────────────┘
                              │
                              │ Deploy to
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      IIS Servers                                 │
│                                                                  │
│  ┌──────────────┐     ┌──────────────┐     ┌──────────────┐   │
│  │   Server 1   │     │   Server 2   │     │   Server N   │   │
│  │              │     │              │     │              │   │
│  │  • App Pools │     │  • App Pools │     │  • App Pools │   │
│  │  • Websites  │     │  • Websites  │     │  • Websites  │   │
│  │  • Bindings  │     │  • Bindings  │     │  • Bindings  │   │
│  │  • Files     │     │  • Files     │     │  • Files     │   │
│  └──────────────┘     └──────────────┘     └──────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

## Detailed Workflow Steps

```
┌──────────────────────────────────────────────────────────────────┐
│  1. TRIGGER                                                       │
├──────────────────────────────────────────────────────────────────┤
│                                                                   │
│  • User clicks "Run workflow" in Actions tab                     │
│  • OR: Push to main branch                                       │
│  • OR: Scheduled cron job                                        │
│  • OR: Called from another workflow                              │
│                                                                   │
└───────────────────────────┬──────────────────────────────────────┘
                            │
                            ▼
┌──────────────────────────────────────────────────────────────────┐
│  2. CHECKOUT PROVIDER REPO                                        │
├──────────────────────────────────────────────────────────────────┤
│                                                                   │
│  actions/checkout@v4                                             │
│  • Clone terraform-provider-iis repository                       │
│  • Fetch all source code                                         │
│  • Prepare for build                                             │
│                                                                   │
└───────────────────────────┬──────────────────────────────────────┘
                            │
                            ▼
┌──────────────────────────────────────────────────────────────────┐
│  3. SETUP GO                                                      │
├──────────────────────────────────────────────────────────────────┤
│                                                                   │
│  actions/setup-go@v5                                             │
│  • Install Go 1.24                                               │
│  • Configure Go environment                                      │
│  • Enable Go module cache                                        │
│                                                                   │
└───────────────────────────┬──────────────────────────────────────┘
                            │
                            ▼
┌──────────────────────────────────────────────────────────────────┐
│  4. BUILD PROVIDER                                                │
├──────────────────────────────────────────────────────────────────┤
│                                                                   │
│  go build -o terraform-provider-iis_v0.1.0                       │
│                                                                   │
│  • Compile Go source code                                        │
│  • Create binary for linux_amd64                                 │
│  • Output: terraform-provider-iis_v0.1.0                         │
│                                                                   │
└───────────────────────────┬──────────────────────────────────────┘
                            │
                            ▼
┌──────────────────────────────────────────────────────────────────┐
│  5. INSTALL PROVIDER                                              │
├──────────────────────────────────────────────────────────────────┤
│                                                                   │
│  Target: ~/.terraform.d/plugins/terraform.local/maxjoehnk/iis/   │
│          0.1.0/linux_amd64/                                      │
│                                                                   │
│  • Create directory structure                                    │
│  • Copy provider binary                                          │
│  • Set executable permissions                                    │
│                                                                   │
└───────────────────────────┬──────────────────────────────────────┘
                            │
                            ▼
┌──────────────────────────────────────────────────────────────────┐
│  6. CONFIGURE TERRAFORM CLI                                       │
├──────────────────────────────────────────────────────────────────┤
│                                                                   │
│  Create: ~/.terraform.d/terraform.rc                             │
│                                                                   │
│  provider_installation {                                         │
│    dev_overrides {                                               │
│      "terraform.local/maxjoehnk/iis" = "<provider_path>"        │
│    }                                                             │
│    direct {                                                      │
│      exclude = ["terraform.local/*/*"]                          │
│    }                                                             │
│  }                                                               │
│                                                                   │
└───────────────────────────┬──────────────────────────────────────┘
                            │
                            ▼
┌──────────────────────────────────────────────────────────────────┐
│  7. SETUP OPENTOFU                                                │
├──────────────────────────────────────────────────────────────────┤
│                                                                   │
│  opentofu/setup-opentofu@v1                                      │
│  • Install OpenTofu (Terraform fork)                             │
│  • Configure tofu CLI                                            │
│  • No wrapper for direct access                                  │
│                                                                   │
└───────────────────────────┬──────────────────────────────────────┘
                            │
                            ▼
┌──────────────────────────────────────────────────────────────────┐
│  8. CREATE CONFIGURATION                                          │
├──────────────────────────────────────────────────────────────────┤
│                                                                   │
│  Files created:                                                  │
│  • iis-config.yaml    (from workflow input)                      │
│  • main.tf            (copied from examples/)                    │
│                                                                   │
│  main.tf loads config:                                           │
│  locals {                                                        │
│    config = yamldecode(file("iis-config.yaml"))                 │
│  }                                                               │
│                                                                   │
└───────────────────────────┬──────────────────────────────────────┘
                            │
                            ▼
┌──────────────────────────────────────────────────────────────────┐
│  9. TERRAFORM INIT                                                │
├──────────────────────────────────────────────────────────────────┤
│                                                                   │
│  tofu init -upgrade                                              │
│                                                                   │
│  • Initialize working directory                                  │
│  • Load provider (from local override)                           │
│  • Setup backend (local by default)                              │
│  • Download any other providers                                  │
│                                                                   │
└───────────────────────────┬──────────────────────────────────────┘
                            │
                            ▼
┌──────────────────────────────────────────────────────────────────┐
│  10. TERRAFORM VALIDATE                                           │
├──────────────────────────────────────────────────────────────────┤
│                                                                   │
│  tofu validate -no-color                                         │
│                                                                   │
│  • Check configuration syntax                                    │
│  • Validate resource definitions                                 │
│  • Verify provider configuration                                 │
│                                                                   │
└───────────────────────────┬──────────────────────────────────────┘
                            │
                            ▼
┌──────────────────────────────────────────────────────────────────┐
│  11. TERRAFORM PLAN                                               │
├──────────────────────────────────────────────────────────────────┤
│                                                                   │
│  tofu plan -no-color -out=tfplan                                 │
│                                                                   │
│  • Connect to IIS servers via API                                │
│  • Read current state                                            │
│  • Calculate changes needed                                      │
│  • Create execution plan                                         │
│  • Save plan to tfplan file                                      │
│                                                                   │
│  Output shows:                                                   │
│  • Resources to add    (green +)                                 │
│  • Resources to change (yellow ~)                                │
│  • Resources to delete (red -)                                   │
│                                                                   │
└───────────────────────────┬──────────────────────────────────────┘
                            │
                            ▼
┌──────────────────────────────────────────────────────────────────┐
│  12. TERRAFORM APPLY (if auto_approve=true)                       │
├──────────────────────────────────────────────────────────────────┤
│                                                                   │
│  tofu apply -auto-approve tfplan                                 │
│                                                                   │
│  For each server:                                                │
│  • Create/Update Application Pools                               │
│  • Create/Update Websites                                        │
│  • Configure Bindings                                            │
│  • Create Directories                                            │
│  • Copy Files                                                    │
│                                                                   │
│  • Update state file                                             │
│  • Report success/failure                                        │
│                                                                   │
└───────────────────────────┬──────────────────────────────────────┘
                            │
                            ▼
┌──────────────────────────────────────────────────────────────────┐
│  13. CAPTURE OUTPUTS                                              │
├──────────────────────────────────────────────────────────────────┤
│                                                                   │
│  tofu output -json                                               │
│                                                                   │
│  Outputs include:                                                │
│  • server1_resources (pools, websites, directories)              │
│  • server2_resources                                             │
│  • server1_certificates                                          │
│  • server2_certificates                                          │
│  • server1_existing_websites                                     │
│  • server2_existing_websites                                     │
│                                                                   │
└───────────────────────────┬──────────────────────────────────────┘
                            │
                            ▼
┌──────────────────────────────────────────────────────────────────┐
│  14. UPLOAD ARTIFACTS                                             │
├──────────────────────────────────────────────────────────────────┤
│                                                                   │
│  actions/upload-artifact@v4                                      │
│                                                                   │
│  Artifacts uploaded:                                             │
│  • terraform.tfstate   (current state)                           │
│  • tfplan              (execution plan)                          │
│                                                                   │
│  Retention: 30 days                                              │
│                                                                   │
└───────────────────────────┬──────────────────────────────────────┘
                            │
                            ▼
┌──────────────────────────────────────────────────────────────────┐
│  15. RETURN TO CALLER                                             │
├──────────────────────────────────────────────────────────────────┤
│                                                                   │
│  Workflow outputs:                                               │
│  • terraform_output (JSON)                                       │
│                                                                   │
│  Accessible in calling workflow via:                             │
│  ${{ needs.deploy.outputs.terraform_output }}                    │
│                                                                   │
└──────────────────────────────────────────────────────────────────┘
```

## Data Flow

```
GitHub Secrets          Workflow Input               Provider Repo
     │                       │                              │
     │                       │                              │
     ▼                       ▼                              ▼
┌─────────┐          ┌──────────────┐              ┌──────────────┐
│ IIS_    │          │ config_yaml  │              │  Go Source   │
│ CONFIG  │─────────>│              │              │    Code      │
└─────────┘          │ provider_    │              └──────────────┘
                     │ version      │                      │
                     │              │                      │
                     │ terraform_   │                      │
                     │ version      │                      ▼
                     │              │              ┌──────────────┐
                     │ auto_approve │              │   Provider   │
                     │              │              │    Binary    │
                     │ destroy      │              └──────────────┘
                     └──────────────┘                      │
                            │                              │
                            └──────────┬───────────────────┘
                                       │
                                       ▼
                            ┌─────────────────────┐
                            │   Terraform/Tofu    │
                            │   Configuration     │
                            └─────────────────────┘
                                       │
                                       ▼
                            ┌─────────────────────┐
                            │    IIS Servers      │
                            │   (via REST API)    │
                            └─────────────────────┘
                                       │
                                       ▼
                            ┌─────────────────────┐
                            │   State File +      │
                            │   Outputs           │
                            └─────────────────────┘
```

## Authentication Flow

```
┌─────────────────────────────────────────────────────────────────┐
│  Configuration (YAML)                                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  servers:                                                        │
│    server1:                                                      │
│      host: "https://server.example.com:55539"                   │
│      ntlm_username: "admin"                                     │
│      ntlm_password: "password"                                  │
│      ntlm_domain: ""                                            │
│                                                                  │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│  IIS Provider (Go)                                               │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  1. Receive NTLM credentials                                    │
│  2. Generate API token automatically                            │
│  3. Store token in provider state                               │
│                                                                  │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│  IIS REST API                                                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Endpoint: https://server.example.com:55539/api                 │
│                                                                  │
│  Authentication: Bearer token                                   │
│  • Generated from NTLM credentials                              │
│  • Auto-renewed if expired                                      │
│  • Transparent to user                                          │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## Multi-Server Deployment Flow

```
Configuration
     │
     ├─── Server 1 Config
     │         │
     │         ├─── Provider Instance 1
     │         │         │
     │         │         ├─── App Pools
     │         │         ├─── Websites
     │         │         ├─── Directories
     │         │         └─── Files
     │         │
     │         └─── Deploy to Server 1
     │
     ├─── Server 2 Config
     │         │
     │         ├─── Provider Instance 2
     │         │         │
     │         │         ├─── App Pools
     │         │         ├─── Websites
     │         │         ├─── Directories
     │         │         └─── Files
     │         │
     │         └─── Deploy to Server 2
     │
     └─── Server N Config
               │
               └─── ...
```

## Error Handling Flow

```
┌─────────────┐
│  Terraform  │
│    Init     │
└──────┬──────┘
       │
       ▼
   Success? ──No──> Fail workflow
       │
      Yes
       │
       ▼
┌─────────────┐
│  Terraform  │
│  Validate   │
└──────┬──────┘
       │
       ▼
   Success? ──No──> Fail workflow
       │
      Yes
       │
       ▼
┌─────────────┐
│  Terraform  │
│    Plan     │
└──────┬──────┘
       │
       ▼
   Success? ──No──> Fail workflow
       │
      Yes
       │
       ▼
  Auto-approve?
       │
   ┌───┴───┐
  Yes     No ──> Stop (plan only)
   │
   ▼
┌─────────────┐
│  Terraform  │
│    Apply    │
└──────┬──────┘
       │
       ▼
   Success? ──No──> Fail workflow + Upload artifacts
       │
      Yes
       │
       ▼
┌─────────────┐
│   Upload    │
│  Artifacts  │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   Return    │
│   Outputs   │
└─────────────┘
```

## Security Layers

```
┌─────────────────────────────────────────────────────────────┐
│  Layer 1: GitHub Secrets                                     │
│  • Encrypted at rest                                         │
│  • Masked in logs                                            │
│  • Access controlled by repository permissions               │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│  Layer 2: Workflow Inputs                                    │
│  • Secrets passed as workflow inputs                         │
│  • Not exposed in code                                       │
│  • Scoped to workflow run                                    │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│  Layer 3: Terraform State                                    │
│  • Sensitive data in state file                              │
│  • Uploaded as GitHub artifact                               │
│  • 30-day retention                                          │
│  • Encrypted in transit and at rest                          │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│  Layer 4: IIS API                                            │
│  • HTTPS/TLS encryption                                      │
│  • NTLM authentication                                       │
│  • Bearer token API calls                                    │
│  • Per-request authorization                                 │
└─────────────────────────────────────────────────────────────┘
```
