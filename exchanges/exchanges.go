package exchanges

import (
	"arbitrade/exchanges/cryptopia"
	"arbitrade/exchanges/generics"
	"errors"
)

//New constructs a new exchange
func New(id string) (generics.ExchangeAPI, error) {
	switch id {
	case "cryptopia":
		exptr := new(cryptopia.Cryptopia)
		exptr.Exchange = new(generics.Exchange)
		exptr.OrderBooks = map[string]*generics.OrderBook{}
		return exptr, nil
	}
	return nil, errors.New("Invalid exchange specifier")
}

//NewMultiple constructs a list of new exchanges
func NewMultiple(id []string) ([]generics.ExchangeAPI, error) {
	exchanges := []generics.ExchangeAPI{}
	for _, _id := range id {
		switch _id {
		case "cryptopia":
			exptr := new(cryptopia.Cryptopia)
			exptr.Exchange = new(generics.Exchange)
			exptr.OrderBooks = map[string]*generics.OrderBook{}
			exchanges = append(exchanges, exptr)
		default:
			return exchanges, errors.New("Invalid exchange specifier")
		}
	}
	return exchanges, nil
}
