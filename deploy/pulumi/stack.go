// Package pulumi provides shared utilities for Pulumi automation API.
package pulumi

import (
	"context"
	"fmt"
	"os"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optpreview"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// StackOptions configures stack creation and management.
type StackOptions struct {
	// ProjectName is the Pulumi project name.
	ProjectName string

	// StackName is the fully qualified stack name.
	StackName string

	// WorkDir is the working directory (defaults to temp dir).
	WorkDir string

	// BackendURL is the Pulumi state backend URL.
	// Examples: "file://~/.pulumi", "s3://bucket/path", "https://api.pulumi.com"
	BackendURL string

	// SecretsProvider configures how secrets are encrypted.
	// Examples: "passphrase", "awskms://alias/my-key"
	SecretsProvider string

	// Config contains stack configuration values.
	Config map[string]string

	// SecretConfig contains secret configuration values.
	SecretConfig map[string]string

	// EnvVars are additional environment variables for the Pulumi process.
	EnvVars map[string]string
}

// Stack wraps a Pulumi automation stack with helper methods.
type Stack struct {
	stack    auto.Stack
	opts     StackOptions
	workDir  string
	isTempWD bool
}

// NewStack creates or selects an existing Pulumi stack.
func NewStack(ctx context.Context, opts StackOptions, program pulumi.RunFunc) (*Stack, error) {
	if opts.ProjectName == "" {
		return nil, fmt.Errorf("pulumi: project name is required")
	}
	if opts.StackName == "" {
		return nil, fmt.Errorf("pulumi: stack name is required")
	}

	workDir := opts.WorkDir
	isTempWD := false
	if workDir == "" {
		var err error
		workDir, err = os.MkdirTemp("", "pulumi-"+opts.ProjectName+"-")
		if err != nil {
			return nil, fmt.Errorf("pulumi: failed to create temp dir: %w", err)
		}
		isTempWD = true
	}

	// Set up environment variables
	for k, v := range opts.EnvVars {
		_ = os.Setenv(k, v)
	}

	// Configure the backend
	project := workspace.Project{
		Name:    tokens.PackageName(opts.ProjectName),
		Runtime: workspace.NewProjectRuntimeInfo("go", nil),
	}

	if opts.BackendURL != "" {
		project.Backend = &workspace.ProjectBackend{
			URL: opts.BackendURL,
		}
	}

	// Create or select stack
	stack, err := auto.UpsertStackInlineSource(ctx, opts.StackName, opts.ProjectName, program,
		auto.Project(project),
		auto.WorkDir(workDir),
	)
	if err != nil {
		if isTempWD {
			_ = os.RemoveAll(workDir)
		}
		return nil, fmt.Errorf("pulumi: failed to create/select stack: %w", err)
	}

	// Set configuration values
	for k, v := range opts.Config {
		if err := stack.SetConfig(ctx, k, auto.ConfigValue{Value: v}); err != nil {
			return nil, fmt.Errorf("pulumi: failed to set config %q: %w", k, err)
		}
	}

	// Set secret configuration values
	for k, v := range opts.SecretConfig {
		if err := stack.SetConfig(ctx, k, auto.ConfigValue{Value: v, Secret: true}); err != nil {
			return nil, fmt.Errorf("pulumi: failed to set secret config %q: %w", k, err)
		}
	}

	return &Stack{
		stack:    stack,
		opts:     opts,
		workDir:  workDir,
		isTempWD: isTempWD,
	}, nil
}

// Up runs a Pulumi up (deploy) operation.
func (s *Stack) Up(ctx context.Context, opts ...optup.Option) (*UpResult, error) {
	result, err := s.stack.Up(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("pulumi up failed: %w", err)
	}

	summary := ResultSummary{}
	if result.Summary.ResourceChanges != nil {
		changes := *result.Summary.ResourceChanges
		summary.Create = changes["create"]
		summary.Update = changes["update"]
		summary.Delete = changes["delete"]
		summary.Same = changes["same"]
	}

	return &UpResult{
		Summary: summary,
		Outputs: convertOutputs(result.Outputs),
	}, nil
}

// Preview runs a Pulumi preview operation.
func (s *Stack) Preview(ctx context.Context, opts ...optpreview.Option) (*PreviewResult, error) {
	result, err := s.stack.Preview(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("pulumi preview failed: %w", err)
	}

	return &PreviewResult{
		Summary: ResultSummary{
			Create: result.ChangeSummary["create"],
			Update: result.ChangeSummary["update"],
			Delete: result.ChangeSummary["delete"],
			Same:   result.ChangeSummary["same"],
		},
	}, nil
}

// Destroy runs a Pulumi destroy operation.
func (s *Stack) Destroy(ctx context.Context, opts ...optdestroy.Option) error {
	_, err := s.stack.Destroy(ctx, opts...)
	if err != nil {
		return fmt.Errorf("pulumi destroy failed: %w", err)
	}
	return nil
}

// Refresh refreshes the stack state.
func (s *Stack) Refresh(ctx context.Context) error {
	_, err := s.stack.Refresh(ctx)
	if err != nil {
		return fmt.Errorf("pulumi refresh failed: %w", err)
	}
	return nil
}

// Outputs returns the stack outputs.
func (s *Stack) Outputs(ctx context.Context) (map[string]string, error) {
	outputs, err := s.stack.Outputs(ctx)
	if err != nil {
		return nil, fmt.Errorf("pulumi outputs failed: %w", err)
	}
	return convertOutputs(outputs), nil
}

// Name returns the stack name.
func (s *Stack) Name() string {
	return s.stack.Name()
}

// Close cleans up stack resources.
func (s *Stack) Close() error {
	if s.isTempWD && s.workDir != "" {
		return os.RemoveAll(s.workDir)
	}
	return nil
}

// UpResult contains the result of a Pulumi up operation.
type UpResult struct {
	Summary ResultSummary
	Outputs map[string]string
}

// PreviewResult contains the result of a Pulumi preview operation.
type PreviewResult struct {
	Summary ResultSummary
}

// ResultSummary contains resource change counts.
type ResultSummary struct {
	Create int
	Update int
	Delete int
	Same   int
}

// Total returns the total number of resources affected.
func (s ResultSummary) Total() int {
	return s.Create + s.Update + s.Delete + s.Same
}

// convertOutputs converts Pulumi outputs to a string map.
func convertOutputs(outputs auto.OutputMap) map[string]string {
	result := make(map[string]string, len(outputs))
	for k, v := range outputs {
		if v.Value != nil {
			result[k] = fmt.Sprintf("%v", v.Value)
		}
	}
	return result
}
