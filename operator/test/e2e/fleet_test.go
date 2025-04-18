package e2e

import (
	"fmt"
	"github.com/unfamousthomas/thesis-operator/test/utils"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Fleet Controller", Ordered, func() {
	Context("Basic Fleet Operations", func() {
		const namespace = "test-fleet-ns"
		const fleetName = "test-fleet"
		const initialReplicas = 3

		verifyPodsForFleet := func() error {
			return checkIfPodsOnlineForFleet(fleetName, namespace)
		}

		BeforeEach(func() {
			By("Creating the fleet namespace")
			_, _ = utils.Run(exec.Command("kubectl", "create", "namespace", namespace))

			By("Creating the fleet manifest file")
			fleetFile, err := os.CreateTemp("", "fleet-*.yaml")
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			defer os.Remove(fleetFile.Name())

			manifest := utils.CreateFleetManifest(fleetName, namespace, example_server_image, initialReplicas, true, "oldest_first")
			_, err = fleetFile.WriteString(manifest)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			err = fleetFile.Close()
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("Applying fleet manifest")
			cmd := exec.Command("kubectl", "apply", "-f", fleetFile.Name())
			_, err = utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("Waiting for fleet to be ready")
			verifyFleetCreation := func() error {
				cmd := exec.Command("kubectl", "get", "fleet", fleetName, "-n", namespace)
				_, err := utils.Run(cmd)
				if err != nil {
					return fmt.Errorf("fleet not found: %w", err)
				}

				serverCmd := exec.Command("kubectl", "get", "servers", "-l", fmt.Sprintf("fleet=%s", fleetName),
					"-n", namespace, "--no-headers")
				_, err = utils.Run(serverCmd)
				if err != nil {
					if strings.Contains(err.Error(), "No resources found") {
						return fmt.Errorf("no servers found for fleet")
					}
					return fmt.Errorf("error checking servers: %w", err)
				}

				return nil
			}
			EventuallyWithOffset(1, verifyFleetCreation, time.Minute, 5*time.Second).Should(Succeed())

			By("Waiting for pods to be ready")
			EventuallyWithOffset(1, verifyPodsForFleet, time.Minute, 5*time.Second).Should(Succeed())

		})

		AfterEach(func() {
			allowFleetServersDelete(fleetName, namespace)

			By("Removing fleet")
			deleteFleet := func() error {
				cmd := exec.Command("kubectl", "delete", "fleet", fleetName, "--namespace", namespace)
				_, err := utils.Run(cmd)
				return err
			}
			EventuallyWithOffset(1, deleteFleet, 2*time.Minute, 5*time.Second).Should(Succeed())

			By("Verifying fleet was deleted")
			verifyFleetDeleted := func() error {
				cmd := exec.Command("kubectl", "get", "fleet", fleetName, "-n", namespace)
				_, err := utils.Run(cmd)
				if err != nil && strings.Contains(err.Error(), "not found") {
					return nil
				}
				return fmt.Errorf("fleet still exists or error: %v", err)
			}
			EventuallyWithOffset(1, verifyFleetDeleted, time.Minute, 5*time.Second).Should(Succeed())
		})

		It("Should create the specified number of servers", func() {
			By("Verifying fleet has created the right number of servers")
			verifyServerCount := func() error {
				cmd := exec.Command("kubectl", "get", "servers", "-l", "fleet="+fleetName, "-n", namespace, "--no-headers")
				output, err := utils.Run(cmd)
				if err != nil {
					return err
				}

				lines := strings.Split(strings.TrimSpace(string(output)), "\n")
				count := 0
				if strings.TrimSpace(string(output)) != "" {
					count = len(lines)
				}

				if count != initialReplicas {
					return fmt.Errorf("expected %d servers, found %d", initialReplicas, count)
				}
				return nil
			}
			EventuallyWithOffset(1, verifyServerCount, 2*time.Minute, 5*time.Second).Should(Succeed())
			By("Check if pods are online")
			EventuallyWithOffset(1, verifyPodsForFleet, time.Minute, 5*time.Second).Should(Succeed())

		})

		It("Should add finalizer to the fleet", func() {
			By("Checking fleet has finalizer")
			getFleetFinalizers := func() (string, error) {
				cmd := exec.Command("kubectl", "get", "fleet", fleetName, "-n", namespace, "-o", "jsonpath={.metadata.finalizers}")
				output, err := utils.Run(cmd)
				//todo this needs to actually check finalizers
				return string(output), err
			}

			Eventually(getFleetFinalizers, time.Minute, 5*time.Second).Should(Equal("[\"fleets.unfamousthomas.me/finalizer\"]"))
		})

		It("Should update status to reflect current number of servers", func() {
			By("Checking fleet status reflects the right server count")
			getFleetReplicas := func() (string, error) {
				cmd := exec.Command("kubectl", "get", "fleet", fleetName, "-n", namespace, "-o", "jsonpath={.status.current_replicas}")
				output, err := utils.Run(cmd)
				//replicas check
				return string(output), err
			}

			Eventually(getFleetReplicas, time.Minute, 5*time.Second).Should(Equal("3"))
			EventuallyWithOffset(1, verifyPodsForFleet, time.Minute, 5*time.Second).Should(Succeed())
		})

		It("Should scale up when replicas are increased", func() {
			By("Updating fleet scale to 5 replicas")
			patchCmd := exec.Command("kubectl", "patch", "fleet", fleetName, "-n", namespace, "--type", "merge", "-p",
				`{"spec":{"scaling":{"replicas":5}}}`)
			_, err := utils.Run(patchCmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("Verifying fleet scales up to 5 servers")
			verifyScaledUpCount := func() error {
				cmd := exec.Command("kubectl", "get", "servers", "-l", "fleet="+fleetName, "-n", namespace, "--no-headers")
				output, err := utils.Run(cmd)
				if err != nil {
					return err
				}

				lines := strings.Split(strings.TrimSpace(string(output)), "\n")
				count := 0
				if strings.TrimSpace(string(output)) != "" {
					count = len(lines)
				}

				if count != 5 {
					return fmt.Errorf("expected 5 servers, found %d", count)
				}
				return nil
			}
			EventuallyWithOffset(1, verifyScaledUpCount, 2*time.Minute, 5*time.Second).Should(Succeed())
			EventuallyWithOffset(1, verifyPodsForFleet, time.Minute, 5*time.Second).Should(Succeed())

			By("Checking status is updated correctly")
			getFleetReplicas := func() (string, error) {
				cmd := exec.Command("kubectl", "get", "fleet", fleetName, "-n", namespace, "-o", "jsonpath={.status.current_replicas}")
				output, err := utils.Run(cmd)
				//replicas check
				return string(output), err
			}

			Eventually(getFleetReplicas, time.Minute, 5*time.Second).Should(Equal("5"))
			EventuallyWithOffset(1, verifyPodsForFleet, time.Minute, 5*time.Second).Should(Succeed())

		})

		It("Should scale down when replicas are decreased", func() {
			By("Updating fleet scale to 2 replicas")
			patchCmd := exec.Command("kubectl", "patch", "fleet", fleetName, "-n", namespace, "--type", "merge", "-p",
				`{"spec":{"scaling":{"replicas":2}}}`)
			_, err := utils.Run(patchCmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			EventuallyWithOffset(1, verifyPodsForFleet, time.Minute, 5*time.Second).Should(Succeed())

			By("Allowing deletion for all servers")
			allowFleetServersDelete(fleetName, namespace)

			By("Verifying fleet scales down to 2 servers")
			verifyScaledDownCount := func() error {
				cmd := exec.Command("kubectl", "get", "servers", "-l", "fleet="+fleetName, "-n", namespace, "--no-headers")
				output, err := utils.Run(cmd)
				if err != nil {
					return err
				}

				lines := strings.Split(strings.TrimSpace(string(output)), "\n")
				count := 0
				if strings.TrimSpace(string(output)) != "" {
					count = len(lines)
				}

				if count != 2 {
					return fmt.Errorf("expected 2 servers, found %d", count)
				}
				return nil
			}
			EventuallyWithOffset(1, verifyScaledDownCount, 2*time.Minute, 5*time.Second).Should(Succeed())

			By("Checking status is updated correctly")
			getFleetReplicas := func() (string, error) {
				cmd := exec.Command("kubectl", "get", "fleet", fleetName, "-n", namespace, "-o", "jsonpath={.status.current_replicas}")
				output, err := utils.Run(cmd)
				return string(output), err
			}

			Eventually(getFleetReplicas, time.Minute, 5*time.Second).Should(Equal("2"))
		})

	})
})

func checkIfPodsOnlineForFleet(fleetName string, namespace string) error {
	By("Getting servers that belong to the fleet")
	cmd := exec.Command("kubectl", "get", "servers", "-l", "fleet="+fleetName, "-n", namespace, "-o", "name")
	output, err := utils.Run(cmd)
	if err != nil {
		return fmt.Errorf("error getting servers: %w", err)
	}

	servers := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, server := range servers {
		serverName := strings.TrimPrefix(server, "server.network.unfamousthomas.me/")
		podName := serverName + "-pod"

		cmd := exec.Command("kubectl", "get", "pod", podName, "-n", namespace, "-o", "jsonpath={.status.phase}")
		output, err := utils.Run(cmd)
		if err != nil {
			return fmt.Errorf("error checking pod %s: %w", podName, err)
		}

		if string(output) != "Running" {
			return fmt.Errorf("pod %s is in %s state, expected Running", podName, string(output))
		}
	}

	return nil
}

func allowFleetServersDelete(fleetName string, namespace string) {
	processedServers := make(map[string]bool)
	overallTimeout := time.After(8 * time.Minute)

	// Keep trying until all servers are processed or gone
	for attempts := 0; attempts < 10; attempts++ {
		select {
		case <-overallTimeout:
			log.Println("Overall timeout reached. Some servers may not have been processed.")
			return
		default:
		}
		// Limit attempts to avoid infinite loops
		By("Getting current servers in the fleet")
		cmd := exec.Command("kubectl", "get", "servers", "-l", "fleet="+fleetName, "-n", namespace, "-o", "name")
		output, err := utils.Run(cmd)
		if err != nil {
			// Check if error is because no servers exist
			if strings.Contains(err.Error(), "No resources found") {
				log.Println("No servers found, all servers have been processed")
				return
			}
			log.Printf("Error getting servers: %s\n", err)
			time.Sleep(5 * time.Second)
			continue
		}

		servers := strings.Split(strings.TrimSpace(string(output)), "\n")
		if len(servers) == 0 {
			log.Println("No servers found, all servers have been processed")
			return
		}

		processedAny := false

		for _, server := range servers {
			serverName := strings.TrimPrefix(server, "server.network.unfamousthomas.me/")

			if processedServers[serverName] {
				continue
			}
			if serverName == "" {
				continue
			}

			podName := serverName + "-pod"
			exists, err := utils.PodExists(podName, namespace)
			if err != nil {
				log.Printf("Could not check if pod %s exists: %s\n", podName, err)
				continue
			}
			if !exists {
				log.Printf("Pod %s does not exist\n", podName)
				processedServers[serverName] = true
				continue
			}

			By(fmt.Sprintf("Getting pod IP for server %s", serverName))
			ip, err := utils.GetPodIP(serverName+"-pod", namespace)
			if err != nil {
				log.Printf("Error getting IP for pod %s: %s\n", podName, err)
				continue
			}
			fmt.Printf("IP for server %s is %s\n", serverName, ip)

			By(fmt.Sprintf("Sending allow delete request to server %s", serverName))
			err = utils.SendAllowDeleteRequest(ip, namespace)
			if err != nil {
				log.Printf("Error sending delete request to %s: %s\n", serverName, err)
				continue
			}

			processedServers[serverName] = true
			processedAny = true
		}

		if !processedAny {
			log.Println("No new servers processed in this iteration")
			time.Sleep(5 * time.Second)
		}

		allProcessed := true
		for _, server := range servers {
			serverName := strings.TrimPrefix(server, "server.network.unfamousthomas.me/")
			if !processedServers[serverName] {
				allProcessed = false
				break
			}
		}

		if allProcessed {
			log.Println("All servers have been processed")
			return
		}

		// Wait a bit between iterations to allow for server deletion
		time.Sleep(2 * time.Second)
	}

	log.Println("Maximum attempts reached, some servers may not have been processed")
}
