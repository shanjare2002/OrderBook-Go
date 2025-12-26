# OrderBook-Go

A high-performance order book implementation in Go using B-trees for efficient price level management and order matching.

## Overview

This project implements a **limit order book** (LOB) with the following characteristics:

- **Data Structure**: B-tree based order levels for O(log n) inserts, deletes, and O(n) ordered traversals
- **Ordering**: 
  - Bids: Descending by price (highest bid first)
  - Asks: Ascending by price (lowest ask first)
- **Matching**: Incoming orders are matched against opposite-side levels with best-price-first priority
- **API**: RESTful HTTP endpoints for order submission and book snapshots

## Project Structure

```
OrderBookProject/
├── main.go                          # Entry point; starts HTTP server on :8080
├── go.mod                           # Module definition
├── src/
│   ├── api/
│   │   └── endpoints.go            # HTTP handlers for orders, users, and book snapshots
│   ├── orderbook/
│   │   └── orderBook.go            # Core order book logic with B-trees
│   ├── orders/
│   │   └── order.go                # Order types (Order, Bid, Ask) with user and asset support
│   └── users/
│       └── user.go                 # User management with balances and asset tracking
├── scripts/
│   ├── demo_scenario.py            # Test/demo script for API testing
│   └── btc_load_test.py            # Load testing script
└── README.md                        # This file
```

## API Endpoints

### `GET /hello`
Basic health check.

**Response:**
```
hello world
```

---

### `GET /getOrderBook`
Retrieve a snapshot of the current order book in JSON format.

**Response (200 OK):**
```json
{
  "bids": [
    { "userId": "uuid", "position": 0, "quantity": 10, "price": 101.5, "ticker": "BTC" },
    { "userId": "uuid", "position": 0, "quantity": 5, "price": 100.2, "ticker": "BTC" }
  ],
  "asks": [
    { "userId": "uuid", "position": 1, "quantity": 3, "price": 99.9, "ticker": "BTC" },
    { "userId": "uuid", "position": 1, "quantity": 7, "price": 100.5, "ticker": "BTC" }
  ]
}
```

**Notes:**
- `bids` array is ordered highest price first (best bids)
- `asks` array is ordered lowest price first (best asks)
- `position`: 0 = BUY, 1 = SELL
- `quantity` and `price` are numeric values
- `userId` identifies the user who placed the order
- `ticker` is the asset being traded (e.g., BTC, ETH)

---

### `GET /getTopOfBook`
Retrieve the best bid and ask from the current order book.

**Response (200 OK):**
```json
{
  "bestBid": { "userId": "uuid", "position": 0, "quantity": 10, "price": 101.5, "ticker": "BTC" },
  "bestAsk": { "userId": "uuid", "position": 1, "quantity": 3, "price": 99.9, "ticker": "BTC" }
}
```

**Notes:**
- Returns the single best bid (highest price) and best ask (lowest price)
- Useful for checking current market conditions quickly

---

---

### `POST /registerUser`
Create a new user in the system.

**Response (200 OK):**
```json
{
  "userId": "550e8400-e29b-41d4-a716-446655440000",
  "balance": {}
}
```

**Fields:**
- `userId`: Unique UUID assigned to the new user
- `balance`: Empty map to store asset balances

**Notes:**
- No request body required
- User is created with an empty balance map
- The returned `userId` must be used for all subsequent operations

---

### `POST /addBalance`
Add or update balance for a specific asset for a registered user.

**Query Parameters:**
- `userId` (required): The UUID of a registered user

**Request Body (JSON):**
```json
{
  "asset": "BTC",
  "amount": 1000.0
}
```

**Fields:**
- `asset`: The asset symbol (e.g., BTC, ETH, USD)
- `amount`: Amount to add to the balance for this asset

**Response (200 OK):**
Returns no response body, only confirms the balance was updated.

**Error Responses:**
- `400 Bad Request` – Missing userId query parameter, invalid JSON, or invalid userId format
- `404 Not Found` – User not found
- `405 Method Not Allowed` – Non-POST request
- `500 Internal Server Error` – Server initialization error

**Notes:**
- If the asset already exists in user's balance, the amount is added to existing balance
- If the asset doesn't exist, it is created with the provided amount
- This endpoint must be called before placing orders that require balances
- **Important:** The order book typically contains exchanges between two assets (USD vs BTC, or USD vs ETH). The assets added to a user's balance via this endpoint determine which trading pair(s) the user can participate in. For example, adding USD and BTC balances allows the user to trade USD/BTC pairs.

---

### `POST /order`
Submit a new limit order. The order is matched against opposite-side levels and any remainder is added to the book.

**IMPORTANT:** User must be registered via `/registerUser` endpoint before placing an order.

**Request Body (JSON):**
```json
{
  "userId": "user-uuid",
  "position": 0,
  "quantity": 50,
  "price": 101.25,
  "ticker": "BTC"
}
```

**Query Parameters:**
- `userId` (required): The UUID of a registered user

**Fields:**
- `userId`: UUID of the registered user placing the order
- `position`: 0 for BUY, 1 for SELL
- `quantity`: Number of units to trade (must be > 0)
- `price`: Limit price in currency units (must be > 0)
- `ticker`: Asset symbol being traded (e.g., BTC, ETH)

**Response (200 OK):**
- `"Order fullfilled fully"` – Entire order was matched.
- `"Partially placed X units"` – Order was partially matched, X units were placed.
- `"Order recieved"` – Order was added to the book without matching.

**Error Responses:**
- `400 Bad Request` – Invalid JSON, missing userId query parameter, invalid values, or insufficient balance
- `404 Not Found` – User not registered. Must call `/registerUser` first
- `405 Method Not Allowed` – Non-POST request
- `500 Internal Server Error` – Server initialization error

**Example Order Flow:**
1. Register user: `POST /registerUser`
2. Add balance: `POST /addBalcne?userId=<uuid>` with asset and amount
3. Place order: `POST /order?userId=<uuid>` with order details
4. Current book: Asks at [99.9, 100.5], Bids at [101.5, 100.2]
5. Submit BUY order: price=100.0, qty=10
   - Matches 3 units at ask 99.9 (remaining qty: 7)
   - Cannot match at ask 100.5 (too high)
   - Remaining 7 units added as a new bid at 100.0

---

### `POST /user`
Create a new user in the system.

**Request Body (JSON):**
```json
{
  "asset": "BTC",
  "amount": 1000.0
}
```

**Fields:**
- `asset`: The asset symbol to initialize balance for (e.g., BTC, ETH)
- `amount`: Initial balance amount for the asset

**Response (200 OK):**
```json
{
  "userId": "new-uuid-string",
  "balance": {
    "BTC": 1000.0
  }
}
```

**Notes:**
- Creates a new user with a unique UUID
- Initializes the user's balance with the provided asset and amount
- Multiple assets can be added through subsequent requests

## Key Features

### User Management
- Each order is associated with a specific user via UUID
- Users maintain individual balances for multiple assets
- Balance tracking enables position management and risk controls

### Top of Book Endpoint
- New `/getTopOfBook` endpoint provides O(1) access to best bid/ask prices
- Efficient for high-frequency monitoring of market conditions

### Order Matching
- Automatic order matching against opposite-side levels
- Best-price-first matching priority
- Partial fill support with remainder added to order book

## Building & Running

### Prerequisites
- Go 1.16+ 
- Dependencies: `github.com/google/btree`, `github.com/google/uuid`

### Build
```bash
go build ./...
```

### Run
```bash
go run main.go
```

The server starts on `http://localhost:8080`.

### Test API with Scripts
Use the provided Python scripts for testing:

```bash
# Demo scenario with various order operations
python3 scripts/demo_scenario.py

# Load testing
python3 scripts/btc_load_test.py
```

## Data Structures

### Order Book
- **Asks**: B-tree sorted ascending by price (best ask = lowest price)
- **Bids**: B-tree sorted descending by price (best bid = highest price)
- Enables O(log n) insertions and deletions with O(1) access to best prices

### User
- Maintains a map of asset balances
- Identified by UUID for order tracking and balance management

### Order
- Links to user via pointer
- Includes position (BUY/SELL), quantity, price, and ticker
- Snapshot support for read-only representations

```bash
# Health check
curl http://localhost:8080/hello

# Get order book
curl http://localhost:8080/getOrderBook

# Submit a BUY order
curl -X POST http://localhost:8080/order \
  -H "Content-Type: application/json" \
  -d '{"position": 0, "quantity": 10, "price": 101.5}'

# Submit a SELL order
curl -X POST http://localhost:8080/order \
  -H "Content-Type: application/json" \
  -d '{"position": 1, "quantity": 5, "price": 99.9}'
```

## Key Components

### `Order` (orders/order.go)
Represents a single order with position (BUY/SELL), quantity, and price.

### `Bid` & `Ask` (orders/order.go)
Wrapper types implementing `btree.Item` with custom ordering:
- **Bid**: `Less(i, j)` returns `true` if price[i] > price[j]` (descending)
- **Ask**: `Less(i, j)` returns `true` if price[i] < price[j]` (ascending)

### `OrderBook` (orderbook/orderBook.go)
Core order book with two B-trees:
- `Bids`: Max-heap style (highest price first)
- `Asks`: Min-heap style (lowest price first)

**Key Methods:**
- `NewOrder(order)`: Matches incoming order and adds remainder to book
- `fillBid(order)`: Matches incoming BUY order against asks
- `fillAsk(order)`: Matches incoming SELL order against bids
- `Snapshot()`: Returns a JSON-serializable view of the current book

## Performance Characteristics

| Operation | Complexity | Notes |
|-----------|-----------|-------|
| Insert/Update level | O(log n) | B-tree rebalancing |
| Delete level | O(log n) | B-tree rebalancing |
| Get level by price | O(log n) | Direct lookup |
| Order matching | O(m log n) | m = matched levels, n = total levels |
| Full traversal (snapshot) | O(n) | Single in-order scan |
| Top-of-book | O(1) | Tree.Min() / Tree.Max() |

## Dependencies

- `github.com/google/btree` – B-tree implementation

Install:
```bash
go get github.com/google/btree
```

