# Contribution Guidelines

[![Contributions Welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](https://github.com/kittipat1413/go-common/issues)

Thank you for considering contributing to the Go-Common! This document provides guidelines for contributing to this project.

## How to Contribute

### 1. Fork and Clone
```bash
git clone https://github.com/YOUR_USERNAME/go-common.git
cd go-common
```

### 2. Create a Branch
```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/issue-description
```

### 3. Make Changes
- Follow existing code patterns and conventions
- Add comprehensive documentation for new features
- Include examples in `/example` directories where applicable
- Write tests for new functionality

### 4. Run Tests and Linting
```bash
# Run all tests
make test

# Generate coverage report
make test-coverage

# Run linting
make lint

# Generate mocks (if needed)
make generate-mock
```

### 5. Submit a Pull Request
- Ensure all tests pass and coverage is maintained
- Update relevant `README` files
- Include clear commit messages
- Reference any related issues

## Code Guidelines

### Documentation
- All public functions and types must have Go doc comments
- Follow the existing documentation style (concise but comprehensive)
- Include practical examples in comments
- Update `README` files for new packages

### Testing
- Maintain test coverage above 90%
- Write table-driven tests where appropriate
- Include both positive and negative test cases
- Use mock interfaces for testing

### Code Style
- Follow standard Go conventions
- Use meaningful variable and function names
- Keep functions focused and single-purpose
- Handle errors appropriately

## Package Structure
When adding new packages, follow the existing structure:
```
framework/
├── your-package/
│   ├── README.md            # Package documentation
│   ├── your_package.go      # Main implementation
│   ├── your_package_test.go # Tests
│   └── example/             # Usage examples
│       └── main.go
```

## Celebrate 🎉
Once your pull request is merged, take a moment to celebrate your contribution! Thank you for helping improve Go-Common.