# Lissto E2E Tests

End-to-end testing for the complete Lissto flow: **CLI → API → Controller**

## Overview

This repository contains e2e tests that validate the full Lissto stack:

- **Blueprint creation** (admin role → global namespace)
- **Stack lifecycle** (user role → prepare → create → update → delete)
- **Image updates** and rollouts
- **Resource cleanup**

## Quick Start

### Prerequisites

- Go 1.22+
- Docker
- kubectl
- Helm 3+

### Run Full E2E Suite

```bash
# Install dependencies and run full e2e tests
make e2e

# Clean up after tests
make e2e-clean
```

### Run Individual Steps

```bash
# Create cluster
make cluster-create

# Deploy Lissto
make deploy

# Download and configure CLI
make download-cli
make setup-cli

# Wait for ready
make wait-ready

# Run tests
make test

# Clean up
make cluster-delete
```

## Test Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│  1. SETUP                                                               │
│     └─► k3d cluster + Helm deploy + CLI download + contexts configured  │
├─────────────────────────────────────────────────────────────────────────┤
│  2. BLUEPRINT CREATE (Admin Role)                                       │
│     └─► lissto blueprint create --repository <test-repo> compose.yaml   │
│     └─► Verify: Blueprint CRD in global namespace with annotations      │
├─────────────────────────────────────────────────────────────────────────┤
│  3. STACK CREATE (User Role)                                            │
│     └─► CLI auto-creates Env on first use                               │
│     └─► lissto stack create <global/blueprint-id>                       │
│         └─► Phase 1: prepare → request_id + resolved images             │
│         └─► Phase 2: create  → Stack CRD + ConfigMap + K8s resources    │
│     └─► Verify: Stack CRD, Deployment, Service, (Ingress)               │
├─────────────────────────────────────────────────────────────────────────┤
│  4. IMAGE UPDATE (User Role)                                            │
│     └─► Update stack with new image tag                                 │
│     └─► Verify: Deployment rollout with new image                       │
├─────────────────────────────────────────────────────────────────────────┤
│  5. CLEANUP                                                             │
│     └─► User deletes stack → verify resources removed                   │
│     └─► Admin deletes blueprint → verify CRD removed                    │
└─────────────────────────────────────────────────────────────────────────┘
```

## Version Selection

You can test specific versions of each component:

```bash
# Test specific versions
make e2e CLI_REF=v0.1.6 API_TAG=v1.0.0 CONTROLLER_TAG=v1.0.0

# Test from main branches
make e2e CLI_REF=main API_TAG=main CONTROLLER_TAG=main

# Use versions from Helm chart defaults
make e2e USE_HELM_VERSIONS=true
```

### Version Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `CLI_REF` | CLI version (tag like `v0.1.6` or `latest` for release, branch name for source build) | `latest` |
| `API_TAG` | API container image tag | `main` |
| `CONTROLLER_TAG` | Controller container image tag | `main` |
| `HELM_REF` | Helm chart git ref | `main` |
| `USE_HELM_VERSIONS` | Use versions from Helm values.yaml | `false` |

## Running Specific Tests

```bash
# Run only blueprint tests
make test-focus FOCUS="Blueprint"

# Run only stack tests
make test-focus FOCUS="Stack"

# List all tests
make test-dry-run
```

## GitHub Actions

Tests run automatically:
- **Nightly** at 2 AM UTC (tests `main` branches)
- **Manual trigger** with version selection

Trigger manually from GitHub Actions UI with custom version inputs.

## Test Fixtures

### Docker Compose Files

- `fixtures/simple-nginx.yaml` - Single service with public image (nginx)
- `fixtures/lissto-stack.yaml` - Dogfooding with Lissto images

### Configuration

- `fixtures/helm-values-test.yaml` - Helm values for test deployment

## Local Development

### Adding New Tests

1. Create test file in `tests/` directory (e.g., `tests/06_new_feature_test.go`)
2. Use Ginkgo BDD style:

```go
var _ = Describe("New Feature", func() {
    It("should do something", func() {
        // Use helpers for CLI and K8s operations
        output, err := helpers.RunCLI("some", "command")
        Expect(err).NotTo(HaveOccurred())

        Eventually(func() bool {
            return helpers.ResourceExists("kind", "name", "namespace")
        }, "30s", "2s").Should(BeTrue())
    })
})
```

### Test Helpers

- `helpers.RunCLI(args...)` - Execute CLI commands
- `helpers.RunCLIAs(role, args...)` - Execute as specific role (admin/user)
- `helpers.ResourceExists(kind, name, namespace)` - Check K8s resource
- `helpers.WaitForResource(kind, name, namespace, timeout)` - Wait for resource

## Troubleshooting

### Cluster Issues

```bash
# Check cluster status
make cluster-status

# Delete and recreate cluster
make cluster-delete cluster-create
```

### Test Failures

```bash
# Check Lissto pods
kubectl get pods -n lissto-system

# Check Lissto logs
kubectl logs -n lissto-system -l app=lissto-api
kubectl logs -n lissto-system -l app=lissto-controller
```

## License

See [LICENSE](LICENSE) for details.
