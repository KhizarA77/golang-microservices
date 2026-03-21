# Go Lang Practice Labs

A structured, hands-on lab series covering Go concurrency, design patterns, backend development, and microservices — progressing from beginner to expert.

---

## Learning Path

| Lab | Directory | Topic | Level |
|-----|-----------|-------|-------|
| 01 | `lab01-goroutines-channels` | Goroutines & Channels | Beginner |
| 02 | `lab02-concurrency-patterns` | Concurrency Patterns | Intermediate |
| 03 | `lab03-worker-pools` | Worker Pools & Pipelines | Advanced |
| 04 | `lab04-creational-patterns` | Creational Design Patterns | Beginner |
| 05 | `lab05-structural-patterns` | Structural Design Patterns | Intermediate |
| 06 | `lab06-behavioral-patterns` | Behavioral Design Patterns | Advanced |
| 07 | `lab07-http-server` | HTTP Server Basics | Beginner |
| 08 | `lab08-rest-api` | REST API Design | Intermediate |
| 09 | `lab09-clean-architecture` | Clean Architecture | Advanced |
| 10 | `lab10-service-communication` | Service-to-Service Communication | Intermediate |
| 11 | `lab11-grpc-services` | gRPC & Protocol Buffers | Advanced |
| 12 | `lab12-event-driven` | Event-Driven Architecture | Advanced |
| 13 | `lab13-capstone` | Capstone: Microservices Platform | Expert |

---

## Prerequisites

- Go 1.21+ installed (`go version` to check)
- Basic Go knowledge: variables, functions, structs, interfaces, error handling
- A code editor — VS Code with the Go extension is recommended

---

## How to Use These Labs

Each lab is a self-contained Go module. The workflow for every lab:

1. **Read the README.md** in the lab directory thoroughly
2. **Study the background** — understand the concepts before coding
3. **Open the starter `.go` files** — read the existing code and `TODO` comments
4. **Implement the TODOs** — fill in the logic yourself
5. **Run and test** your implementation
6. **Move to the next lab** once things work

```bash
cd lab01-goroutines-channels
go mod tidy
go run .
```

---

## Lab Themes

### Concurrency (Labs 01–03)
Go's concurrency model is built around goroutines and channels — lightweight, composable primitives. These labs take you from launching goroutines to building production-grade concurrent systems.

### Design Patterns (Labs 04–06)
Classic GoF patterns adapted to Go's idioms. Go doesn't have classes but has interfaces, composition, and first-class functions — which lead to elegant pattern implementations.

### Backend Development (Labs 07–09)
Building HTTP servers and REST APIs using Go's standard library. Clean architecture separates concerns into layers so code stays testable and maintainable.

### Microservices (Labs 10–13)
Splitting functionality into independent services that communicate over the network. Each service owns its data, scales independently, and the capstone ties everything together.

---

## Recommended Study Order

Follow the labs in numbered order. Concepts build on each other:
- Lab 01–03 before Lab 13 (capstone uses concurrency heavily)
- Lab 07–09 before Lab 10–11 (service comms assumes HTTP knowledge)
- Lab 04–06 can be done alongside Labs 07–09

---

## Tools You'll Need

```bash
# Verify Go is installed
go version

# For gRPC labs (Lab 11)
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# protoc compiler - install from https://grpc.io/docs/protoc-installation/
```

Good luck — Go is worth mastering.
