# AgentKit for AWS Pulumi v0.3.0 Release Notes

**Release Date:** April 20, 2026

This release adds the AWS Lightsail Container Service deployment provider, implementing the new `deploy.Provider` interface from agentkit v0.6.0.

## Highlights

- **AWS Lightsail deployment provider** for containerized agent deployments
- **Pulumi automation utilities** for programmatic stack management

## New Features

### Lightsail Deployment Provider

Deploy agents to AWS Lightsail Container Service with a simple import:

```go
import (
    "github.com/plexusone/agentkit/deploy"
    _ "github.com/plexusone/agentkit-aws-pulumi/deploy/providers/lightsail"
)

cfg, _ := deploy.LoadDeployConfig("deploy.yaml")
provider, _ := deploy.GetProvider(cfg)
defer provider.Close()

status, _ := provider.Deploy(ctx, cfg)
fmt.Println(status.Outputs["serviceUrl"])
```

### Configuration

Configure deployments via YAML:

```yaml
stack:
  name: my-agent-prod
  project: my-project
  region: us-east-1
  image:
    repository: myorg/myagent
    tag: v1.0.0
  resources:
    port: 8080
  health_check:
    path: /health

provider: lightsail
```

### Lightsail Power Levels

| Power | vCPU | Memory |
|-------|------|--------|
| nano | 0.25 | 512 MB |
| micro | 0.5 | 1 GB |
| small | 1 | 2 GB |
| medium | 2 | 4 GB |
| large | 4 | 8 GB |
| xlarge | 8 | 16 GB |

### Pulumi Utilities

The `deploy/pulumi` package provides automation utilities:

```go
import "github.com/plexusone/agentkit-aws-pulumi/deploy/pulumi"

stack, err := pulumi.NewStack(ctx, pulumi.StackOptions{
    ProjectName: "my-project",
    StackName:   "prod",
    BackendURL:  "s3://my-bucket/pulumi",
    Config: map[string]string{
        "aws:region": "us-east-1",
    },
}, program)

result, err := stack.Up(ctx)
fmt.Printf("Created: %d, Updated: %d\n", result.Summary.Create, result.Summary.Update)
```

## Added

- `deploy/pulumi/stack.go` - Pulumi automation utilities (Stack, StackOptions, UpResult)
- `deploy/providers/lightsail/lightsail.go` - Lightsail Provider implementation
- `deploy/providers/lightsail/config.go` - Lightsail-specific configuration
- `deploy/providers/lightsail/init.go` - Auto-registration via init()

## Dependencies

| Package | From | To |
|---------|------|-----|
| `github.com/plexusone/agentkit` | v0.5.0 | v0.6.0 |

## Installation

```bash
go get github.com/plexusone/agentkit-aws-pulumi@v0.3.0
```

## License

MIT
