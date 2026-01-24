# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.0] - 2026-01-24

### Added
- Add Claude LLM provider support alongside OpenAI with clean interface abstraction
- Input validation for job names with `isValidJobName()` helper function
  - Restricts job names to alphanumeric characters, dashes, underscores, and dots
  - Maximum length validation (100 characters)
  - Prevents path traversal and injection attacks
- Tool name whitelisting to prevent unauthorized MCP tool execution
- Build number validation with type checking and range validation
- Context cancellation support for graceful shutdown
- Memory leak prevention with conversation history limiting (max 50 messages)
- System message preservation during history trimming
- Comprehensive production readiness TODO list in README
- Security warnings and best practices in documentation
- Troubleshooting section in README
- Known limitations documentation
- **Makefile** with common development and deployment commands
- **.env.example** template for environment configuration
- **CHANGELOG.md** for tracking version history

### Changed
- Improved context management with `context.WithCancel()` instead of plain `context.Background()`
- Enhanced parameter validation for `analyze_logs` tool
- Refactored history trimming logic to execute after message append
- Updated README with clearer setup instructions and security considerations
- Improved error messages for invalid inputs

### Fixed
- **Security:** Command injection vulnerability through unvalidated job names
- **Security:** Replaced dynamic content with static messages to prevent log injection
- **Security:** Type assertion panic from missing parameter validation
- **Memory:** Unbounded conversation history growth causing memory leaks
- **Performance:** Unnecessary map allocations in hot loop path
- **Reliability:** Context not properly cancellable, preventing graceful shutdown
- Race condition in context usage (creating new context per iteration)

### Security
- Use static messages without including response data in conversation history
- Added input sanitization for all user-controlled parameters
- Implemented tool execution whitelisting
- Added parameter type validation to prevent runtime panics
- Prevented command injection via job name validation
- Improved resource management with bounded history

## [0.2.0] - 2025-09-14

### Added

- `analyze_logs` capability for AI-powered log analysis
- Helper function that combines log fetching with LLM analysis
- Support for troubleshooting build failures with AI assistance

## [0.1.0] - 2025-09-13

### Added
- MCP client-server stack implementation
- Jenkins integration via API token
- Single LLM support - OpenAI GPT-4
- Basic MCP tools: `trigger_job`, `get_build_status`, `get_console_log`
- Docker/Podman containerization
- Initial documentation and setup instructions

---

## Migration Guide

### Upgrading from 0.2.0 to 0.3.0

**Breaking Changes:**
- Job names now validated and must contain only alphanumeric characters, dashes, underscores, and dots
- Maximum job name length is 100 characters

**Action Required:**
1. Verify your Jenkins job names comply with the new validation rules
2. Update any job names that contain invalid characters
3. If you have custom MCP tools, add them to the `allowedTools` whitelist

**Example:**
```go
// Old job names that will now fail:
"my job"        // Contains space
"job/path"      // Contains slash
"job@special"   // Contains @

// Valid job names:
"my-job"
"my_job"
"my.job"
"job123"
```

---

## Upgrade Notes

N/A

## Contributors

- StackBalancer

---

## License

This project is licensed under the GNU General Public License v3.0. See [LICENSE](LICENSE) for details.