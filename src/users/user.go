package user

import "github.com/google/uuid"

type Asset string

type BalanceRequest struct {
	Asset  Asset `json:"asset"`
	Amount float64 `json:"amount"`
}

type User struct { 
	UserId uuid.UUID  `json:"userId"`
	Balance map[Asset]float64 `json:"balance"`
}

func (user *User) AddBalance(asset Asset, amount float64) {
	
	_, contains := user.Balance[asset]
	if contains == true {
		user.Balance[asset] += amount
	}else{
		user.Balance[asset] = amount
	}
}


func NewUser() User {
	newId := uuid.New()
	emptyMap := make(map[Asset]float64)
	return User{newId, emptyMap}
}