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
	"os"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unfamousthomas/thesis-operator/test/utils"
)

const namespace = "loputoo-system"

var controllerPodName string

const projectimage = "example.com/loputoo:v0.0.1"
const example_server_image = "example.com/example-server:v0.0.1"
const sidecar_image = "ghcr.io/unfamousthomas/sidecar:latest"

var _ = Describe("controller", Ordered, func() {
	BeforeAll(func() {

		By("installing the cert-manager")
		Expect(utils.InstallCertManager()).To(Succeed())

		By("creating manager namespace")
		cmd := exec.Command("kubectl", "create", "ns", namespace)
		_, _ = utils.Run(cmd)

		By("building the manager(Operator) image")
		cmd = exec.Command("make", "docker-build", fmt.Sprintf("IMG=%s", projectimage), fmt.Sprintf("SERVER_IMG=%s", example_server_image), fmt.Sprintf("SIDECAR_IMG=%s", sidecar_image))
		_, err := utils.Run(cmd)
		ExpectWithOffset(1, err).NotTo(HaveOccurred())

		By("loading the the manager(Operator) image on Kind")
		err = utils.LoadImageToKindClusterWithName(projectimage)
		ExpectWithOffset(1, err).NotTo(HaveOccurred())

		By("loading sidecar image on Kind")
		err = utils.LoadImageToKindClusterWithName(sidecar_image)
		ExpectWithOffset(1, err).NotTo(HaveOccurred())

		By("loading example server image on Kind")
		err = utils.LoadImageToKindClusterWithName(example_server_image)
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
				"-n", namespace,
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
				"-n", namespace,
			)
			status, err := utils.Run(cmd)
			ExpectWithOffset(2, err).NotTo(HaveOccurred())
			if string(status) != "Running" {
				return fmt.Errorf("controller pod in %s status", status)
			}

			// Check if all containers are ready
			cmd = exec.Command("kubectl", "get",
				"pods", controllerPodName, "-o", "jsonpath={.status.containerStatuses[*].ready}",
				"-n", namespace,
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
				"-n", namespace, "-o", "jsonpath={.spec.clusterIP}")
			svcIP, err := utils.Run(cmd)
			if err != nil || string(svcIP) == "" {
				return fmt.Errorf("webhook service not ready")
			}

			// Check that webhook endpoints are ready
			cmd = exec.Command("kubectl", "get", "endpoints", "loputoo-webhook-service",
				"-n", namespace, "-o", "jsonpath={.subsets[0].addresses[0].ip}")
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

	//AfterAll(func() {
	//	By("undeploying the controller-manager")
	//	cmd := exec.Command("make", "undeploy", fmt.Sprintf("IMG=%s", projectimage))
	//	_, err := utils.Run(cmd)
	//	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	//
	//	By("uninstalling the cert-manager bundle")
	//	utils.UninstallCertManager()
	//
	//	By("removing manager namespace")
	//	cmd = exec.Command("kubectl", "delete", "ns", namespace)
	//	_, _ = utils.Run(cmd)
	//})

	Context("Server", func() {
		const serverName = "test-server"

		BeforeAll(func() {
			By("Creating the server manifest file")
			serverFile, err := os.CreateTemp("", "server-*.yaml")
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			defer os.Remove(serverFile.Name())
			manifest := utils.CreateServerManifest(serverName, namespace, example_server_image)
			_, err = serverFile.WriteString(manifest)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			err = serverFile.Close()
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("Applying server manifest")
			cmd := exec.Command("kubectl", "apply", "-f", serverFile.Name())
			_, err = utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			By("Waiting for server pod to be ready")
			verifyServerDeployment := func() error {
				cmd = exec.Command("kubectl", "get", "pods", serverName+"-pod", "-n", namespace, "-o", "jsonpath={.status.phase}")
				output, err := utils.Run(cmd)
				if err != nil {
					return err
				}
				if string(output) != "Running" {
					return fmt.Errorf("server pod in %s status", string(output))
				}
				return nil
			}
			EventuallyWithOffset(1, verifyServerDeployment, 2*time.Minute, 5*time.Second).Should(Succeed())
		})
		AfterAll(func() {
			By("Remove server")
			deleteServer := func() error {
				By("Creating the server manifest file")
				serverFile, err := os.CreateTemp("", "server-*.yaml")
				ExpectWithOffset(1, err).NotTo(HaveOccurred())
				defer os.Remove(serverFile.Name())
				manifest := utils.CreateServerManifest(serverName, namespace, example_server_image)
				_, err = serverFile.WriteString(manifest)
				ExpectWithOffset(1, err).NotTo(HaveOccurred())
				err = serverFile.Close()
				ExpectWithOffset(1, err).NotTo(HaveOccurred())

				By("Check if server exists")
				cmd := exec.Command("kubectl", "get", "pods", serverName+"-pod", "-n", namespace)
				output, err := utils.Run(cmd)
				if err != nil || strings.TrimSpace(string(output)) == "" {
					// Server doesn't exist, we're done
					return nil
				}

				By("Send request")
				//In this controller, we first need to get the internal ip
				ip, err := utils.GetPodIP(serverName+"-pod", namespace)
				ExpectWithOffset(1, err).NotTo(HaveOccurred()) //Something has gone wrong, as we did not get an error with get but this fails
				err = utils.SendAllowDeleteRequest(ip, namespace)
				ExpectWithOffset(1, err).NotTo(HaveOccurred())

				By("Server exists, deleting it now")
				//So if we get here the server still exists and we need to figure out how to delete it.
				cmd = exec.Command("kubectl", "delete", "-f", serverFile.Name())
				_, err = utils.Run(cmd)
				ExpectWithOffset(1, err).NotTo(HaveOccurred()) //Deletion was triggered, however finalizers block it

				// Verify server is actually gone
				verifyCmd := exec.Command("kubectl", "get", "server", serverName, "--ignore-not-found")
				output, err = utils.Run(verifyCmd)
				if err != nil || strings.TrimSpace(string(output)) == "" {
					// Server is gone
					return nil
				}

				return fmt.Errorf("server still exists after deletion attempt")
			}
			EventuallyWithOffset(1, deleteServer, 2*time.Minute, 5*time.Second).Should(Succeed())
		})
		It("Finalizers match", func() {
			//TODO
			serverFinalizers := exec.Command("kubectl", "get", "server", serverName, "-n", namespace, "-o", "jsonpath={.metadata.finalizers}")
			output, err := utils.Run(serverFinalizers)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			ExpectWithOffset(1, string(output)).Should(Equal("[\"servers.unfamousthomas.me/finalizer\"]"))

			podFinalizers := exec.Command("kubectl", "get", "pod", serverName+"-pod", "-n", namespace, "-o", "jsonpath={.metadata.finalizers}")
			output, err = utils.Run(podFinalizers)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			ExpectWithOffset(1, string(output)).Should(Equal("[\"servers.unfamousthomas.me/finalizer\"]"))
		})
	})
})
