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
	var userBlueprintID string
	var globalBlueprintID string
	var userNamespace string
	ctx := context.Background()

	BeforeAll(func() {
		var err error
		k8s, err = helpers.NewK8sClient()
		Expect(err).NotTo(HaveOccurred(), "Failed to create K8s client")

		cli = helpers.NewCLIRunner()
		userNamespace = helpers.GetUserNamespace("e2e-user")

		By("Creating user blueprint for stack tests")
		fixturePath := helpers.GetFixturePath(helpers.FixtureSimpleNginx)
		output, err := cli.BlueprintCreate(fixturePath, helpers.TestRepository)
		Expect(err).NotTo(HaveOccurred(), "User blueprint creation should succeed")
		userBlueprintID = helpers.ExtractBlueprintID(output)
		GinkgoWriter.Printf("User blueprint: %s\n", userBlueprintID)

		By("Creating global blueprint for stack tests")
		output, err = cli.BlueprintCreateGlobal(fixturePath, helpers.TestRepository, "main")
		Expect(err).NotTo(HaveOccurred(), "Global blueprint creation should succeed")
		globalBlueprintID = helpers.ExtractBlueprintID(output)
		GinkgoWriter.Printf("Global blueprint: %s\n", globalBlueprintID)
	})

	Describe("Stack Creation Restrictions", func() {
		It("should not allow admin to create stacks", func() {
			By("Attempting to create stack as admin from global blueprint")
			_, err := cli.RunAsAdmin("stack", "create", globalBlueprintID)
			Expect(err).To(HaveOccurred(),
				"Admin should NOT be able to create stacks")
		})

		It("should not allow deploy to create stacks", func() {
			By("Attempting to create stack as deploy from global blueprint")
			_, err := cli.RunAsDeploy("stack", "create", globalBlueprintID)
			Expect(err).To(HaveOccurred(),
				"Deploy should NOT be able to create stacks")
		})
	})

	Describe("Stack Creation from User Blueprint", Ordered, func() {
		var stackID string
		var stackName string

		It("should create stack from user's own blueprint", func() {
			By("Creating stack from user blueprint")
			output, err := cli.StackCreate(userBlueprintID)
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

			GinkgoWriter.Printf("Created stack from user blueprint: %s (name: %s)\n", stackID, stackName)
		})

		It("should exist in user namespace", func() {
			By("Verifying stack exists in user namespace")
			Eventually(func() bool {
				return k8s.StackExists(ctx, userNamespace, stackName)
			}, 60*time.Second, 2*time.Second).Should(BeTrue(),
				"Stack should exist in user namespace: %s", userNamespace)
		})

		It("should have correct blueprint reference", func() {
			By("Checking Stack spec")
			stack, err := k8s.GetStack(ctx, userNamespace, stackName)
			Expect(err).NotTo(HaveOccurred())

			spec, found := stack.Object["spec"].(map[string]interface{})
			Expect(found).To(BeTrue(), "Stack should have spec")

			bpRef, _ := spec["blueprintReference"].(string)
			Expect(bpRef).To(Equal(userBlueprintID), "Stack should reference user blueprint")
		})

		It("should reach Running phase", func() {
			By("Waiting for stack to be Running")
			Eventually(func() string {
				phase, _ := k8s.GetStackPhase(ctx, userNamespace, stackName)
				return phase
			}, 120*time.Second, 5*time.Second).Should(Equal("Running"),
				"Stack should reach Running phase")
		})

		AfterAll(func() {
			By("Cleaning up user blueprint stack for next test")
			if stackName != "" {
				_, err := cli.StackDelete(stackName)
				if err != nil {
					GinkgoWriter.Printf("Warning: failed to delete stack %s: %v\n", stackName, err)
				}

				// Wait for stack to be deleted
				Eventually(func() bool {
					return !k8s.StackExists(ctx, userNamespace, stackName)
				}, 60*time.Second, 5*time.Second).Should(BeTrue(),
					"Stack should be deleted")

				GinkgoWriter.Printf("Cleaned up stack: %s\n", stackName)
			}
		})
	})

	Describe("Stack Creation from Global Blueprint", Ordered, func() {
		var stackID string
		var stackName string

		It("should create stack from global blueprint", func() {
			By("Creating stack from global blueprint")
			output, err := cli.StackCreate(globalBlueprintID)
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

			GinkgoWriter.Printf("Created stack from global blueprint: %s (name: %s)\n", stackID, stackName)
		})

		It("should exist in user namespace", func() {
			By("Verifying stack exists in user namespace")
			Eventually(func() bool {
				return k8s.StackExists(ctx, userNamespace, stackName)
			}, 60*time.Second, 2*time.Second).Should(BeTrue(),
				"Stack should exist in user namespace: %s", userNamespace)
		})

		It("should have correct blueprint reference", func() {
			By("Checking Stack spec")
			stack, err := k8s.GetStack(ctx, userNamespace, stackName)
			Expect(err).NotTo(HaveOccurred())

			spec, found := stack.Object["spec"].(map[string]interface{})
			Expect(found).To(BeTrue(), "Stack should have spec")

			bpRef, _ := spec["blueprintReference"].(string)
			Expect(bpRef).To(Equal(globalBlueprintID), "Stack should reference global blueprint")
		})

		It("should reach Running phase", func() {
			By("Waiting for stack to be Running")
			Eventually(func() string {
				phase, _ := k8s.GetStackPhase(ctx, userNamespace, stackName)
				return phase
			}, 120*time.Second, 5*time.Second).Should(Equal("Running"),
				"Stack should reach Running phase")
		})

		It("should create Deployment for web service", func() {
			By("Waiting for Deployment to be created")
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
	})

	Describe("Stack Listing", func() {
		It("should list stacks", func() {
			output, err := cli.StackList()
			Expect(err).NotTo(HaveOccurred(), "Stack list should succeed")
			Expect(output).NotTo(BeEmpty(), "Stack list should return data")
		})
	})

	// Store info for subsequent tests
	AfterAll(func() {
		GinkgoWriter.Printf("User blueprint: %s\n", userBlueprintID)
		GinkgoWriter.Printf("Global blueprint: %s\n", globalBlueprintID)
		GinkgoWriter.Printf("User namespace: %s\n", userNamespace)
	})
})
