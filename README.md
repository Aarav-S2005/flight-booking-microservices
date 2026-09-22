# ✈️ Cloud-Native Flight Booking & Reservation Microservices Platform

A high-performance, distributed, event-driven flight booking, route discovery, seat allocation, and payment processing ecosystem built with **Go (Golang)**, **PostgreSQL**, **RabbitMQ**, and **Redis**.

---

## 🏛️ Architecture Overview

The system is designed following modern microservices architectural patterns with clear domain boundaries, event-driven choreographies for asynchronous decoupling, lock-free in-memory route searching, and strict ACID transaction guarantees for inventory and seat management.

### System Block Diagram
![System Block Diagram](./docs/block%20diagram.png)

### End-to-End User Flow
![User Flow](./docs/User%20Flow.png)

---

## 🚀 Core Microservices

### 1. `services/auth-service`
* **Domain:** Identity, Authentication, and User Profile Management.
* **Security:** Asymmetric cryptographic signing using **ECDSA (ES256)** private keys, distributing public keys across peripheral services for zero-latency local token verification without centralized auth bottlenecks.
* **Storage:** PostgreSQL with salted **bcrypt** password hashing.

### 2. `services/flight-service`
* **Domain:** Flight Discovery, Schedule Management, and Real-Time Routing.
* **Engine:** In-memory graph-based **Depth-First Search (DFS)** for multi-segment route planning, calculating dynamic layovers, pricing discounts, and segment validations.
* **Concurrency:** Lock-free, read-optimized **Snapshot Registry** maintaining precomputed sort indices (price, duration, departure time) to achieve sub-millisecond query latencies.
* **Eventing:** Consumes seat decrement events to keep the live in-memory registry synchronized with database persistence.

### 3. `services/booking-service`
* **Domain:** Core Ticket Booking & Transactional Coordination.
* **Data Integrity:** Employs pessimistic row-level locking (`SELECT ... FOR UPDATE`) with optimistic version checks to guarantee zero double-booking under extreme concurrent traffic.
* **Resilience:** Integrates `singleflight.Group` to coalesce duplicate in-flight downstream lookups against the flight service.
* **Events:** Emits `BookingConfirmedForPayment`, `BookingConfirmedForReservation`, and `FlightSeatUpdated` events via RabbitMQ topic exchanges.

### 4. `services/payment-service`
* **Domain:** Payment Validation, Settlement, and Timeout Enforcement.
* **Workflows:** Consumes async payment initiation events from RabbitMQ and supports direct payment submission with time-window validation (e.g., payment expiry verification against booking creation timestamps).
* **Reliability:** Idempotent database records using composite unique constraints (`UNIQUE(user_id, booking_id)`).

### 5. `services/reservation-service`
* **Domain:** Aircraft Geometry, Interactive Seat Layout, and Seat Assignment.
* **Seat Matrix:** Pre-seeded with realistic aircraft seat maps (Boeing 737/747/777/787, Airbus A320/A330/A350/A380) including row, column, and total seat configurations.
* **Concurrency:** Transactional seat claiming using PostgreSQL row locks and unique constraints (`UNIQUE (flight_id, column_allocated, seat_number)`), enforcing time-window policies (seat selection opens 48h and closes 4h prior to departure).

### 6. `services/notification-service`
* **Domain:** Multi-Channel Async User Communications.
* **Engine:** Headless event processor consuming lifecycle events across all domains to compile and dispatch responsive HTML emails via SMTP with persistence tracking in MongoDB/PostgreSQL.

### 7. `shared`
* **Workspace Core:** Centralized Go workspace modules including:
  * `shared/rabbitmq`: Connection pooling, topology declaration, consumer worker pools, and DLQ handling.
  * `shared/middlewares`: High-performance slog structured logger, request tracing, and JWT auth verifier/authenticator.
  * `shared/keys`: PEM parser and ECDSA key loaders.
  * `shared/db`: Postgres (`pgxpool`), Redis, and Mongo initializers.
  * `shared/app-error`: RFC-compliant standardized error taxonomy.

---

## 🛠️ Tech Stack & Engineering Practices

| Layer | Technology | Purpose |
| :--- | :--- | :--- |
| **Language** | Go (Golang 1.22+) | High-throughput, concurrent microservices |
| **HTTP Router** | `go-chi/chi/v5` | Lightweight, idiomatic HTTP routing |
| **Authentication** | `go-chi/jwtauth/v5` (ES256) | Asymmetric ECDSA JWT authentication |
| **Message Broker**| RabbitMQ (AMQP 0-9-1) | Event-driven pub/sub choreography |
| **Primary Databases** | PostgreSQL (`jackc/pgx/v5`) | ACID transactions, row-level locks (`FOR UPDATE`) |
| **In-Memory Cache**| Redis (`go-redis/v9`) | Hot booking state caching & fast verification |
| **HTTP Client** | `go-resty/resty/v2` | Inter-service sync HTTP client |
| **Concurrency** | `golang.org/x/sync/singleflight` | Request deduplication & cache-stampede prevention |

---

## 🔑 Key Engineering Highlights

* **Event-Driven Choreography:** Fully decoupled microservices communicating asynchronously via RabbitMQ topic exchange (`booking.events.v1`), eliminating synchronous cascading failures.
* **Concurrency & Zero Double-Booking:** Strict atomic database transactions using `SELECT ... FOR UPDATE` row locks combined with seat counter versioning.
* **Lock-Free Read Snapshots:** The flight discovery service maintains copy-on-write immutable snapshot registries for high-throughput query handling with zero mutex lock contention.
* **Asymmetric Security Model:** Authentication service signs tokens using a private ECDSA key (`ES256`), while downstream microservices verify signatures independently using public keys without roundtrips to the auth service.
* **Distributed Request Coalescing:** Eliminates duplicate upstream flight metadata lookups using `singleflight` request merging.

---

## 📁 Repository Structure

```
flight-booking-microservices/
├── docs/                                # Architecture & User flow diagrams
│   ├── block diagram.png
│   └── User Flow.png
├── secret/                              # ECDSA ES256 key pairs
│   ├── ec-priv.pem
│   └── ec-pub.pem
├── services/
│   ├── auth-service/                    # User authentication & token issuance
│   ├── booking-service/                 # Transactional ticket booking engine
│   ├── flight-service/                  # Route discovery & in-memory snapshot registry
│   ├── notification-service/            # Asynchronous email notifications
│   ├── payment-service/                 # Payment processing & status checks
│   └── reservation-service/             # Seat matrix & seat claiming engine
├── shared/                              # Shared Go packages
│   ├── app-error/
│   ├── db/
│   ├── keys/
│   ├── middlewares/
│   ├── rabbitmq/
│   └── utility/
├── go.work                              # Go Multi-Module Workspace
└── README.md
```

---

## 🚦 Getting Started & Local Setup

### Prerequisites
* **Go** 1.22 or higher
* **PostgreSQL** 15+
* **RabbitMQ** 3.12+
* **Redis** 7+

### 1. Clone the repository
```bash
git clone https://github.com/Aarav-S2005/flight-booking-microservices.git
cd flight-booking-microservices
```

### 2. Generate ECDSA Key Pair (if not present)
```bash
mkdir -p secret
openssl ecparam -name prime256v1 -genkey -noout -out secret/ec-priv.pem
openssl ec -in secret/ec-priv.pem -pubout -out secret/ec-pub.pem
```

### 3. Configure Environment Variables
Each microservice contains a `.env` configuration file specifying database DSNs, ports, and RabbitMQ connection strings:
```env
PORT=8080
POSTGRES_DSN=postgres://postgres:password@localhost:5432/flight_booking?sslmode=disable
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
REDIS_URL=localhost:6379
PUBLIC_JWT_SECRET_PATH=../../secret/ec-pub.pem
PRIVATE_JWT_SECRET_PATH=../../secret/ec-priv.pem
```

### 4. Run Services
Run each service using Go workspace support:
```bash
# In separate terminal tabs or processes:
go run ./services/auth-service/cmd/server/main.go
go run ./services/flight-service/cmd/server/main.go
go run ./services/booking-service/cmd/server/main.go
go run ./services/reservation-service/cmd/server/main.go
go run ./services/payment-service/cmd/server/main.go
go run ./services/notification-service/cmd/server/main.go
```

---

## 📜 License
This project is licensed under the [MIT License](LICENSE).
