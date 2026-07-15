# Market Stream

A real-time cryptocurrency market data engine built in Go.

The application connects to Binance's WebSocket API, converts live trade events into OHLC candles, stores completed candles in MySQL, broadcasts them to connected clients over WebSocket, and visualizes them using TradingView Lightweight Charts. 

---

## Project Status

This is the **final state of the project**. It implements a single-timeframe
pipeline end-to-end (ingestion → candle aggregation → persistence →
broadcast → chart). A multi-timeframe engine was attempted and reverted in
favor of this simpler, working version.

Everything listed under [Not Implemented](#not-implemented) was scoped out
and left for a future iteration — it is not part of this build.

---

## Features

- Live Binance WebSocket integration
- Trade → OHLC candle aggregation
- Time-bucketed candle generation
- WebSocket broadcasting
- MySQL persistence
- In-memory rolling history
- REST API for historical candles
- Live candlestick chart
- Snapshot + Live chart loading

---

## Architecture

```text
        Binance WebSocket
               │
               ▼
        WebSocket Client
               │
               ▼
         Trade Pipeline
               │
               ▼
        Candle Engine
               │
      Completed Candle
        ┌──────┼──────┐
        ▼      ▼      ▼
    History  MySQL  WebSocket
      Store   Store   Server
                        │
                        ▼
                  Browser Client
```

---

## Project Structure

```text
cmd/
  server/

internal/
  binance/
  candle/
  history/
  model/
  mysql/
  pipeline/
  repository/
  ws/

web/
  css/
  js/
  index.html
```

---

## Data Flow

```text
Binance Trade
      │
      ▼
AggTrade
      │
      ▼
Trade
      │
      ▼
Pipeline
      │
      ▼
Candle Engine
      │
Completed Candle
      │
 ├── Save to MySQL
 ├── Save to History
 └── Broadcast to Browser
```

---

## REST API

| Method | Endpoint | Description |
|---------|----------|-------------|
| GET | `/` | Frontend |
| GET | `/ws` | WebSocket endpoint |
| GET | `/history` | In-memory history |
| GET | `/candles?limit=100` | Recent candles |

---

## Tech Stack

- Go
- Gorilla WebSocket
- MySQL
- TradingView Lightweight Charts
- HTML / CSS / JavaScript

---

## Run

```bash
go mod download
```

Create database

```sql
CREATE DATABASE marketstream;
```

Create table

```bash
mysql -u root -p marketstream < internal/mysql/schema.sql
```

Run

```bash
go run ./cmd/server
```

Open

```
http://localhost:8080
```

---

## Completed

- Binance WebSocket Client
- Candle Engine (single timeframe)
- Pipeline
- WebSocket Server
- Live Chart
- History API
- MySQL Storage

---

## Not Implemented

These were considered but not built in this version:

- Multi-timeframe engine (attempted, reverted — see [Project Status](#project-status))
- Technical indicators (EMA, SMA, VWAP, RSI, MACD)
- Symbol switching
- Trader dashboard
- Order book
- Docker
- Metrics
- Graceful shutdown

---

## Learning Objectives

This project focuses on learning backend system design concepts:

- WebSockets
- Streaming systems
- Time-series processing
- Repository pattern
- Event pipelines
- REST APIs
- MySQL persistence
- Domain modeling
- Concurrent Go applications

---
