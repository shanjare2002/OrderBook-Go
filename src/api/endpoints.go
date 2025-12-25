package endpoints

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/shanjare2002/OrderBook-Go/src/orderbook"
	"github.com/shanjare2002/OrderBook-Go/src/orders"
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

func add_order(w http.ResponseWriter, r *http.Request) {
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
	NewOrderBook.NewOrder(&newOrder)

	w.WriteHeader(http.StatusOK)
	if newOrder.Quantity == 0 {
		w.Write([]byte("Order fullfilled fully"))
		return
	}
	if(org > newOrder.Quantity){
			fmt.Fprintf(w, "Partially placed %v units", org - newOrder.Quantity)
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
	return handler
}
