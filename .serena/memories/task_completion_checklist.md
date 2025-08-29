# Task Completion Guidelines for Hub Project

## Code Quality Checklist
When completing any coding task, ensure these steps are followed:

### 1. Code Formatting and Linting
```bash
# Always run these commands before considering a task complete
go fmt ./...
go vet ./...
```

### 2. Build Verification
```bash
# Ensure the code builds successfully
go build -v ./...
```

### 3. Dependency Management
```bash
# Clean up dependencies if new ones were added
go mod tidy
```

### 4. Testing (When Applicable)
```bash
# Run relevant tests (currently has configuration issues)
go test -v .
# OR run specific tests
go test -v -run TestSpecificFunction
```

## Specific Guidelines for Hub Development

### When Adding New Features
1. **Follow the established patterns**: Use the same handler function signatures and configuration struct patterns
2. **Add appropriate error handling**: Return meaningful errors with context
3. **Consider both volatile and persistent messaging patterns** if applicable
4. **Update documentation**: Add examples to README.md if adding public APIs
5. **Add tests**: Create test cases following the existing test structure

### When Modifying Configuration
1. **Update Options struct** in `hub.go` if needed
2. **Maintain backward compatibility** when possible
3. **Update default option functions** if new fields are added
4. **Consider clustering implications** for new configuration options

### When Working with Messaging Patterns
1. **Choose appropriate pattern**: Volatile vs Persistent vs KV vs Object Store
2. **Handle ACK semantics correctly** for JetStream operations
3. **Implement proper error handlers**
4. **Consider performance implications** of chosen patterns

### File Organization
- **Core messaging**: Use `core.go`
- **Persistent messaging**: Use `jetstream.go`  
- **Storage operations**: Use `kv_store.go` or `object_store.go`
- **Configuration**: Add to `hub.go`
- **Size utilities**: Use `size.go`
- **Authentication**: Use `auth.go`

## Final Verification Steps
1. Code compiles without errors
2. Code is properly formatted (`go fmt`)
3. Static analysis passes (`go vet`)
4. Dependencies are clean (`go mod tidy`)
5. Documentation is updated if public APIs changed
6. Tests pass (when test configuration issues are resolved)

## Known Issues to Consider
- Current tests fail due to NATS server configuration requiring system account for leaf nodes and gateways
- Integration tests need function name updates (`DefaultOptions` vs `DefaultNodeOptions`)
- Consider these when working on testing-related tasks