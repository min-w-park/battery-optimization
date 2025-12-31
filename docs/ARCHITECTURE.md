# System Architecture

## High-Level Service Architecture

```mermaid
graph TB
    subgraph External
        AEMO[AEMO Market Operator]
        Operator[Human Operator]
        Battery[Battery Hardware]
    end

    subgraph "Battery Optimization Platform"
        subgraph "Core Services (✅ Implemented)"
            Asset[Asset Management<br/>Service<br/>✅ M2]
            Market[Market Data<br/>Service<br/>✅ M3]
            Telemetry[Telemetry<br/>Service<br/>✅ M5]
            Device[Device Interface<br/>Service<br/>✅ M5]
            Bidding[Bidding<br/>Service<br/>✅ M6]
        end

        subgraph "Future Services"
            Economics[Economics<br/>Service<br/>⏳ Future]
        end

        subgraph "Infrastructure"
            NATS[NATS Event Bus]
            AssetDB[(Asset DB<br/>PostgreSQL)]
            MarketDB[(Market DB<br/>PostgreSQL)]
            TelemetryDB[(Telemetry DB<br/>PostgreSQL)]
        end
    end

    %% External connections
    Operator -->|Register Battery| Asset
    Operator -->|Approve Commands| Bidding
    AEMO -->|Price Forecast| Market
    AEMO -->|FCAS Dispatch| Market
    Battery <-->|Telemetry/Commands| Device

    %% Service to DB
    Asset --> AssetDB
    Market --> MarketDB
    Telemetry --> TelemetryDB

    %% Event flows
    Asset -.->|BatteryRegistered| NATS
    Market -.->|PriceUpdated| NATS
    Telemetry -.->|StateChanged| NATS
    Device -.->|ChargingStarted| NATS
    Bidding -.->|CommandIssued| NATS
    Economics -.->|EconomicsCalculated| NATS

    NATS -.->|Events| Asset
    NATS -.->|Events| Market
    NATS -.->|Events| Telemetry
    NATS -.->|Events| Device
    NATS -.->|Events| Bidding
    NATS -.->|Events| Economics

    style NATS fill:#f9f,stroke:#333,stroke-width:4px
    style Asset fill:#bbf,stroke:#333
    style Market fill:#bbf,stroke:#333
    style Telemetry fill:#bbf,stroke:#333
    style Device fill:#bbf,stroke:#333
    style Bidding fill:#bbf,stroke:#333
    style Economics fill:#bbf,stroke:#333
```

## Service Boundaries (Domain-Driven Design)

```mermaid
graph LR
    subgraph "Asset Management Context"
        A1[Battery Aggregate]
        A2[Specifications]
        A3[Constraints]
    end

    subgraph "Market Context"
        M1[Price Data]
        M2[Forecasts]
        M3[AEMO Integration]
    end

    subgraph "Telemetry Context"
        T1[Real-time State]
        T2[SoC Tracking]
        T3[Hardware Adapters]
    end

    subgraph "Trading Context"
        B1[Bidding Strategy]
        B2[FCAS Contracts]
        B3[Opportunity Detection]
    end

    subgraph "Economics Context"
        E1[Revenue Calculation]
        E2[Conflict Resolution]
        E3[ROI Analysis]
    end

    A1 -.->|BatteryRegistered| B1
    M1 -.->|PriceUpdated| B1
    T1 -.->|StateChanged| B1
    B1 -.->|ConflictDetected| E1
    E1 -.->|EconomicsCalculated| B1
```

## Event Flow: Charging Scenario

```mermaid
sequenceDiagram
    participant Market
    participant Bidding
    participant Economics
    participant Device
    participant Battery

    Market->>Bidding: PriceUpdated (low price)
    Bidding->>Bidding: Analyze opportunity
    Bidding->>Economics: ChargingOpportunityDetected
    Economics->>Economics: Calculate ROI
    Economics->>Bidding: Revenue projection
    
    alt Manual/Semi-Auto Mode
        Bidding->>Operator: Request approval
        Operator->>Bidding: Approve
    end

    Bidding->>Device: ChargingCommandIssued
    
    alt Conflict Exists
        Device->>Bidding: ConflictDetected (FCAS active)
        Bidding->>Economics: CalculateEconomics
        Economics->>Bidding: ConflictResolutionSuggested
        Bidding->>Device: ConflictResolved
    end

    Device->>Battery: Start charging
    Device->>Bidding: ChargingStarted
    
    loop Every 1 second
        Battery->>Device: Telemetry update
        Device->>Bidding: BatteryStateChanged
    end

    Battery->>Device: Target SoC reached
    Device->>Bidding: ChargingCompleted
```

## Event Flow: FCAS Dispatch

```mermaid
sequenceDiagram
    participant AEMO
    participant Market
    participant Bidding
    participant Device
    participant Battery

    Note over Bidding: FCAS Contract Active (RAISE)
    
    AEMO->>Market: FCAS Dispatch Signal (RAISE)
    Market->>Bidding: FcasDispatchReceived
    
    Bidding->>Bidding: Validate contract
    Bidding->>Device: Reduce charging / Start discharging
    
    Device->>Battery: Adjust power output
    Battery->>Device: Power adjusted
    Device->>Bidding: BatteryStateChanged
    
    Note over Device,Battery: Response within 1 second
    
    Device->>Bidding: FcasDispatchCompleted (COMPLIANT)
    
    AEMO->>Market: Dispatch ended
    Market->>Bidding: Resume normal operations
```

## Data Flow: DB-per-Service Pattern

```mermaid
graph TB
    subgraph "Asset Management Service"
        AssetLogic[Business Logic]
        AssetRepo[Repository]
        AssetDB[(Asset DB)]
        AssetLogic --> AssetRepo
        AssetRepo --> AssetDB
    end

    subgraph "Market Data Service"
        MarketLogic[Business Logic]
        MarketRepo[Repository]
        MarketDB[(Market DB)]
        MarketLogic --> MarketRepo
        MarketRepo --> MarketDB
    end

    subgraph "Telemetry Service"
        TelemetryLogic[Business Logic]
        TelemetryRepo[Repository]
        TelemetryDB[(Telemetry DB)]
        TelemetryLogic --> TelemetryRepo
        TelemetryRepo --> TelemetryDB
    end

    NATS[NATS Event Bus]

    AssetLogic -.->|Events| NATS
    MarketLogic -.->|Events| NATS
    TelemetryLogic -.->|Events| NATS

    NATS -.->|Events| AssetLogic
    NATS -.->|Events| MarketLogic
    NATS -.->|Events| TelemetryLogic

    style NATS fill:#f9f,stroke:#333,stroke-width:3px
    style AssetDB fill:#bfb,stroke:#333
    style MarketDB fill:#bfb,stroke:#333
    style TelemetryDB fill:#bfb,stroke:#333
```

**Key Principle**: No direct database access between services. All communication via events.

## Technology Stack

| Layer | Technology | Purpose |
|-------|-----------|---------|
| **Language** | Go 1.23 | Service implementation |
| **Event Bus** | NATS 2.10 | Asynchronous messaging |
| **Database** | PostgreSQL 18 | Per-service data storage |
| **Containers** | Docker + Docker Compose | Local development |
| **Orchestration** | Kubernetes (future) | Production deployment |
| **Observability** | Prometheus + Grafana (future) | Metrics and monitoring |
| **Testing** | Go testing + testify | Unit and integration tests |

## Design Patterns Applied

### 1. Domain-Driven Design (DDD)
- **Bounded Contexts**: Each service owns its domain
- **Aggregates**: Battery, Price, Contract
- **Repositories**: Abstract data access
- **Domain Events**: First-class communication mechanism

### 2. Event-Driven Architecture
- **Event Sourcing**: State changes captured as events
- **Eventual Consistency**: Services update asynchronously
- **Outbox Pattern** (future): Reliable event publishing

### 3. Hexagonal Architecture (Ports & Adapters)
- **Ports**: Interfaces defined by domain
- **Adapters**: 
  - Primary: HTTP handlers, event subscribers
  - Secondary: Database repositories, NATS publishers

### 4. Anti-Corruption Layer
- **BatteryAdapter**: Isolate hardware-specific details
- **CustomAttributes**: Manufacturer-specific data segregation
- **Domain Model**: Clean, vendor-agnostic

## Deployment View (Future)

```mermaid
graph TB
    subgraph "Kubernetes Cluster"
        subgraph "Namespace: battery-optimization"
            subgraph "Services"
                AssetPod[Asset Management<br/>Deployment]
                MarketPod[Market Data<br/>Deployment]
                TelemetryPod[Telemetry<br/>Deployment]
                DevicePod[Device Interface<br/>Deployment]
                BiddingPod[Bidding<br/>Deployment]
                EconomicsPod[Economics<br/>Deployment]
            end

            subgraph "Infrastructure"
                NATSCluster[NATS Cluster<br/>StatefulSet]
                PostgresOperator[PostgreSQL<br/>Operator]
            end

            subgraph "Observability"
                Prometheus[Prometheus]
                Grafana[Grafana]
                Jaeger[Jaeger Tracing]
            end
        end
    end

    subgraph "External"
        Ingress[Ingress Controller]
        LoadBalancer[Load Balancer]
    end

    LoadBalancer --> Ingress
    Ingress --> AssetPod
    Ingress --> MarketPod

    AssetPod -.-> NATSCluster
    MarketPod -.-> NATSCluster
    TelemetryPod -.-> NATSCluster
    DevicePod -.-> NATSCluster
    BiddingPod -.-> NATSCluster
    EconomicsPod -.-> NATSCluster

    AssetPod --> PostgresOperator
    MarketPod --> PostgresOperator
    TelemetryPod --> PostgresOperator

    Prometheus -.->|Scrape| AssetPod
    Prometheus -.->|Scrape| MarketPod
    Grafana --> Prometheus
```

**Note**: Kubernetes deployment is future work. Current focus is local Docker Compose development.

## Security Considerations (Future)

- **Authentication**: JWT tokens for API access
- **Authorization**: RBAC for operator permissions
- **Secrets Management**: Kubernetes secrets or Vault
- **Network Policies**: Service-to-service communication restrictions
- **Encryption**: TLS for all communication

## Scalability Considerations

### Current (MVP - Educational Project)
- Single instance per service
- 5 microservices implemented (Asset Management, Market Data, Telemetry, Device Interface, Bidding)
- Suitable for 1-10 batteries
- Development/demo environment
- Event-driven architecture with NATS
- DB-per-service pattern (3 PostgreSQL databases)

### Future Production
- **Horizontal scaling**: Multiple service instances
- **Event partitioning**: NATS JetStream with partitioning
- **Database sharding**: Partition by battery site
- **Caching**: Redis for frequently accessed data
- **Rate limiting**: Protect against event storms

## Next Steps

See [PLANNING.md](../PLANNING.md) for current milestone progress and upcoming tasks.