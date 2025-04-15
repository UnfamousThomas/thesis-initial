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
	"log"
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

var _ = Describe("controller", Ordered, func() {
	Context("Server", func() {
		const serverName = "test-server"

		BeforeEach(func() {
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
		AfterEach(func() {
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
				cmd := exec.Command("kubectl", "get", "servers", serverName, "-n", namespace)
				output, err := utils.Run(cmd)
				log.Printf("Error checking: %s", err)
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

				By("Check if server still exists")
				cmd = exec.Command("kubectl", "get", "servers", "-n", namespace)
				output, err = utils.Run(cmd)
				if err != nil || strings.TrimSpace(string(output)) == "" {
					// Server doesn't exist, we're done
					return nil
				}

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
				} else {
					log.Printf("Output: %s\n", output)
				}

				return fmt.Errorf("server still exists after deletion attempt")
			}
			EventuallyWithOffset(1, deleteServer, 2*time.Minute, 5*time.Second).Should(Succeed())
		})

		It("Finalizers match", func() {
			serverFinalizers := exec.Command("kubectl", "get", "server", serverName, "-n", namespace, "-o", "jsonpath={.metadata.finalizers}")
			output, err := utils.Run(serverFinalizers)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			ExpectWithOffset(1, string(output)).Should(Equal("[\"servers.unfamousthomas.me/finalizer\"]"))

			podFinalizers := exec.Command("kubectl", "get", "pod", serverName+"-pod", "-n", namespace, "-o", "jsonpath={.metadata.finalizers}")
			output, err = utils.Run(podFinalizers)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			ExpectWithOffset(1, string(output)).Should(Equal("[\"servers.unfamousthomas.me/finalizer\"]"))
		})

		It("Don't delete with finalizers", func() {
			// Delete the server resource without waiting for full deletion
			deleteCmd := exec.Command("kubectl", "delete", "server", serverName, "-n", namespace, "--wait=false")
			_, err := utils.Run(deleteCmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			// Give some time for the deletion to be processed
			time.Sleep(3 * time.Second)

			By("Verify if server is now in deleting state")
			checkCmd := exec.Command("kubectl", "get", "server", serverName, "-n", namespace)
			_, err = utils.Run(checkCmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			// Verify the resource has a deletion timestamp (is in terminating state)
			deletionTimestampCmd := exec.Command("kubectl", "get", "server", serverName, "-n", namespace,
				"-o", "jsonpath={.metadata.deletionTimestamp}")
			output, err := utils.Run(deletionTimestampCmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			ExpectWithOffset(1, string(output)).NotTo(BeEmpty())

			By("Verify if server pod is in deleting state")
			podDeleteCmd := exec.Command("kubectl", "get", "pod", serverName+"-pod", "-n", namespace)
			_, err = utils.Run(podDeleteCmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			podDeletionTimestampCmd := exec.Command("kubectl", "get", "pod", serverName+"-pod", "-n", namespace,
				"-o", "jsonpath={.metadata.deletionTimestamp}")
			output, err = utils.Run(podDeletionTimestampCmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			ExpectWithOffset(1, string(output)).NotTo(BeEmpty())

			By("Remove finalizers")
			serverPatchCmd := exec.Command("kubectl", "patch", "server", serverName, "-n", namespace,
				"--type", "json", "-p", "[{\"op\": \"remove\", \"path\": \"/metadata/finalizers\"}]")
			_, err = utils.Run(serverPatchCmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			podPatchCmd := exec.Command("kubectl", "patch", "pod", serverName+"-pod", "-n", namespace,
				"--type", "json", "-p", "[{\"op\": \"remove\", \"path\": \"/metadata/finalizers\"}]")
			_, err = utils.Run(podPatchCmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("Check if server is deleted")
			// Wait for actual deletion to complete after removing finalizers
			Eventually(func() error {
				cmd := exec.Command("kubectl", "get", "server", serverName, "-n", namespace)
				_, err := utils.Run(cmd)
				if strings.Contains(err.Error(), "not found") {
					return nil
				}
				return err
			}, "30s", "2s").Should(Not(HaveOccurred()))
		})

		It("Creates a pod when server is created", func() {
			// Check that pod exists
			podName := serverName + "-pod"
			podCmd := exec.Command("kubectl", "get", "pod", podName, "-n", namespace)
			_, err := utils.Run(podCmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			// Verify pod properties
			labelCmd := exec.Command("kubectl", "get", "pod", podName, "-n", namespace,
				"-o", "jsonpath={.metadata.labels.app}")
			output, err := utils.Run(labelCmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			ExpectWithOffset(1, string(output)).Should(Equal(serverName))
		})

		It("Sets proper status conditions", func() {
			condCmd := exec.Command("kubectl", "get", "server", serverName, "-n", namespace,
				"-o", "jsonpath={.status.conditions[?(@.type==\"PodCreated\")].status}")
			output, err := utils.Run(condCmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			ExpectWithOffset(1, string(output)).Should(Equal("True"))
		})

		It("Handles deletion properly when allowed", func() {
			// First get the Pod IP
			ip, err := utils.GetPodIP(serverName+"-pod", namespace)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			// Mock the service to allow deletion
			err = utils.SendAllowDeleteRequest(ip, namespace)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			// Delete the server
			deleteCmd := exec.Command("kubectl", "delete", "server", serverName, "-n", namespace)
			_, err = utils.Run(deleteCmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			// Verify both resources are gone
			Eventually(func() error {
				cmd := exec.Command("kubectl", "get", "server", serverName, "-n", namespace)
				_, err := utils.Run(cmd)
				if err == nil {
					return fmt.Errorf("server still exists")
				}
				return nil
			}, "60s", "5s").Should(Succeed())

			Eventually(func() error {
				cmd := exec.Command("kubectl", "get", "pod", serverName+"-pod", "-n", namespace)
				_, err := utils.Run(cmd)
				if err == nil {
					return fmt.Errorf("pod still exists")
				}
				return nil
			}, "60s", "5s").Should(Succeed())
		})

		It("Blocks deletion when not allowed", func() {
			// Try to delete the server
			deleteCmd := exec.Command("kubectl", "delete", "server", serverName, "-n", namespace, "--wait=false")
			_, err := utils.Run(deleteCmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			time.Sleep(5 * time.Second)

			// Verify server is still there (in terminating state)
			cmd := exec.Command("kubectl", "get", "server", serverName, "-n", namespace)
			_, err = utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
		})
	})
})
