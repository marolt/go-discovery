package collector

import (
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/marolt/go-discovery/pkg/model"
)

// CollectSystemInfo gathers basic system information
func CollectSystemInfo(report *model.DiscoveryReport, logger *log.Logger) {
	logger.Println("Collecting system information")

	// Get hostname
	hostname, err := os.Hostname()
	if err != nil {
		logger.Printf("Error getting hostname: %v", err)
		hostname = "unknown"
	}
	report.Hostname = hostname

	// Get OS information based on the platform
	switch runtime.GOOS {
	case "linux":
		collectLinuxInfo(report, logger)
	case "darwin":
		collectMacOSInfo(report, logger)
	case "windows":
		collectWindowsInfo(report, logger)
	default:
		logger.Printf("Unsupported operating system: %s", runtime.GOOS)
		report.SystemInfo = model.SystemInfo{
			OSName:    runtime.GOOS,
			OSVersion: "unknown",
			Kernel:    runtime.GOARCH,
		}
	}

	logger.Printf("Collected system info: OS=%s, Version=%s, Kernel=%s",
		report.SystemInfo.OSName, report.SystemInfo.OSVersion, report.SystemInfo.Kernel)
}

// collectLinuxInfo gathers information specific to Linux systems
func collectLinuxInfo(report *model.DiscoveryReport, logger *log.Logger) {
	// Initialize with defaults
	osName := "Linux"
	osVersion := "unknown"

	// Try to get distribution info from /etc/os-release (systemd standard)
	if data, err := os.ReadFile("/etc/os-release"); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			key := parts[0]
			value := strings.Trim(parts[1], "\"'")

			switch key {
			case "NAME":
				osName = value
			case "VERSION_ID":
				osVersion = value
			}
		}
	} else {
		logger.Printf("Failed to read /etc/os-release: %v", err)

		// Try alternative files for distribution identification
		osInfo := tryAlternativeDistroFiles(logger)
		if osInfo.name != "" {
			osName = osInfo.name
		}
		if osInfo.version != "" {
			osVersion = osInfo.version
		}
	}

	// Get kernel version
	kernel := getKernelVersion(logger)

	report.SystemInfo = model.SystemInfo{
		OSName:    osName,
		OSVersion: osVersion,
		Kernel:    kernel,
	}
}

// distroInfo holds distribution name and version
type distroInfo struct {
	name    string
	version string
}

// tryAlternativeDistroFiles attempts to identify Linux distribution using alternative files
func tryAlternativeDistroFiles(logger *log.Logger) distroInfo {
	info := distroInfo{}

	// Check for /etc/lsb-release (Ubuntu and some derivatives)
	if data, err := os.ReadFile("/etc/lsb-release"); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			key := parts[0]
			value := strings.Trim(parts[1], "\"'")

			switch key {
			case "DISTRIB_ID":
				info.name = value
			case "DISTRIB_RELEASE":
				info.version = value
			}
		}
		if info.name != "" {
			return info
		}
	}

	// Check for /etc/redhat-release (RHEL, CentOS, Fedora)
	if data, err := os.ReadFile("/etc/redhat-release"); err == nil {
		text := strings.TrimSpace(string(data))
		// Parse "CentOS Linux release 7.9.2009 (Core)" format
		if parts := strings.SplitN(text, "release", 2); len(parts) == 2 {
			info.name = strings.TrimSpace(parts[0])
			versionPart := strings.TrimSpace(parts[1])
			if spaceParts := strings.SplitN(versionPart, " ", 2); len(spaceParts) > 0 {
				info.version = spaceParts[0]
			}
		} else {
			info.name = text
		}
		return info
	}

	// Check for /etc/debian_version (Debian-based)
	if data, err := os.ReadFile("/etc/debian_version"); err == nil {
		info.name = "Debian"
		info.version = strings.TrimSpace(string(data))
		return info
	}

	// Check if /etc/issue exists and use it as fallback
	if data, err := os.ReadFile("/etc/issue"); err == nil {
		text := strings.TrimSpace(string(data))
		if text != "" {
			// Just use the first line
			firstLine := strings.SplitN(text, "\n", 2)[0]
			info.name = firstLine
			// We can't reliably extract version from here
		}
		return info
	}

	logger.Println("Could not identify Linux distribution from standard files")
	return info
}

// collectMacOSInfo gathers information specific to macOS systems
func collectMacOSInfo(report *model.DiscoveryReport, logger *log.Logger) {
	osName := "macOS"
	osVersion := "unknown"

	// Get macOS version using sw_vers
	cmd := exec.Command("sw_vers", "-productVersion")
	if output, err := cmd.Output(); err == nil {
		osVersion = strings.TrimSpace(string(output))
	} else {
		logger.Printf("Error getting macOS version: %v", err)
	}

	// Get kernel version
	kernel := getKernelVersion(logger)

	report.SystemInfo = model.SystemInfo{
		OSName:    osName,
		OSVersion: osVersion,
		Kernel:    kernel,
	}
}

// collectWindowsInfo gathers information specific to Windows systems
func collectWindowsInfo(report *model.DiscoveryReport, logger *log.Logger) {
	osName := "Windows"
	osVersion := "unknown"
	kernel := "unknown"

	// Get Windows version using PowerShell
	cmd := exec.Command("powershell", "-Command", "(Get-CimInstance Win32_OperatingSystem).Caption")
	if output, err := cmd.Output(); err == nil {
		osName = strings.TrimSpace(string(output))
	} else {
		logger.Printf("Error getting Windows OS name: %v", err)
	}

	// Get Windows version
	versionCmd := exec.Command("powershell", "-Command", "(Get-CimInstance Win32_OperatingSystem).Version")
	if output, err := versionCmd.Output(); err == nil {
		osVersion = strings.TrimSpace(string(output))
	} else {
		logger.Printf("Error getting Windows version: %v", err)
	}

	// Get kernel/build information
	buildCmd := exec.Command("powershell", "-Command", "(Get-CimInstance Win32_OperatingSystem).BuildNumber")
	if output, err := buildCmd.Output(); err == nil {
		kernel = strings.TrimSpace(string(output))
	} else {
		logger.Printf("Error getting Windows build number: %v", err)
	}

	report.SystemInfo = model.SystemInfo{
		OSName:    osName,
		OSVersion: osVersion,
		Kernel:    kernel,
	}
}

// getKernelVersion gets the kernel version using uname
func getKernelVersion(logger *log.Logger) string {
	cmd := exec.Command("uname", "-r")
	output, err := cmd.Output()
	if err != nil {
		logger.Printf("Error getting kernel version: %v", err)
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}
