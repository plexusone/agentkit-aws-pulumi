package lightsail

import (
	"github.com/plexusone/agentkit/deploy"
)

func init() {
	// Register the Lightsail provider with high priority (100) as it's the default.
	deploy.RegisterProvider(deploy.ProviderLightsail, newProvider, 100)
}

// newProvider is the factory function for creating Lightsail providers.
func newProvider(cfg *deploy.DeployConfig) (deploy.Provider, error) {
	return New(cfg)
}
