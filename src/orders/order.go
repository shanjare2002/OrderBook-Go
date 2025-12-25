package orders
import("github.com/google/btree")
type Position int
type Quanity int
type Price float64

const (
	BUY Position = iota
	SELL
)

type Order struct {
	Position Position `json:"position"`
	Quantity Quanity  `json:"quantity"`
	Price    Price    `json:"price"`
}

type Bid struct {Order}
func (b Bid) Less(t btree.Item) bool {
	return b.Price > t.(Bid).Price
}

type Ask struct {Order}
func(a Ask) Less(t btree.Item) bool {
	return a.Price < t.(Ask).Price
}


func NewOrder(position Position, quantity Quanity, price Price) Order {
	return Order{position, quantity, price}
}
