# Contributing to Atlas

Thanks for your interest in contributing! Please read this guide before opening issues or submitting pull requests.

## Project Lead

**[@Atlas51215](https://github.com/Atlas51215)** is the project owner and lead. All significant design decisions go through them. When in doubt, open an issue and ask before investing time in a large change.

## Getting Started

1. Fork the repository and clone your fork.
2. Create a feature branch from `main`:
   ```bash
   git checkout -b feat/your-feature-name
   ```
3. Make your changes, write tests, and verify everything passes:
   ```bash
   go test -v ./...
   ```
4. Open a pull request against `main`.

## How to Report a Bug

- Search existing issues first to avoid duplicates.
- Open a new issue with a clear title and a description that includes:
  - What you expected to happen
  - What actually happened
  - Steps to reproduce
  - Go version and OS

## How to Suggest a Feature

- Open an issue describing the feature and the problem it solves.
- Wait for feedback from [@Atlas51215](https://github.com/Atlas51215) before starting implementation. This avoids wasted effort on changes that won't be accepted.

## Pull Request Guidelines

- Keep PRs focused — one feature or fix per PR.
- Write or update tests to cover your changes.
- All tests must pass (`go test -v ./...`) before requesting review.
- Follow existing code style and conventions (see `CLAUDE.md` for stack details).
- Write a clear PR description explaining what changed and why.
- The project lead has final say on merging.

## Commit Messages

Use short, imperative-style commit messages:
```
add category query functions
fix migration runner skipping applied files
update home handler to return JSON
```

## Code of Conduct

Be respectful and constructive. Harassment or hostile behavior will result in removal from the project.
