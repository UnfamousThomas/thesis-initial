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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unfamousthomas/thesis-operator/test/utils"
)

var _ = Describe("Server Controller", Ordered, func() {
	Context("Server", func() {
		const namespace = "server-controller-test"
		BeforeAll(func() {
			By("Creating a namespace")
			_, _ = utils.Run(exec.Command("kubectl", "create", "namespace", namespace))
		})

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
			By("Creating the server manifest file")
			serverFile, err := os.CreateTemp("", "server-*.yaml")
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			defer os.Remove(serverFile.Name())
			manifest := utils.CreateServerManifest(serverName, namespace, example_server_image)
			_, err = serverFile.WriteString(manifest)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			err = serverFile.Close()
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			deleteServer := func() error {

				By("Get IP")
				ip, err := utils.GetPodIP(serverName+"-pod", namespace)
				if err != nil {
					if utils.IsNotFound(err.Error()) {
						log.Printf("Resource not found, returning nil")
						return nil
					}
					return err
				}

				By("Server exists, deleting it now")
				cmd := exec.Command("kubectl", "delete", "-f", serverFile.Name(), "--wait=false")
				_, _ = utils.Run(cmd)
				err = utils.SendAllowDeleteRequest(ip, namespace)
				if err != nil {
					log.Printf("WARNING! Failed to send server %s the delete request. This could be destructive: %s\n", serverName, err)
				}

				By("Check if the server still exists")
				// Verify server is actually gone
				verifyCmd := exec.Command("kubectl", "get", "server", serverName, "-n", namespace)
				output, err := utils.Run(verifyCmd)

				if utils.IsNotFound(string(output)) {
					log.Printf("Resource not found, returning nil")
					return nil
				}

				if err != nil {
					if utils.IsNotFound(err.Error()) {
						return nil
					}
					return err
				}

				return fmt.Errorf("server still exists after deletion attempt")
			}

			Eventually(deleteServer, 2*time.Minute, 5*time.Second).Should(Succeed())
		})

		It("Blocks deletion when not allowed", func() {
			// Try to delete the server
			By("try to delete server")
			deleteCmd := exec.Command("kubectl", "delete", "server", serverName, "-n", namespace, "--wait=false")
			_, err := utils.Run(deleteCmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("Wait 5 seconds")
			time.Sleep(5 * time.Second)

			By("Verify server was not deleted")
			// Verify server is still there (in terminating state)
			cmd := exec.Command("kubectl", "get", "server", serverName, "-n", namespace)
			_, err = utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
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

		It("Creates a pod when server is created", func() {
			podName := serverName + "-pod"
			podCmd := exec.Command("kubectl", "get", "pod", podName, "-n", namespace)
			_, err := utils.Run(podCmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			labelCmd := exec.Command("kubectl", "get", "pod", podName, "-n", namespace,
				"-o", "jsonpath={.metadata.labels.server}")
			output, err := utils.Run(labelCmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			ExpectWithOffset(1, string(output)).
				Should(Equal(serverName))
		})

		It("Don't delete with finalizers", func() {
			By("Trigger deletion")
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
				"-o", `jsonpath="{.metadata.deletionTimestamp}"`)
			output, err := utils.Run(deletionTimestampCmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			ExpectWithOffset(1, string(output)).NotTo(BeEmpty()) //Server deletion should be triggered

			By("Verify if server pod is in deleting state")
			podDeleteCmd := exec.Command("kubectl", "get", "pod", serverName+"-pod", "-n", namespace)
			_, err = utils.Run(podDeleteCmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			podDeletionTimestampCmd := exec.Command("kubectl", "get", "pod", serverName+"-pod", "-n", namespace,
				"-o", "jsonpath={.metadata.deletionTimestamp}")
			output, err = utils.Run(podDeletionTimestampCmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			ExpectWithOffset(1, string(output)).To(BeEmpty()) //Pod deletion should not be triggered
		})

		It("Handles deletion properly when allowed", func() {
			By("Send delete allow request")
			ip, err := utils.GetPodIP(serverName+"-pod", namespace)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			err = utils.SendAllowDeleteRequest(ip, namespace)
			if err != nil {
				fmt.Printf("Server %s failed to send delete request. This could have disastrous consequenses. Error: %s\n", serverName, err)
			}

			By("Trigger deletion")
			deleteCmd := exec.Command("kubectl", "delete", "server", serverName, "-n", namespace)
			_, err = utils.Run(deleteCmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("Checking if server was deleted")
			Eventually(func() error {
				cmd := exec.Command("kubectl", "get", "server", serverName, "-n", namespace)
				output, err := utils.Run(cmd)
				if utils.IsNotFound(err.Error()) || utils.IsNotFound(string(output)) {
					return nil
				}
				return fmt.Errorf("server still exists")
			}, "60s", "5s").Should(Succeed())

			Eventually(func() error {
				cmd := exec.Command("kubectl", "get", "pod", serverName+"-pod", "-n", namespace)
				_, err := utils.Run(cmd)
				if err != nil {
					if utils.IsNotFound(err.Error()) {
						return nil
					}
					return err
				}
				return fmt.Errorf("pod still exists")
			}, "60s", "5s").Should(Succeed())
		})
	})
})
