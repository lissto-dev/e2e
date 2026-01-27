package tests

import (
	"context"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/lissto-dev/e2e/tests/helpers"
)

var _ = Describe("Blueprint Lifecycle", Ordered, func() {
	var k8s *helpers.K8sClient
	var cli *helpers.CLIRunner
	var userBlueprintID string
	var userBlueprintName string
	var globalBlueprintID string
	var globalBlueprintName string
	var userNamespace string
	ctx := context.Background()

	BeforeAll(func() {
		var err error
		k8s, err = helpers.NewK8sClient()
		Expect(err).NotTo(HaveOccurred(), "Failed to create K8s client")

		cli = helpers.NewCLIRunner()
		userNamespace = helpers.GetUserNamespace("e2e-user")
	})

	Describe("Blueprint Creation (User Role)", func() {
		It("should create blueprint in user namespace", func() {
			By("Creating blueprint as user")
			fixturePath := helpers.GetFixturePath(helpers.FixtureSimpleNginx)

			output, err := cli.BlueprintCreate(fixturePath, helpers.TestRepository)
			Expect(err).NotTo(HaveOccurred(), "Blueprint creation should succeed: %s", output)

			userBlueprintID = helpers.ExtractBlueprintID(output)
			Expect(userBlueprintID).NotTo(BeEmpty(), "Blueprint ID should be returned")

			// Extract blueprint name from scoped ID (format: namespace/name)
			parts := strings.Split(userBlueprintID, "/")
			if len(parts) == 2 {
				userBlueprintName = parts[1]
			} else {
				userBlueprintName = userBlueprintID
			}

			GinkgoWriter.Printf("Created user blueprint: %s (name: %s)\n", userBlueprintID, userBlueprintName)
		})

		It("should exist in user namespace", func() {
			By("Verifying blueprint exists in user namespace")
			Eventually(func() bool {
				return k8s.BlueprintExists(ctx, userNamespace, userBlueprintName)
			}, 30*time.Second, 2*time.Second).Should(BeTrue(),
				"Blueprint should exist in user namespace: %s", userNamespace)
		})
	})

	Describe("Blueprint Creation (Deploy Role)", func() {
		It("should create global blueprint with --branch flag", func() {
			By("Creating blueprint as deploy with --branch main")
			fixturePath := helpers.GetFixturePath(helpers.FixtureSimpleNginx)

			output, err := cli.BlueprintCreateGlobal(fixturePath, helpers.TestRepository, "main")
			Expect(err).NotTo(HaveOccurred(), "Global blueprint creation should succeed: %s", output)

			globalBlueprintID = helpers.ExtractBlueprintID(output)
			Expect(globalBlueprintID).NotTo(BeEmpty(), "Blueprint ID should be returned")

			// Extract blueprint name from scoped ID
			parts := strings.Split(globalBlueprintID, "/")
			if len(parts) == 2 {
				globalBlueprintName = parts[1]
			} else {
				globalBlueprintName = globalBlueprintID
			}

			GinkgoWriter.Printf("Created global blueprint: %s (name: %s)\n", globalBlueprintID, globalBlueprintName)
		})

		It("should exist in global namespace", func() {
			By("Verifying blueprint exists in global namespace")
			Eventually(func() bool {
				return k8s.BlueprintExists(ctx, helpers.GlobalNamespace, globalBlueprintName)
			}, 30*time.Second, 2*time.Second).Should(BeTrue(),
				"Blueprint should exist in global namespace")
		})

		It("should have repository annotation", func() {
			By("Checking repository annotation")
			Eventually(func() string {
				repo, _ := k8s.GetBlueprintAnnotation(ctx, helpers.GlobalNamespace, globalBlueprintName, "lissto.dev/repository")
				return repo
			}, 10*time.Second, 2*time.Second).Should(ContainSubstring("lissto-dev/e2e"),
				"Blueprint should have repository annotation")
		})
	})

	Describe("Blueprint Creation Restrictions", func() {
		It("should not allow admin to create blueprints", func() {
			By("Attempting to create blueprint as admin")
			fixturePath := helpers.GetFixturePath(helpers.FixtureSimpleNginx)

			_, err := cli.RunAsAdmin("blueprint", "create", fixturePath, "--repository", helpers.TestRepository)
			Expect(err).To(HaveOccurred(),
				"Admin should NOT be able to create blueprints")
		})
	})

	Describe("Blueprint Listing", func() {
		It("should list blueprints including user and global", func() {
			output, err := cli.BlueprintList()
			Expect(err).NotTo(HaveOccurred(), "Blueprint list should succeed")

			// User should see their own blueprint
			Expect(output).To(ContainSubstring(userBlueprintName),
				"Blueprint list should contain user's blueprint")

			// User should also see global blueprints
			Expect(output).To(ContainSubstring(globalBlueprintName),
				"Blueprint list should contain global blueprint")
		})
	})

	Describe("Blueprint Get", func() {
		It("should get the user blueprint", func() {
			output, err := cli.BlueprintGet(userBlueprintID)
			Expect(err).NotTo(HaveOccurred(), "Blueprint get should succeed")
			Expect(output).NotTo(BeEmpty(), "Blueprint get should return data")
		})

		It("should get the global blueprint", func() {
			output, err := cli.BlueprintGet(globalBlueprintID)
			Expect(err).NotTo(HaveOccurred(), "Blueprint get should succeed")
			Expect(output).NotTo(BeEmpty(), "Blueprint get should return data")
		})
	})

	// Store blueprint IDs for use in stack tests
	AfterAll(func() {
		GinkgoWriter.Printf("User blueprint ID for stack tests: %s\n", userBlueprintID)
		GinkgoWriter.Printf("Global blueprint ID for stack tests: %s\n", globalBlueprintID)
	})
})

// Export for other tests
var SharedUserBlueprintID string
var SharedGlobalBlueprintID string
