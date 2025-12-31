// Package agentcore provides Pulumi components for AgentCore deployments on AWS.
package agentcore

import (
	"github.com/agentplexus/agentkit/platforms/agentcore/iac"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// StackBuilder provides a fluent interface for building AgentCore stacks.
type StackBuilder struct {
	config iac.StackConfig
}

// NewStackBuilder creates a new stack builder.
func NewStackBuilder(stackName string) *StackBuilder {
	return &StackBuilder{
		config: iac.StackConfig{
			StackName: stackName,
			Agents:    []iac.AgentConfig{},
			Tags:      make(map[string]string),
		},
	}
}

// WithDescription sets the stack description.
func (b *StackBuilder) WithDescription(description string) *StackBuilder {
	b.config.Description = description
	return b
}

// WithAgent adds an agent to the stack.
func (b *StackBuilder) WithAgent(config iac.AgentConfig) *StackBuilder {
	b.config.Agents = append(b.config.Agents, config)
	return b
}

// WithAgents adds multiple agents to the stack.
func (b *StackBuilder) WithAgents(configs ...iac.AgentConfig) *StackBuilder {
	b.config.Agents = append(b.config.Agents, configs...)
	return b
}

// WithSimpleAgent adds an agent with minimal configuration.
func (b *StackBuilder) WithSimpleAgent(name, containerImage string) *StackBuilder {
	return b.WithAgent(iac.DefaultAgentConfig(name, containerImage))
}

// WithDefaultAgent sets an agent as the default.
func (b *StackBuilder) WithDefaultAgent(name, containerImage string) *StackBuilder {
	config := iac.DefaultAgentConfig(name, containerImage)
	config.IsDefault = true
	return b.WithAgent(config)
}

// WithVPC configures VPC settings.
func (b *StackBuilder) WithVPC(config *iac.VPCConfig) *StackBuilder {
	b.config.VPC = config
	return b
}

// WithExistingVPC uses an existing VPC.
func (b *StackBuilder) WithExistingVPC(vpcID string, subnetIDs []string) *StackBuilder {
	b.config.VPC = &iac.VPCConfig{
		VPCID:     vpcID,
		SubnetIDs: subnetIDs,
	}
	return b
}

// WithNewVPC creates a new VPC with the specified CIDR.
func (b *StackBuilder) WithNewVPC(cidr string, maxAZs int) *StackBuilder {
	b.config.VPC = &iac.VPCConfig{
		CreateVPC:          true,
		VPCCidr:            cidr,
		MaxAZs:             maxAZs,
		EnableVPCEndpoints: true,
	}
	return b
}

// WithSecrets configures secrets management.
func (b *StackBuilder) WithSecrets(config *iac.SecretsConfig) *StackBuilder {
	b.config.Secrets = config
	return b
}

// WithSecretValues creates secrets with the provided values.
func (b *StackBuilder) WithSecretValues(values map[string]string) *StackBuilder {
	b.config.Secrets = &iac.SecretsConfig{
		CreateSecrets: true,
		SecretValues:  values,
	}
	return b
}

// WithObservability configures observability.
func (b *StackBuilder) WithObservability(config *iac.ObservabilityConfig) *StackBuilder {
	b.config.Observability = config
	return b
}

// WithOpik configures Opik observability.
func (b *StackBuilder) WithOpik(project string, apiKeySecretARN string) *StackBuilder {
	b.config.Observability = &iac.ObservabilityConfig{
		Provider:             "opik",
		Project:              project,
		APIKeySecretARN:      apiKeySecretARN,
		EnableCloudWatchLogs: true,
		LogRetentionDays:     30,
	}
	return b
}

// WithLangfuse configures Langfuse observability.
func (b *StackBuilder) WithLangfuse(project string, apiKeySecretARN string) *StackBuilder {
	b.config.Observability = &iac.ObservabilityConfig{
		Provider:             "langfuse",
		Project:              project,
		APIKeySecretARN:      apiKeySecretARN,
		EnableCloudWatchLogs: true,
		LogRetentionDays:     30,
	}
	return b
}

// WithCloudWatchOnly configures CloudWatch-only observability.
func (b *StackBuilder) WithCloudWatchOnly(retentionDays int) *StackBuilder {
	b.config.Observability = &iac.ObservabilityConfig{
		Provider:             "cloudwatch",
		EnableCloudWatchLogs: true,
		LogRetentionDays:     retentionDays,
	}
	return b
}

// WithIAM configures IAM settings.
func (b *StackBuilder) WithIAM(config *iac.IAMConfig) *StackBuilder {
	b.config.IAM = config
	return b
}

// WithExistingRole uses an existing IAM role.
func (b *StackBuilder) WithExistingRole(roleARN string) *StackBuilder {
	b.config.IAM = &iac.IAMConfig{
		RoleARN: roleARN,
	}
	return b
}

// WithBedrockModels restricts Bedrock access to specific models.
func (b *StackBuilder) WithBedrockModels(modelIDs ...string) *StackBuilder {
	if b.config.IAM == nil {
		b.config.IAM = iac.DefaultIAMConfig()
	}
	b.config.IAM.BedrockModelIDs = modelIDs
	return b
}

// WithTags adds tags to all resources.
func (b *StackBuilder) WithTags(tags map[string]string) *StackBuilder {
	for k, v := range tags {
		b.config.Tags[k] = v
	}
	return b
}

// WithTag adds a single tag.
func (b *StackBuilder) WithTag(key, value string) *StackBuilder {
	b.config.Tags[key] = value
	return b
}

// WithRemovalPolicy sets the removal policy.
func (b *StackBuilder) WithRemovalPolicy(policy string) *StackBuilder {
	b.config.RemovalPolicy = policy
	return b
}

// RetainOnDelete sets the removal policy to retain.
func (b *StackBuilder) RetainOnDelete() *StackBuilder {
	return b.WithRemovalPolicy("retain")
}

// DestroyOnDelete sets the removal policy to destroy.
func (b *StackBuilder) DestroyOnDelete() *StackBuilder {
	return b.WithRemovalPolicy("destroy")
}

// Config returns the current configuration.
func (b *StackBuilder) Config() iac.StackConfig {
	return b.config
}

// Validate validates the current configuration.
func (b *StackBuilder) Validate() error {
	b.config.ApplyDefaults()
	return b.config.Validate()
}

// Build creates the AgentCore stack.
func (b *StackBuilder) Build(ctx *pulumi.Context) (*AgentCoreStack, error) {
	return NewAgentCoreStack(ctx, b.config)
}

// MustBuild creates the AgentCore stack, panicking on error.
func (b *StackBuilder) MustBuild(ctx *pulumi.Context) *AgentCoreStack {
	stack, err := b.Build(ctx)
	if err != nil {
		panic(err)
	}
	return stack
}

// AgentBuilder provides a fluent interface for building agent configurations.
type AgentBuilder struct {
	config iac.AgentConfig
}

// NewAgentBuilder creates a new agent builder.
func NewAgentBuilder(name, containerImage string) *AgentBuilder {
	return &AgentBuilder{
		config: iac.DefaultAgentConfig(name, containerImage),
	}
}

// WithDescription sets the agent description.
func (b *AgentBuilder) WithDescription(description string) *AgentBuilder {
	b.config.Description = description
	return b
}

// WithMemory sets the memory allocation in MB.
func (b *AgentBuilder) WithMemory(memoryMB int) *AgentBuilder {
	b.config.MemoryMB = memoryMB
	return b
}

// WithTimeout sets the timeout in seconds.
func (b *AgentBuilder) WithTimeout(timeoutSeconds int) *AgentBuilder {
	b.config.TimeoutSeconds = timeoutSeconds
	return b
}

// WithEnvironment sets environment variables.
func (b *AgentBuilder) WithEnvironment(env map[string]string) *AgentBuilder {
	for k, v := range env {
		b.config.Environment[k] = v
	}
	return b
}

// WithEnvVar adds a single environment variable.
func (b *AgentBuilder) WithEnvVar(key, value string) *AgentBuilder {
	b.config.Environment[key] = value
	return b
}

// WithSecrets adds secret ARNs.
func (b *AgentBuilder) WithSecrets(secretARNs ...string) *AgentBuilder {
	b.config.SecretsARNs = append(b.config.SecretsARNs, secretARNs...)
	return b
}

// AsDefault marks this agent as the default.
func (b *AgentBuilder) AsDefault() *AgentBuilder {
	b.config.IsDefault = true
	return b
}

// Build returns the agent configuration.
func (b *AgentBuilder) Build() iac.AgentConfig {
	return b.config
}
