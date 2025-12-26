package orderbook

import (
	"fmt"
	"strings"

	"github.com/google/btree"
	"github.com/google/uuid"
	"github.com/shanjare2002/OrderBook-Go/src/orders"
	user "github.com/shanjare2002/OrderBook-Go/src/users"
)

type OrderBook struct {
	Users []*user.User
	Bids  *btree.BTree
	Asks  *btree.BTree
}

type BookSnapshot struct {
	Bids []orders.OrderSnapshot `json:"bids"`
	Asks []orders.OrderSnapshot `json:"asks"`
}

type TopOfBook struct {
	Buy  *orders.OrderSnapshot `json:"buy,omitempty"`
	Sell *orders.OrderSnapshot `json:"sell,omitempty"`
}

func NewOrderBook() *OrderBook {
	bids := btree.New(16)
	asks := btree.New(16)
	return &OrderBook{
		Bids: bids,
		Asks: asks,
	}
}

func (orderbook *OrderBook) GetUser(userId uuid.UUID) (*user.User, error) {
	for _, User := range orderbook.Users {
		if User.UserId == userId {
			return User, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

func (orderbook *OrderBook) RegisterUser() user.User {
	newUser := user.NewUser()
	orderbook.Users = append(orderbook.Users, &newUser)
	return newUser
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
		if newBid.Price < currAsk.Order.Price {
			return false
		}
		if currAsk.Order.Quantity > newBid.Quantity {
			currAsk.Order.Quantity -= newBid.Quantity
			orderbook.Asks.ReplaceOrInsert(currAsk)
			orderbook.swapAssests(newBid.User, currAsk.User, currAsk.Price, currAsk.Quantity, currAsk.Ticker)
			newBid.Quantity = 0
			return false
		}
		newBid.Quantity -= currAsk.Order.Quantity
		orderbook.swapAssests(newBid.User, currAsk.User, currAsk.Price, currAsk.Quantity, currAsk.Ticker)
		orderbook.Asks.Delete(currAsk)
		return true
	})
	if newBid.Quantity > 0 {
		orderbook.addBid(newBid)
	}

}

func (orderbook *OrderBook) swapAssests(buyer *user.User, seller *user.User, price orders.Price, quantity orders.Quanity, asset user.Asset) {
	usd := float64(price) * float64(quantity)
	buyer.Balance["USD"] -= usd
	seller.Balance["USD"] += usd
	buyer.Balance[asset] += float64(quantity)
	seller.Balance[asset] -= float64(quantity)
}

func (orderbook *OrderBook) fillAsk(newAsk *orders.Order) {
	orderbook.Bids.Ascend(func(it btree.Item) bool {
		currBid := it.(orders.Bid)
		if newAsk.Price > currBid.Order.Price {
			return false
		}
		if currBid.Order.Quantity > newAsk.Quantity {
			currBid.Order.Quantity -= newAsk.Quantity
			orderbook.swapAssests(currBid.User, newAsk.User, currBid.Price, currBid.Quantity, currBid.Ticker)
			orderbook.Bids.ReplaceOrInsert(currBid)
			newAsk.Quantity = 0
			return false
		}
		newAsk.Quantity -= currBid.Order.Quantity
		orderbook.swapAssests(currBid.User, newAsk.User, currBid.Price, currBid.Quantity, currBid.Ticker)
		orderbook.Bids.Delete(currBid)
		return true
	})
	if newAsk.Quantity > 0 {
		orderbook.addAsk(newAsk)
	}

}

func (orderbook *OrderBook) NewOrder(newOrder *orders.Order) {
	switch newOrder.Position {
	case orders.BUY:
		orderbook.fillBid(newOrder)
	case orders.SELL:
		orderbook.fillAsk(newOrder)
	}

}

func (ob *OrderBook) Snapshot() BookSnapshot {
	snap := BookSnapshot{
		Asks: make([]orders.OrderSnapshot, 0, ob.Asks.Len()),
		Bids: make([]orders.OrderSnapshot, 0, ob.Bids.Len()),
	}

	ob.Asks.Ascend(func(it btree.Item) bool {
		snap.Asks = append(snap.Asks, it.(orders.Ask).Order.Snapshot())
		return true
	})
	ob.Bids.Ascend(func(it btree.Item) bool {
		snap.Bids = append(snap.Bids, it.(orders.Bid).Order.Snapshot())
		return true
	})
	return snap
}

func (ob *OrderBook) Top() TopOfBook {
	res := TopOfBook{}
	ob.Asks.Ascend(func(it btree.Item) bool {
		snap := it.(orders.Ask).Order.Snapshot()
		res.Buy = &snap
		return false
	})
	ob.Bids.Ascend(func(it btree.Item) bool {
		snap := it.(orders.Bid).Order.Snapshot()
		res.Sell = &snap
		return false 
	})
	return res
}

func (orderbook *OrderBook) addBid(order *orders.Order) {
	key := orders.Bid{Order: &orders.Order{Price: order.Price}}
	existing := orderbook.Bids.Get(key)

	if existing != nil {
		bid := existing.(orders.Bid)
		bid.Order.Quantity += order.Quantity
		orderbook.Bids.ReplaceOrInsert(bid)
	} else {
		orderbook.Bids.ReplaceOrInsert(orders.Bid{Order: order})
	}

}

func (orderbook *OrderBook) addAsk(order *orders.Order) {

	key := orders.Ask{Order: &orders.Order{Price: order.Price}}
	existing := orderbook.Asks.Get(key)

	if existing != nil {
		ask := existing.(orders.Ask)
		ask.Order.Quantity += order.Quantity
		orderbook.Asks.ReplaceOrInsert(ask)
	} else {
		orderbook.Asks.ReplaceOrInsert(orders.Ask{Order: order}) // O(log n)
	}
}
