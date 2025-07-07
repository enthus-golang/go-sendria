# Contributing to Sendria Go Client

Thank you for your interest in contributing to the Sendria Go client! This document provides guidelines and instructions for contributing to this project.

## Code of Conduct

By participating in this project, you agree to abide by our code of conduct. Please be respectful and professional in all interactions.

## How to Contribute

### Reporting Issues

- Check if the issue has already been reported
- Use the issue templates when available
- Provide clear and detailed information about the issue
- Include steps to reproduce the problem
- Mention your Go version and operating system

### Submitting Pull Requests

1. Fork the repository
2. Create a new branch for your feature or fix:
   ```bash
   git checkout -b feature/your-feature-name
   ```
3. Make your changes following the coding standards
4. Write or update tests as necessary
5. Ensure all tests pass:
   ```bash
   make test
   ```
6. Run linters:
   ```bash
   make lint
   ```
7. Commit your changes with a clear commit message
8. Push to your fork and submit a pull request

### Coding Standards

- Follow standard Go conventions and idioms
- Use `gofmt` and `goimports` to format your code
- Write clear, self-documenting code with comments where necessary
- Ensure all exported types, functions, and methods have documentation comments
- Keep functions small and focused on a single task
- Handle errors appropriately
- Write tests for new functionality

### Testing

- Write unit tests for all new functionality
- Ensure tests are clear and well-documented
- Use table-driven tests where appropriate
- Mock external dependencies appropriately
- Aim for high test coverage

### Documentation

- Update README.md if you're adding new features
- Add examples for new functionality
- Keep documentation clear and concise
- Update API documentation comments

## Development Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/enthus-golang/sendria.git
   cd sendria
   ```

2. Install dependencies:
   ```bash
   make deps
   ```

3. Install development tools:
   ```bash
   make tools
   ```

4. Run tests:
   ```bash
   make test
   ```

5. Run linters:
   ```bash
   make lint
   ```

## Pull Request Process

1. Update the README.md with details of changes if applicable
2. Update the examples if you've added new functionality
3. Ensure all tests pass and coverage is maintained
4. The PR will be merged once you have the sign-off of at least one maintainer

## Release Process

Releases are automated through GitHub Actions when a new tag is pushed. Only maintainers can create releases.

## Questions?

If you have questions, feel free to open an issue for discussion.