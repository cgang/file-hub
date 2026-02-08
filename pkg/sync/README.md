# Sync Protocol

This directory contains the Protocol Buffer definitions for the File Hub Sync Protocol, which is designed to complement the existing WebDAV implementation. The protocol is optimized for mobile devices to upload files to a home server using protobuf over HTTP/2.

## Files

- `sync.proto`: The main Protocol Buffer definition file containing all message types and service definitions for the sync protocol.
- `sync.pb.go`: Generated Go code from the protobuf definitions (not tracked in git).

## Features

- Chunk-based file transfers for handling unstable internet connections
- Version-based delta synchronization to minimize data transfer
- Efficient binary serialization using Protocol Buffers
- Optimized for mobile device usage patterns
- Lightweight authentication mechanisms
- Batch operations support

## Building

The protobuf Go code is automatically generated during the build process. When you run `make build`, the protobuf code will be generated if needed. Alternatively, you can manually generate the code with:

```bash
make gen-proto  # Note: This target was removed; generation now happens automatically during build
```

However, the preferred approach is to let the build system handle code generation automatically through Make's dependency mechanism.