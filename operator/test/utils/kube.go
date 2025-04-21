package utils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func GetPodIP(podName, namespace string) (string, error) {
	cmd := exec.Command("kubectl", "get", "pod", podName, "-n", namespace, "-o", "jsonpath={.status.podIP}")
	output, err := Run(cmd)
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func SendAllowDeleteRequest(ip string, namespace string) error {
	// Create a temporary pod with curl installed
	// This is to avoid issues with networking in KinD clusters
	if ip == "" {
		return fmt.Errorf("ip is invalid")
	}
	curlPodYaml := fmt.Sprintf(`
apiVersion: v1
kind: Pod
metadata:
  name: curl-pod
  namespace: %s
spec:
  containers:
  - name: curl
    image: curlimages/curl
    command: ["sleep", "300"]
  restartPolicy: Never
`, namespace)

	// Apply the curl pod
	tmpFile, err := os.CreateTemp("", "curl-pod-*.yaml")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(curlPodYaml); err != nil {
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}

	cmd := exec.Command("kubectl", "apply", "-f", tmpFile.Name())
	if _, err := Run(cmd); err != nil {
		return err
	}

	// Wait for pod to be ready
	waitCmd := exec.Command("kubectl", "wait", "--for=condition=Ready", "--timeout=30s", "-n", namespace, "pod/curl-pod")
	if _, err := Run(waitCmd); err != nil {
		return err
	}

	// JSON payload for the request
	payload := `{"allowed":true}`

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Execute the curl command from inside the pod with the JSON payload
	curlCmd := exec.CommandContext(ctx, "kubectl", "exec", "-n", namespace, "curl-pod", "--",
		"curl", "-X", "POST",
		"-H", "Content-Type: application/json",
		"-d", payload,
		"--connect-timeout", "10", // 10 seconds connection timeout
		"--max-time", "20", // 20 seconds maximum time for operation
		fmt.Sprintf("http://%s:%d/allow_delete", ip, 8080))
	_, err = Run(curlCmd)

	// Clean up
	deleteCmd := exec.Command("kubectl", "delete", "pod", "-n", namespace, "curl-pod", "--grace-period=0", "--force")
	_, _ = Run(deleteCmd) // Ignore errors during cleanup

	return err
}

func PodExists(podName, namespace string) (bool, error) {
	cmd := exec.Command("kubectl", "get", "pod", podName, "-n", namespace, "--ignore-not-found")
	output, err := Run(cmd)
	if err != nil {
		return false, fmt.Errorf("error checking pod existence: %w", err)
	}

	return strings.TrimSpace(string(output)) != "", nil
}
