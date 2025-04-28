package e2e

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/unfamousthomas/thesis-operator/test/utils"
	"log"
	"os/exec"
)

var _ = Context("Service", Ordered, func() {
	var serviceIp string

	BeforeAll(func() {
		By("Get pod name with label app=http-controller-service")
		getPodCmd := exec.Command("kubectl", "get", "pods", "-l", "app=http-controller-service", "-n", "default", "-o", "jsonpath={.items[0].metadata.name}")
		output, err := utils.Run(getPodCmd)
		ExpectWithOffset(1, err).NotTo(HaveOccurred())

		podName := string(output)
		log.Printf("Found Pod: %s", podName)

		By("Get IP")
		serviceIp, err = utils.GetPodIP(podName, "default")
		if err != nil {
			if utils.IsNotFound(err.Error()) {
				log.Printf("Resource not found, returning nil")
				return
			}
		}
	})

	It("It should create and remove server using the service", func() {
		fmt.Println(serviceIp)
		//Setup the test
	})

	It("Should modify the pod labels using the service", func() {
		//Setup the test
	})

	It("Should create and remove fleet using the service", func() {
		//Setup the test
	})

	It("Should create and remove the gametype using the service", func() {
		//Setup the test
	})

	It("Should create and remove the scaler using the service", func() {
		//Setup the test
	})
})
