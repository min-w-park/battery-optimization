# Battery Optimization System

A microservices-based battery optimization platform for energy market participation, built with Go and event-driven architecture.

## 🎯 Project Goal

This project demonstrates modern software engineering practices applied to battery energy storage system (BESS) optimization:

- **Domain-Driven Design (DDD)**: Clear bounded contexts and service boundaries
- **Event-Driven Architecture**: Asynchronous communication with eventual consistency
- **Test-Driven Development (TDD)**: Tests written first, driving implementation
- **Microservices**: Independent deployability with DB-per-service pattern
- **Modern Practices**: Continuous delivery mindset, observability, fast feedback

## 🏗️ Architecture Overview

The system consists of several microservices:

1. **Asset Management Service** - Battery specifications and constraints
2. **Market Data Service** - Energy market pricing data
3. **Telemetry Service** - Real-time battery state monitoring
4. **Device Interface Service** - Hardware abstraction layer
5. **Bidding Service** - Real-time bidding decisions based on market conditions

Services communicate via events using NATS, maintaining loose coupling and independent deployability.

## 🔋 Domain Context

This system simulates battery optimization for the Australian National Electricity Market (NEM), focusing on:

- Energy arbitrage (buy low, sell high)
- State of Charge (SoC) management
- Power and ramp rate constraints
- Real-time market participation

## 🚀 Quick Start

```bash
# Start all services
docker-compose up

# Run tests
make test

# View logs
docker-compose logs -f
```

## 📚 Documentation

- [Planning & Milestones](./PLANNING.md)
- [Domain Events](./EVENTS.md)
- [Architecture Decisions](./docs/architecture/)

## 🎓 Learning Focus

This project prioritizes:
1. High-level architectural patterns over language-specific syntax
2. Understanding trade-offs in distributed systems
3. Modern engineering practices (TDD, CD, observability)
4. Domain-driven design in a real-world context

## 🏢 Inspiration

Inspired by battery optimization platforms in the Australian energy market, such as Hachiko and Evergen.

## 📝 Current Status

See [PLANNING.md](./PLANNING.md) for current progress and next steps.

## 🔗 Tech Stack

- **Language**: Go
- **Event Bus**: NATS
- **Database**: PostgreSQL (per service)
- **Containers**: Docker & Docker Compose
- **Testing**: Go testing package with table-driven tests
