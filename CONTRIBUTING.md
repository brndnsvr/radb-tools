# Contributing to RADb API Client

Thank you for your interest in contributing to the RADb API Client! This document provides guidelines and information for contributors.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [How to Contribute](#how-to-contribute)
- [Development Process](#development-process)
- [Pull Request Process](#pull-request-process)
- [Coding Standards](#coding-standards)
- [Testing Requirements](#testing-requirements)
- [Documentation](#documentation)
- [Community](#community)

## Code of Conduct

### Our Pledge

We are committed to providing a welcoming and inclusive environment for all contributors, regardless of experience level, gender identity, sexual orientation, disability, personal appearance, body size, race, ethnicity, age, religion, or nationality.

### Our Standards

**Positive behavior includes:**
- Using welcoming and inclusive language
- Being respectful of differing viewpoints
- Gracefully accepting constructive criticism
- Focusing on what is best for the community
- Showing empathy towards other community members

**Unacceptable behavior includes:**
- Harassment, trolling, or discriminatory comments
- Personal or political attacks
- Publishing others' private information
- Other conduct which could reasonably be considered inappropriate

### Enforcement

Instances of unacceptable behavior may be reported to the project team. All complaints will be reviewed and investigated promptly and fairly.

## Getting Started

### Prerequisites

- Go 1.21 or later
- Git
- Basic understanding of Go programming
- Familiarity with command-line interfaces

### Development Setup

1. **Fork the repository** on GitHub

2. **Clone your fork:**
   ```bash
   git clone https://github.com/YOUR_USERNAME/radb-client.git
   cd radb-client
   ```

3. **Add upstream remote:**
   ```bash
   git remote add upstream https://github.com/example/radb-client.git
   ```

4. **Install dependencies:**
   ```bash
   make deps
   ```

5. **Build and test:**
   ```bash
   make build
   make test
   ```

## How to Contribute

### Ways to Contribute

**Code Contributions:**
- Bug fixes
- New features
- Performance improvements
- Code refactoring

**Non-Code Contributions:**
- Documentation improvements
- Bug reports
- Feature requests
- User support
- Testing and feedback

### Reporting Bugs

**Before submitting:**
- Check existing issues to avoid duplicates
- Verify the bug with the latest version
- Collect necessary information

**Bug report should include:**
1. **Title:** Clear, concise description
2. **Environment:** OS, version, Go version
3. **Steps to reproduce:** Detailed steps
4. **Expected behavior:** What should happen
5. **Actual behavior:** What actually happens
6. **Logs:** Debug output if available
7. **Additional context:** Any other relevant information

**Example bug report:**
```markdown
## Bug: Route create fails with IPv6 prefix

**Environment:**
- OS: Ubuntu 22.04
- Version: radb-client 1.0.0
- Go version: 1.21.0

**Steps to reproduce:**
1. Create file: `{"route":"2001:db8::/32","origin":"AS64500",...}`
2. Run: `radb-client route create route.json`

**Expected:** Route should be created
**Actual:** Error: "Invalid route format"

**Debug output:**
```
[attach debug output]
```
```

### Suggesting Features

**Before submitting:**
- Check if feature already requested
- Consider if it fits project scope
- Think about implementation

**Feature request should include:**
1. **Title:** Clear feature description
2. **Problem:** What problem does it solve?
3. **Proposed solution:** How should it work?
4. **Alternatives:** Other approaches considered
5. **Use cases:** Real-world scenarios
6. **Additional context:** Any other information

**Example feature request:**
```markdown
## Feature: Support for route6 objects

**Problem:**
Currently, IPv6 routes must be created manually through the web interface.

**Proposed Solution:**
Add `route6` commands similar to `route` commands.

**Use Cases:**
- Manage IPv6 allocations programmatically
- Automate IPv6 route updates
- Maintain consistency between IPv4 and IPv6 routes

**Additional Context:**
RADb API already supports route6 objects via /RADB/route6 endpoint.
```

## Development Process

### Branching Strategy

**Main branches:**
- `main` - Stable, production-ready code
- `develop` - Integration branch for features (if used)

**Feature branches:**
- `feature/feature-name` - New features
- `fix/bug-description` - Bug fixes
- `docs/improvement` - Documentation
- `refactor/component` - Code refactoring
- `test/test-description` - Tests only

### Workflow

1. **Update your fork:**
   ```bash
   git checkout main
   git fetch upstream
   git merge upstream/main
   git push origin main
   ```

2. **Create feature branch:**
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Make changes:**
   ```bash
   # Write code
   vim internal/api/client.go

   # Add tests
   vim internal/api/client_test.go

   # Run tests
   make test

   # Format code
   make fmt

   # Lint
   make lint
   ```

4. **Commit changes:**
   ```bash
   git add .
   git commit -m "Add support for IPv6 route validation"
   ```

5. **Push to your fork:**
   ```bash
   git push origin feature/your-feature-name
   ```

6. **Create pull request** on GitHub

### Commit Messages

**Format:**
```
<type>: <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `style`: Formatting, no code change
- `refactor`: Code refactoring
- `test`: Adding tests
- `chore`: Maintenance

**Example:**
```
feat: Add IPv6 route validation

Implements validation for IPv6 routes with proper CIDR
notation checking. Validates against RFC 4291 format.

- Add validateIPv6CIDR function
- Add unit tests for edge cases
- Update documentation

Fixes #123
```

**Guidelines:**
- Use present tense: "Add feature" not "Added feature"
- First line: 50 characters or less
- Body: wrap at 72 characters
- Reference issues: "Fixes #123", "Closes #456"
- Explain what and why, not how

## Pull Request Process

### Before Submitting

**Checklist:**
- [ ] Code follows project style guidelines
- [ ] Tests added/updated and passing
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Commits are well-formed
- [ ] Branch is up to date with main

### Pull Request Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix (non-breaking change fixing an issue)
- [ ] New feature (non-breaking change adding functionality)
- [ ] Breaking change (fix or feature causing existing functionality to change)
- [ ] Documentation update

## Testing
Describe testing performed:
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing performed

## Checklist
- [ ] Code follows project style
- [ ] Self-review completed
- [ ] Comments added for complex code
- [ ] Documentation updated
- [ ] No new warnings generated
- [ ] Tests added and passing
- [ ] CHANGELOG.md updated

## Related Issues
Fixes #(issue number)

## Screenshots (if applicable)
Add screenshots for UI changes
```

### Review Process

**What reviewers look for:**
1. **Correctness:** Does it work as intended?
2. **Testing:** Adequate test coverage?
3. **Code quality:** Clear, maintainable code?
4. **Documentation:** Sufficient documentation?
5. **Style:** Follows project conventions?
6. **Performance:** No performance regressions?

**Response time:**
- Initial review: Within 3 business days
- Follow-up: Within 2 business days
- Approval: At least one maintainer approval required

### After Review

**Address feedback:**
```bash
# Make requested changes
vim internal/api/client.go

# Commit
git commit -m "Address review comments"

# Push
git push origin feature/your-feature-name
```

**When approved:**
- Squash commits if requested
- Maintainer will merge

## Coding Standards

### Go Style Guide

Follow official Go style:
- [Effective Go](https://go.dev/doc/effective_go)
- [Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

### Project Conventions

**Naming:**
```go
// Good
func GetRoute(prefix string) (*Route, error)
func validateIPv4CIDR(cidr string) bool

// Bad
func get_route(prefix string) (*Route, error)
func ValidateIPV4CIDR(cidr string) bool
```

**Error Handling:**
```go
// Good
if err != nil {
    return fmt.Errorf("failed to create route: %w", err)
}

// Bad
if err != nil {
    return err  // No context
}
```

**Comments:**
```go
// Good
// GetRoute retrieves a route object by prefix.
// Returns ErrNotFound if the route doesn't exist.
func GetRoute(prefix string) (*Route, error) {

// Bad
// get route
func GetRoute(prefix string) (*Route, error) {
```

### Code Formatting

```bash
# Format code
make fmt

# Or manually
gofmt -w .
goimports -w .
```

### Linting

```bash
# Run linter
make lint

# Or manually
golangci-lint run ./...

# Fix auto-fixable issues
golangci-lint run --fix ./...
```

## Testing Requirements

### Test Coverage

**Minimum requirements:**
- Overall coverage: > 80%
- New features: > 90%
- Bug fixes: 100% of affected code

**Check coverage:**
```bash
make test-coverage
go tool cover -html=coverage.txt
```

### Writing Tests

**Unit tests:**
```go
func TestManager_SaveSnapshot(t *testing.T) {
    // Arrange
    tmpDir := t.TempDir()
    manager, err := NewManager(tmpDir)
    require.NoError(t, err)

    testData := map[string]string{"key": "value"}

    // Act
    err = manager.SaveSnapshot("test", testData)

    // Assert
    require.NoError(t, err)
    assert.FileExists(t, filepath.Join(tmpDir, "test.json"))
}
```

**Table-driven tests:**
```go
func TestValidateIPv4(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    bool
        wantErr bool
    }{
        {"valid", "192.0.2.0/24", true, false},
        {"invalid_cidr", "192.0.2.0/33", false, true},
        {"no_cidr", "192.0.2.0", false, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ValidateIPv4(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.want, got)
            }
        })
    }
}
```

### Integration Tests

```go
// +build integration

func TestIntegration_RouteOperations(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    client := setupTestClient(t)
    // ... integration test logic
}
```

**Run integration tests:**
```bash
go test -tags=integration ./...
```

## Documentation

### Documentation Requirements

**All contributions should include:**
- Code comments (GoDoc format)
- README updates (if needed)
- User guide updates (if needed)
- API documentation updates
- CHANGELOG.md entry

### Writing Documentation

**Documentation style:**
- Clear and concise
- Examples for complex features
- Cross-references to related docs
- Keep up to date with code

**GoDoc format:**
```go
// Package api provides a client for the RADb API.
package api

// Client represents an RADb API client.
type Client struct {
    // ...
}

// NewClient creates a new RADb API client.
//
// Example:
//   client := api.NewClient("https://api.radb.net", "RADB", username, apiKey)
func NewClient(baseURL, source, username, apiKey string) *Client {
    // ...
}
```

### Updating CHANGELOG

**Format:**
```markdown
## [Unreleased]

### Added
- New feature X (#123)

### Changed
- Modified behavior Y (#124)

### Fixed
- Bug Z (#125)

### Deprecated
- Feature A (will be removed in v2.0)

### Removed
- Old feature B

### Security
- Security improvement C
```

## Community

### Communication Channels

**GitHub:**
- Issues: Bug reports and feature requests
- Discussions: Questions and general discussions
- Pull Requests: Code contributions

**Getting Help:**
- Check documentation first
- Search existing issues
- Ask in Discussions
- Be patient and respectful

### Recognition

**Contributors will be:**
- Listed in CONTRIBUTORS file
- Mentioned in release notes
- Thanked in project documentation

**Significant contributors may:**
- Become project maintainers
- Get commit access
- Shape project direction

### License

By contributing, you agree that your contributions will be licensed under the same license as the project (MIT License).

## Questions?

If you have questions about contributing:
1. Check this guide
2. Review existing issues and PRs
3. Ask in GitHub Discussions
4. Contact maintainers

Thank you for contributing to the RADb API Client!
