// Package agentcore provides Pulumi components for AgentCore deployments on AWS.
package agentcore

import (
	"github.com/agentplexus/agentkit/platforms/agentcore/iac"
)

// Re-export config types from agentkit for convenience.
type (
	StackConfig         = iac.StackConfig
	AgentConfig         = iac.AgentConfig
	VPCConfig           = iac.VPCConfig
	SecretsConfig       = iac.SecretsConfig
	ObservabilityConfig = iac.ObservabilityConfig
	IAMConfig           = iac.IAMConfig
)

// Re-export config loading functions from agentkit.
var (
	// LoadStackConfigFromFile loads a StackConfig from a JSON or YAML file.
	LoadStackConfigFromFile = iac.LoadStackConfigFromFile

	// LoadStackConfigFromJSON parses a StackConfig from JSON data.
	LoadStackConfigFromJSON = iac.LoadStackConfigFromJSON

	// LoadStackConfigFromYAML parses a StackConfig from YAML data.
	LoadStackConfigFromYAML = iac.LoadStackConfigFromYAML

	// JSONConfigExample returns an example JSON configuration.
	JSONConfigExample = iac.JSONConfigExample

	// YAMLConfigExample returns an example YAML configuration.
	YAMLConfigExample = iac.YAMLConfigExample

	// WriteExampleConfig writes an example configuration file.
	WriteExampleConfig = iac.WriteExampleConfig

	// Default config functions
	DefaultAgentConfig          = iac.DefaultAgentConfig
	DefaultVPCConfig            = iac.DefaultVPCConfig
	DefaultObservabilityConfig  = iac.DefaultObservabilityConfig
	DefaultIAMConfig            = iac.DefaultIAMConfig
	ValidMemoryValues           = iac.ValidMemoryValues
	ValidObservabilityProviders = iac.ValidObservabilityProviders
)
