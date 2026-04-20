# AgentKit for AWS Pulumi v0.2.0 Release Notes

**Release Date:** March 2, 2026

Organization migration release: module path and all dependencies migrated from `github.com/agentplexus` to `github.com/plexusone`.

## Highlights

- **Module path migration** from `github.com/agentplexus/agentkit-aws-pulumi` to `github.com/plexusone/agentkit-aws-pulumi`
- **Shared CI/CD workflows** migrated to `plexusone/.github` for standardized pipeline management

## Breaking Changes

### Module Path Changed

The Go module path has changed:

```diff
- github.com/agentplexus/agentkit-aws-pulumi
+ github.com/plexusone/agentkit-aws-pulumi
```

### Upgrade Steps

1. Update your `go.mod`:

   ```bash
   go get github.com/plexusone/agentkit-aws-pulumi@v0.2.0
   ```

2. Update all import paths:

   ```diff
   - "github.com/agentplexus/agentkit-aws-pulumi/agentcore"
   + "github.com/plexusone/agentkit-aws-pulumi/agentcore"
   ```

3. Run `go mod tidy` to clean up dependencies

## Changed

- Updated dependency to `github.com/plexusone/agentkit` v0.5.0 (also migrated from agentplexus)
- Docker image references updated from `ghcr.io/agentplexus/` to `ghcr.io/plexusone/`

## Infrastructure

CI/CD workflows migrated to shared workflows from `plexusone/.github`:

- `go-ci.yaml` - Build and test workflow
- `go-lint.yaml` - golangci-lint workflow
- `go-sast-codeql.yaml` - CodeQL security scanning

## Dependencies

| Package | From | To |
|---------|------|-----|
| `github.com/plexusone/agentkit` | 0.4.0 | 0.5.0 |
| `github.com/pulumi/pulumi/sdk/v3` | 3.x | 3.224.0 |

## Related Migrations

This release is part of the organization-wide migration from `agentplexus` to `plexusone`:

| Module | Version |
|--------|---------|
| [agentkit](https://github.com/plexusone/agentkit) | v0.5.0 |
| [assistantkit](https://github.com/plexusone/assistantkit) | v0.11.0 |
| [agent-team-release](https://github.com/plexusone/agent-team-release) | v0.8.0 |
| **agentkit-aws-pulumi** | **v0.2.0** |

## Installation

```bash
go get github.com/plexusone/agentkit-aws-pulumi@v0.2.0
```

## License

MIT
