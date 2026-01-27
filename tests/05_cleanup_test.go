package tests

import (
	"context"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/lissto-dev/e2e/tests/helpers"
)

var _ = Describe("Cleanup", Ordered, func() {
	var k8s *helpers.K8sClient
	var cli *helpers.CLIRunner
	var userBlueprintID string
	var userBlueprintName string
	var globalBlueprintID string
	var globalBlueprintName string
	var stackID string
	var stackName string
	var userNamespace string
	ctx := context.Background()

	BeforeAll(func() {
		var err error
		k8s, err = helpers.NewK8sClient()
		Expect(err).NotTo(HaveOccurred(), "Failed to create K8s client")

		cli = helpers.NewCLIRunner()
		userNamespace = helpers.GetUserNamespace("e2e-user")

		By("Creating resources to clean up")
		fixturePath := helpers.GetFixturePath(helpers.FixtureSimpleNginx)

		// Create user blueprint (goes to user namespace)
		output, err := cli.BlueprintCreate(fixturePath, helpers.TestRepository)
		Expect(err).NotTo(HaveOccurred(), "User blueprint creation should succeed")
		userBlueprintID = helpers.ExtractBlueprintID(output)
		parts := strings.Split(userBlueprintID, "/")
		if len(parts) == 2 {
			userBlueprintName = parts[1]
		} else {
			userBlueprintName = userBlueprintID
		}

		// Create global blueprint (goes to global namespace)
		output, err = cli.BlueprintCreateGlobal(fixturePath, helpers.TestRepository, "main")
		Expect(err).NotTo(HaveOccurred(), "Global blueprint creation should succeed")
		globalBlueprintID = helpers.ExtractBlueprintID(output)
		parts = strings.Split(globalBlueprintID, "/")
		if len(parts) == 2 {
			globalBlueprintName = parts[1]
		} else {
			globalBlueprintName = globalBlueprintID
		}

		// Create stack from user blueprint
		output, err = cli.StackCreate(userBlueprintID)
		Expect(err).NotTo(HaveOccurred(), "Stack creation should succeed")
		stackID = helpers.ExtractStackID(output)
		parts = strings.Split(stackID, "/")
		if len(parts) == 2 {
			stackName = parts[1]
		} else {
			stackName = stackID
		}

		// Wait for stack to be ready
		Eventually(func() bool {
			return k8s.StackExists(ctx, userNamespace, stackName)
		}, 60*time.Second, 5*time.Second).Should(BeTrue())

		GinkgoWriter.Printf("Created user blueprint: %s (name: %s)\n", userBlueprintID, userBlueprintName)
		GinkgoWriter.Printf("Created global blueprint: %s (name: %s)\n", globalBlueprintID, globalBlueprintName)
		GinkgoWriter.Printf("Created stack: %s (name: %s)\n", stackID, stackName)
	})

	Describe("Stack Deletion (User Role)", func() {
		It("should delete the stack", func() {
			By("Deleting stack via CLI")
			output, err := cli.StackDelete(stackName)
			Expect(err).NotTo(HaveOccurred(), "Stack deletion should succeed: %s", output)
		})

		It("should remove Stack CRD", func() {
			By("Waiting for Stack CRD to be deleted")
			Eventually(func() bool {
				return !k8s.StackExists(ctx, userNamespace, stackName)
			}, 60*time.Second, 5*time.Second).Should(BeTrue(),
				"Stack CRD should be deleted")
		})

		It("should remove owned Deployment", func() {
			By("Waiting for Deployment to be deleted")
			deploymentName := stackName + "-web"
			Eventually(func() bool {
				return !k8s.DeploymentExists(ctx, userNamespace, deploymentName)
			}, 60*time.Second, 5*time.Second).Should(BeTrue(),
				"Deployment should be deleted with stack")
		})

		It("should remove owned Service", func() {
			By("Waiting for Service to be deleted")
			serviceName := stackName + "-web"
			Eventually(func() bool {
				return !k8s.ServiceExists(ctx, userNamespace, serviceName)
			}, 60*time.Second, 5*time.Second).Should(BeTrue(),
				"Service should be deleted with stack")
		})
	})

	Describe("User Blueprint Deletion", func() {
		It("should allow user to delete their own blueprint", func() {
			By("Deleting user blueprint via CLI as user")
			output, err := cli.RunAsUser("blueprint", "delete", userBlueprintID)
			Expect(err).NotTo(HaveOccurred(), "User should be able to delete their own blueprint: %s", output)
		})

		It("should remove user Blueprint CRD", func() {
			By("Waiting for Blueprint CRD to be deleted")
			Eventually(func() bool {
				return !k8s.BlueprintExists(ctx, userNamespace, userBlueprintName)
			}, 30*time.Second, 2*time.Second).Should(BeTrue(),
				"User Blueprint CRD should be deleted")
		})
	})

	Describe("Global Blueprint Deletion (Admin Role)", func() {
		It("should not allow user to delete global blueprint", func() {
			By("Attempting to delete global blueprint as user")
			_, err := cli.RunAsUser("blueprint", "delete", globalBlueprintID)
			Expect(err).To(HaveOccurred(),
				"User should NOT be able to delete global blueprints")
		})

		It("should delete the global blueprint as admin", func() {
			By("Deleting global blueprint via CLI as admin")
			output, err := cli.BlueprintDelete(globalBlueprintID)
			Expect(err).NotTo(HaveOccurred(), "Admin blueprint deletion should succeed: %s", output)
		})

		It("should remove global Blueprint CRD", func() {
			By("Waiting for Blueprint CRD to be deleted")
			Eventually(func() bool {
				return !k8s.BlueprintExists(ctx, helpers.GlobalNamespace, globalBlueprintName)
			}, 30*time.Second, 2*time.Second).Should(BeTrue(),
				"Global Blueprint CRD should be deleted")
		})
	})

	Describe("Final Verification", func() {
		It("should have clean state", func() {
			By("Verifying stack is gone")
			Expect(k8s.StackExists(ctx, userNamespace, stackName)).To(BeFalse())

			By("Verifying user blueprint is gone")
			Expect(k8s.BlueprintExists(ctx, userNamespace, userBlueprintName)).To(BeFalse())

			By("Verifying global blueprint is gone")
			Expect(k8s.BlueprintExists(ctx, helpers.GlobalNamespace, globalBlueprintName)).To(BeFalse())
		})
	})
})
