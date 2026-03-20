# Contributing to hocon2

## Development Setup

```bash
git clone https://github.com/o3co/hocon2.git
cd hocon2
make all
```

### Requirements

- Go 1.25+
- [golangci-lint](https://golangci-lint.run/welcome/install/)

## Branch Strategy

- `master` — release branch (protected: PR + CI required)
- `develop` — default work branch

## Workflow

1. Create a feature branch from `develop`
2. Make changes with tests
3. Run `make all` to verify
4. Open a PR to `develop`

## Commit Style

Use [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` new feature
- `fix:` bug fix
- `test:` test changes
- `docs:` documentation
- `chore:` maintenance
- `refactor:` code restructuring

## Testing

```bash
make test           # run all tests
go test ./... -v    # verbose output
```

### Golden Tests

Test data lives in `testdata/`. Each `.hocon` file has corresponding output files (`.json`, `.yaml`, `.toml`, `.properties`). To add a test case, create the input and all expected outputs.
