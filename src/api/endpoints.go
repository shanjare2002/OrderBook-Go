package endpoints

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/shanjare2002/OrderBook-Go/src/orderbook"
	"github.com/shanjare2002/OrderBook-Go/src/orders"
	user "github.com/shanjare2002/OrderBook-Go/src/users"
)

var NewOrderBook *orderbook.OrderBook

func hello_handler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("hello world"))

}

func get_orderbook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(NewOrderBook.Snapshot()); err != nil {
		http.Error(w, "failed to encode order book", http.StatusInternalServerError)
		return
	}
}

func get_top_of_book(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(NewOrderBook.Top()); err != nil {
		http.Error(w, "failed to encode top of book", http.StatusInternalServerError)
		return
	}
}

func add_user(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if NewOrderBook == nil {
		http.Error(w, "order book not initialized", http.StatusInternalServerError)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()
	newUser := NewOrderBook.RegisterUser()
	if err := json.NewEncoder(w).Encode(newUser); err != nil {
		http.Error(w, "failed to encode order book", http.StatusInternalServerError)
		return
	}
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	if err := json.NewEncoder(w).Encode(NewOrderBook.Users); err != nil {
		http.Error(w, "failed to encode order book", http.StatusInternalServerError)
		return
	}
}

func addBalance(w http.ResponseWriter, r *http.Request) {
	var balanceReq user.BalanceRequest
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if NewOrderBook == nil {
		http.Error(w, "order book not initialized", http.StatusInternalServerError)
		return
	}

	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&balanceReq); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	userId := r.URL.Query().Get("userId")

	if userId == "" {
		http.Error(w, "cannot find userId", http.StatusBadRequest)
		return
	}

	userUUID, err := uuid.Parse(userId)
	if err != nil {
		http.Error(w, "invalid userId format", http.StatusBadRequest)
		return
	}

	currUser, err := NewOrderBook.GetUser(userUUID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	fmt.Println(currUser.Balance)
	currUser.AddBalance(balanceReq.Asset, balanceReq.Amount)
}

func add_order(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if NewOrderBook == nil {
		http.Error(w, "order book not initialized", http.StatusInternalServerError)
		return
	}

	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var newOrder orders.Order
	if err := dec.Decode(&newOrder); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		http.Error(w, "body must contain a single JSON object", http.StatusBadRequest)
		return
	}

	if newOrder.Position != orders.BUY && newOrder.Position != orders.SELL {
		http.Error(w, "invalid position", http.StatusBadRequest)
		return
	}
	if newOrder.Quantity <= 0 {
		http.Error(w, "quantity must be > 0", http.StatusBadRequest)
		return
	}
	if newOrder.Price <= 0 {
		http.Error(w, "price must be > 0", http.StatusBadRequest)
		return
	}

	org := newOrder.Quantity
	userId := r.URL.Query().Get("userId")

	if userId == "" {
		http.Error(w, "cannot find userId", http.StatusBadRequest)
		return
	}

	userUUID, err := uuid.Parse(userId)
	if err != nil {
		http.Error(w, "invalid userId format", http.StatusBadRequest)
		return
	}

	currUser, err := NewOrderBook.GetUser(userUUID)
	if err != nil {
		http.Error(w, "user must be registered before placing an order. Use /registerUser endpoint first", http.StatusNotFound)
		return
	}

	newOrder.User = currUser

	if newOrder.Position == orders.BUY && currUser.Balance["USD"] < float64(newOrder.Price)*float64(newOrder.Quantity) {
		http.Error(w, "insufficient balance USD", http.StatusBadRequest)
		return
	}

	if newOrder.Position == orders.SELL && currUser.Balance[newOrder.Ticker] < float64(newOrder.Quantity) {
		http.Error(w, "insufficient balance don't have ticker in your portfolio", http.StatusBadRequest)
		return
	}

	NewOrderBook.NewOrder(&newOrder)

	w.WriteHeader(http.StatusOK)
	if newOrder.Quantity == 0 {
		w.Write([]byte("Order fullfilled fully"))
		return
	}
	if org > newOrder.Quantity {
		fmt.Fprintf(w, "Partially placed %v units", org-newOrder.Quantity)
		return
	}
	w.Write([]byte("Order recieved"))
}

func NewService() http.Handler {
	NewOrderBook = orderbook.NewOrderBook()
	handler := http.NewServeMux()
	handler.HandleFunc("/hello", hello_handler)
	handler.HandleFunc("/order", add_order)
	handler.HandleFunc("/getOrderBook", get_orderbook)
	handler.HandleFunc("/topOfBook", get_top_of_book)
	handler.HandleFunc("/getUsers", getUsers)
	handler.HandleFunc("/registerUser", add_user)
	handler.HandleFunc("/addBalance", addBalance)
	return handler
}
