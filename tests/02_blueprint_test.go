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
	var blueprintID string
	var blueprintName string
	ctx := context.Background()

	BeforeAll(func() {
		var err error
		k8s, err = helpers.NewK8sClient()
		Expect(err).NotTo(HaveOccurred(), "Failed to create K8s client")

		cli = helpers.NewCLIRunner()
	})

	Describe("Blueprint Creation (Admin Role)", func() {
		It("should create a blueprint from simple-nginx compose", func() {
			By("Creating blueprint with repository override")
			fixturePath := helpers.GetFixturePath(helpers.FixtureSimpleNginx)

			output, err := cli.BlueprintCreate(fixturePath, helpers.TestRepository)
			Expect(err).NotTo(HaveOccurred(), "Blueprint creation should succeed: %s", output)

			blueprintID = helpers.ExtractBlueprintID(output)
			Expect(blueprintID).NotTo(BeEmpty(), "Blueprint ID should be returned")

			By("Extracting blueprint name from scoped ID")
			// Blueprint ID format: namespace/name
			parts := strings.Split(blueprintID, "/")
			Expect(len(parts)).To(Equal(2), "Blueprint ID should be in format namespace/name")
			blueprintName = parts[1]

			GinkgoWriter.Printf("Created blueprint: %s (name: %s)\n", blueprintID, blueprintName)
		})

		It("should create blueprint in global namespace", func() {
			By("Verifying blueprint exists in global namespace")
			Eventually(func() bool {
				return k8s.BlueprintExists(ctx, helpers.GlobalNamespace, blueprintName)
			}, 30*time.Second, 2*time.Second).Should(BeTrue(),
				"Blueprint should exist in global namespace")
		})

		It("should have correct annotations", func() {
			By("Checking repository annotation")
			Eventually(func() string {
				repo, _ := k8s.GetBlueprintAnnotation(ctx, helpers.GlobalNamespace, blueprintName, "lissto.dev/repository")
				return repo
			}, 10*time.Second, 2*time.Second).Should(ContainSubstring("lissto-dev/e2e"),
				"Blueprint should have repository annotation")

			By("Checking services annotation")
			services, err := k8s.GetBlueprintAnnotation(ctx, helpers.GlobalNamespace, blueprintName, "lissto.dev/services")
			Expect(err).NotTo(HaveOccurred())
			Expect(services).NotTo(BeEmpty(), "Blueprint should have services annotation")
		})
	})

	Describe("Blueprint Listing", func() {
		It("should list blueprints including the created one", func() {
			output, err := cli.BlueprintList()
			Expect(err).NotTo(HaveOccurred(), "Blueprint list should succeed")
			Expect(output).To(ContainSubstring(blueprintName),
				"Blueprint list should contain created blueprint")
		})
	})

	Describe("Blueprint Get", func() {
		It("should get the created blueprint", func() {
			output, err := cli.BlueprintGet(blueprintID)
			Expect(err).NotTo(HaveOccurred(), "Blueprint get should succeed")
			Expect(output).NotTo(BeEmpty(), "Blueprint get should return data")
		})
	})

	Describe("Blueprint Idempotency", func() {
		It("should return same ID for duplicate content", func() {
			By("Creating blueprint with same content")
			fixturePath := helpers.GetFixturePath(helpers.FixtureSimpleNginx)

			output, err := cli.BlueprintCreate(fixturePath, helpers.TestRepository)
			Expect(err).NotTo(HaveOccurred(), "Duplicate blueprint creation should succeed")

			duplicateID := helpers.ExtractBlueprintID(output)
			Expect(duplicateID).To(Equal(blueprintID),
				"Duplicate content should return same blueprint ID (deduplication)")
		})
	})

	// Store blueprint ID for use in stack tests
	AfterAll(func() {
		// Store in environment for other tests
		GinkgoWriter.Printf("Blueprint ID for stack tests: %s\n", blueprintID)
		// Note: In real implementation, you'd store this in a shared state
	})
})

// Export for other tests
var SharedBlueprintID string
