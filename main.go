package main

import (
	"arbitrade/exchanges"
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
	orderBooks, err := exchange.GetOrderBooks()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(orderBooks)
}
