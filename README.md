# net-cat

Simple TCP chat server written in Go.

## Overview

`net-cat` starts a TCP server that accepts multiple client connections and broadcasts chat messages between them. Each client enters a username when connecting, and all joined clients see chat history and live messages.

## Project Structure

- `main.go` - application entry point
- `internal/server/server.go` - server logic, connection handling, message broadcasting
- `internal/server/client.go` - client representation and message sending

## Requirements

- Go 1.26.2

## Run

```sh
go run main.go
```
or
```sh
go run main.go <port>
```
