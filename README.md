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

**Prerequisites**: Docker, Docker Compose

```bash
# Start all infrastructure (PostgreSQL 18 + NATS)
docker-compose up -d

# Check service health
docker-compose ps

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

For detailed setup and development commands, see [CLAUDE.md](./CLAUDE.md) and [QUICKSTART.md](./docs/QUICKSTART.md).

## 📚 Documentation

**Core Documentation**:
- **[PLANNING.md](./PLANNING.md)** - Milestone tracking and project status (M0 ✅, M1 ✅)
- **[EVENTS.md](./EVENTS.md)** - Complete event catalog (25 events with schemas and flow diagrams)
- **[ARCHITECTURE.md](./docs/ARCHITECTURE.md)** - System architecture with mermaid diagrams
- **[STRUCTURE.md](./docs/STRUCTURE.md)** - Go project structure and development workflow
- **[QUICKSTART.md](./docs/QUICKSTART.md)** - Quick start guide for development setup
- **[CLAUDE.md](./CLAUDE.md)** - Guidance for Claude Code (development commands, patterns)
- **[CONTRIBUTING.md](./CONTRIBUTING.md)** - **Development philosophy and TDD workflow**

**Development Guides**:
- **[CONTRIBUTING.md](./CONTRIBUTING.md)** - Development philosophy and TDD workflow (Kent Beck style) - **READ FIRST**
- **[Domain Guide](./.claude/skills/BATTERY-DOMAIN-SKILL.md)** - Battery domain concepts and validation rules

**M2 Milestone Documentation**:
- **[M2 Overview](./docs/milestones/M2-OVERVIEW.md)** - Big picture and learning objectives
- **[M2 Domain Spec](./docs/milestones/M2-DOMAIN-SPEC.md)** - Battery aggregate and validation rules
- **[M2 API Spec](./docs/milestones/M2-API-SPEC.md)** - REST endpoints and DTOs
- **[M2 Checklist](./docs/milestones/M2-CHECKLIST.md)** - Step-by-step implementation guide

## 🎓 Learning Focus

This project prioritizes:
1. High-level architectural patterns over language-specific syntax
2. Understanding trade-offs in distributed systems
3. Modern engineering practices (TDD, CD, observability)
4. Domain-driven design in a real-world context

## 🏢 Inspiration

Inspired by battery optimization platforms in the Australian energy market, such as Hachiko and Evergen.

## 📝 Current Status

**✅ Completed Milestones**:
- **M0**: Project Setup - Infrastructure running (Docker Compose, PostgreSQL 18, NATS)
- **M1**: Event Storming - 25 events documented with schemas and flow diagrams
- **M2**: Asset Management Service - Production-ready REST API with TDD and Hexagonal Architecture
  - 79.8% test coverage (domain: 96.7%)
  - 3 RESTful endpoints (POST, GET, LIST)
  - PostgreSQL persistence with automatic migrations
  - Docker deployment ready
  - Comprehensive documentation

**🚧 Next Up**: M3 (Market Data Service)

See [PLANNING.md](./PLANNING.md) for detailed milestone tracking and next steps.

## 🔗 Tech Stack

- **Language**: Go 1.21+
- **Event Bus**: NATS 2.10
- **Database**: PostgreSQL 18 (DB-per-service: 3 independent instances)
- **Containers**: Docker & Docker Compose
- **Testing**: Go testing package with table-driven tests (TDD approach)

## 🎯 Business Value Alignment

### 1. Solid Optimization Fundamentals (Current Focus)
**Foundation First Approach**: Before bankability or fast deployment can deliver value, the optimization engine must work correctly.

This project builds:
- Accurate battery state management and constraints
- Real-time market data integration
- Optimal charging/discharging decisions
- FCAS (Frequency Control Ancillary Services) support
- Event-driven architecture for scalability

**Why this matters**: Predictable revenue (bankability) is only possible with consistent, reliable optimization.

---

### 2. Future Enhancement: Bankability Layer

Once optimization fundamentals are solid, the next phase would add financial predictability features:

#### Revenue Predictability
- Historical revenue tracking per operation
- Multi-scenario forecasting (conservative, expected, optimistic)
- Confidence intervals for investor reporting
- Revenue volatility analysis

#### Investment Metrics
- ROI, IRR, Payback period calculations
- Risk-adjusted returns
- Expected vs. actual performance tracking

#### Lender Dashboard
- Standardized financial reporting for banks/investors
- Compliance documentation
- Real-time performance vs. forecast
- "Bankability score" calculation

**Business Impact**: Makes C&I batteries investable by providing lenders with predictable revenue projections backed by proven optimization performance.

---

### 3. Future Enhancement: Fast Deployment via AI Simulation

#### The Pre-Sales Problem
Traditional battery project sales require:
- Manual site modeling (weeks)
- Custom revenue projections
- Multiple sales meetings
- High touch sales process

This limits deployment velocity.

#### The Solution: Public AI-Driven Simulation Tool
A self-service tool that transforms the sales pipeline:

**User Flow:**
1. **Upload historical data** (CSV from any battery/solar monitor)
2. **AI normalizes data** (handles different formats, fills gaps)
3. **Run optimization** (same engine as live operations)
4. **See results instantly**: "You could have earned $45,000 more last year"

**Architecture Advantage:**
This project's **core optimization engine** is designed to be data-source agnostic:
```
┌─────────────────────────────┐
│ Simulation Interface        │ ← Future: CSV upload
├─────────────────────────────┤
│ Live Operation Interface    │ ← Current: Real-time APIs
├─────────────────────────────┤
│ Optimization Engine (CORE)  │ ← This project
│ • Market analysis           │
│ • Constraint validation     │
│ • Revenue calculation       │
│ • FCAS integration          │
└─────────────────────────────┘
```

**Dual Usage:**
- **Simulation Mode**: Historical data → "What if" analysis (pre-sales)
- **Live Mode**: Real-time telemetry → Actual optimization (production)

**Business Impact:**
- **Lead Magnet**: Prospects engage immediately with value
- **Sales Efficiency**: Automate pre-sales modeling
- **Trust Through Transparency**: Show actual historical performance
- **Scalability**: Small team handles thousands of sites

#### AI Translation Layer (Future)
- **Data Preprocessing**: Normalize CSV columns from different OEMs
- **Gap Filling**: Use historical weather/market data to complete datasets
- **Quality Detection**: Flag poor quality data before simulation
- **Privacy**: De-identification and secure upload protocols

---

## 🏗️ Design Philosophy

**Build the Foundation Right:**
This project focuses on creating a robust optimization engine that:
1. **Works correctly** (accurate decisions)
2. **Scales efficiently** (handles multiple sites)
3. **Remains flexible** (adapts to different use cases)

Once this foundation is solid:
- **Bankability features** add financial predictability on top
- **Simulation tools** reuse the same optimization logic for pre-sales
- **Multi-OEM support** (via hardware abstraction layer) enables universal deployment

**Key Insight**: The optimization engine is the "brain" - whether analyzing historical data (simulation) or controlling live batteries (production), the core logic remains the same. Build it once, use it everywhere.
