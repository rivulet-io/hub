# Hub Code Style and Conventions

## Naming Conventions
- **Package Level**: `hub` (single, lowercase)
- **Constants**: PascalCase (e.g., `HubClusterName`, `SizeBytes`)
- **Structs**: PascalCase (e.g., `Hub`, `Options`, `KeyValueStoreConfig`)
- **Methods**: PascalCase for exported, camelCase for unexported
- **Variables**: camelCase (e.g., `opts`, `tempDir`)
- **Config Structs**: Descriptive with "Config" suffix (e.g., `PersistentConfig`, `ObjectStoreConfig`)

## Function Signatures
- **Factory Functions**: `New*` prefix (e.g., `NewHub`, `NewSizeFromMegabytes`)
- **Default Constructors**: `Default*Options` pattern (e.g., `DefaultNodeOptions`, `DefaultGatewayOptions`)
- **CRUD Operations**: Clear action verbs (e.g., `CreateOrUpdatePersistent`, `PutToKeyValueStore`)

## Error Handling
- Return `error` as last return value
- Use `fmt.Errorf` for error wrapping with `%w` verb
- Descriptive error messages with context
- Validate inputs and return meaningful errors

## Handler Function Patterns
```go
// Volatile messaging handler
func(subject string, msg []byte) (response []byte, reply bool)

// JetStream handler  
func(subject string, msg []byte) (response []byte, reply bool, ack bool)

// Error handler
func(error)
```

## Configuration Patterns
- Use struct-based configuration (not functional options)
- Provide sensible defaults via `Default*Options()` functions
- Group related options in sub-structs when appropriate
- Use `Size` type for byte quantities with human-readable constructors

## Testing Conventions
- Test file naming: `*_test.go`
- Test function naming: `Test*` prefix describing the feature
- Use table-driven tests for multiple scenarios
- Defer cleanup operations
- Create temporary directories for file-based tests
- Use channels and timeouts for async operations testing

## Documentation Style
- Package-level documentation in README.md with comprehensive examples
- Exported functions and types should have Go doc comments
- Include usage examples in documentation
- Document complex configuration options with use cases