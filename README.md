# Go RPC Generator

A CLI tool to generate Go gRPC code from proto files, with built-in version management for generator plugins.

## Prerequisites

- Go 1.20+
- `protoc` compiler installed and in your PATH.

## Installation

```bash
git clone <repo>
cd go-rpc-gen
go install
```

## Usage

```bash
go-rpc-gen -proto-dir=path/to/protos -out-dir=path/to/output
```

### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-proto-dir` | Directory containing .proto files | `.` |
| `-out-dir` | Output directory for generated files | `.` |
| `-go-version` | Version of `protoc-gen-go` | `latest` |
| `-grpc-version` | Version of `protoc-gen-go-grpc` | `latest` |
