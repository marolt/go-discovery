package report

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/marolt/go-discovery/pkg/model"
)

// WriteReport outputs the report in the requested format
func WriteReport(report *model.DiscoveryReport, outputFile, format string, logger *log.Logger) error {
	logger.Printf("Writing report to %s in %s format", outputFile, format)

	var data []byte
	var err error

	switch format {
	case "json":
		data, err = json.MarshalIndent(report, "", "  ")
	case "yaml":
		data, err = yaml.Marshal(report)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}

	if err != nil {
		return fmt.Errorf("error marshaling report: %v", err)
	}

	if err := os.WriteFile(outputFile, data, 0644); err != nil {
		return fmt.Errorf("error writing report file: %v", err)
	}

	return nil
}
