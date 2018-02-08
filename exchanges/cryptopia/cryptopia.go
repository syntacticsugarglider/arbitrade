package cryptopia

import (
	"arbitrade/exchanges/generics"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

//Symbol represents a single symbol returned from the Cryptopia API
type Symbol struct {
	ID                   int
	Name                 string
	Symbol               string
	Algorithm            string
	WithdrawFee          float64
	MinWithdraw          float64
	MinBaseTrade         float64
	IsTipEnabled         bool
	MinTip               float64
	DepositConfirmations int
	Status               string
	StatusMessage        string
	ListingStatus        string
}

//Market represents a single market returned from the Cryptopia API
type Market struct {
	TradePairID    int32 `json:"TradePairID"`
	Label          string
	AskPrice       float64
	BidPrice       float64
	Low            float64
	High           float64
	Volume         float64
	LastPrice      float64
	BuyVolume      float64
	SellVolume     float64
	Change         float64
	Open           float64
	Close          float64
	BaseVolume     float64
	BaseBuyVolume  float64
	BaseSellVolume float64
}

//TradePair represents a single trade pair returned from the Cryptopia API
type TradePair struct {
	ID               int32 `json:"Id"`
	Label            string
	Currency         string
	Symbol           string
	BaseCurrency     string
	BaseSymbol       string
	Status           string
	StatusMessage    string
	TradeFee         float64
	MinimumTrade     float64
	MaximumTrade     float64
	MinimumBaseTrade float64
	MaximumBaseTrade float64
	MinimumPrice     float64
	MaximumPrice     float64
}

//Symbols represents all Cryptopia symbols
type Symbols struct {
	Success bool
	Message string
	Data    []Symbol
}

//Markets represents all Cryptopia markets
type Markets struct {
	Success bool
	Message string
	Data    []Market
}

//TradePairs represents all Cryptopia trade pairs
type TradePairs struct {
	Success bool
	Message string
	Data    []TradePair
}

//Order represents a Cryptopia buy or sell order
type Order struct {
	TradePairID int32 `json:"TradePairId"`
	Label       string
	Price       float64
	Volume      float64
	Total       float64
}

//OrderBookData represents the two-part order book structure from a single-book request to the Cryptopia API
type OrderBookData struct {
	Buy  []Order
	Sell []Order
}

//OrderBooksData represents the order book structure from a multi-book request to the Cryptopia API
type OrderBooksData struct {
	TradePairID int32 `json:"tradePairId"`
	Market      string
	Buy         []Order
	Sell        []Order
}

//OrderBook represents a single-order-book Cryptopia API request
type OrderBook struct {
	Success bool
	Message string
	Data    OrderBookData
}

//OrderBooks represents a multi-order-book Cryptopia API request
type OrderBooks struct {
	Success bool
	Message string
	Data    []OrderBooksData
}

//Cryptopia represents the Cryptopia API and exchange structure
type Cryptopia struct {
	*generics.Exchange
}

//GetMarkets returns and internally updates the market data of the Cryptopia exchange
func (c *Cryptopia) GetMarkets() (map[string]*generics.Market, error) {
	if c.Symbols == nil || len(c.Symbols) == 0 {
		return nil, errors.New("Exchange symbol data not populated")
	}
	_markets := *new(Markets)
	err := generics.Fetch("https://www.cryptopia.co.nz/api/GetMarkets", &_markets)
	if err != nil {
		return nil, err
	}
	if !_markets.Success {
		return nil, errors.New("Market request to Cryptopia returned with declared failure")
	}
	markets := map[string]*generics.Market{}
	for _, market := range _markets.Data {
		markets[market.Label] = &generics.Market{
			SymbolPair: market.Label,
			Data:       market,
			Active:     true,
			Orders: &generics.OrderBook{
				Buy:  []generics.Order{},
				Sell: []generics.Order{},
			},
		}
		c.OrderBooks[market.Label] = markets[market.Label].Orders
	}
	_tradepairs := *new(TradePairs)
	err = generics.Fetch("https://www.cryptopia.co.nz/api/GetTradePairs", &_tradepairs)
	if err != nil {
		return nil, err
	}
	if !_tradepairs.Success {
		return nil, errors.New("Trade pairs request to Cryptopia returned with declared failure")
	}
	for _, pair := range _tradepairs.Data {
		_ = TradePair{}
		_, ok := c.Symbols[pair.Symbol]
		if ok {
			_, ok = c.Symbols[pair.BaseSymbol]
		}
		markets[pair.Label].Active = ok && markets[pair.Label].Active && pair.Status == "OK" && c.Symbols[pair.Symbol].Active && c.Symbols[pair.BaseSymbol].Active
		markets[pair.Label].MinTrade = pair.MinimumTrade
		markets[pair.Label].MaxTrade = pair.MaximumTrade
		markets[pair.Label].MinBaseTrade = pair.MinimumBaseTrade
		markets[pair.Label].MaxBaseTrade = pair.MaximumBaseTrade
		markets[pair.Label].Symbol = c.Symbols[pair.Symbol]
		markets[pair.Label].BaseSymbol = c.Symbols[pair.BaseSymbol]
		markets[pair.Label].TradeFee = pair.TradeFee
		markets[pair.Label].MaxPrice = pair.MaximumPrice
		markets[pair.Label].MinPrice = pair.MinimumPrice
	}
	c.Markets = markets
	return markets, nil
}

//GetSymbols returns and internally updates the symbol data of the Cryptopia exchange
func (c *Cryptopia) GetSymbols() (map[string]*generics.Symbol, error) {
	_symbols := *new(Symbols)
	err := generics.Fetch("https://www.cryptopia.co.nz/api/GetCurrencies", &_symbols)
	if err != nil {
		return nil, err
	}
	if !_symbols.Success {
		return nil, errors.New("Symbols request to Cryptopia returned with declared failure")
	}
	symbols := map[string]*generics.Symbol{}
	for _, symbol := range _symbols.Data {
		symbols[symbol.Symbol] = &generics.Symbol{
			Data:         symbol,
			WithdrawFee:  symbol.WithdrawFee,
			WithdrawMin:  symbol.MinWithdraw,
			MinBaseTrade: symbol.MinBaseTrade,
			Symbol:       symbol.Symbol,
			Active:       symbol.Status == "OK" && symbol.ListingStatus == "Active",
			Name:         symbol.Name,
		}
	}
	c.Symbols = symbols
	return symbols, nil
}

//GetOrderBooks returns and internally updates the order book data of the Cryptopia exchange
func (c *Cryptopia) GetOrderBooks(t ...[]string) (map[string]*generics.OrderBook, error) {
	_targets := t
	targets := []string{}
	if len(_targets) > 1 {
		return nil, errors.New("Too many arguments passed to Cryptopia GetOrderBooks")
	}
	if len(_targets) == 0 {
		targets = make([]string, len(c.Markets))
		i := 0
		for k := range c.Markets {
			targets[i] = k
			i++
		}
	} else {
		targets = _targets[0]
	}
	ids := []int32{}
	for _, target := range targets {
		id, ok := c.Markets[target]
		if !ok {
			return nil, errors.New("Invalid symbol passed to Cryptopia GetOrderBooks")
		}
		ids = append(ids, id.Data.(Market).TradePairID)
	}
	_ids := []string{}
	for _, _id := range ids {
		_ids = append(_ids, strconv.FormatInt(int64(_id), 10))
	}
	_orderBooks := *new(OrderBooks)
	_rids := []string{}
	if len(_ids) > 29 {
		_rids = targets[29:]
		_ids = _ids[0:29]
	}
	err := errors.New("")
	done := make(chan struct{})
	orderBooks := map[string]*generics.OrderBook{}
	go func(chan struct{}) {
		err = generics.Fetch(fmt.Sprintf("https://www.cryptopia.co.nz/api/GetMarketOrderGroups/%s", strings.Join(_ids, "-")), &_orderBooks)
		if err != nil {
			return
		}
		if !_orderBooks.Success {
			err = errors.New("Order books request to Cryptopia returned with declared failure")
			return
		}
		timestamp := time.Now().UTC()
		for _, book := range _orderBooks.Data {
			_ = OrderBooksData{}
			c.OrderBooks[strings.Replace(book.Market, "_", "/", -1)].Timestamp = timestamp
			c.OrderBooks[strings.Replace(book.Market, "_", "/", -1)].Buy = []generics.Order{}
			for _, buy := range book.Buy {
				c.OrderBooks[strings.Replace(book.Market, "_", "/", -1)].Buy = append(c.OrderBooks[strings.Replace(book.Market, "_", "/", -1)].Buy, generics.Order{
					Price:  buy.Price,
					Volume: buy.Volume,
				})
			}
			c.OrderBooks[strings.Replace(book.Market, "_", "/", -1)].Sell = []generics.Order{}
			for _, sell := range book.Sell {
				c.OrderBooks[strings.Replace(book.Market, "_", "/", -1)].Sell = append(c.OrderBooks[strings.Replace(book.Market, "_", "/", -1)].Sell, generics.Order{
					Price:  sell.Price,
					Volume: sell.Volume,
				})
			}
			orderBooks[strings.Replace(book.Market, "_", "/", -1)] = c.OrderBooks[strings.Replace(book.Market, "_", "/", -1)]
		}
		done <- struct{}{}
	}(done)
	if len(_rids) > 1 {
		rOrderBooks, err := c.GetOrderBooks(_rids)
		for k, v := range rOrderBooks {
			orderBooks[k] = v
		}
		if err != nil {
			return nil, err
		}
		<-done
		if err != nil {
			return nil, err
		}
		return orderBooks, nil
	}
	<-done
	if err != nil {
		return nil, err
	}
	return orderBooks, nil
}

//GetOrderBook returns and internally updates the order book data of a single market within the Cryptopia exchange
func (c *Cryptopia) GetOrderBook(symbol string) (*generics.OrderBook, error) {
	_, ok := c.Markets[symbol]
	if !ok {
		return nil, errors.New("Invalid symbol passed to Cryptopia GetOrderBook")
	}
	_orderBook := *new(OrderBook)
	err := generics.Fetch(fmt.Sprintf("https://www.cryptopia.co.nz/api/GetMarketOrders/%d", c.Markets[symbol].Data.(Market).TradePairID), &_orderBook)
	if err != nil || !_orderBook.Success {
		if err != nil {
			return nil, err
		}
		return nil, errors.New("Order book request to Cryptopia returned with declared failure")
	}
	c.Markets[symbol].Orders.Timestamp = time.Now().UTC()
	_done := make(chan struct{})
	go func(_done chan struct{}) {
		for _, value := range _orderBook.Data.Sell {
			c.Markets[symbol].Orders.Sell = append(c.Markets[symbol].Orders.Sell, generics.Order{
				Price:  value.Price,
				Volume: value.Volume,
			})
		}
		_done <- struct{}{}
	}(_done)
	done := make(chan struct{})
	go func(done chan struct{}) {
		for _, value := range _orderBook.Data.Buy {
			c.Markets[symbol].Orders.Buy = append(c.Markets[symbol].Orders.Buy, generics.Order{
				Price:  value.Price,
				Volume: value.Volume,
			})
		}
		done <- struct{}{}
	}(done)
	<-_done
	<-done
	return c.Markets[symbol].Orders, nil
}

//GetExchange returns the internal exchange object of the Cryptopia
func (c *Cryptopia) GetExchange() *generics.Exchange {
	return c.Exchange
}
