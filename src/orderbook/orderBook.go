package orderbook

import (
	"fmt"
	"strings"
	"github.com/google/btree"
	"github.com/shanjare2002/OrderBook-Go/src/orders"
)

type OrderBook struct {
	Bids *btree.BTree
	Asks *btree.BTree
}

type BookSnapshot struct {
	Bids []orders.Order `json:"bids"`
	Asks []orders.Order `json:"asks"`
}

func NewOrderBook() *OrderBook {
	bids := btree.New(16)
	asks := btree.New(16)
	return &OrderBook{
		Bids: bids,
		Asks: asks,
	}
}

func (orderbook OrderBook) String() string {
	var bids strings.Builder
	var asks strings.Builder

	orderbook.Bids.Ascend(func(it btree.Item) bool {
		currBid := it.(orders.Bid).Order
		fmt.Fprintf(&bids, "Bid at %v for quantity: %v \n", currBid.Price, currBid.Quantity)
		return true
	})
	orderbook.Asks.Ascend(func(it btree.Item) bool {
		currAsk := it.(orders.Ask).Order
		fmt.Fprintf(&asks, "Ask at %v for quantity: %v \n", currAsk.Price, currAsk.Quantity)
		return true
	})
	return fmt.Sprintf("Asks: \n%s \nBids:\n%s", asks.String(), bids.String())
}
func (orderbook *OrderBook) fillBid(newBid *orders.Order) {
	orderbook.Asks.Ascend(func(it btree.Item) bool {
		currAsk := it.(orders.Ask)
		if newBid.Price < currAsk.Order.Price{
			return false
		}
		if currAsk.Order.Quantity > newBid.Quantity{
			currAsk.Order.Quantity -= newBid.Quantity
			orderbook.Asks.ReplaceOrInsert(currAsk)
			newBid.Quantity = 0
			return false
		}
		newBid.Quantity -= currAsk.Order.Quantity
		orderbook.Asks.Delete(currAsk)
		return true 
	})
	if newBid.Quantity > 0 {
		orderbook.addBid(newBid)
	}

}
func (orderbook *OrderBook) fillAsk(newAsk *orders.Order) {
	orderbook.Bids.Ascend(func(it btree.Item) bool {
		currBid:= it.(orders.Bid)
		if newAsk.Price > currBid.Order.Price{
			return false
		}
		if currBid.Order.Quantity > newAsk.Quantity{
			currBid.Order.Quantity -= newAsk.Quantity
			orderbook.Bids.ReplaceOrInsert(currBid)
			newAsk.Quantity = 0
			return false
		}
		newAsk.Quantity -= currBid.Order.Quantity
		orderbook.Bids.Delete(currBid)
		return true 
	})
	if newAsk.Quantity > 0 {
		orderbook.addAsk(newAsk)
	}

}

func (orderbook *OrderBook) NewOrder(newOrder *orders.Order){
	switch newOrder.Position {
	case orders.BUY:
		orderbook.fillBid(newOrder)
	case orders.SELL:
		orderbook.fillAsk(newOrder)
	}
	
}

func (ob OrderBook) Snapshot() BookSnapshot {
	snap := BookSnapshot{}
	ob.Asks.Ascend(func(it btree.Item) bool {
		snap.Asks = append(snap.Asks, it.(orders.Ask).Order)
		return true
	})
	ob.Bids.Ascend(func(it btree.Item) bool {
		snap.Bids = append(snap.Bids, it.(orders.Bid).Order)
		return true
	})
	return snap
}

func (orderbook *OrderBook) addBid(order *orders.Order) {
	key := orders.Bid{Order: orders.Order{Price: order.Price}}
	existing := orderbook.Bids.Get(key)

	if existing != nil {
		bid := existing.(orders.Bid)
		bid.Order.Quantity += order.Quantity
		orderbook.Bids.ReplaceOrInsert(bid)
	} else {
		newBid := orders.Bid{Order: orders.Order{Position: order.Position, Quantity: order.Quantity, Price: order.Price}}
		orderbook.Bids.ReplaceOrInsert(newBid)
	}

}

func (orderbook *OrderBook) addAsk(order *orders.Order) {

	key := orders.Ask{Order: orders.Order{Price: order.Price}}
	existing := orderbook.Asks.Get(key)

	if existing != nil {
		ask := existing.(orders.Ask)
		ask.Order.Quantity += order.Quantity
		orderbook.Asks.ReplaceOrInsert(ask)
	} else {
		orderbook.Asks.ReplaceOrInsert(orders.Ask{Order: *order}) // O(log n)
	}
}
