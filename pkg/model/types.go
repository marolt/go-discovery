package model

import "time"

// SystemInfo contains basic information about the system
type SystemInfo struct {
	OSName    string `json:"os_name" yaml:"os_name"`
	OSVersion string `json:"os_version" yaml:"os_version"`
	Kernel    string `json:"kernel" yaml:"kernel"`
}

// WebServer represents a detected web server
type WebServer struct {
	Type          string   `json:"type" yaml:"type"`
	Status        string   `json:"status" yaml:"status"`
	ConfigFile    string   `json:"config_file" yaml:"config_file"`
	DocumentRoots []string `json:"document_roots" yaml:"document_roots"`
}

// Database represents a detected database server
type Database struct {
	Type          string `json:"type" yaml:"type"`
	Service       string `json:"service,omitempty" yaml:"service,omitempty"`
	Status        string `json:"status" yaml:"status"`
	ConfigFile    string `json:"config_file" yaml:"config_file"`
	DataDirectory string `json:"data_directory" yaml:"data_directory"`
}

// DockerContainer represents a detected docker container
type DockerContainer struct {
	ContainerID    string   `json:"container_id" yaml:"container_id"`
	Name           string   `json:"name" yaml:"name"`
	Image          string   `json:"image" yaml:"image"`
	Ports          []string `json:"ports" yaml:"ports"`
	Volumes        []string `json:"volumes" yaml:"volumes"`
	Networks       []string `json:"networks" yaml:"networks"`
	ManagedBy      string   `json:"managed_by" yaml:"managed_by"`
	ComposeProject string   `json:"compose_project" yaml:"compose_project"`
	ComposeService string   `json:"compose_service" yaml:"compose_service"`
	ComposeFile    string   `json:"compose_file" yaml:"compose_file"`
}

// DiscoveryReport represents the complete system discovery report
type DiscoveryReport struct {
	Timestamp        string            `json:"timestamp" yaml:"timestamp"`
	Hostname         string            `json:"hostname" yaml:"hostname"`
	SystemInfo       SystemInfo        `json:"system_info" yaml:"system_info"`
	WebServers       []WebServer       `json:"web_servers" yaml:"web_servers"`
	Databases        []Database        `json:"databases" yaml:"databases"`
	DockerContainers []DockerContainer `json:"docker_containers" yaml:"docker_containers"`
}

// NewDiscoveryReport creates a new discovery report with timestamp set
func NewDiscoveryReport() *DiscoveryReport {
	return &DiscoveryReport{
		Timestamp: time.Now().Format(time.RFC3339),
	}
}
