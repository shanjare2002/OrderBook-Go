#!/usr/bin/env python3
import json
import os
import sys
from urllib import request, parse, error

BASE_URL = os.environ.get("BASE_URL", "http://localhost:8080")


def _http_request(method: str, path: str, payload: dict | None = None):
    url = BASE_URL + path
    data = None
    headers = {"Content-Type": "application/json"}
    if payload is not None:
        data = json.dumps(payload).encode("utf-8")
    req = request.Request(url, data=data, method=method, headers=headers)
    try:
        with request.urlopen(req, timeout=5) as resp:
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


def main():
    print(f"Using BASE_URL={BASE_URL}")

    # 1) Register two users
    print("Registering users...")
    s1, u1_body = post_json("/registerUser")
    s2, u2_body = post_json("/registerUser")
    if s1 != 200 or s2 != 200:
        print(f"Failed to register users: {s1} {u1_body} / {s2} {u2_body}")
        sys.exit(1)
    user1 = json.loads(u1_body)
    user2 = json.loads(u2_body)
    uid1 = user1.get("userId")
    uid2 = user2.get("userId")
    print(f"User1: {uid1}\nUser2: {uid2}")

    # 2) Add balances (USD) and stock (AAPL)
    print("Funding users...")
    balances = [
        (uid1, {"asset": "USD", "amount": 50000}),
        (uid2, {"asset": "USD", "amount": 25000}),
        (uid1, {"asset": "AAPL", "amount": 200}),
        (uid2, {"asset": "AAPL", "amount": 100}),
    ]
    
    for uid, payload in balances:
        status, body = post_json(f"/addBalcne?userId={parse.quote(uid)}", payload)
        if status != 200:
            print(f"Failed to add balance for {uid}: {status} {body}")
            sys.exit(1)
    print("Balances updated.")

    # 3) Place orders
    print("Placing orders...")
    orders = [
        (uid1, {"position": 0, "quantity": 50, "price": 150.25, "ticker": "AAPL"}),
        (uid1, {"position": 0, "quantity": 30, "price": 149.80, "ticker": "AAPL"}),
        (uid2, {"position": 1, "quantity": 40, "price": 150.00, "ticker": "AAPL"}),
        (uid2, {"position": 1, "quantity": 20, "price": 151.00, "ticker": "AAPL"}),
    ]
    for uid, payload in orders:
        status, body = post_json(f"/order?userId={parse.quote(uid)}", payload)
        if status != 200:
            print(f"Order failed for {uid}: {status} {body}")
            sys.exit(1)
        print(f"Order for {uid}: {body}")

    # 4) Fetch order book snapshot
    print("\nOrder book:")
    status, body = get_json("/getOrderBook")
    if status != 200:
        print(f"Failed to fetch order book: {status} {body}")
        sys.exit(1)
    try:
        parsed = json.loads(body)
        print(json.dumps(parsed, indent=2))
    except json.JSONDecodeError:
        print(body)

    # 5) List users
    print("\nUsers:")
    status, body = get_json("/getUsers")
    if status != 200:
        print(f"Failed to fetch users: {status} {body}")
        sys.exit(1)
    try:
        parsed = json.loads(body)
        print(json.dumps(parsed, indent=2))
    except json.JSONDecodeError:
        print(body)


if __name__ == "__main__":
    main()
