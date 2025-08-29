# Hub Development Commands

## Essential Go Commands

### Building and Dependencies
```bash
# Build the project
go build -v ./...

# Download and tidy dependencies
go mod tidy

# Download dependencies
go mod download

# Verify dependencies
go mod verify
```

### Code Quality
```bash
# Format code
go fmt ./...

# Static analysis
go vet ./...

# Run tests (Note: Current tests have configuration issues)
go test -v .

# Run tests with coverage
go test -v -cover ./...

# Run tests with race detection
go test -v -race ./...
```

### Testing
```bash
# Run unit tests
go test -v .

# Run integration tests (needs fixing)
cd tests && go run main.go

# Run specific test
go test -v -run TestSpecificFunction

# Run benchmarks
go test -v -bench=.
```

## Project-Specific Commands

### Development Workflow
```bash
# Complete development check
go mod tidy && go fmt ./... && go vet ./... && go build -v ./...

# Quick validation
go fmt ./... && go vet ./...

# Install as module dependency (when using in other projects)
go get github.com/snowmerak/hub
```

## System Commands (macOS)

### Basic Unix Commands
```bash
# File operations
ls -la          # List files with details
find . -name "*.go"  # Find Go files
grep -r "pattern" .  # Search for patterns

# Directory operations
cd /path/to/dir
pwd
mkdir -p path/to/dir

# Git operations
git status
git add .
git commit -m "message"
git push
git pull

# File viewing
cat filename
less filename
head -n 10 filename
tail -n 10 filename
```

### Development Tools
```bash
# Check Go version
go version

# Check Go environment
go env

# Clean module cache
go clean -modcache

# Check for updates
go list -u -m all
```

## Notes
- Tests currently fail due to NATS server configuration issues (leaf nodes and gateways require system account)
- Integration tests in `tests/` directory need to be updated to use correct function names
- No Makefile present - using standard Go toolchain commands
- Project uses Apache 2.0 License