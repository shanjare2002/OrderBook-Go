package main

import (
	"net/http"
	"log"
	endpoints "github.com/shanjare2002/OrderBook-Go/src/api"
)

func main(){
	defer func() {
        if r := recover(); r != nil {
            log.Println("PANIC:", r)
        }
    }()
	service := endpoints.NewService()
	
	http.ListenAndServe(":8080", service)
}