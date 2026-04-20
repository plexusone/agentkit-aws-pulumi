// Package lightsail provides a deployment provider for AWS Lightsail Container Service.
package lightsail

import (
	"context"
	"fmt"
	"strconv"
	"time"

	pulumiutil "github.com/plexusone/agentkit-aws-pulumi/deploy/pulumi"
	"github.com/plexusone/agentkit/deploy"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/lightsail"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Provider implements the deploy.Provider interface for AWS Lightsail.
type Provider struct {
	cfg    *deploy.DeployConfig
	lsCfg  *Config
	closed bool
}

// Verify interface compliance at compile time.
var _ deploy.Provider = (*Provider)(nil)

// New creates a new Lightsail deployment provider.
func New(cfg *deploy.DeployConfig) (*Provider, error) {
	lsCfg := DefaultConfig()

	// Parse provider-specific config if provided
	if cfg.ProviderConfig != nil {
		if c, ok := cfg.ProviderConfig.(*Config); ok {
			lsCfg = c
		}
	}

	// Override region from stack config if set
	if cfg.Stack.Region != "" {
		lsCfg.Region = cfg.Stack.Region
	}

	// Validate configuration
	if err := lsCfg.Validate(); err != nil {
		return nil, fmt.Errorf("lightsail: invalid config: %w", err)
	}

	return &Provider{
		cfg:   cfg,
		lsCfg: lsCfg,
	}, nil
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return string(deploy.ProviderLightsail)
}

// Capabilities returns the provider capabilities.
func (p *Provider) Capabilities() deploy.Capabilities {
	spec := p.lsCfg.GetPowerSpec()
	return deploy.Capabilities{
		AutoScaling:        false, // Lightsail has manual scaling only
		CustomDomain:       true,
		HTTPS:              true, // Lightsail provides HTTPS via default endpoint
		VPC:                false,
		SecretsIntegration: false, // Use environment variables
		Preview:            true,
		Rollback:           false,
		MaxMemoryMB:        spec.MemoryMB,
	}
}

// Deploy deploys the configuration to AWS Lightsail.
func (p *Provider) Deploy(ctx context.Context, cfg *deploy.DeployConfig) (*deploy.DeploymentStatus, error) {
	startTime := time.Now()

	status := &deploy.DeploymentStatus{
		StackName: cfg.Stack.Name,
		State:     deploy.StateInProgress,
		Provider:  p.Name(),
		StartTime: startTime,
		Outputs:   make(map[string]string),
	}

	// Create the Pulumi stack
	stack, err := pulumiutil.NewStack(ctx, pulumiutil.StackOptions{
		ProjectName: cfg.Stack.Project,
		StackName:   cfg.Stack.Name,
		BackendURL:  cfg.PulumiBackend,
		Config: map[string]string{
			"aws:region": p.lsCfg.Region,
		},
	}, p.createProgram(cfg))
	if err != nil {
		status.State = deploy.StateFailed
		status.Error = err.Error()
		status.EndTime = time.Now()
		status.Duration = status.EndTime.Sub(startTime)
		return status, fmt.Errorf("lightsail: failed to create stack: %w", err)
	}
	defer func() { _ = stack.Close() }()

	// Run the deployment
	result, err := stack.Up(ctx)
	if err != nil {
		status.State = deploy.StateFailed
		status.Error = err.Error()
		status.EndTime = time.Now()
		status.Duration = status.EndTime.Sub(startTime)
		return status, fmt.Errorf("lightsail: deployment failed: %w", err)
	}

	// Update status with results
	status.State = deploy.StateSucceeded
	status.EndTime = time.Now()
	status.Duration = status.EndTime.Sub(startTime)
	status.Outputs = result.Outputs

	// Add resource summary
	status.Resources = []deploy.Resource{
		{
			Type:  "aws:lightsail:ContainerService",
			Name:  p.getServiceName(cfg),
			State: "created",
		},
	}

	return status, nil
}

// Status returns the current deployment status.
func (p *Provider) Status(ctx context.Context, stackName string) (*deploy.DeploymentStatus, error) {
	stack, err := pulumiutil.NewStack(ctx, pulumiutil.StackOptions{
		ProjectName: p.cfg.Stack.Project,
		StackName:   stackName,
		BackendURL:  p.cfg.PulumiBackend,
	}, func(ctx *pulumi.Context) error {
		// Empty program for status check
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("lightsail: failed to get stack: %w", err)
	}
	defer func() { _ = stack.Close() }()

	outputs, err := stack.Outputs(ctx)
	if err != nil {
		return nil, fmt.Errorf("lightsail: failed to get outputs: %w", err)
	}

	return &deploy.DeploymentStatus{
		StackName: stackName,
		State:     deploy.StateSucceeded,
		Provider:  p.Name(),
		Outputs:   outputs,
	}, nil
}

// Destroy removes all resources associated with a deployment.
func (p *Provider) Destroy(ctx context.Context, stackName string) error {
	stack, err := pulumiutil.NewStack(ctx, pulumiutil.StackOptions{
		ProjectName: p.cfg.Stack.Project,
		StackName:   stackName,
		BackendURL:  p.cfg.PulumiBackend,
		Config: map[string]string{
			"aws:region": p.lsCfg.Region,
		},
	}, p.createProgram(p.cfg))
	if err != nil {
		return fmt.Errorf("lightsail: failed to get stack: %w", err)
	}
	defer func() { _ = stack.Close() }()

	if err := stack.Destroy(ctx); err != nil {
		return fmt.Errorf("lightsail: destroy failed: %w", err)
	}

	return nil
}

// Close releases resources held by the provider.
func (p *Provider) Close() error {
	p.closed = true
	return nil
}

// createProgram creates the Pulumi program for Lightsail deployment.
func (p *Provider) createProgram(cfg *deploy.DeployConfig) pulumi.RunFunc {
	return func(ctx *pulumi.Context) error {
		serviceName := p.getServiceName(cfg)

		// Determine power level
		power := p.lsCfg.Power
		if power == "" {
			power = "nano"
		}

		// Determine scale
		scale := p.lsCfg.Scale
		if scale == 0 {
			scale = 1
		}

		// Create container service
		service, err := lightsail.NewContainerService(ctx, serviceName, &lightsail.ContainerServiceArgs{
			Name:       pulumi.String(serviceName),
			Power:      pulumi.String(power),
			Scale:      pulumi.Int(scale),
			IsDisabled: pulumi.Bool(p.lsCfg.IsDisabled),
			Tags:       pulumi.ToStringMap(p.mergeTags(cfg)),
		})
		if err != nil {
			return fmt.Errorf("failed to create container service: %w", err)
		}

		// Build container configuration
		containers := lightsail.ContainerServiceDeploymentVersionContainerArray{
			lightsail.ContainerServiceDeploymentVersionContainerArgs{
				ContainerName: pulumi.String("app"),
				Image:         pulumi.String(p.getImageURI(cfg)),
				Ports: pulumi.StringMap{
					strconv.Itoa(cfg.Stack.Resources.Port): pulumi.String("HTTP"),
				},
				Environment: pulumi.ToStringMap(cfg.Stack.Environment),
			},
		}

		// Configure health check
		healthCheck := lightsail.ContainerServiceDeploymentVersionPublicEndpointHealthCheckArgs{
			Path:               pulumi.String("/health"),
			IntervalSeconds:    pulumi.Int(30),
			TimeoutSeconds:     pulumi.Int(5),
			HealthyThreshold:   pulumi.Int(2),
			UnhealthyThreshold: pulumi.Int(3),
		}
		if cfg.Stack.HealthCheck != nil {
			if cfg.Stack.HealthCheck.Path != "" {
				healthCheck.Path = pulumi.String(cfg.Stack.HealthCheck.Path)
			}
			if cfg.Stack.HealthCheck.IntervalSeconds > 0 {
				healthCheck.IntervalSeconds = pulumi.Int(cfg.Stack.HealthCheck.IntervalSeconds)
			}
			if cfg.Stack.HealthCheck.TimeoutSeconds > 0 {
				healthCheck.TimeoutSeconds = pulumi.Int(cfg.Stack.HealthCheck.TimeoutSeconds)
			}
			if cfg.Stack.HealthCheck.HealthyThreshold > 0 {
				healthCheck.HealthyThreshold = pulumi.Int(cfg.Stack.HealthCheck.HealthyThreshold)
			}
			if cfg.Stack.HealthCheck.UnhealthyThreshold > 0 {
				healthCheck.UnhealthyThreshold = pulumi.Int(cfg.Stack.HealthCheck.UnhealthyThreshold)
			}
		}

		// Create deployment
		_, err = lightsail.NewContainerServiceDeploymentVersion(ctx, serviceName+"-deployment", &lightsail.ContainerServiceDeploymentVersionArgs{
			ServiceName: service.Name,
			Containers:  containers,
			PublicEndpoint: lightsail.ContainerServiceDeploymentVersionPublicEndpointArgs{
				ContainerName: pulumi.String("app"),
				ContainerPort: pulumi.Int(cfg.Stack.Resources.Port),
				HealthCheck:   &healthCheck,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to create deployment: %w", err)
		}

		// Export outputs
		ctx.Export("serviceName", service.Name)
		ctx.Export("serviceUrl", service.Url)
		ctx.Export("state", service.State)
		ctx.Export("region", pulumi.String(p.lsCfg.Region))

		return nil
	}
}

// getServiceName returns the Lightsail service name.
func (p *Provider) getServiceName(cfg *deploy.DeployConfig) string {
	if p.lsCfg.ServiceName != "" {
		return p.lsCfg.ServiceName
	}
	return cfg.Stack.Name
}

// getImageURI returns the container image URI.
func (p *Provider) getImageURI(cfg *deploy.DeployConfig) string {
	tag := cfg.Stack.Image.Tag
	if tag == "" {
		tag = "latest"
	}
	if cfg.Stack.Image.Digest != "" {
		return cfg.Stack.Image.Repository + "@" + cfg.Stack.Image.Digest
	}
	return cfg.Stack.Image.Repository + ":" + tag
}

// mergeTags merges provider tags with stack tags.
func (p *Provider) mergeTags(cfg *deploy.DeployConfig) map[string]string {
	tags := make(map[string]string)
	for k, v := range cfg.Stack.Tags {
		tags[k] = v
	}
	for k, v := range p.lsCfg.Tags {
		tags[k] = v
	}
	// Add default tags
	tags["managed-by"] = "agentkit-deploy"
	tags["provider"] = "lightsail"
	return tags
}
