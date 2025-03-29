# go-discovery

A comprehensive system discovery and inventory tool written in Go. This tool scans your system and collects information about the operating system, installed web servers, databases, and running Docker containers.

## Features

- **System Information**: Identifies OS name, version, and kernel details
- **Web Server Discovery**: Detects and analyzes Apache, Nginx, Lighttpd, and Caddy installations
- **Database Detection**: Identifies installed database servers
- **Docker Analysis**: Discovers running Docker containers including those managed by Docker Compose
- **Flexible Output**: Generates reports in YAML or JSON format

## Installation

### Pre-built Binaries

Download the pre-built binary for your platform from the [releases page](https://github.com/marolt/go-discovery/releases).

### Building from Source

Requirements:
- Go 1.24.0 or later
- Make (optional, for using the Makefile)

```bash
# Clone the repository
git clone https://github.com/marolt/go-discovery.git
cd go-discovery

# Build using Make
make build

# Or build directly with Go
go build -o bin/discovery ./cmd/discovery
```

## Usage

Run the binary with optional flags:

```bash
./discovery [options]
```

### Command-line Options

| Option | Default | Description |
|--------|---------|-------------|
| `-format` | `yaml` | Output format: `yaml` or `json` |
| `-output` | `system_discovery_report.[yaml|json]` | Output file path |
| `-log` | `system_discovery.log` | Log file path |
| `-stdout` | `true` | Log to stdout as well as log file |

### Examples

Basic usage:
```bash
./discovery
```

Generate JSON output:
```bash
./discovery -format json -output inventory.json
```

Silent operation (logs only to file):
```bash
./discovery -stdout=false
```

## Output Example

The tool generates a structured report similar to:

```yaml
timestamp: "2023-08-13T15:04:05Z"
hostname: "server01"
system_info:
  os_name: "Ubuntu"
  os_version: "22.04"
  kernel: "5.15.0-56-generic"
web_servers:
  - type: "Nginx"
    status: "Running"
    config_file: "/etc/nginx/nginx.conf"
    document_roots: 
      - "/var/www/html"
databases:
  - type: "MySQL"
    service: "mysql"
    status: "Running"
    config_file: "/etc/mysql/mysql.conf.d/mysqld.cnf"
    data_directory: "/var/lib/mysql"
docker_containers:
  - container_id: "a1b2c3d4e5f6"
    name: "webapp"
    image: "nginx:latest"
    ports:
      - "80:80"
    volumes:
      - "/data:/app/data"
    networks:
      - "frontend"
    managed_by: "docker-compose"
    compose_project: "mywebapp"
    compose_service: "web"
    compose_file: "/opt/mywebapp/docker-compose.yml"
```

## Development

### Project Structure

```
go-discovery/
├── cmd/
│   └── discovery/      # Application entry point
│       └── main.go
├── pkg/
│   ├── collector/      # System information collectors
│   │   ├── system.go
│   │   ├── webserver.go
│   │   ├── database.go
│   │   └── docker.go
│   ├── model/          # Data structures
│   │   └── types.go
│   └── report/         # Report generation
│       └── writer.go
├── Makefile            # Build automation
└── README.md           # This file
```

### Available Make Targets

- `make build` - Build the application
- `make clean` - Clean build artifacts
- `make test` - Run tests
- `make fmt` - Format code
- `make lint` - Lint code
- `make build-all` - Build for multiple platforms
- `make install` - Install binary to GOPATH
- `make help` - Show available commands

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request
