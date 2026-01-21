package tests

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Lissto E2E Test Suite")
}

var _ = BeforeSuite(func() {
	By("Setting up e2e test environment")
	// Any global setup can go here
})

var _ = AfterSuite(func() {
	By("Tearing down e2e test environment")
	// Any global cleanup can go here
})
