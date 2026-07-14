# Market Stream

A real-time stock market (kinda) data engine written in Go. It connects to
Binance's public WebSocket API, converts individual trade events into
time-bucketed OHLC candles, broadcasts completed candles to connected
browsers over WebSocket, persists them to MySQL, and renders them with
TradingView's Lightweight Charts.

The project exists primarily to work through backend streaming-system
concerns — domain modeling, an in-process event pipeline, a candle
aggregation engine, concurrent WebSocket fan-out, and a small REST/history
layer — rather than to be a full trading UI.

---

## Features

- **Binance WebSocket client** (`internal/binance`) — connects to a Binance
  `aggTrade` stream and reads raw trade messages.
- **Wire-to-domain conversion** — Binance's `AggTrade` payload (string
  prices/quantities, Unix millis) is converted into a clean internal
  `model.Trade` (typed floats, `time.Time`).
- **Trade → Candle aggregation** (`internal/candle`) — a stateful engine
  that folds trades into OHLCV candles.
- **Fixed time-bucketed candle generation** — candles are bucketed into
  30-second windows (`time.Truncate`-based).
- **Live WebSocket broadcasting** (`internal/ws`) — every completed candle
  is pushed as JSON to all connected browser clients.
- **Rolling in-memory history** (`internal/history`) — keeps the most
  recent 500 candles in memory for fast reads.
- **MySQL persistence** (`internal/mysql`) — every completed candle is
  durably stored.
- **REST endpoints** for serving the frontend, a full in-memory history
  snapshot, and a limit-based candle query with MySQL fallback.
- **Browser frontend** built on TradingView Lightweight Charts — renders a
  candlestick series with a volume histogram beneath it.
- **Snapshot + live chart architecture** — the browser loads an initial
  candle snapshot over REST, then applies live updates from the WebSocket
  connection.

---

## Architecture

```text
                Binance WebSocket
              (wss://stream.binance.com)
                        │
                        ▼
                 binance.Client
              (Connect / ReadTrade)
                        │
               AggTrade (wire format)
                        │
                        ▼
                  model.Trade
               (domain conversion)
                        │
                        ▼
                pipeline.Pipeline
                 (ProcessTrade)
                        │
                        ▼
                 candle.Engine
            (OHLCV bucketing, 30s tf)
                        │
                completed candle?
                        │
          ┌─────────────┼─────────────┐
          ▼              ▼             ▼
   history.Store    mysql.Store    ws.Server
  (in-memory ring,  (durable       (JSON broadcast
   capacity 500)     persistence)   to all clients)
                                        │
                                        ▼
                                Browser Clients
                         (websocket.js → chart.js)
```

Only **completed** candles travel past the candle engine. Raw trades are
consumed internally to update the in-progress candle but are not broadcast
to clients — the frontend never sees tick-by-tick trades, only finished
30-second bars.

---

## Project Structure

```
cmd/
  server/            entrypoint: wires client, engine, stores, and pipeline together
internal/
  binance/           Binance WebSocket client (connect + read + decode)
  candle/            stateful OHLCV aggregation engine
  history/           fixed-capacity in-memory candle buffer
  model/             domain types (Trade, AggTrade, Candle) + wire→domain conversion
  mysql/             MySQL connection setup and candle persistence queries
  pipeline/          glues trade ingestion to the candle engine and its output sinks
  repository/        CandleRepository interface (not currently wired to any concrete store)
  ws/                WebSocket server, client registry, broadcast, REST handlers
web/
  index.html         page shell, loads Lightweight Charts from a CDN
  css/style.css       styling for the chart card and ticker header
  js/app.js           boot sequence: create chart → load history → connect WS
  js/chart.js          chart/series creation, candle + volume rendering
  js/websocket.js       WebSocket client and REST history fetch
```

**Notes on responsibility boundaries:**

- `internal/model` has no dependencies on other internal packages — it's
  the shared vocabulary (`Trade`, `Candle`, `AggTrade`) used everywhere
  else.
- `internal/pipeline` is the only package that depends on `candle`,
  `history`, `mysql`, and `ws` together. It is the composition point
  between ingestion and the three output sinks.
- `internal/repository` defines a `CandleRepository` interface
  (`SaveCandle`, `GetCandles`), but its method signatures don't match
  `mysql.Store`'s actual (error-returning) methods, and nothing in the
  codebase references the interface. It reads as a planned abstraction
  that was never finished — see [Code Quality Review](#code-quality-review).

---

## Data Flow

One trade's journey through the system:

1. **Trade arrives** — `binance.Client.ReadTrade()` blocks on
   `conn.ReadMessage()` over the Binance WebSocket connection.
2. **JSON parsing** — the raw message is unmarshaled into
   `model.AggTrade`, matching Binance's `aggTrade` wire format
   (`e`, `E`, `a`, `s`, `p`, `q`, `T`, ...).
3. **Domain conversion** — `AggTrade.ToTrade()` parses the string
   price/quantity into `float64` and converts the millisecond timestamp
   into `time.Time`, producing a `model.Trade`.
4. **Pipeline** — `main.go`'s read loop hands the trade to
   `pipeline.Pipeline.ProcessTrade`.
5. **Candle Engine** — `candle.Engine.Update(trade)` either starts a new
   candle bucket, updates the in-progress candle's High/Low/Close/Volume,
   or (if the trade's timestamp has crossed the bucket boundary) finalizes
   the current candle and starts the next one.
6. **Completed Candle** — only emitted when a bucket boundary is crossed;
   otherwise `Update` returns `(nil, false)` and nothing further happens.
7. **Fan-out** — on completion, the pipeline, in order:
   - saves the candle to MySQL (`mysql.Store.SaveCandle`)
   - appends it to the in-memory history ring (`history.Store.SaveCandle`)
   - broadcasts it to every connected WebSocket client
     (`ws.Server.Broadcast`)
8. **Browser** — `websocket.js` receives the JSON message, and
   `chart.js.addCandle()` pushes it into the candlestick and volume
   series and updates the OHLC ticker readout.

---

## Candle Engine

`internal/candle/engine.go` holds a single in-progress `*model.Candle` and
a fixed `timeframe` (currently `30 * time.Second`, set in `cmd/server/main.go`).

- **Time bucketing** — a candle's start time is `trade.TradeTime.Truncate(timeframe)`;
  its end time is `start + timeframe`. This aligns candles to fixed wall-clock
  boundaries rather than "30 seconds from the first trade."
- **OHLC generation** — the first trade in a bucket sets `Open == High ==
  Low == Close` to that trade's price. Subsequent trades update `High`
  (if greater), `Low` (if lower), and always overwrite `Close`.
- **Volume accumulation** — each trade's `Quantity` is added to the
  candle's running `Volume`.
- **Candle completion** — detected when an incoming trade's timestamp is
  no longer before the current candle's `EndTime`. At that point the
  finished candle is returned (with `completed == true`) and a new candle
  is immediately started from the triggering trade.
- **Single-symbol, single-timeframe** — the engine holds one candle at a
  time; it has no concept of multiple symbols or multiple concurrent
  timeframes. Running more than one market or interval requires more than
  one `Engine` instance (not currently wired up in `main.go`).

---

## REST API

All endpoints are served from `ws.Server.Start()` on `:8080`.

### `GET /`
Serves the static frontend from the `web/` directory via
`http.FileServer` (`index.html`, `css/`, `js/`).

### `GET /ws`
Upgrades the HTTP connection to a WebSocket. See
[WebSocket Protocol](#websocket-protocol).

### `GET /history`
Returns the entire in-memory history buffer (up to 500 candles) as a JSON
array, oldest to newest. Used by the frontend on page load to draw the
initial chart.

```json
[
  { "Symbol": "BNBUSDT", "Open": 572.10, "High": 572.55, "Low": 571.80,
    "Close": 572.30, "Volume": 128.4471,
    "StartTime": "2026-07-15T10:15:30Z", "EndTime": "2026-07-15T10:16:00Z" }
]
```

### `GET /candles?limit=N`
Returns up to `N` most recent candles.

- `limit` is required and must parse as an integer — a missing or
  non-numeric value returns `400 invalid limit`.
- `limit` is served from the in-memory history store first.
- **`limit` cannot exceed the history store's capacity (500)** — if it
  does, the handler returns `400` immediately (`ErrLimitExceedsCapacity`)
  without ever falling back to MySQL, even though MySQL may hold more
  history. This is a known ceiling, not a deliberate cap on total
  retrievable history.
- If the in-memory store has fewer than `limit` candles (e.g. shortly
  after server start), the handler queries MySQL for `limit` candles and
  uses that result set instead.

---

## WebSocket Protocol

- **Endpoint:** `ws://<host>:8080/ws`
- **Server → Client:** whenever the candle engine completes a bucket, the
  server marshals the `model.Candle` directly to JSON (no envelope, no
  message type field) and sends it as a text frame to every connected
  client. There is no per-client subscription or symbol filtering — every
  client receives every broadcast candle.
- **Client → Server:** the server reads any message sent by the browser,
  logs it, and replies with a static `"Hello from server!"` text frame.
  This is a connection-liveness stub, not a real client-to-server
  protocol — there is currently no way for a client to subscribe to a
  symbol or request a different timeframe.

Example broadcast payload (field names match the Go struct — no `json`
tags are defined on `model.Candle`):

```json
{
  "Symbol": "BNBUSDT",
  "Open": 572.10,
  "High": 572.55,
  "Low": 571.80,
  "Close": 572.30,
  "Volume": 128.4471,
  "StartTime": "2026-07-15T10:16:00Z",
  "EndTime": "2026-07-15T10:16:30Z"
}
```

---

## Database

Schema: `internal/mysql/schema.sql`

```sql
CREATE TABLE IF NOT EXISTS candles(
    id          bigint          auto_increment  primary key,
    symbol      varchar(20)     not null,
    open        double          not null,
    high        double          not null,
    low         double          not null,
    close       double          not null,
    volume      double          not null,
    start_time  datetime        not null,
    end_time    datetime        not null,

    INDEX idx_symbol (symbol),
    INDEX idx_start_time (start_time)
);
```

- `id` — surrogate primary key; candles are insert-only, never updated.
- `symbol` — indexed so future multi-symbol queries can filter without a
  full table scan; unused today since only one symbol is ever ingested.
- `open` / `high` / `low` / `close` / `volume` — the aggregated OHLCV
  values produced by the candle engine.
- `start_time` / `end_time` — the candle's bucket boundaries;
  `start_time` is indexed to support the `ORDER BY start_time` used by
  `GetCandles` and future range queries.
- There is **no unique constraint** on `(symbol, start_time)` and no
  `timeframe` column — a server restart mid-bucket or a future
  multi-timeframe engine could insert duplicate or ambiguous rows. See
  [Code Quality Review](#code-quality-review).

---

## Frontend

Plain HTML/CSS/JS, no build step, no framework.

- **Lightweight Charts** — loaded from the `unpkg` CDN in `index.html`;
  `chart.js` creates a candlestick series and a histogram series (volume)
  sharing one chart instance, styled to resemble a typical trading UI.
- **History loading** — `websocket.js.loadHistoryFromServer()` fetches
  `/history` and hands the array to `chart.js.loadHistory()`, which bulk
  loads both series via `setData()`.
- **Live updates** — `websocket.js.connectWebSocket()` opens
  `ws://localhost:8080/ws`; each incoming message is parsed and passed to
  `chart.js.addCandle()`, which calls `series.update()` on both the
  candle and volume series and refreshes the OHLC ticker readout.
- **Snapshot + Live architecture** — `app.js` sequences these
  deliberately on load: `createChart()` → `await loadHistoryFromServer()`
  → `connectWebSocket()`, so the chart has its historical shape before
  live ticks start arriving.
- The WebSocket URL (`ws://localhost:8080/ws`) and the header labels
  (`BNBUSDT`, `30s`) are hardcoded in the frontend rather than derived
  from the server, so the UI will silently mismatch reality if the
  backend's symbol, timeframe, or host/port change.

---

## Technologies Used

- Go (module targets `go 1.26.5` per `go.mod`)
- [Gorilla WebSocket](https://github.com/gorilla/websocket)
- MySQL (via [go-sql-driver/mysql](https://github.com/go-sql-driver/mysql))
- [TradingView Lightweight Charts](https://tradingview.github.io/lightweight-charts/)
- HTML / CSS / vanilla JavaScript
- Binance WebSocket API

---

## Running the Project

### Prerequisites
- Go (a recent 1.2x toolchain)
- A running MySQL server

### 1. Install dependencies
```bash
go mod download
```

### 2. Create the database and schema
```bash
mysql -u root -p -e "CREATE DATABASE marketstream;"
mysql -u root -p marketstream < internal/mysql/schema.sql
```

### 3. Configure the database connection
The MySQL DSN is currently **hardcoded** in `cmd/server/main.go`:

```go
mysqlStore, err := mysql.NewStore(
    "root:2159@tcp(localhost:3306)/marketstream?parseTime=true",
)
```

Edit this line to match your local MySQL user/password before running the
server. (Externalizing this into an environment variable or flag is on the
roadmap — see below.)

### 4. Run the server
```bash
go run ./cmd/server
```
This connects to Binance's `bnbusdt@aggTrade` stream and starts the HTTP/WebSocket server on `:8080`.

### 5. Open the chart
Navigate to [http://localhost:8080](http://localhost:8080) in a browser.

---

## Future Roadmap

**Implemented**
- Single-symbol Binance ingestion
- Trade → candle aggregation (fixed 30s timeframe)
- Live WebSocket candle broadcast
- In-memory rolling history + MySQL persistence
- REST history/candle endpoints
- Browser chart with snapshot + live updates

**In Progress**
- `internal/repository.CandleRepository` — an interface exists but isn't
  implemented by `mysql.Store` (signature mismatch) or referenced
  anywhere, suggesting an unfinished move toward a proper repository
  abstraction.

**Planned**
- Multi-timeframe engine (currently one hardcoded 30s bucket)
- Technical indicators
- Trader dashboard
- Multiple symbols (engine, schema, and WS protocol are all currently
  single-symbol)
- Order book support
- Configuration management (env vars/flags instead of hardcoded DSN,
  symbol, timeframe, and port)
- Docker packaging
- Metrics/observability
- Graceful shutdown (signal handling; currently the ingestion loop runs
  forever and calls `log.Fatal` on the first read error)

---

## Learning Outcomes

This project is a working example of several backend concepts:

- **WebSockets** — both as a client (Binance) and a server (browser
  clients), including connection lifecycle and broadcast fan-out.
- **Streaming/event-driven systems** — an unbounded trade stream reduced
  to a bounded, meaningful event (a completed candle).
- **Time-series aggregation** — wall-clock-aligned bucketing, OHLCV
  accumulation, and completion detection.
- **Domain modeling** — a clean separation between the external wire
  format (`AggTrade`) and the internal domain type (`Trade`).
- **Pipeline/composition pattern** — `pipeline.Pipeline` composes
  independent components (engine, history, MySQL, WS server) without
  those components knowing about each other.
- **REST API design** — snapshot endpoints with pagination-like `limit`
  semantics and a fallback between a fast in-memory store and durable
  storage.
- **MySQL persistence** — schema design with indexes chosen for the
  actual query patterns (`ORDER BY start_time`, filter by `symbol`).
- **Repository pattern (partially)** — an interface was introduced to
  decouple storage from the pipeline, though it isn't fully wired up yet.

---

## Code Quality Review

**Strengths**
- Clear separation between wire format (`AggTrade`), domain model
  (`Trade`/`Candle`), and transport — this is the kind of boundary that
  pays off if a second exchange or data source is ever added.
- The candle engine's completion detection (`!trade.TradeTime.Before(current.EndTime)`)
  correctly handles gaps in trade activity — a candle still completes
  even if no trade arrives exactly on the boundary.
- The pipeline's fan-out (MySQL → history → broadcast) keeps each sink
  independent and easy to reason about in isolation.
- Small, focused packages with a single responsibility each.

**Weaknesses & Risks**
- **Concurrent map access on `ws.Server.clients`.** `handleWS` (one
  goroutine per client) reads/writes `s.clients` on connect/disconnect,
  while `Broadcast` (called from the trade-ingestion goroutine) iterates
  and deletes from the same map with no mutex. Under concurrent
  connections this is a data race and can crash the process
  (`fatal error: concurrent map iteration and map write`).
- **Hardcoded configuration.** MySQL credentials, the Binance symbol, the
  candle timeframe, and the HTTP port are all literal values in
  `cmd/server/main.go` / `ws/server.go`. There's no environment- or
  flag-based configuration, which also means credentials currently live
  in source.
- **No graceful shutdown.** The main loop calls `log.Fatal` on the first
  `ReadMessage`/`ReadTrade` error, so a single Binance disconnect kills
  the whole server (including active browser WebSocket clients) with no
  reconnect logic and no `SIGINT`/`SIGTERM` handling to close DB/WS
  connections cleanly.
- **`GET /candles?limit=N` has an unreachable fallback for large limits.**
  Because `history.Store.GetCandles` rejects `limit > capacity` before
  MySQL is ever consulted, requesting more candles than fit in memory
  (500) always fails, even though MySQL may have the data.
- **No tests.** There are no `_test.go` files anywhere in the module;
  the candle engine's boundary logic in particular would benefit from
  table-driven tests given how easy off-by-one bucket errors are to
  introduce.
- **Dead/inconsistent code.** `pipeline.broadcast(trade)` is defined but
  never called; `ws.Server.home` is unreachable (superseded by the static
  file server); `internal/repository.CandleRepository` isn't implemented
  by anything; `cmd/server/main.go` has several commented-out blocks left
  from earlier iterations.
- **No idempotency at the storage layer.** `mysql.Store.SaveCandle` is a
  plain `INSERT` with no unique constraint on `(symbol, start_time)`, so
  a process restart mid-timeframe could produce duplicate rows for the
  same bucket.

**Architecture / Scalability Considerations**
- The candle engine, history store, and even the WebSocket broadcast are
  all single-symbol by construction (one `*model.Candle`, one flat
  history slice, one client map with no channel/topic concept). Adding
  symbols means keying most of these by symbol, not just adding a field.
- `history.Store`'s overflow handling (`s.candles = s.candles[1:]`) is a
  slice re-slice, not a true ring buffer — acceptable at capacity 500,
  but it re-anchors the backing array's start index on every insert past
  capacity, which is wasteful at larger capacities or higher symbol
  counts.
- `ws.Server.Broadcast` writes to each client synchronously in a loop; a
  slow or stalled client can block broadcasting to every other client.
  A per-client outbound channel/goroutine would decouple this.

**Production Readiness**
Not production-ready today, and not intended to be as-is — this is a
learning-focused backend project. Before it could run unattended, the
top priorities would be: fixing the `clients` map race, removing hardcoded
credentials, adding reconnect/backoff for the Binance client, and adding
graceful shutdown. None of these are architectural rewrites — they're
targeted fixes on top of a design that is otherwise sound.
