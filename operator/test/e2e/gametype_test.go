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

var _ = Describe("GameType Controller", Ordered, func() {
	Context("Basic GameType Operations", func() {
		const namespace = "test-gametype-ns"
		const gameTypeName = "test-gametype"
		const initialReplicas = 3

		BeforeEach(func() {
			By("Creating the gametype namespace")
			_, _ = utils.Run(exec.Command("kubectl", "create", "namespace", namespace))

			By("Creating the gametype manifest file")
			gameTypeFile, err := os.CreateTemp("", "gametype-*.yaml")
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			defer os.Remove(gameTypeFile.Name())

			manifest := utils.CreateGameTypeManifest(gameTypeName, namespace, example_server_image, initialReplicas, true, "oldest_first")
			_, err = gameTypeFile.WriteString(manifest)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			err = gameTypeFile.Close()
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("Applying gametype manifest")
			cmd := exec.Command("kubectl", "apply", "-f", gameTypeFile.Name())
			_, err = utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("Waiting for gametype to be ready")
			Eventually(func() error {
				// Verify GameType exists
				cmd := exec.Command("kubectl", "get", "gametype", gameTypeName, "-n", namespace)
				_, err := utils.Run(cmd)
				if err != nil {
					return fmt.Errorf("gametype not found: %w", err)
				}

				// Verify Fleet exists for the GameType
				getFleetCmd := exec.Command("kubectl", "get", "fleet", "-n", namespace, "--no-headers", "-l", "type="+gameTypeName)
				fleetOutput, err := utils.Run(getFleetCmd)
				if err != nil {
					return fmt.Errorf("fleet for gametype %s not found", gameTypeName)
				}
				if utils.IsNotFound(string(fleetOutput)) {
					return fmt.Errorf("no fleets found for game %s", gameTypeName)
				}

				// Verify Servers exist for the Fleet
				fleetName := strings.Fields(string(fleetOutput))[0]
				getServersCmd := exec.Command("kubectl", "get", "servers", "-n", namespace, "--no-headers", "-l", "fleet="+fleetName)
				serverOutput, err := utils.Run(getServersCmd)
				if err != nil {
					return fmt.Errorf("servers for gametype %s not found", gameTypeName)
				}
				if utils.IsNotFound(string(serverOutput)) {
					return fmt.Errorf("no servers found for game %s", gameTypeName)
				}

				serverCount := len(strings.Split(strings.TrimSpace(string(serverOutput)), "\n"))
				if serverCount == 0 {
					return fmt.Errorf("no servers found for fleet %s", fleetName)
				}

				log.Printf("GameType %s, Fleet %s and %d Servers for fleet are ready", gameTypeName, fleetName, serverCount)
				return nil
			}).Should(Succeed(), "Failed to verify game resources within timeout")
		})

		AfterEach(func() {
			By("Triggering gametype deletion")
			cmd := exec.Command("kubectl", "delete", "gametype", gameTypeName, "--namespace", namespace, "--wait=false")
			_, err := utils.Run(cmd)
			if err != nil {
				if utils.IsNotFound(err.Error()) {
					return
				}
				Expect(err).NotTo(HaveOccurred())
			}

			By("Allowing server deletion")
			allowGameTypeServersDelete(gameTypeName, namespace)

			time.Sleep(time.Second * 3)
			By("Verifying gametype was deleted")
			verifyGameTypeDeleted := func() error {
				cmd := exec.Command("kubectl", "get", "gametype", gameTypeName, "-n", namespace)
				_, err := utils.Run(cmd)
				if err != nil && strings.Contains(err.Error(), "not found") {
					return nil
				}
				return fmt.Errorf("gametype still exists or error: %v", err)
			}
			EventuallyWithOffset(1, verifyGameTypeDeleted, time.Minute, 5*time.Second).Should(Succeed())
		})

		It("Should create a fleet for the gametype", func() {
			By("Verifying fleet is created")
			verifyFleetCreation := func() error {
				cmd := exec.Command("kubectl", "get", "fleet", "-n", namespace, "--no-headers")
				output, err := utils.Run(cmd)
				if err != nil {
					if strings.Contains(err.Error(), "No resources found") {
						return fmt.Errorf("no fleet found for gametype")
					}
					return fmt.Errorf("error checking fleets: %w", err)
				}
				if strings.TrimSpace(string(output)) == "" {
					return fmt.Errorf("no fleet found for gametype")
				}
				return nil
			}
			EventuallyWithOffset(1, verifyFleetCreation, time.Minute, 5*time.Second).Should(Succeed())
		})

		It("Should add finalizer to the gametype", func() {
			By("Checking gametype has finalizer")
			getGameTypeFinalizers := func() (string, error) {
				cmd := exec.Command("kubectl", "get", "gametype", gameTypeName, "-n", namespace, "-o", "jsonpath={.metadata.finalizers}")
				output, err := utils.Run(cmd)
				return string(output), err
			}

			Eventually(getGameTypeFinalizers, time.Minute, 5*time.Second).Should(Equal("[\"gametype.unfamousthomas.me/finalizer\"]"))
		})

		It("Should create a new fleet when fleet spec changes", func() {
			By("Updating gametype fleet spec")
			patchCmd := exec.Command("kubectl", "patch", "gametype", gameTypeName, "-n", namespace, "--type", "merge", "-p",
				`{"spec":{"fleetSpec":{"updateStrategy":"newest_first"}}}`)
			_, err := utils.Run(patchCmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("Verifying new fleet is created")
			verifyFleetCount := func() error {
				cmd := exec.Command("kubectl", "get", "fleet", "-n", namespace, "--no-headers")
				output, err := utils.Run(cmd)
				if err != nil {
					return err
				}

				lines := strings.Split(strings.TrimSpace(string(output)), "\n")
				count := 0
				if strings.TrimSpace(string(output)) != "" {
					count = len(lines)
				}

				if count < 1 {
					return fmt.Errorf("expected at least 1 fleet, found %d", count)
				}
				return nil
			}
			EventuallyWithOffset(1, verifyFleetCount, 2*time.Minute, 5*time.Second).Should(Succeed())

			By("Verifying old fleet is deleted when new one is ready")
			Eventually(func() error {
				cmd := exec.Command("kubectl", "get", "fleet", "-n", namespace, "--no-headers")
				output, err := utils.Run(cmd)
				if err != nil {
					if strings.Contains(err.Error(), "No resources found") {
						return fmt.Errorf("no fleets found")
					}
					return err
				}

				lines := strings.Split(strings.TrimSpace(string(output)), "\n")
				count := 0
				if strings.TrimSpace(string(output)) != "" {
					count = len(lines)
				}

				if count > 1 {
					return fmt.Errorf("expected only 1 fleet to remain, found %d", count)
				}
				return nil
			}, 3*time.Minute, 5*time.Second).Should(Succeed())
		})
	})
})

func allowGameTypeServersDelete(gameName string, namespace string) {
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
		cmd := exec.Command("kubectl", "get", "servers", "-l", "type="+gameName, "-n", namespace, "-o", "name")
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
