package orders

import (
	"github.com/google/btree"
	"github.com/google/uuid"
	user "github.com/shanjare2002/OrderBook-Go/src/users"
)

type Position int
type Quanity int
type Price float64

const (
	BUY Position = iota
	SELL
)


type Order struct {
	User     *user.User
	Position Position   `json:"position"`
	Quantity Quanity    `json:"quantity"`
	Price    Price      `json:"price"`
	Ticker   user.Asset `json:"ticker"`
}

type OrderSnapshot struct {
	UserId   uuid.UUID  `json:"userId"`
	Position Position   `json:"position"`
	Quantity Quanity    `json:"quantity"`
	Price    Price      `json:"price"`
	Ticker   user.Asset `json:"ticker"`
}

type Bid struct { *Order }
func (b Bid) Less(t btree.Item) bool {
	return b.Price > t.(Bid).Price
}

type Ask struct { *Order }
func(a Ask) Less(t btree.Item) bool {
	return a.Price < t.(Ask).Price
}

func NewOrder(user *user.User, position Position, quantity Quanity, price Price, ticker user.Asset) Order {
	return Order{user, position, quantity, price, ticker}
}

func (o Order) Snapshot() OrderSnapshot {
	var uid uuid.UUID
	if o.User != nil {
		uid = o.User.UserId
	}
	return OrderSnapshot{
		UserId:   uid,
		Position: o.Position,
		Quantity: o.Quantity,
		Price:    o.Price,
		Ticker:   o.Ticker,
	}
}
