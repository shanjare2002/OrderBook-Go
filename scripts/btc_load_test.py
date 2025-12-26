#!/usr/bin/env python3
import argparse
import json
import os
import random
import sys
import time
from urllib import request, parse, error

import matplotlib.pyplot as plt

DEFAULT_BASE_URL = os.environ.get("BASE_URL", "http://localhost:8080")

# --- HTTP helpers ---

def _http_request(method: str, path: str, payload: dict | None = None):
    url = BASE_URL + path
    data = None
    headers = {"Content-Type": "application/json"}
    if payload is not None:
        data = json.dumps(payload).encode("utf-8")
    req = request.Request(url, data=data, method=method, headers=headers)
    try:
        with request.urlopen(req, timeout=10) as resp:
            body = resp.read().decode("utf-8")
            return resp.status, body
    except error.HTTPError as e:
        return e.code, e.read().decode("utf-8", errors="ignore")
    except Exception as e:
        print(f"Request to {url} failed: {e}", file=sys.stderr)
        sys.exit(1)


def post_json(path: str, payload: dict | None = None):
    return _http_request("POST", path, payload)


def get_json(path: str):
    return _http_request("GET", path)

# --- API helpers ---

def register_user() -> str:
    status, body = post_json("/registerUser")
    if status != 200:
        raise RuntimeError(f"registerUser failed: {status} {body}")
    user = json.loads(body)
    return user["userId"]


def add_balance(user_id: str, asset: str, amount: float):
    status, body = post_json(f"/addBalance?userId={parse.quote(user_id)}", {"asset": asset, "amount": amount})
    if status != 200:
        raise RuntimeError(f"addBalcne failed for {user_id}: {status} {body}")


def place_order(user_id: str, position: int, quantity: int, price: float, ticker: str = "BTC"):
    payload = {
        "position": position,  # 0=BUY, 1=SELL
        "quantity": quantity,
        "price": price,
        "ticker": ticker,
    }
    status, body = post_json(f"/order?userId={parse.quote(user_id)}", payload)
    if status != 200:
        print(f"Order failed for {user_id}: {status} {body}")
    return status, body

# --- Scenario ---

def get_top_of_book():
    """Fetch current best buy/sell prices."""
    status, body = get_json("/topOfBook")
    if status == 200:
        try:
            data = json.loads(body)
            buy = data.get("buy")
            sell = data.get("sell")
            buy_price = float(buy["price"]) if buy and "price" in buy else None
            sell_price = float(sell["price"]) if sell and "price" in sell else None
            return buy_price, sell_price
        except json.JSONDecodeError:
            return None, None
    return None, None


def seed_ladder(user_ids, sellers, base_price: float, step: float, levels: int, min_qty: int, max_qty: int):
    print(f"Seeding ladders: levels={levels}, step={step}, qty=[{min_qty},{max_qty}]")
    buyers = user_ids
    for i in range(1, levels + 1):
        price = round(base_price - i * step, 2)
        uid = random.choice(buyers)
        qty = random.randint(min_qty, max_qty)
        place_order(uid, 0, qty, price, ticker="BTC")
    for i in range(1, levels + 1):
        price = round(base_price + i * step, 2)
        uid = random.choice(sellers) if sellers else random.choice(user_ids)
        qty = random.randint(min_qty, max_qty)
        place_order(uid, 1, qty, price, ticker="BTC")


def main():
    parser = argparse.ArgumentParser(description="BTC/USD market load generator")
    parser.add_argument("--base-url", default=DEFAULT_BASE_URL)
    parser.add_argument("--users", type=int, default=12)
    parser.add_argument("--orders", type=int, default=800)
    parser.add_argument("--base", type=float, default=30000.0, help="BTC base price")
    parser.add_argument("--spread", type=float, default=1200.0, help="± price randomization range")
    parser.add_argument("--max-qty", type=int, default=8)
    parser.add_argument("--seed-levels", type=int, default=12, help="initial ladder levels per side")
    parser.add_argument("--seed-step", type=float, default=50.0, help="price step between ladder levels")
    parser.add_argument("--seed-min-qty", type=int, default=2)
    parser.add_argument("--seed-max-qty", type=int, default=10)
    parser.add_argument("--wave-period", type=int, default=60, help="orders per wave cycle")
    parser.add_argument("--trend", type=str, default="uptrend", choices=["uptrend", "downtrend", "sideways", "volatile"], help="Price pattern: uptrend, downtrend, sideways, volatile")
    parser.add_argument("--trend-rate", type=float, default=0.5, help="Trend strength per order (in dollars)")
    parser.add_argument("--volatility", type=float, default=1.5, help="Volatility multiplier (higher = wilder price swings)")
    args = parser.parse_args()

    global BASE_URL
    BASE_URL = args.base_url

    print(f"Using BASE_URL={BASE_URL}")
    print(f"Users={args.users}, Orders={args.orders}, BTC base={args.base}, spread=±{args.spread}")

    user_ids = [register_user() for _ in range(args.users)]
    print(f"Registered users: {len(user_ids)}")

    sellers = []
    for i, uid in enumerate(user_ids):
        add_balance(uid, "USD", 2500000000.0)
        if i % 2 == 0:
            add_balance(uid, "BTC", 10000.0)
            sellers.append(uid)
    print(f"Balances funded: USD for all, BTC for sellers ({len(sellers)}/{len(user_ids)}).")

    seed_ladder(
        user_ids,
        sellers,
        base_price=args.base,
        step=args.seed_step,
        levels=args.seed_levels,
        min_qty=args.seed_min_qty,
        max_qty=args.seed_max_qty,
    )

    print("Placing orders...")
    random.seed(42)
    price_samples = []
    current_price = args.base
    momentum = 0  # For momentum-based trending
    
    for n in range(args.orders):
        in_buy_wave = (n // args.wave_period) % 2 == 0
        if in_buy_wave:
            position = 0 if random.random() < 0.65 else 1
        else:
            position = 1 if random.random() < 0.65 else 0

        if position == 1:
            uid = random.choice(sellers) if sellers else random.choice(user_ids)
        else:
            uid = random.choice(user_ids)

        qty = max(args.seed_min_qty, min(args.max_qty, int(abs(random.gauss(3, 2))) or 1))
        
        # Generate realistic price patterns
        if args.trend == "uptrend":
            # Strong upward drift with momentum
            momentum = momentum * 0.7 + random.gauss(0, 0.3)  # Smooth momentum
            trend_component = args.trend_rate + momentum * 0.5
            noise = random.gauss(0, args.spread / args.volatility)
            current_price = current_price + trend_component + noise
        elif args.trend == "downtrend":
            # Strong downward drift
            momentum = momentum * 0.7 + random.gauss(0, 0.3)
            trend_component = -args.trend_rate + momentum * 0.5
            noise = random.gauss(0, args.spread / args.volatility)
            current_price = current_price + trend_component + noise
        elif args.trend == "sideways":
            # Mean-reversion around base (support/resistance)
            deviation = current_price - args.base
            reversion_force = -deviation * 0.02  # Pull back toward center
            noise = random.gauss(0, args.spread / args.volatility)
            current_price = current_price + reversion_force + noise
        elif args.trend == "volatile":
            # Wild price swings with no clear direction
            noise = random.gauss(0, args.spread)
            current_price = current_price + noise
        
        # Prevent price from going negative
        current_price = max(100, current_price)
        
        status, body = place_order(uid, position, qty, round(current_price, 2))

        if n % 100 == 0:
            print(f"[{n}/{args.orders}] user={uid[:8]} pos={'BUY' if position==0 else 'SELL'} qty={qty} price={current_price:.2f} → {status}")
            buy_price, sell_price = get_top_of_book()
            if buy_price is not None and sell_price is not None:
                spread = buy_price - sell_price
                print(f"  💰 Top of Book: buy=${buy_price:.2f}, sell=${sell_price:.2f}, spread=${spread:.2f}")
                price_samples.append((n, buy_price, sell_price))
        time.sleep(0.005)

    print("Orders placement complete.")
    
    if price_samples:
        print(f"\nPrice evolution ({len(price_samples)} samples):")
        for order_n, buy_price, sell_price in price_samples:
            spread = buy_price - sell_price
            print(f"  Order #{order_n:5d}: buy=${buy_price:8.2f}, sell=${sell_price:8.2f}, spread=${spread:8.2f}")
        
        # Plot price evolution
        print("\nGenerating plot...")
        orders = [s[0] for s in price_samples]
        buys = [s[1] for s in price_samples]
        sells = [s[2] for s in price_samples]
        
        fig, ax = plt.subplots(figsize=(12, 6))
        ax.plot(orders, buys, marker='o', color='green', label='Buy (lowest ask)', linewidth=2)
        ax.plot(orders, sells, marker='s', color='red', label='Sell (highest bid)', linewidth=2)
        ax.fill_between(orders, buys, sells, alpha=0.2, color='gray', label='Spread')
        ax.set_xlabel("Order Count", fontsize=12)
        ax.set_ylabel("Price (USD)", fontsize=12)
        ax.set_title("BTC/USD Top of Book Evolution", fontsize=14, fontweight='bold')
        ax.legend(loc='best', fontsize=10)
        ax.grid(True, alpha=0.3)
        plt.tight_layout()
        
        # Save and show
        plot_file = "btc_price_evolution.png"
        plt.savefig(plot_file, dpi=150)
        print(f"Plot saved to {plot_file}")
        plt.show()

    # status, body = get_json("/getOrderBook")
    # if status == 200:
    #     try:
    #         snap = json.loads(body)
    #         print("\nOrder Book Snapshot:")
    #         print(json.dumps(snap, indent=2))
    #     except json.JSONDecodeError:
    #         print(body)
    # else:
    #     print(f"Failed to fetch order book: {status} {body}")

    status, body = get_json("/getUsers")
    if status == 200:
        try:
            users = json.loads(body)
            print(f"\nTotal users: {len(users)}")
        except json.JSONDecodeError:
            print(body)


if __name__ == "__main__":
    main()
