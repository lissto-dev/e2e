package tests

import (
	"context"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/lissto-dev/e2e/tests/helpers"
)

var _ = Describe("Image Update", Ordered, func() {
	var k8s *helpers.K8sClient
	var cli *helpers.CLIRunner
	var blueprintID string
	var stackID string
	var stackName string
	var userNamespace string
	var originalImage string
	ctx := context.Background()

	BeforeAll(func() {
		var err error
		k8s, err = helpers.NewK8sClient()
		Expect(err).NotTo(HaveOccurred(), "Failed to create K8s client")

		cli = helpers.NewCLIRunner()
		userNamespace = helpers.GetUserNamespace("e2e-user")

		By("Ensuring environment exists for stack creation")
		err = cli.EnsureEnv(helpers.TestEnvName)
		Expect(err).NotTo(HaveOccurred(), "Environment creation should succeed")

		By("Creating blueprint for image update test")
		fixturePath := helpers.GetFixturePath(helpers.FixtureSimpleNginx)
		output, err := cli.BlueprintCreate(fixturePath, helpers.TestRepository)
		Expect(err).NotTo(HaveOccurred())
		blueprintID = helpers.ExtractBlueprintID(output)

		By("Creating stack")
		output, err = cli.StackCreate(blueprintID)
		Expect(err).NotTo(HaveOccurred())
		stackID = helpers.ExtractStackID(output)

		parts := strings.Split(stackID, "/")
		if len(parts) == 2 {
			stackName = parts[1]
		} else {
			stackName = stackID
		}

		By("Waiting for stack to be ready")
		deploymentName := stackName + "-web"
		Eventually(func() bool {
			return k8s.DeploymentReady(ctx, userNamespace, deploymentName)
		}, 120*time.Second, 5*time.Second).Should(BeTrue())

		By("Recording original image")
		originalImage, err = k8s.GetDeploymentImage(ctx, userNamespace, deploymentName)
		Expect(err).NotTo(HaveOccurred())
		GinkgoWriter.Printf("Original image: %s\n", originalImage)
	})

	Describe("Image Update Flow", func() {
		It("should update stack with new image tag", func() {
			Skip("Image update CLI command needs to be implemented")

			// This test would:
			// 1. Call CLI to update the stack with a new image
			// 2. Verify the Stack CRD spec.images is updated
			// 3. Verify the Deployment gets the new image

			// Example:
			// output, err := cli.RunAsUser("stack", "update", stackName,
			//     "--image", "web=nginx:1.25")
			// Expect(err).NotTo(HaveOccurred())
		})

		It("should trigger deployment rollout on image change", func() {
			Skip("Image update CLI command needs to be implemented")

			// After image update:
			// deploymentName := stackName + "-web"
			// Eventually(func() string {
			//     img, _ := k8s.GetDeploymentImage(ctx, userNamespace, deploymentName)
			//     return img
			// }, 60*time.Second, 5*time.Second).Should(Equal("nginx:1.25"))
		})
	})

	Describe("Stack Spec Update", func() {
		It("should reflect image changes in Stack CRD", func() {
			Skip("Image update CLI command needs to be implemented")

			// Verify Stack CRD spec.images contains updated image info
			// stack, err := k8s.GetStack(ctx, userNamespace, stackName)
			// Expect(err).NotTo(HaveOccurred())
			// spec := stack.Object["spec"].(map[string]interface{})
			// images := spec["images"].(map[string]interface{})
			// webImage := images["web"].(map[string]interface{})
			// Expect(webImage["image"]).To(ContainSubstring("nginx:1.25"))
		})
	})

	AfterAll(func() {
		// Cleanup will be done in cleanup tests
		GinkgoWriter.Printf("Image update test complete. Stack: %s\n", stackName)
	})
})
