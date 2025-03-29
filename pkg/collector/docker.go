package collector

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/marolt/go-discovery/pkg/model"
)

// DetectDockerContainers identifies Docker containers and their configuration
func DetectDockerContainers(report *model.DiscoveryReport, logger *log.Logger) {
	logger.Println("Detecting Docker containers")

	// Check if Docker is installed
	if _, err := exec.LookPath("docker"); err != nil {
		logger.Println("Docker is not installed or not in PATH")
		return
	}

	// Check if Docker daemon is running
	statusCmd := exec.Command("docker", "info")
	if err := statusCmd.Run(); err != nil {
		logger.Printf("Docker is installed but daemon is not running: %v", err)
		return
	}

	// Get list of running containers
	psCmd := exec.Command("docker", "ps", "--format", "{{.ID}}")
	output, err := psCmd.Output()
	if err != nil {
		logger.Printf("Error getting running containers: %v", err)
		return
	}

	// Process each container
	containerIDs := strings.Fields(string(output))
	for _, containerID := range containerIDs {
		container := model.DockerContainer{
			ContainerID: containerID,
			Ports:       []string{},
			Volumes:     []string{},
			Networks:    []string{},
		}

		// Get container name
		nameCmd := exec.Command("docker", "inspect", "--format", "{{.Name}}", containerID)
		nameOutput, err := nameCmd.Output()
		if err == nil {
			container.Name = strings.TrimSpace(strings.TrimPrefix(string(nameOutput), "/"))
		} else {
			logger.Printf("Error getting name for container %s: %v", containerID, err)
			container.Name = "unknown"
		}

		// Get container image
		imageCmd := exec.Command("docker", "inspect", "--format", "{{.Config.Image}}", containerID)
		imageOutput, err := imageCmd.Output()
		if err == nil {
			container.Image = strings.TrimSpace(string(imageOutput))
		} else {
			logger.Printf("Error getting image for container %s: %v", containerID, err)
			container.Image = "unknown"
		}

		// Get port mappings
		portsCmd := exec.Command("docker", "inspect", "--format", "{{range $p, $conf := .NetworkSettings.Ports}} {{$p}}->{{(index $conf 0).HostPort}} {{end}}", containerID)
		portsOutput, err := portsCmd.Output()
		if err == nil {
			portsList := strings.Split(strings.TrimSpace(string(portsOutput)), " ")
			for _, port := range portsList {
				if port != "" {
					container.Ports = append(container.Ports, port)
				}
			}
		} else {
			logger.Printf("Error getting ports for container %s: %v", containerID, err)
		}

		// Get volume mappings
		volumesCmd := exec.Command("docker", "inspect", "--format", "{{range .Mounts}}{{.Source}}:{{.Destination}} {{end}}", containerID)
		volumesOutput, err := volumesCmd.Output()
		if err == nil {
			volumesList := strings.Fields(strings.TrimSpace(string(volumesOutput)))
			container.Volumes = volumesList
		} else {
			logger.Printf("Error getting volumes for container %s: %v", containerID, err)
		}

		// Get networks
		networksCmd := exec.Command("docker", "inspect", "--format", "{{range $net, $conf := .NetworkSettings.Networks}}{{$net}} {{end}}", containerID)
		networksOutput, err := networksCmd.Output()
		if err == nil {
			networksList := strings.Fields(strings.TrimSpace(string(networksOutput)))
			container.Networks = networksList
		} else {
			logger.Printf("Error getting networks for container %s: %v", containerID, err)
		}

		// Check if container is managed by Docker Compose
		composeProjectCmd := exec.Command("docker", "inspect", "--format", "{{index .Config.Labels \"com.docker.compose.project\"}}", containerID)
		composeProjectOutput, err := composeProjectCmd.Output()
		composeProject := strings.TrimSpace(string(composeProjectOutput))

		composeServiceCmd := exec.Command("docker", "inspect", "--format", "{{index .Config.Labels \"com.docker.compose.service\"}}", containerID)
		composeServiceOutput, err := composeServiceCmd.Output()
		composeService := strings.TrimSpace(string(composeServiceOutput))

		// Determine management type and set related fields
		if composeProject != "" || composeService != "" {
			container.ManagedBy = "docker-compose"
			container.ComposeProject = composeProject
			container.ComposeService = composeService

			// Find compose file location
			composeFileCmd := exec.Command("docker", "inspect", "--format", "{{index .Config.Labels \"com.docker.compose.project.working_dir\"}}", containerID)
			composeFileOutput, err := composeFileCmd.Output()
			if err == nil {
				workingDir := strings.TrimSpace(string(composeFileOutput))
				container.ComposeFile = filepath.Join(workingDir, "docker-compose.yml")

				// Check if the file exists, try .yaml extension if not
				if _, err := os.Stat(container.ComposeFile); os.IsNotExist(err) {
					yamlPath := filepath.Join(workingDir, "docker-compose.yaml")
					if _, err := os.Stat(yamlPath); err == nil {
						container.ComposeFile = yamlPath
					}
				}
			} else {
				logger.Printf("Error getting compose file for container %s: %v", containerID, err)
				container.ComposeFile = "unknown"
			}
		} else {
			container.ManagedBy = "standalone"
			container.ComposeProject = "-"
			container.ComposeService = "-"
			container.ComposeFile = "-"
		}

		// Add container to report
		report.DockerContainers = append(report.DockerContainers, container)
	}

	// Also check if Docker Compose is installed and find compose files
	if _, err := exec.LookPath("docker-compose"); err == nil || isDockerComposeV2Available() {
		logger.Println("Docker Compose is installed")
		findComposeFiles(logger)
	}

	logger.Printf("Detected %d Docker containers", len(report.DockerContainers))
}

// isDockerComposeV2Available checks if Docker Compose V2 is available
func isDockerComposeV2Available() bool {
	cmd := exec.Command("docker", "compose", "version")
	return cmd.Run() == nil
}

// findComposeFiles looks for docker-compose files in common locations
func findComposeFiles(logger *log.Logger) {
	// Common paths to search for compose files
	paths := []string{"/opt", "/srv", "/home"}

	for _, path := range paths {
		err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return filepath.SkipDir
			}

			// Skip deep directories to improve performance
			if strings.Count(path, "/") > 5 {
				return filepath.SkipDir
			}

			// Check if file is a compose file
			if !info.IsDir() && (info.Name() == "docker-compose.yml" || info.Name() == "docker-compose.yaml") {
				logger.Printf("Found docker-compose file: %s", path)
			}
			return nil
		})

		if err != nil {
			logger.Printf("Error searching for docker-compose files in %s: %v", path, err)
		}
	}
}
