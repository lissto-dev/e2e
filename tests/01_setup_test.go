package tests

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/lissto-dev/e2e/tests/helpers"
)

var _ = Describe("Setup Validation", Ordered, func() {
	var k8s *helpers.K8sClient
	var cli *helpers.CLIRunner
	ctx := context.Background()

	BeforeAll(func() {
		var err error
		k8s, err = helpers.NewK8sClient()
		Expect(err).NotTo(HaveOccurred(), "Failed to create K8s client")

		cli = helpers.NewCLIRunner()
	})

	Describe("Kubernetes Cluster", func() {
		It("should have lissto-system namespace", func() {
			Expect(k8s.NamespaceExists(ctx, helpers.LisstoSystemNamespace)).To(BeTrue(),
				"lissto-system namespace should exist")
		})

		It("should have CRDs registered", func() {
			// Check that we can list CRDs (they exist)
			_, err := k8s.ListBlueprints(ctx, helpers.LisstoSystemNamespace)
			Expect(err).NotTo(HaveOccurred(), "Blueprint CRD should be registered")

			_, err = k8s.ListStacks(ctx, helpers.LisstoSystemNamespace)
			Expect(err).NotTo(HaveOccurred(), "Stack CRD should be registered")
		})
	})

	Describe("Lissto Components", func() {
		It("should have API deployment ready", func() {
			Eventually(func() bool {
				return k8s.DeploymentReady(ctx, helpers.LisstoSystemNamespace, "lissto-api")
			}, 60*time.Second, 5*time.Second).Should(BeTrue(),
				"API deployment should be ready")
		})

		It("should have Controller deployment ready", func() {
			Eventually(func() bool {
				return k8s.DeploymentReady(ctx, helpers.LisstoSystemNamespace, "lissto-controller")
			}, 60*time.Second, 5*time.Second).Should(BeTrue(),
				"Controller deployment should be ready")
		})
	})

	Describe("CLI Configuration", func() {
		It("should have admin context configured", func() {
			output, err := cli.Run("context", "use", helpers.RoleAdmin)
			Expect(err).NotTo(HaveOccurred(), "Should switch to admin context: %s", output)
		})

		It("should have deploy context configured", func() {
			output, err := cli.Run("context", "use", helpers.RoleDeploy)
			Expect(err).NotTo(HaveOccurred(), "Should switch to deploy context: %s", output)
		})

		It("should have user context configured", func() {
			output, err := cli.Run("context", "use", helpers.RoleUser)
			Expect(err).NotTo(HaveOccurred(), "Should switch to user context: %s", output)
		})

		It("should be able to list blueprints as admin", func() {
			_, err := cli.RunAsAdmin("blueprint", "list")
			Expect(err).NotTo(HaveOccurred(), "Admin should be able to list blueprints")
		})
	})

	Describe("Test Fixtures", func() {
		It("should have simple-nginx fixture", func() {
			Expect(helpers.FixtureExists(helpers.FixtureSimpleNginx)).To(BeTrue(),
				"simple-nginx.yaml fixture should exist")
		})

		It("should have multi-service fixture", func() {
			Expect(helpers.FixtureExists(helpers.FixtureMultiService)).To(BeTrue(),
				"multi-service.yaml fixture should exist")
		})
	})
})
