# AegisBox — Secure Ephemeral Code Execution Engine

AegisBox is an isolated, ephemeral code execution engine designed for educational and platform engineering exploration. Conceptualized after modern online judges, automated grading systems, and sandboxed cloud compute engines, AegisBox leverages native Linux kernel primitives to safely execute untrusted code.

---

## 1. Project Goals

1. **Native Linux Isolation**: Understand and implement containerization fundamentals from scratch without Docker (PID/Mount/Network/User namespaces, Cgroups v2, Seccomp, and ephemeral root filesystems).
2. **Deterministic Resource Enforcement**: Prevent rogue executions, memory leaks, fork bombs, and CPU hogging using granular Linux cgroups v2.
3. **Defense in Depth**: Layered containment ensuring the sandboxed process perceives itself as an isolated guest with strictly zero host filesystem or network access.
4. **Observable & Extensible**: Modular runtime design supporting multiple languages (starting with Python 3), structured JSON telemetry, and RESTful execution lifecycle APIs.
5. **Engineering Rigor**: Production-ready project structure, strict validation, comprehensive testing, and automated CI/CD.

---

## 2. Current Project State: Milestone 1

> **Current Milestone**: **Milestone 1 — Project Foundation & Domain Contracts**

In Milestone 1, we establish the core project layout, domain types, configuration system, and automated CI pipeline. No sandbox primitives are executed yet; this foundation guarantees architectural stability for subsequent isolation layers.

### Project Layout

```text
aegisbox/
├── .github/
│   └── workflows/
│       └── ci.yml          # GitHub Actions CI (gofmt, vet, test, build)
├── cmd/
│   └── aegisbox/
│       └── main.go         # Application CLI entrypoint & health check
├── configs/
│   └── config.yaml         # Default runtime and server configuration
├── internal/
│   ├── api/                # REST API routes & handlers (Milestone 8)
│   ├── config/             # Configuration parsing, validation & environment loading
│   ├── executor/           # Domain contracts (ExecutionRequest, ExecutionResult, Status)
│   ├── runtime/            # Runtime adapters (Python, etc.) (Milestone 2)
│   └── sandbox/            # Namespaces, cgroups v2, rootfs, and seccomp (Milestones 3-6)
├── scripts/                # Rootfs generation and operational scripts
├── tests/                  # Integration and foundation test suite
├── .gitignore              # Git ignore rules for Go and build outputs
├── go.mod                  # Go module definition
├── Makefile                # Multi-platform build and test targets
└── README.md               # Project documentation
```

---

## 3. Development & Runtime Architecture

AegisBox is built using a **dual-environment development model**:

```text
[ Windows Development Workstation ]
  ├── VS Code / PowerShell
  ├── Go Toolchain & Git
  └── Local Verification / Unit Tests
            │
            ▼ (git push)
[ GitHub Repository & Actions CI ]
  ├── gofmt verification
  ├── go vet static analysis
  ├── go test -v -race (Linux runner)
  └── binary compilation & smoke checks
            │
            ▼ (deploy)
[ Ubuntu Linux Runtime Target ]
  ├── Linux Kernel (cgroups v2, namespaces, seccomp)
  ├── AegisBox Systemd Service
  └── Ephemeral Sandboxed User Processes
```

- **Development**: Windows with Go, Git, and VS Code for authoring and unit-testing core logic.
- **Runtime Host**: Ubuntu Linux machine where Linux kernel mechanisms (`unshare`, `cgroups v2`, `seccomp`, `pivot_root`) are natively available.

---

## 4. Domain Contracts

### Execution Request

```json
{
  "language": "python",
  "code": "print('Hello, AegisBox!')",
  "timeout_ms": 1000,
  "max_mem_mb": 64,
  "max_processes": 10
}
```

### Execution Result

```json
{
  "execution_id": "exec-9f8a3c",
  "status": "COMPLETED",
  "stdout": "Hello, AegisBox!\n",
  "stderr": "",
  "exit_code": 0,
  "execution_time_ms": 18,
  "memory_usage_bytes": 14680064
}
```

### Supported Execution Statuses

- `COMPLETED`: Normal termination with exit code 0 or handled script exit.
- `RUNTIME_ERROR`: Script encountered an unhandled exception or non-zero exit code.
- `TIME_LIMIT_EXCEEDED`: Wall-clock execution exceeded configured timeout.
- `OOM_KILLED`: Exceeded memory allocation and terminated by kernel cgroup OOM killer.
- `PROCESS_LIMIT_EXCEEDED`: Process/thread creation exceeded `pids.max`.
- `START_ERROR`: Failure to initialize the sandbox environment.
- `SANDBOX_ERROR`: Internal containment error or cleanup failure.
- `UNSUPPORTED_LANGUAGE`: Requested runtime is not registered.

---

## 5. Local Development & Testing

### Building Locally

```bash
# Build the binary
go build -v -o bin/aegisbox ./cmd/aegisbox

# Run health diagnostics
./bin/aegisbox -health

# Check version
./bin/aegisbox -version
```

### Running Tests

```bash
# Run all unit and foundation tests
go test -v ./...

# Run tests with race detection (Linux / WSL)
go test -v -race ./...

# Run static analysis
go vet ./...

# Verify code formatting
gofmt -l .
```

Using `make` (if Make is installed):

```bash
make fmt-check
make vet
make test
make build
```

---

## 6. Continuous Integration (CI)

Every push and pull request triggers `.github/workflows/ci.yml`:
1. **Formatting Check**: Verifies all code complies with `gofmt`.
2. **Static Analysis**: Runs `go vet` to catch potential bugs and structural errors.
3. **Unit Tests**: Runs the test suite with `-race` enabled on a native Linux runner.
4. **Compilation**: Compiles the `aegisbox` executable.
5. **Smoke Execution**: Executes `./bin/aegisbox -version` and `./bin/aegisbox -health` to confirm binary functionality.

---

## 7. Roadmap

- [x] **Milestone 1**: Project Foundation & Domain Contracts
- [ ] **Milestone 2**: Ephemeral Rootfs & Runtime Abstraction
- [ ] **Milestone 3**: Process Isolation & Linux Namespaces
- [ ] **Milestone 4**: Cgroups v2 Resource Controller
- [ ] **Milestone 5**: Filesystem Isolation & Ephemeral Workspaces
- [ ] **Milestone 6**: Syscall Defense with Seccomp
- [ ] **Milestone 7**: Execution Engine & Process Guardian
- [ ] **Milestone 8**: REST API & Observability
- [ ] **Milestone 9**: End-to-End Testing & Security Verification
- [ ] **Milestone 10**: Linux Systemd Deployment & CI/CD Pipeline
