# Contributing to gosqlite

We welcome contributions to the `gosqlite` project! By contributing, you agree to abide by the [Code of Conduct](CODE_OF_CONDUCT.md).

## How to Contribute

### Reporting Bugs

If you find a bug, please open an issue on our issue tracker. Before opening a new issue, please search existing issues to see if your bug has already been reported.

When reporting a bug, please include:

*   A clear and concise description of the bug.
*   Steps to reproduce the behavior.
*   Expected behavior.
*   Actual behavior.
*   Any relevant error messages or stack traces.
*   Your operating system and Go version.

### Suggesting Enhancements

We welcome suggestions for new features or improvements. Please open an issue on our issue tracker to propose an enhancement. Describe the enhancement, why it would be useful, and any potential implementation details.

### Contributing Code

1.  **Fork the repository.**
2.  **Create a new branch** for your feature or bug fix: `git checkout -b feature/your-feature-name` or `git checkout -b bugfix/your-bug-fix`.
3.  **Make your changes.** Ensure your code adheres to the project's [Coding Standards](gemini.md#coding-standards).
4.  **Write tests** for your changes. All new features and bug fixes should have corresponding tests.
5.  **Run tests** to ensure everything passes: `go test ./...`.
6.  **Ensure `CGO_ENABLED=0 go test ./...` passes.** This verifies that your changes do not introduce any CGO or C tool-chain dependencies.
7.  **Update documentation** as necessary.
8.  **Commit your changes** with a clear and concise commit message.
9.  **Push your branch** to your forked repository.
10. **Open a Pull Request (PR)** to the `main` branch of the upstream repository. Provide a clear description of your changes and reference any related issues.

## Development Setup

To set up your development environment:

1.  **Clone the repository:** `git clone https://github.com/your-org/gosqlite.git`
2.  **Navigate to the project directory:** `cd gosqlite`
3.  **Run tests:** `go test ./...`

## Code Style

We follow standard Go formatting and style guidelines. Please run `go fmt ./...` and `go vet ./...` before submitting your code.

## Licensing

By contributing to `gosqlite`, you agree that your contributions will be licensed under the [Apache 2.0 License](LICENSE).
