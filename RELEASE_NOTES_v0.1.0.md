# AgentKit for AWS Pulumi v0.1.0 Release Notes

**Release Date:** December 31, 2025

Initial release of Pulumi components for deploying agentkit-based agents to AWS Bedrock AgentCore.

## Highlights

- **Native Go Pulumi components** for AgentCore deployment
- **Fluent builder API** matching agentkit-aws-cdk patterns
- **Shared configuration schema** with agentkit core and agentkit-aws-cdk
- **Full AWS resource management** via Pulumi state

## Features

### Pulumi Components

Native Go components for AWS Bedrock AgentCore:

```go
import (
    "github.com/agentplexus/agentkit-aws-pulumi/agentcore"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        _, err := agentcore.NewStackBuilder("my-agents").
            WithAgents(research, orchestration).
            WithOpik("my-project", secretARN).
            WithTags(map[string]string{"Environment": "production"}).
            Build(ctx)
        return err
    })
}
```

### Fluent Builders

- `AgentBuilder` - Configure individual agents with memory, timeout, environment variables
- `StackBuilder` - Compose agents with VPC, observability, and IAM settings

### Configuration Loading

Load stack configuration from JSON or YAML files:

```go
pulumi.Run(func(ctx *pulumi.Context) error {
    _, err := agentcore.NewStackFromFile(ctx, "config.yaml")
    return err
})
```

### AWS Resources Created

- **VPC** with public/private subnets across multiple AZs
- **NAT Gateway** for private subnet internet access
- **Security Groups** for agent network isolation
- **IAM Roles** with Bedrock and Secrets Manager access
- **CloudWatch Log Groups** with configurable retention

## Deployment Approaches

| Approach | Description |
|----------|-------------|
| **Go Constructs** | Full type safety, IDE support, complex logic |
| **JSON/YAML Config** | Configuration-driven with minimal Go wrapper |

## Configuration

Uses shared configuration schema from `agentkit/platforms/agentcore/iac/`:

```yaml
stackName: my-agents
agents:
  - name: research
    containerImage: ghcr.io/example/research:latest
    memoryMB: 512
    timeoutSeconds: 30
  - name: orchestration
    containerImage: ghcr.io/example/orchestration:latest
    memoryMB: 1024
    isDefault: true
vpc:
  createVPC: true
  vpcCidr: 10.0.0.0/16
observability:
  provider: opik
  project: my-agents
```

## Dependencies

- **340 transitive packages** (native Go Pulumi SDK)
- Requires `github.com/agentplexus/agentkit` for shared configuration
- No Node.js runtime required (unlike CDK)

## Comparison with agentkit-aws-cdk

| Feature | agentkit-aws-cdk | agentkit-aws-pulumi |
|---------|------------------|---------------------|
| Runtime | Node.js (jsii) | Native Go |
| State management | CloudFormation | Pulumi Cloud/S3/Local |
| Dependencies | 21 | 340 |
| Preview command | `cdk diff` | `pulumi preview` |
| Deploy command | `cdk deploy` | `pulumi up` |

## Related Modules

| Module | Purpose |
|--------|---------|
| [agentkit](https://github.com/agentplexus/agentkit) | Core library, shared IaC config, pure CloudFormation |
| [agentkit-aws-cdk](https://github.com/agentplexus/agentkit-aws-cdk) | AWS CDK constructs |
| [agentkit-terraform](https://github.com/agentplexus/agentkit-terraform) | Terraform modules (planned) |

## Installation

```bash
go get github.com/agentplexus/agentkit-aws-pulumi
```

## Prerequisites

1. Pulumi CLI: `curl -fsSL https://get.pulumi.com | sh`
2. AWS credentials configured
3. Pulumi project initialized: `pulumi new go --name my-agents`

## License

MIT
