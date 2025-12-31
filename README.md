# AgentKit for AWS Pulumi

[![Build Status][build-status-svg]][build-status-url]
[![Lint Status][lint-status-svg]][lint-status-url]
[![Go Report Card][goreport-svg]][goreport-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![License][license-svg]][license-url]

Pulumi components for deploying [agentkit](https://github.com/agentplexus/agentkit)-based agents to AWS Bedrock AgentCore.

## Scope

This module provides **Pulumi** components for AWS. For other IaC tools:

| IaC Tool | Module | Dependencies |
|----------|--------|--------------|
| **AWS CDK** | [agentkit-aws-cdk](https://github.com/agentplexus/agentkit-aws-cdk) | 21 |
| **Pulumi** | `agentkit-aws-pulumi` (this module) | 340 |
| **Terraform** | [agentkit-terraform](https://github.com/agentplexus/agentkit-terraform) | 0 (HCL only) |
| **CloudFormation** | [agentkit](https://github.com/agentplexus/agentkit) (core) | 0 extra |

All modules share the same YAML/JSON configuration schema from `agentkit/platforms/agentcore/iac/`.

## Architecture

```
agentkit/                              # Core library (shared config)
├── platforms/agentcore/iac/
│   ├── config.go                      # Shared config structs
│   ├── loader.go                      # JSON/YAML loading
│   └── cloudformation.go              # Pure CF generator

agentkit-aws-pulumi/                   # Pulumi AWS components (this module)
├── agentcore/
│   ├── stack.go                       # Pulumi resources
│   ├── builder.go                     # Fluent builders
│   └── loader.go                      # Config re-exports
```

## Installation

```bash
go get github.com/agentplexus/agentkit-aws-pulumi
```

## Two Deployment Approaches

| Approach | Code Required | Best For |
|----------|---------------|----------|
| [1. Go Constructs](#1-go-constructs) | Full Go code | Type safety, IDE support, complex logic |
| [2. JSON/YAML Config](#2-jsonyaml-config) | Minimal Go | Configuration-driven deployments |

---

## 1. Go Constructs

Type-safe Go code with full IDE support and compile-time validation.

```go
package main

import (
	"github.com/agentplexus/agentkit-aws-pulumi/agentcore"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Build agents with fluent API
		research := agentcore.NewAgentBuilder("research", "ghcr.io/example/research:latest").
			WithMemory(512).
			WithTimeout(30).
			Build()

		orchestration := agentcore.NewAgentBuilder("orchestration", "ghcr.io/example/orchestration:latest").
			WithMemory(1024).
			WithTimeout(300).
			AsDefault().
			Build()

		// Build stack
		_, err := agentcore.NewStackBuilder("my-agents").
			WithDescription("My AgentCore deployment").
			WithAgents(research, orchestration).
			WithNewVPC("10.0.0.0/16", 2).
			WithOpik("my-project", "arn:aws:secretsmanager:us-east-1:123456789:secret:opik-key").
			WithTags(map[string]string{"Environment": "production"}).
			Build(ctx)

		return err
	})
}
```

**Deploy:**
```bash
pulumi up
```

---

## 2. JSON/YAML Config

Minimal Go wrapper that loads configuration from JSON or YAML files.

**main.go** (never changes):
```go
package main

import (
	"github.com/agentplexus/agentkit-aws-pulumi/agentcore"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		_, err := agentcore.NewStackFromFile(ctx, "config.yaml")
		return err
	})
}
```

**config.yaml**:
```yaml
stackName: my-agents
description: My AgentCore deployment

agents:
  - name: research
    containerImage: ghcr.io/example/research:latest
    memoryMB: 512
    timeoutSeconds: 30

  - name: orchestration
    containerImage: ghcr.io/example/orchestration:latest
    memoryMB: 1024
    timeoutSeconds: 300
    isDefault: true

vpc:
  createVPC: true
  vpcCidr: 10.0.0.0/16
  maxAZs: 2

observability:
  provider: opik
  project: my-project
  enableCloudWatchLogs: true

tags:
  Environment: production
```

**Deploy:**
```bash
pulumi up
```

---

## Configuration Reference

Configuration uses the same schema as [agentkit-aws-cdk](https://github.com/agentplexus/agentkit-aws-cdk). See the agentkit-aws-cdk README for full configuration reference.

### Quick Reference

| Config | Required | Description |
|--------|----------|-------------|
| `stackName` | Yes | Stack name |
| `agents` | Yes | List of agents |
| `vpc` | No | VPC config (creates new by default) |
| `observability` | No | Monitoring config |
| `iam` | No | IAM config |
| `tags` | No | Resource tags |

---

## Prerequisites

1. **Install Pulumi CLI**:
   ```bash
   curl -fsSL https://get.pulumi.com | sh
   ```

2. **Configure AWS credentials**:
   ```bash
   aws configure
   # or
   export AWS_ACCESS_KEY_ID=xxx
   export AWS_SECRET_ACCESS_KEY=xxx
   ```

3. **Initialize Pulumi project**:
   ```bash
   pulumi new go --name my-agents
   ```

---

## Comparison with agentkit-aws-cdk

| Feature | agentkit-aws-cdk | agentkit-aws-pulumi |
|---------|------------------|---------------------|
| Runtime | Node.js (jsii) | Native Go |
| State | CloudFormation | Pulumi Cloud/S3/Local |
| Dependencies | 21 | 340 |
| Multi-cloud | AWS only | AWS (this module) |
| Preview | `cdk diff` | `pulumi preview` |

---

## License

MIT

 [build-status-svg]: https://github.com/agentplexus/agentkit-aws-pulumi/actions/workflows/ci.yaml/badge.svg?branch=main
 [build-status-url]: https://github.com/agentplexus/agentkit-aws-pulumi/actions/workflows/ci.yaml
 [lint-status-svg]: https://github.com/agentplexus/agentkit-aws-pulumi/actions/workflows/lint.yaml/badge.svg?branch=main
 [lint-status-url]: https://github.com/agentplexus/agentkit-aws-pulumi/actions/workflows/lint.yaml
 [goreport-svg]: https://goreportcard.com/badge/github.com/agentplexus/agentkit-aws-pulumi
 [goreport-url]: https://goreportcard.com/report/github.com/agentplexus/agentkit-aws-pulumi
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/agentplexus/agentkit-aws-pulumi
 [docs-godoc-url]: https://pkg.go.dev/github.com/agentplexus/agentkit-aws-pulumi
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/agentplexus/agentkit-aws-pulumi/blob/master/LICENSE
