package helpers

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	// RoleAdmin is the admin role for read/delete operations
	RoleAdmin = "e2e-admin"
	// RoleDeploy is the deploy role for global blueprint creation
	RoleDeploy = "e2e-deploy"
	// RoleUser is the user role for stack and user blueprint operations
	RoleUser = "e2e-user"
)

// CLIRunner provides methods to execute Lissto CLI commands
type CLIRunner struct {
	binaryPath string
	configPath string
}

// NewCLIRunner creates a new CLI runner
func NewCLIRunner() *CLIRunner {
	// Try to find lissto binary
	binaryPath := "lissto"
	if path, err := exec.LookPath("lissto"); err == nil {
		binaryPath = path
	}

	// Use default config path
	configPath := filepath.Join(os.Getenv("HOME"), ".config", "lissto")

	return &CLIRunner{
		binaryPath: binaryPath,
		configPath: configPath,
	}
}

// Run executes a CLI command with the current context
func (r *CLIRunner) Run(args ...string) (string, error) {
	return r.runCommand(args...)
}

// RunAs executes a CLI command as a specific role (admin or user)
func (r *CLIRunner) RunAs(role string, args ...string) (string, error) {
	// Switch context first
	if _, err := r.runCommand("context", "use", role); err != nil {
		return "", fmt.Errorf("failed to switch to context %s: %w", role, err)
	}
	return r.runCommand(args...)
}

// RunAsAdmin executes a CLI command as admin
func (r *CLIRunner) RunAsAdmin(args ...string) (string, error) {
	return r.RunAs(RoleAdmin, args...)
}

// RunAsDeploy executes a CLI command as deploy
func (r *CLIRunner) RunAsDeploy(args ...string) (string, error) {
	return r.RunAs(RoleDeploy, args...)
}

// RunAsUser executes a CLI command as user
func (r *CLIRunner) RunAsUser(args ...string) (string, error) {
	return r.RunAs(RoleUser, args...)
}

// GetCurrentContext returns the current CLI context
func (r *CLIRunner) GetCurrentContext() (string, error) {
	output, err := r.runCommand("context", "current")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// runCommand executes the CLI binary with given arguments
func (r *CLIRunner) runCommand(args ...string) (string, error) {
	cmd := exec.Command(r.binaryPath, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String()

	if err != nil {
		// Include stderr in error message for debugging
		return output, fmt.Errorf("command failed: %w\nstderr: %s\nstdout: %s",
			err, stderr.String(), output)
	}

	return output, nil
}

// BlueprintCreate creates a blueprint from a compose file as user (goes to user namespace)
func (r *CLIRunner) BlueprintCreate(composePath, repository string) (string, error) {
	args := []string{"blueprint", "create", composePath}
	if repository != "" {
		args = append(args, "--repository", repository)
	}
	return r.RunAsUser(args...)
}

// BlueprintCreateGlobal creates a global blueprint using deploy role with --branch flag
func (r *CLIRunner) BlueprintCreateGlobal(composePath, repository, branch string) (string, error) {
	args := []string{"blueprint", "create", composePath}
	if repository != "" {
		args = append(args, "--repository", repository)
	}
	if branch != "" {
		args = append(args, "--branch", branch)
	}
	return r.RunAsDeploy(args...)
}

// BlueprintList lists all blueprints
func (r *CLIRunner) BlueprintList() (string, error) {
	return r.RunAsUser("blueprint", "list")
}

// BlueprintGet gets a specific blueprint
func (r *CLIRunner) BlueprintGet(name string) (string, error) {
	return r.RunAsUser("blueprint", "get", name)
}

// BlueprintDelete deletes a blueprint
func (r *CLIRunner) BlueprintDelete(name string) (string, error) {
	return r.RunAsAdmin("blueprint", "delete", name)
}

// EnvCreate creates an environment for the user
func (r *CLIRunner) EnvCreate(name string) (string, error) {
	return r.RunAsUser("env", "create", name)
}

// EnvUse selects an environment
func (r *CLIRunner) EnvUse(name string) (string, error) {
	return r.RunAsUser("env", "use", name)
}

// EnvExists checks if an environment exists (returns true if it does)
func (r *CLIRunner) EnvExists(name string) bool {
	_, err := r.RunAsUser("env", "get", name)
	return err == nil
}

// EnsureEnv creates an environment if it doesn't exist and selects it
func (r *CLIRunner) EnsureEnv(name string) error {
	// Try to use it first (will fail if doesn't exist)
	if _, err := r.EnvUse(name); err == nil {
		return nil
	}
	// Create it
	if _, err := r.EnvCreate(name); err != nil {
		return err
	}
	// Select it
	_, err := r.EnvUse(name)
	return err
}

// StackCreate creates a stack from a blueprint
func (r *CLIRunner) StackCreate(blueprintID string) (string, error) {
	return r.RunAsUser("stack", "create", blueprintID)
}

// StackList lists all stacks
func (r *CLIRunner) StackList() (string, error) {
	return r.RunAsUser("stack", "list")
}

// StackGet gets a specific stack
func (r *CLIRunner) StackGet(name string) (string, error) {
	return r.RunAsUser("stack", "get", name)
}

// StackDelete deletes a stack
func (r *CLIRunner) StackDelete(name string) (string, error) {
	return r.RunAsUser("stack", "delete", name)
}

// ExtractBlueprintID extracts the blueprint ID from create output
func ExtractBlueprintID(output string) string {
	// Look for "ID: xxx" pattern in output
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "ID:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	// If no ID: prefix found, return trimmed output (might be just the ID)
	return strings.TrimSpace(output)
}

// ExtractStackID extracts the stack ID from create output
func ExtractStackID(output string) string {
	return ExtractBlueprintID(output) // Same format
}
