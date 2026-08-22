# AegisBox

AegisBox is a learning project for building a secure ephemeral code execution engine from Linux primitives.

The goal is to understand how container and sandbox technologies work internally, rather than simply relying on Docker as an abstraction.

## Project Goal

AegisBox accepts untrusted source code, executes it inside an isolated environment, collects the execution result, and destroys the environment after execution.

Conceptually:

Client
  ↓
Execution Engine
  ↓
Linux Isolation
  ↓
Code Execution
  ↓
Result Collection
  ↓
Sandbox Cleanup

## What I'm Learning

- Linux processes
- Process IDs and parent/child processes
- Linux namespaces
- Mount namespaces
- PID namespaces
- User namespaces
- Network namespaces
- Linux cgroups v2
- Filesystem isolation
- read-only root filesystems
- tmpfs
- seccomp
- Linux capabilities
- Process resource limits
- Go process management
- REST API design
- Worker architecture
- Observability

## Current Progress

### Phase 1 — Process Execution

- [x] Create Go module
- [x] Create basic execution engine
- [x] Execute Python from Go
- [x] Capture stdout
- [x] Capture stderr
- [x] Observe parent/child process relationship
- [ ] PID namespace
- [ ] Mount namespace
- [ ] User namespace
- [ ] Network namespace
- [ ] cgroups v2
- [ ] seccomp
- [ ] Resource limits
- [ ] REST API
- [ ] Worker queue
- [ ] Observability

## Architecture

The final architecture is planned to look like:

Client
  ↓
REST API
  ↓
Job Queue
  ↓
Execution Worker
  ↓
Sandbox
  ├── PID Namespace
  ├── Mount Namespace
  ├── User Namespace
  ├── Network Namespace
  ├── cgroups v2
  └── seccomp
  ↓
Execution Result

## Status

🚧 Work in Progress

This repository documents my learning process while building the execution engine from Linux primitives.
