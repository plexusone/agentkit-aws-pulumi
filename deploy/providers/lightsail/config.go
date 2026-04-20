package lightsail

// Config contains AWS Lightsail-specific configuration.
type Config struct {
	// Region is the AWS region (e.g., "us-east-1").
	Region string `yaml:"region" json:"region"`

	// ServiceName is the Lightsail container service name.
	ServiceName string `yaml:"service_name" json:"service_name"`

	// Power is the container service power level.
	// Valid values: "nano", "micro", "small", "medium", "large", "xlarge".
	// nano: 0.25 vCPU, 512 MB RAM
	// micro: 0.5 vCPU, 1 GB RAM
	// small: 1 vCPU, 2 GB RAM
	// medium: 2 vCPU, 4 GB RAM
	// large: 4 vCPU, 8 GB RAM
	// xlarge: 8 vCPU, 16 GB RAM
	Power string `yaml:"power" json:"power"`

	// Scale is the number of container instances (1-20).
	Scale int `yaml:"scale" json:"scale"`

	// IsDisabled sets whether the service is disabled.
	IsDisabled bool `yaml:"is_disabled" json:"is_disabled"`

	// PublicDomainNames maps custom domains to container names.
	// Key: container name, Value: list of domain names.
	PublicDomainNames map[string][]string `yaml:"public_domain_names" json:"public_domain_names"`

	// Tags are resource tags.
	Tags map[string]string `yaml:"tags" json:"tags"`
}

// PowerSpec describes the resources for a power level.
type PowerSpec struct {
	VCPU     float64
	MemoryMB int
}

// PowerSpecs maps power levels to their resource specifications.
var PowerSpecs = map[string]PowerSpec{
	"nano":   {VCPU: 0.25, MemoryMB: 512},
	"micro":  {VCPU: 0.5, MemoryMB: 1024},
	"small":  {VCPU: 1, MemoryMB: 2048},
	"medium": {VCPU: 2, MemoryMB: 4096},
	"large":  {VCPU: 4, MemoryMB: 8192},
	"xlarge": {VCPU: 8, MemoryMB: 16384},
}

// DefaultConfig returns a Config with default values.
func DefaultConfig() *Config {
	return &Config{
		Region: "us-east-1",
		Power:  "nano",
		Scale:  1,
	}
}

// Validate checks that the configuration is valid.
func (c *Config) Validate() error {
	if c.Power != "" {
		if _, ok := PowerSpecs[c.Power]; !ok {
			return &ConfigError{Field: "power", Message: "invalid power level"}
		}
	}

	if c.Scale < 0 || c.Scale > 20 {
		return &ConfigError{Field: "scale", Message: "scale must be between 1 and 20"}
	}

	return nil
}

// GetPowerSpec returns the resource specification for the current power level.
func (c *Config) GetPowerSpec() PowerSpec {
	power := c.Power
	if power == "" {
		power = "nano"
	}
	return PowerSpecs[power]
}

// ConfigError represents a configuration validation error.
type ConfigError struct {
	Field   string
	Message string
}

// Error implements the error interface.
func (e *ConfigError) Error() string {
	return "lightsail: config " + e.Field + ": " + e.Message
}
