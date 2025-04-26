/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e

import (
	"fmt"
	"github.com/unfamousthomas/thesis-operator/test/utils"
	"os/exec"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const projectimage = "example.com/loputoo:v0.0.1"
const example_server_image = "nginx:latest" //Just a random image to use as a "fake server"
const sidecar_image = "ghcr.io/unfamousthomas/sidecar:latest"

const serverName = "test-server"
const systemns = "loputoo-system"

var controllerPodName string

var _ = BeforeSuite(func() {
	By("Setting up timeout")
	By("resetting the Kind cluster")
	cmd := exec.Command("kind", "delete", "cluster", "--name", "kind")
	_, _ = utils.Run(cmd)
	cmd = exec.Command("kind", "create", "cluster", "--name", "kind")
	_, err := utils.Run(cmd)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	By("installing the cert-manager")
	Expect(utils.InstallCertManager()).To(Succeed())

	By("creating manager systemns")
	cmd = exec.Command("kubectl", "create", "ns", systemns)
	_, _ = utils.Run(cmd)

	By("building the manager(Operator) image")
	cmd = exec.Command("make", "docker-build", fmt.Sprintf("IMG=%s", projectimage), fmt.Sprintf("SERVER_IMG=%s", example_server_image), fmt.Sprintf("SIDECAR_IMG=%s", sidecar_image))
	_, err = utils.Run(cmd)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	By("loading the the manager(Operator) image on Kind")
	err = utils.LoadImageToKindClusterWithName(projectimage)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	By("loading sidecar image on Kind")
	err = utils.LoadImageToKindClusterWithName(sidecar_image)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	By("installing CRDs")
	cmd = exec.Command("make", "install")
	_, err = utils.Run(cmd)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	By("deploying the controller-manager")
	cmd = exec.Command("make", "deploy", fmt.Sprintf("IMG=%s", projectimage))
	_, err = utils.Run(cmd)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	By("validating that the controller-manager pod is running as expected")
	verifyControllerUp := func() error {
		// Get pod name

		cmd = exec.Command("kubectl", "get",
			"pods", "-l", "control-plane=controller-manager",
			"-o", "go-template={{ range .items }}"+
				"{{ if not .metadata.deletionTimestamp }}"+
				"{{ .metadata.name }}"+
				"{{ \"\\n\" }}{{ end }}{{ end }}",
			"-n", systemns,
		)

		podOutput, err := utils.Run(cmd)
		ExpectWithOffset(2, err).NotTo(HaveOccurred())
		podNames := utils.GetNonEmptyLines(string(podOutput))
		if len(podNames) != 1 {
			return fmt.Errorf("expect 1 controller pods running, but got %d", len(podNames))
		}
		controllerPodName = podNames[0]
		ExpectWithOffset(2, controllerPodName).Should(ContainSubstring("controller-manager"))

		// Validate pod status
		cmd = exec.Command("kubectl", "get",
			"pods", controllerPodName, "-o", "jsonpath={.status.phase}",
			"-n", systemns,
		)
		status, err := utils.Run(cmd)
		ExpectWithOffset(2, err).NotTo(HaveOccurred())
		if string(status) != "Running" {
			return fmt.Errorf("controller pod in %s status", status)
		}

		// Check if all containers are ready
		cmd = exec.Command("kubectl", "get",
			"pods", controllerPodName, "-o", "jsonpath={.status.containerStatuses[*].ready}",
			"-n", systemns,
		)
		readyStatus, err := utils.Run(cmd)
		ExpectWithOffset(2, err).NotTo(HaveOccurred())
		readyStatuses := strings.Split(string(readyStatus), " ")
		for _, status := range readyStatuses {
			if status != "true" {
				return fmt.Errorf("not all containers in controller pod are ready")
			}
		}

		// Check that webhook service is ready
		cmd = exec.Command("kubectl", "get", "svc", "loputoo-webhook-service",
			"-n", systemns, "-o", "jsonpath={.spec.clusterIP}")
		svcIP, err := utils.Run(cmd)
		if err != nil || string(svcIP) == "" {
			return fmt.Errorf("webhook service not ready")
		}

		// Check that webhook endpoints are ready
		cmd = exec.Command("kubectl", "get", "endpoints", "loputoo-webhook-service",
			"-n", systemns, "-o", "jsonpath={.subsets[0].addresses[0].ip}")
		endpointIP, err := utils.Run(cmd)
		if err != nil || string(endpointIP) == "" {
			return fmt.Errorf("webhook endpoints not ready")
		}

		// Verify webhook configurations are present
		cmd = exec.Command("kubectl", "get", "mutatingwebhookconfiguration",
			"loputoo-mutating-webhook-configuration")
		_, err = utils.Run(cmd)
		if err != nil {
			return fmt.Errorf("mutating webhook configuration not found")
		}

		cmd = exec.Command("kubectl", "get", "validatingwebhookconfiguration",
			"loputoo-validating-webhook-configuration")
		_, err = utils.Run(cmd)
		if err != nil {
			return fmt.Errorf("validating webhook configuration not found")
		}

		// Let's give the webhook a bit more time to be fully ready
		// This helps with race conditions that aren't caught by the other checks
		time.Sleep(5 * time.Second)

		return nil
	}
	EventuallyWithOffset(1, verifyControllerUp, time.Minute, time.Second).Should(Succeed())
})

var _ = AfterSuite(func() {
	By("undeploying the controller-manager")
	cmd := exec.Command("make", "undeploy", fmt.Sprintf("IMG=%s", projectimage))
	_, err := utils.Run(cmd)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	By("uninstalling the cert-manager bundle")
	utils.UninstallCertManager()

	By("removing manager systemns")
	cmd = exec.Command("kubectl", "delete", "ns", systemns)
	_, _ = utils.Run(cmd)
})

// Run e2e tests using the Ginkgo runner.
func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	_, _ = fmt.Fprintf(GinkgoWriter, "Starting loputoo suite\n")
	RunSpecs(t, "e2e suite")
}
