# 📡 AGENTS.md — NetDash

## Overview

**NetDash** is a lightweight, modular network visibility and monitoring service designed for home lab environments.

It discovers devices on a local network, tracks their availability over time, and exposes that data via a simple web UI and API.

The system is intentionally:

* minimal in dependencies
* easy to extend
* structured for incremental evolution into a full observability tool

---

## Core Responsibilities

NetDash currently provides:

1. **Device Discovery**

   * Active subnet scanning (TCP probing)
   * (Planned) ARP table ingestion
   * (Planned) mDNS/ZeroConf discovery

2. **State Tracking**

   * Tracks device online/offline status
   * Maintains last seen timestamps

3. **Event System**

   * Emits events when:

     * new devices are discovered
     * devices go online/offline

4. **Persistence (SQLite)**

   * `events` table → alert history
   * `device_uptime` table → state transitions

5. **Web Interface**

   * Simple polling-based UI
   * Displays current device state

---

## Architecture

```
cmd/server        → application entrypoint
internal/
  api/            → HTTP handlers
  alerts/         → event pipeline + persistence
  config/         → configuration + label loading
  device/         → core domain (models + store)
  discovery/      → network scanning + discovery workers
  storage/        → SQLite initialization
static/           → frontend UI
```

---

## Key Design Principles

### 1. Single Source of Truth

The `device.Store` is the authoritative state for all devices.

All discovery mechanisms MUST update devices via:

```go
store.Update(ip, func(d *Device) { ... })
```

Never mutate device state outside the store.

---

### 2. Event-Driven Side Effects

State changes trigger events via a channel:

```go
store.SetAlertChannel(alertManager.Channel())
```

Agents should:

* emit events, not directly write logs or DB records
* let the `alerts.Manager` handle persistence

---

### 3. Pluggable Discovery

Each discovery method runs independently:

* `scanner.go` → active probing
* `arp.go` → passive MAC discovery (stubbed)
* `mdns.go` → service discovery (stubbed)

New discovery methods should:

* run as goroutines
* call `store.Update(...)`
* avoid blocking

---

### 4. Eventually Consistent State

The system is not transactional.

Expect:

* temporary inconsistencies
* delayed updates
* race-tolerant design

Do NOT introduce heavy locking or synchronization beyond the store.

---

## Data Model

### Device (core entity)

```go
type Device struct {
  IP       string
  LastSeen time.Time
  Online   bool
}
```

⚠️ This is intentionally minimal right now.

Future extensions may include:

* MAC address
* hostname
* vendor
* open ports
* device type (fingerprinting)

---

## Database Schema

### events

```
id INTEGER
timestamp DATETIME
message TEXT
```

### device_uptime

```
ip TEXT
last_change DATETIME
status INTEGER (1=online, 0=offline)
```

---

## Known Limitations

Agents should be aware:

1. **Scanner is simplistic**

   * Only probes TCP:80
   * Misses many devices

2. **No hostname resolution**

3. **No MAC/vendor detection**

4. **No uptime aggregation queries**

5. **UI is polling-based (not real-time push)**

6. **ARP + mDNS are stubbed**

---

## Safe Extension Guidelines

### ✅ Good Changes

* Add new discovery modules
* Extend `Device` struct
* Add new API endpoints
* Improve scanning logic
* Add DB queries (read-only or append-only)

---

### ⚠️ Be Careful With

* Modifying `device.Store` locking behavior
* Blocking inside goroutines
* Writing directly to DB outside `alerts` or store
* Changing event semantics

---

### ❌ Avoid

* Introducing heavy frameworks
* Tight coupling between packages
* Long-running blocking operations in HTTP handlers
* Replacing polling UI without fallback

---

## Recommended Next Improvements

High-impact areas for future agents:

### 1. Device Enrichment

* MAC address via ARP parsing
* Vendor lookup (OUI database)
* Reverse DNS hostnames

---

### 2. Better Discovery

* ICMP ping support
* Multi-port probing
* mDNS + SSDP implementation

---

### 3. Uptime Analytics

* Aggregate uptime % from `device_uptime`
* Detect flapping devices
* Expose metrics API

---

### 4. Real-Time UI

* WebSocket server
* Push updates instead of polling

---

### 5. Observability

* Prometheus metrics endpoint
* Structured logging

---

## Mental Model for Agents

Think of NetDash as:

> A **state engine** (device store)
>
> * **event pipeline** (alerts)
> * **pluggable inputs** (discovery)
> * **simple outputs** (API/UI)

Do not treat it like:

* a static scanner
* a monolithic app
* a fully consistent database system

---

## How to Run

```bash
go mod init netdash
go get github.com/mattn/go-sqlite3
go run ./cmd/server
```

Open:

```
http://localhost:8080
```

---

## Final Guidance

When extending this system:

* Prefer **small, composable additions**
* Keep **state centralized**
* Emit **events instead of side effects**
* Optimize for **observability over perfection**

If you’re unsure where to add logic:

* discovery → finding devices
* device/store → state transitions
* alerts → persistence + history
* api → exposure to UI

---

This project is intentionally a foundation.
Build on it—don’t fight it.
