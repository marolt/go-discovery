package collector

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/marolt/go-discovery/pkg/model"
)

// DetectWebServers identifies installed web servers
func DetectWebServers(report *model.DiscoveryReport, logger *log.Logger) {
	logger.Println("Detecting web servers")

	// Detect Apache web server
	detectApache(report, logger)

	// Detect Nginx web server
	detectNginx(report, logger)

	// Detect Lighttpd web server
	detectLighttpd(report, logger)

	// Detect Caddy web server
	detectCaddy(report, logger)

	logger.Printf("Detected %d web servers", len(report.WebServers))
}

// detectApache checks for Apache web server
func detectApache(report *model.DiscoveryReport, logger *log.Logger) {
	var webServer model.WebServer
	webServer.Type = "Apache"
	webServer.Status = "Not Installed"
	webServer.DocumentRoots = []string{}

	// Check if Apache is installed
	apacheExecNames := []string{"apache2", "httpd"}
	apacheInstalled := false

	for _, execName := range apacheExecNames {
		path, err := exec.LookPath(execName)
		if err == nil {
			apacheInstalled = true
			logger.Printf("Found Apache executable at %s", path)
			break
		}
	}

	if !apacheInstalled {
		logger.Println("Apache web server not found")
		return
	}

	// Check if Apache service is running
	webServer.Status = getServiceStatus("apache2", "httpd", logger)

	// Find Apache config file using command first, then fall back to predefined paths
	webServer.ConfigFile = getWebServerConfigFromCommand("apache", logger)

	if webServer.ConfigFile == "" {
		configFilePaths := []string{
			"/etc/apache2/apache2.conf",          // Debian/Ubuntu
			"/etc/apache2/httpd.conf",            // SUSE
			"/etc/httpd/conf/httpd.conf",         // RHEL/CentOS/Fedora
			"/usr/local/etc/apache24/httpd.conf", // FreeBSD
			"/opt/homebrew/etc/httpd/httpd.conf", // macOS (Homebrew ARM64)
			"/usr/local/etc/httpd/httpd.conf",    // macOS (Homebrew Intel)
			"/etc/apache2/httpd.conf",            // macOS (Built-in)
		}

		webServer.ConfigFile = findExistingFile(configFilePaths, logger)
	}

	// Extract document roots from config
	if webServer.ConfigFile != "" {
		docRoots := extractApacheDocumentRoots(webServer.ConfigFile, logger)
		webServer.DocumentRoots = append(webServer.DocumentRoots, docRoots...)

		// Look for included config files with DocumentRoot directives
		includeFiles := extractApacheIncludeFiles(webServer.ConfigFile, logger)
		for _, includeFile := range includeFiles {
			additionalDocRoots := extractApacheDocumentRoots(includeFile, logger)
			webServer.DocumentRoots = append(webServer.DocumentRoots, additionalDocRoots...)
		}
	}

	// Add Apache to the report if installed
	report.WebServers = append(report.WebServers, webServer)
	logger.Printf("Detected Apache web server: status=%s, config=%s", webServer.Status, webServer.ConfigFile)
}

// detectNginx checks for Nginx web server
func detectNginx(report *model.DiscoveryReport, logger *log.Logger) {
	var webServer model.WebServer
	webServer.Type = "Nginx"
	webServer.Status = "Not Installed"
	webServer.DocumentRoots = []string{}

	// Check if Nginx is installed
	nginxPath, err := exec.LookPath("nginx")
	if err != nil {
		logger.Println("Nginx web server not found")
		return
	}
	logger.Printf("Found Nginx executable at %s", nginxPath)

	// Check if Nginx service is running
	webServer.Status = getServiceStatus("nginx", "", logger)

	// Find Nginx config file using nginx -t first, then fall back to predefined paths
	webServer.ConfigFile = getWebServerConfigFromCommand("nginx", logger)

	if webServer.ConfigFile == "" {
		configFilePaths := []string{
			"/etc/nginx/nginx.conf",              // Most Linux distros
			"/usr/local/etc/nginx/nginx.conf",    // FreeBSD, macOS (Homebrew Intel)
			"/opt/homebrew/etc/nginx/nginx.conf", // macOS (Homebrew ARM64)
		}

		webServer.ConfigFile = findExistingFile(configFilePaths, logger)
	}

	// Extract document roots (root directives) from nginx config
	if webServer.ConfigFile != "" {
		docRoots := extractNginxDocumentRoots(webServer.ConfigFile, logger)
		webServer.DocumentRoots = append(webServer.DocumentRoots, docRoots...)

		// Look for included config files with root directives
		includeFiles := extractNginxIncludeFiles(webServer.ConfigFile, logger)
		for _, includeFile := range includeFiles {
			additionalDocRoots := extractNginxDocumentRoots(includeFile, logger)
			webServer.DocumentRoots = append(webServer.DocumentRoots, additionalDocRoots...)
		}
	}

	// Add Nginx to the report
	report.WebServers = append(report.WebServers, webServer)
	logger.Printf("Detected Nginx web server: status=%s, config=%s",
		webServer.Status, webServer.ConfigFile)
}

// detectLighttpd checks for Lighttpd web server
func detectLighttpd(report *model.DiscoveryReport, logger *log.Logger) {
	var webServer model.WebServer
	webServer.Type = "Lighttpd"
	webServer.Status = "Not Installed"
	webServer.DocumentRoots = []string{}

	// Check if Lighttpd is installed
	lighttpdPath, err := exec.LookPath("lighttpd")
	if err != nil {
		logger.Println("Lighttpd web server not found")
		return
	}
	logger.Printf("Found Lighttpd executable at %s", lighttpdPath)

	// Check if Lighttpd service is running
	webServer.Status = getServiceStatus("lighttpd", "", logger)

	// Find Lighttpd config file using command first, then fall back to predefined paths
	webServer.ConfigFile = getWebServerConfigFromCommand("lighttpd", logger)

	if webServer.ConfigFile == "" {
		configFilePaths := []string{
			"/etc/lighttpd/lighttpd.conf",           // Most Linux distros
			"/usr/local/etc/lighttpd/lighttpd.conf", // FreeBSD, macOS (Homebrew)
		}

		webServer.ConfigFile = findExistingFile(configFilePaths, logger)
	}

	// Extract document roots from Lighttpd config
	if webServer.ConfigFile != "" {
		if data, err := os.ReadFile(webServer.ConfigFile); err == nil {
			content := string(data)
			// Extract "server.document-root" values
			lines := strings.Split(content, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "server.document-root") {
					parts := strings.SplitN(line, "=", 2)
					if len(parts) == 2 {
						docRoot := strings.Trim(strings.TrimSpace(parts[1]), "\"'")
						webServer.DocumentRoots = append(webServer.DocumentRoots, docRoot)
					}
				}
			}
		}
	}

	// Add Lighttpd to the report
	report.WebServers = append(report.WebServers, webServer)
	logger.Printf("Detected Lighttpd web server: status=%s, config=%s",
		webServer.Status, webServer.ConfigFile)
}

// detectCaddy checks for Caddy web server
func detectCaddy(report *model.DiscoveryReport, logger *log.Logger) {
	var webServer model.WebServer
	webServer.Type = "Caddy"
	webServer.Status = "Not Installed"
	webServer.DocumentRoots = []string{}

	// Check if Caddy is installed
	caddyPath, err := exec.LookPath("caddy")
	if err != nil {
		logger.Println("Caddy web server not found")
		return
	}
	logger.Printf("Found Caddy executable at %s", caddyPath)

	// Check if Caddy service is running
	webServer.Status = getServiceStatus("caddy", "", logger)

	// Find Caddy config file using command first, then fall back to predefined paths
	webServer.ConfigFile = getWebServerConfigFromCommand("caddy", logger)

	if webServer.ConfigFile == "" {
		configFilePaths := []string{
			"/etc/caddy/Caddyfile",           // Most Linux distros
			"/usr/local/etc/caddy/Caddyfile", // FreeBSD, macOS (Homebrew)
			"/etc/caddy/caddy.conf",          // Alternative name
		}

		webServer.ConfigFile = findExistingFile(configFilePaths, logger)
	}

	// Extract document roots from Caddy config
	if webServer.ConfigFile != "" {
		if data, err := os.ReadFile(webServer.ConfigFile); err == nil {
			content := string(data)
			// Look for root directives
			lines := strings.Split(content, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "root") {
					parts := strings.Fields(line)
					if len(parts) >= 2 {
						docRoot := parts[1]
						webServer.DocumentRoots = append(webServer.DocumentRoots, docRoot)
					}
				}
			}
		}
	}

	// Add Caddy to the report
	report.WebServers = append(report.WebServers, webServer)
	logger.Printf("Detected Caddy web server: status=%s, config=%s",
		webServer.Status, webServer.ConfigFile)
}

// getServiceStatus checks if a service is running
func getServiceStatus(serviceName string, alternativeServiceName string, logger *log.Logger) string {
	switch runtime.GOOS {
	case "linux":
		// First try systemctl
		cmd := exec.Command("systemctl", "is-active", serviceName)
		output, err := cmd.Output()
		if err == nil && strings.TrimSpace(string(output)) == "active" {
			return "Running"
		}

		// Check alternative service name if provided
		if alternativeServiceName != "" {
			cmd = exec.Command("systemctl", "is-active", alternativeServiceName)
			output, err = cmd.Output()
			if err == nil && strings.TrimSpace(string(output)) == "active" {
				return "Running"
			}
		}

		// Then try service command
		cmd = exec.Command("service", serviceName, "status")
		if err := cmd.Run(); err == nil {
			return "Running"
		}

		// Check for service process
		processCmd := exec.Command("pgrep", "-x", serviceName)
		if err := processCmd.Run(); err == nil {
			return "Running"
		}

		if alternativeServiceName != "" {
			processCmd = exec.Command("pgrep", "-x", alternativeServiceName)
			if err := processCmd.Run(); err == nil {
				return "Running"
			}
		}

		return "Installed but not running"

	case "darwin":
		// macOS - try launchctl
		cmd := exec.Command("launchctl", "list")
		output, err := cmd.Output()
		if err == nil && strings.Contains(string(output), serviceName) {
			return "Running"
		}

		// Check brew services
		brewCmd := exec.Command("brew", "services", "list")
		brewOutput, err := brewCmd.Output()
		if err == nil && strings.Contains(string(brewOutput), serviceName+" started") {
			return "Running"
		}

		// Check process
		processCmd := exec.Command("pgrep", "-x", serviceName)
		if err := processCmd.Run(); err == nil {
			return "Running"
		}

		return "Installed but not running"

	case "windows":
		// Windows - use SC query
		cmd := exec.Command("sc", "query", serviceName)
		output, err := cmd.Output()
		if err == nil && strings.Contains(string(output), "RUNNING") {
			return "Running"
		}

		return "Installed but not running"

	default:
		return "Unknown status"
	}
}

// findExistingFile tries a list of file paths and returns the first one that exists
func findExistingFile(paths []string, logger *log.Logger) string {
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			logger.Printf("Found configuration file at %s", path)
			return path
		}
	}
	logger.Println("Configuration file not found")
	return ""
}

// getWebServerConfigFromCommand tries to determine config file location using server's built-in commands
func getWebServerConfigFromCommand(serverType string, logger *log.Logger) string {
	var cmd *exec.Cmd
	var configPath string

	switch serverType {
	case "nginx":
		// Nginx provides -t flag to test config and show the path
		cmd = exec.Command("nginx", "-t")
		output, err := cmd.CombinedOutput() // Using CombinedOutput because nginx -t writes to stderr
		if err == nil || strings.Contains(string(output), "successful") {
			// Extract config path from the output
			outputStr := string(output)
			// Look for the pattern like "the configuration file /etc/nginx/nginx.conf syntax is ok"
			if idx := strings.Index(outputStr, "configuration file "); idx > 0 {
				rest := outputStr[idx+len("configuration file "):]
				if endIdx := strings.Index(rest, " "); endIdx > 0 {
					configPath = rest[:endIdx]
					logger.Printf("Found Nginx config path using nginx -t: %s", configPath)
					return configPath
				}
			}
		} else {
			logger.Printf("Failed to get config path from nginx -t: %v", err)
		}

	case "apache":
		// Try apache2ctl -V or httpd -V to get config details
		execNames := []string{"apache2ctl", "httpd", "apache2"}

		for _, execName := range execNames {
			if execName == "apache2" {
				cmd = exec.Command(execName, "-V")
			} else {
				cmd = exec.Command(execName, "-V")
			}

			output, err := cmd.Output()
			if err == nil {
				outputStr := string(output)

				// Look for SERVER_CONFIG_FILE
				configFile := extractValueFromApacheOutput(outputStr, "SERVER_CONFIG_FILE", logger)
				serverRoot := extractValueFromApacheOutput(outputStr, "HTTPD_ROOT", logger)

				if configFile != "" {
					// If config file is relative, combine with server root
					if !filepath.IsAbs(configFile) && serverRoot != "" {
						configPath = filepath.Join(serverRoot, configFile)
					} else {
						configPath = configFile
					}

					logger.Printf("Found Apache config path using %s -V: %s", execName, configPath)
					return configPath
				}
			} else {
				logger.Printf("Failed to get config path from %s -V: %v", execName, err)
			}
		}

	case "lighttpd":
		// Try lighttpd -p to print the parsed config
		cmd = exec.Command("lighttpd", "-p", "-f", "/etc/lighttpd/lighttpd.conf")
		_, err := cmd.Output()
		if err == nil {
			// If successful with default path
			configPath = "/etc/lighttpd/lighttpd.conf"
			logger.Printf("Verified Lighttpd config at: %s", configPath)
			return configPath
		}

		// Try alternative paths
		cmd = exec.Command("lighttpd", "-p", "-f", "/usr/local/etc/lighttpd/lighttpd.conf")
		_, err = cmd.Output()
		if err == nil {
			configPath = "/usr/local/etc/lighttpd/lighttpd.conf"
			logger.Printf("Verified Lighttpd config at: %s", configPath)
			return configPath
		}

	case "caddy":
		// Try to find Caddy config through environment or common paths
		// Caddy v2 often stores its config in /etc/caddy/Caddyfile
		cmd = exec.Command("caddy", "version")
		output, err := cmd.Output()
		if err == nil && strings.Contains(string(output), "v2") {
			// For Caddy v2, try to validate the default config locations
			locations := []string{
				"/etc/caddy/Caddyfile",
				"/usr/local/etc/caddy/Caddyfile",
			}

			for _, loc := range locations {
				validateCmd := exec.Command("caddy", "validate", "--adapter", "caddyfile", loc)
				if err := validateCmd.Run(); err == nil {
					configPath = loc
					logger.Printf("Validated Caddy config at: %s", configPath)
					return configPath
				}
			}
		}
	}

	logger.Printf("Could not determine %s config path from commands, will try predefined paths", serverType)
	return ""
}

// extractValueFromApacheOutput extracts a key value from Apache -V output
func extractValueFromApacheOutput(output string, key string, logger *log.Logger) string {
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, key) {
			start := strings.Index(line, key+"=\"")
			if start > 0 {
				start = start + len(key) + 2 // Skip past key=" part
				end := strings.Index(line[start:], "\"")
				if end > 0 {
					value := line[start : start+end]
					logger.Printf("Found %s: %s", key, value)
					return value
				}
			}
		}
	}
	return ""
}

// extractApacheDocumentRoots extracts DocumentRoot directives from Apache config
func extractApacheDocumentRoots(configFile string, logger *log.Logger) []string {
	var docRoots []string

	// Read the configuration file
	data, err := os.ReadFile(configFile)
	if err != nil {
		logger.Printf("Error reading Apache config file %s: %v", configFile, err)
		return docRoots
	}

	// Parse the file content to find DocumentRoot directives
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "DocumentRoot") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				// Remove quotes from the document root if present
				docRoot := strings.Trim(parts[1], "\"'")
				docRoots = append(docRoots, docRoot)
			}
		}
	}

	return docRoots
}

// extractApacheIncludeFiles finds all included/imported configuration files in Apache config
func extractApacheIncludeFiles(configFile string, logger *log.Logger) []string {
	var includeFiles []string

	// Read the configuration file
	data, err := os.ReadFile(configFile)
	if err != nil {
		logger.Printf("Error reading Apache config file %s: %v", configFile, err)
		return includeFiles
	}

	// Parse the file content to find Include directives
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Include") || strings.HasPrefix(line, "IncludeOptional") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				includePattern := strings.Trim(parts[1], "\"'")

				// Handle relative paths
				if !filepath.IsAbs(includePattern) {
					includePattern = filepath.Join(filepath.Dir(configFile), includePattern)
				}

				// If it's a directory, find all .conf files
				fileInfo, err := os.Stat(includePattern)
				if err == nil && fileInfo.IsDir() {
					files, err := filepath.Glob(filepath.Join(includePattern, "*.conf"))
					if err == nil {
						includeFiles = append(includeFiles, files...)
					}
				} else {
					// Handle wildcards
					matches, err := filepath.Glob(includePattern)
					if err == nil {
						includeFiles = append(includeFiles, matches...)
					}
				}
			}
		}
	}

	return includeFiles
}

// extractNginxDocumentRoots extracts root directives from Nginx config
func extractNginxDocumentRoots(configFile string, logger *log.Logger) []string {
	var docRoots []string

	// Read the configuration file
	data, err := os.ReadFile(configFile)
	if err != nil {
		logger.Printf("Error reading Nginx config file %s: %v", configFile, err)
		return docRoots
	}

	// Parse the file content to find root directives
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "root") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				// Remove trailing semicolon and quotes
				docRoot := strings.TrimRight(strings.Trim(parts[1], "\"'"), ";")
				docRoots = append(docRoots, docRoot)
			}
		}
	}

	return docRoots
}

// extractNginxIncludeFiles finds all included configuration files in Nginx config
func extractNginxIncludeFiles(configFile string, logger *log.Logger) []string {
	var includeFiles []string

	// Read the configuration file
	data, err := os.ReadFile(configFile)
	if err != nil {
		logger.Printf("Error reading Nginx config file %s: %v", configFile, err)
		return includeFiles
	}

	// Parse the file content to find include directives
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "include") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				// Remove trailing semicolon and quotes
				includePattern := strings.TrimRight(strings.Trim(parts[1], "\"'"), ";")

				// Handle relative paths
				if !filepath.IsAbs(includePattern) {
					includePattern = filepath.Join(filepath.Dir(configFile), includePattern)
				}

				// Handle wildcards
				matches, err := filepath.Glob(includePattern)
				if err == nil {
					includeFiles = append(includeFiles, matches...)
				}
			}
		}
	}

	return includeFiles
}
