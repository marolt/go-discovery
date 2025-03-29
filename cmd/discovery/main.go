package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/marolt/go-discovery/pkg/collector"
	"github.com/marolt/go-discovery/pkg/report"
)

// Version information set by build process
var (
	version   = "dev"
	buildTime = "unknown"
)

func main() {
	// Add a version flag
	showVersion := flag.Bool("version", false, "Show version information")

	// Parse command line flags
	outputFormat := flag.String("format", "yaml", "Output format: yaml or json")
	outputFile := flag.String("output", "", "Output file (default: system_discovery_report.[yaml|json])")
	logFile := flag.String("log", "system_discovery.log", "Log file")
	logToStdout := flag.Bool("stdout", true, "Log to stdout as well as log file")
	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Printf("go-discovery version %s (built at %s)\n", version, buildTime)
		return
	}

	// Setup logging
	logWriter, err := os.Create(*logFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating log file: %v\n", err)
		os.Exit(1)
	}
	defer logWriter.Close()

	var writer io.Writer
	if *logToStdout {
		writer = io.MultiWriter(logWriter, os.Stdout)
	} else {
		writer = logWriter
	}
	// Create a logger
	logger := log.New(writer, "", log.LstdFlags)

	logger.Println("Starting system discovery")

	// Create the discovery report
	discoveryReport := collector.RunDiscovery(logger)

	// Determine output file name if not specified
	if *outputFile == "" {
		*outputFile = fmt.Sprintf("system_discovery_report.%s", *outputFormat)
	}

	// Write the report
	if err := report.WriteReport(discoveryReport, *outputFile, *outputFormat, logger); err != nil {
		logger.Fatalf("Failed to write report: %v", err)
	}

	logger.Println("System discovery completed successfully")
}
