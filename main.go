package main

import (
	"arbitrade/db"
	"arbitrade/exchanges"
	"arbitrade/exchanges/generics"
	"log"
)

func main() {
	exchange, err := exchanges.New("cryptopia")
	if err != nil {
		log.Fatal(err)
	}
	_, err = exchange.GetSymbols()
	if err != nil {
		log.Fatal(err)
	}
	_, err = exchange.GetMarkets()
	if err != nil {
		log.Fatal(err)
	}
	manager, err := db.NewManager([]generics.ExchangeAPI{exchange})
	if err != nil {
		log.Fatal(err)
	}
	manager.Start()
	defer manager.Close()
}
