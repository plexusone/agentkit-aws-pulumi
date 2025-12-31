// Example: Basic AgentCore deployment with Pulumi
//
// This example demonstrates deploying a multi-agent system to AWS using Pulumi.
//
// Deploy with:
//
//	pulumi up
package main

import (
	"github.com/agentplexus/agentkit-aws-pulumi/agentcore"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Build agent configurations using the fluent builder API
		research := agentcore.NewAgentBuilder("research", "ghcr.io/agentplexus/stats-research:latest").
			WithDescription("Research agent - web search via Serper").
			WithMemory(512).
			WithTimeout(30).
			WithEnvVar("LOG_LEVEL", "info").
			Build()

		synthesis := agentcore.NewAgentBuilder("synthesis", "ghcr.io/agentplexus/stats-synthesis:latest").
			WithDescription("Synthesis agent - extract statistics from URLs").
			WithMemory(1024).
			WithTimeout(120).
			Build()

		verification := agentcore.NewAgentBuilder("verification", "ghcr.io/agentplexus/stats-verification:latest").
			WithDescription("Verification agent - validate sources").
			WithMemory(512).
			WithTimeout(60).
			Build()

		orchestration := agentcore.NewAgentBuilder("orchestration", "ghcr.io/agentplexus/stats-orchestration:latest").
			WithDescription("Orchestration agent - coordinate workflow").
			WithMemory(512).
			WithTimeout(300).
			AsDefault().
			Build()

		// Build the stack using the fluent builder API
		_, err := agentcore.NewStackBuilder("stats-agent-team").
			WithDescription("Statistics research and verification multi-agent system").
			WithAgents(research, synthesis, verification, orchestration).
			WithNewVPC("10.0.0.0/16", 2).
			WithOpik("stats-agent-team", "arn:aws:secretsmanager:us-east-1:123456789:secret:opik-key").
			WithTags(map[string]string{
				"Project":     "stats-agent-team",
				"Environment": "production",
				"Team":        "ai-platform",
			}).
			Build(ctx)

		return err
	})
}
