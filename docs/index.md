# AgentKit for AWS Pulumi

Pulumi components for deploying [agentkit](https://github.com/plexusone/agentkit)-based agents to AWS.

## Features

- **Native Go Pulumi components** for AWS Bedrock AgentCore deployment
- **AWS Lightsail Container Service** deployment provider
- **Fluent builder API** matching agentkit-aws-cdk patterns
- **Configuration loading** from JSON or YAML files

## Installation

```bash
go get github.com/plexusone/agentkit-aws-pulumi
```

## Quick Start

### Lightsail Container Deployment

Deploy agents to AWS Lightsail Container Service:

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

### AgentCore with Go Constructs

Type-safe Go code with full IDE support:

```go
import (
    "github.com/plexusone/agentkit-aws-pulumi/agentcore"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        research := agentcore.NewAgentBuilder("research", "ghcr.io/example/research:latest").
            WithMemory(512).
            Build()

        _, err := agentcore.NewStackBuilder("my-agents").
            WithAgents(research).
            WithNewVPC("10.0.0.0/16", 2).
            Build(ctx)

        return err
    })
}
```

### AgentCore with YAML Config

Configuration-driven deployments with minimal code:

```go
func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        _, err := agentcore.NewStackFromFile(ctx, "config.yaml")
        return err
    })
}
```

## Related Modules

| Module | Purpose |
|--------|---------|
| [agentkit](https://github.com/plexusone/agentkit) | Core agent framework |
| [agentkit-aws-cdk](https://github.com/plexusone/agentkit-aws-cdk) | AWS CDK constructs |
| [agentkit-terraform](https://github.com/plexusone/agentkit-terraform) | Terraform modules |

## Documentation

- [AWS Deployment Guide](aws-deployment-guide.md) - Detailed deployment options
- [Releases](releases/v0.3.0.md) - Release notes and changelog

## License

MIT
