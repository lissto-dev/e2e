package tests

import (
	"context"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/lissto-dev/e2e/tests/helpers"
)

var _ = Describe("Stack Lifecycle", Ordered, func() {
	var k8s *helpers.K8sClient
	var cli *helpers.CLIRunner
	var blueprintID string
	var stackID string
	var stackName string
	var userNamespace string
	ctx := context.Background()

	BeforeAll(func() {
		var err error
		k8s, err = helpers.NewK8sClient()
		Expect(err).NotTo(HaveOccurred(), "Failed to create K8s client")

		cli = helpers.NewCLIRunner()

		// User namespace is dev-<username>, where username is e2e-user
		userNamespace = helpers.GetUserNamespace("e2e-user")

		By("Creating a blueprint first (using admin)")
		fixturePath := helpers.GetFixturePath(helpers.FixtureSimpleNginx)
		output, err := cli.BlueprintCreate(fixturePath, helpers.TestRepository)
		Expect(err).NotTo(HaveOccurred(), "Blueprint creation should succeed for stack tests")
		blueprintID = helpers.ExtractBlueprintID(output)
		GinkgoWriter.Printf("Using blueprint: %s\n", blueprintID)
	})

	Describe("Stack Creation (User Role)", func() {
		It("should not allow admin to create stacks", func() {
			By("Attempting to create stack as admin")
			_, err := cli.RunAsAdmin("stack", "create", blueprintID)
			Expect(err).To(HaveOccurred(),
				"Admin should NOT be able to create stacks")
		})

		It("should create a stack from blueprint as user", func() {
			By("Creating stack from global blueprint")
			output, err := cli.StackCreate(blueprintID)
			Expect(err).NotTo(HaveOccurred(), "Stack creation should succeed: %s", output)

			stackID = helpers.ExtractStackID(output)
			Expect(stackID).NotTo(BeEmpty(), "Stack ID should be returned")

			// Extract stack name from ID
			parts := strings.Split(stackID, "/")
			if len(parts) == 2 {
				stackName = parts[1]
			} else {
				stackName = stackID
			}

			GinkgoWriter.Printf("Created stack: %s (name: %s)\n", stackID, stackName)
		})

		It("should create stack in user namespace", func() {
			By("Verifying stack exists in user namespace")
			Eventually(func() bool {
				return k8s.StackExists(ctx, userNamespace, stackName)
			}, 60*time.Second, 2*time.Second).Should(BeTrue(),
				"Stack should exist in user namespace: %s", userNamespace)
		})

		It("should create Stack CRD with correct spec", func() {
			By("Checking Stack spec")
			stack, err := k8s.GetStack(ctx, userNamespace, stackName)
			Expect(err).NotTo(HaveOccurred())

			spec, found, _ := stack.Object["spec"].(map[string]interface{})
			Expect(found).To(BeTrue(), "Stack should have spec")

			bpRef, _ := spec["blueprintReference"].(string)
			Expect(bpRef).To(Equal(blueprintID), "Stack should reference correct blueprint")
		})
	})

	Describe("Stack Resources", func() {
		It("should create Deployment for web service", func() {
			By("Waiting for Deployment to be created")
			// Deployment name is typically the service name from compose
			deploymentName := stackName + "-web"

			Eventually(func() bool {
				return k8s.DeploymentExists(ctx, userNamespace, deploymentName)
			}, 60*time.Second, 5*time.Second).Should(BeTrue(),
				"Deployment should be created for web service")
		})

		It("should create Service for web service", func() {
			By("Waiting for Service to be created")
			serviceName := stackName + "-web"

			Eventually(func() bool {
				return k8s.ServiceExists(ctx, userNamespace, serviceName)
			}, 30*time.Second, 2*time.Second).Should(BeTrue(),
				"Service should be created for web service")
		})

		It("should have ConfigMap with manifests", func() {
			By("Checking for manifests ConfigMap")
			// ConfigMap name is typically stack-name + suffix
			Eventually(func() bool {
				// Try common naming patterns
				if k8s.ConfigMapExists(ctx, userNamespace, stackName+"-manifests") {
					return true
				}
				return k8s.ConfigMapExists(ctx, userNamespace, stackName)
			}, 30*time.Second, 2*time.Second).Should(BeTrue(),
				"Manifests ConfigMap should exist")
		})
	})

	Describe("Stack Status", func() {
		It("should reach Running phase", func() {
			By("Waiting for stack to be Running")
			Eventually(func() string {
				phase, _ := k8s.GetStackPhase(ctx, userNamespace, stackName)
				return phase
			}, 120*time.Second, 5*time.Second).Should(Equal("Running"),
				"Stack should reach Running phase")
		})

		It("should have ready deployment", func() {
			By("Checking deployment readiness")
			deploymentName := stackName + "-web"

			Eventually(func() bool {
				return k8s.DeploymentReady(ctx, userNamespace, deploymentName)
			}, 120*time.Second, 5*time.Second).Should(BeTrue(),
				"Deployment should be ready")
		})
	})

	Describe("Stack Listing", func() {
		It("should list stacks including the created one", func() {
			output, err := cli.StackList()
			Expect(err).NotTo(HaveOccurred(), "Stack list should succeed")
			Expect(output).To(ContainSubstring(stackName),
				"Stack list should contain created stack")
		})
	})

	// Store stack info for subsequent tests
	AfterAll(func() {
		GinkgoWriter.Printf("Stack ID for cleanup tests: %s\n", stackID)
		GinkgoWriter.Printf("Stack name: %s\n", stackName)
		GinkgoWriter.Printf("User namespace: %s\n", userNamespace)
	})
})
