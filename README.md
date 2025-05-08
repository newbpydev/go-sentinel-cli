# Go Sentinel

Go Sentinel is an open source, Go-native CLI utility that supercharges your test-driven development (TDD) workflow by automatically watching your Go source files, rerunning tests on changes, and presenting concise, actionable feedback in your terminal. 

## What Problem Does Go Sentinel Solve?

Manual test execution slows down TDD and feedback loops. Existing tools are often language-agnostic, slow, or lack Go-native ergonomics. Go Sentinel solves this by:
- **Automatically detecting file changes** in your Go project (excluding vendor and generated files)
- **Debouncing rapid events** to avoid redundant test runs
- **Running `go test -json`** per package for accurate, real-time results
- **Parsing and summarizing test output** with color-coded, human-friendly CLI feedback
- **Providing keyboard shortcuts** for rerun, filtering failures, and quitting
- **Ensuring stability** with robust error handling and structured logging

## Key Features
- Fast, recursive file watching using [fsnotify](https://github.com/fsnotify/fsnotify)
- Debounced test execution per package
- Real-time, colored summary of test results
- Minimal, intuitive keybindings (Enter: rerun, f: filter failures, q: quit)
- Structured logging (Uber's zap)
- Robust: recovers from panics, keeps the watcher alive
- Configurable debounce interval, color, verbosity, and package scope

## Project Structure

```
/go-sentinel
│
├── docs/
│   ├── RESEARCH.md           # High-level & detailed design, rationale, technical notes
│   └── assets/               # Architecture diagrams and images
│
├── (src/)                    # (To be created) Main Go source code for CLI tool
│
└── README.md                 # This file
```

> **Note:** The core Go implementation is planned; see `docs/RESEARCH.md` for architectural details and rationale.

## Usage (Planned)
- Install via Go or download binary (instructions coming soon)
- Run in your project root: `go-sentinel`
- Use keybindings to control test runs interactively
- Configure via flags or config file (YAML/JSON)

## Open Source & Contribution Guidelines
Go Sentinel is an open source project and welcomes contributions! To get involved:
- **Read the [docs/RESEARCH.md](docs/RESEARCH.md)** for design context and technical approach
- **Open issues** for bugs, feature requests, or questions
- **Submit pull requests** with clear descriptions and tests
- **Follow Go best practices** for code style and documentation
- **Add/update tests** for any code changes
- **Document public flags and configuration**
- **Use semantic versioning** ([semver.org](https://semver.org/))

## Best Practices (from the Community)
- Always provide `--help` and `--version` flags
- Document all public flags and configuration files
- Use clear, predictable exit codes
- Prefer human-friendly output by default, but support machine-readable formats (e.g., JSON)
- Keep output concise and support color toggling for accessibility
- Support easy updates (planned: `go-sentinel update`)
- Follow [CLI best practices](https://github.com/arturtamborski/cli-best-practices)

## License
MIT License. See [LICENSE](LICENSE) for details.

## Acknowledgments
- Inspired by Go TDD workflows and community feedback
- Uses [fsnotify](https://github.com/fsnotify/fsnotify), [zap](https://github.com/uber-go/zap), and other open source libraries

## Roadmap
- [ ] Implement core CLI watcher and debounce logic
- [ ] Integrate test runner and output parser
- [ ] Build interactive CLI UI
- [ ] Add configuration and extensibility
- [ ] Release initial stable version

---

For detailed design, see [`docs/RESEARCH.md`](docs/RESEARCH.md). Contributions and feedback are welcome!
