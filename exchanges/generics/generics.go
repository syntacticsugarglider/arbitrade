package generics

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

//Order represents an unfulfilled order
type Order struct {
	Price  float64
	Volume float64
}

//OrderBook represents a cryptocurrency market order book
type OrderBook struct {
	Buy       []Order
	Sell      []Order
	Timestamp time.Time
}

//Symbol represents a cryptocurrency symbol
type Symbol struct {
	Symbol       string
	WithdrawFee  float64
	WithdrawMin  float64
	MinBaseTrade float64
	Data         ResponseData
	Active       bool
	Name         string
}

//Market represents a cryptocurrency market
type Market struct {
	SymbolPair   string
	Orders       *OrderBook
	Data         ResponseData
	Active       bool
	Symbol       *Symbol
	BaseSymbol   *Symbol
	MinTrade     float64
	MinBaseTrade float64
	MaxTrade     float64
	MaxBaseTrade float64
	TradeFee     float64
	MinPrice     float64
	MaxPrice     float64
}

//ResponseData represents a semantic marker for the full API response data of a call
type ResponseData interface{}

//ExchangeAPI represents the operations performable on a cryptocurrency exchange
type ExchangeAPI interface {
	GetMarkets() (map[string]*Market, error)
	GetSymbols() (map[string]*Symbol, error)
	GetOrderBooks(...[]string) (map[string]*OrderBook, error)
	GetOrderBook(string) (*OrderBook, error)
	GetExchange() *Exchange
}

//Exchange represents all data stored that pertains to a cryptocurrency exchange
type Exchange struct {
	Name       string
	Markets    map[string]*Market
	Symbols    map[string]*Symbol
	OrderBooks map[string]*OrderBook
}

//Fetch wraps an HTTP request and unmarshals the result
func Fetch(url string, target interface{}) error {
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, target)
	if err != nil {
		return err
	}
	return nil
}
