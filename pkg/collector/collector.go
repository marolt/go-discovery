package collector

import (
	"log"

	"github.com/marolt/go-discovery/pkg/model"
)

// RunDiscovery performs the complete system discovery process
func RunDiscovery(logger *log.Logger) *model.DiscoveryReport {
	// Create the report structure
	report := model.NewDiscoveryReport()

	// Collect system information
	CollectSystemInfo(report, logger)

	// Detect web servers
	DetectWebServers(report, logger)

	// Detect databases
	DetectDatabases(report, logger)

	// Detect Docker containers
	DetectDockerContainers(report, logger)

	return report
}
