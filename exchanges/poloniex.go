package poloniex

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

//Symbol represents a single symbol returned from the Poloniex API
type Symbol struct {
	id                   int
	name                 string
	txFee                string
	minConf              string
	depositAddress       float64
	disabled             int
	delisted         	 int
}

//Market represents a single market returned from the Poloniex API
type Market struct {
	id		       int
	last     	   string
	lowestAsk	   string
	highestBid	   string
	percentChange  string
	baseVolume	   string
	quoteVolume	   string
	isFrozen	   string
}

//TradePair represents a single trade pair returned from the Poloniex API
type TradePair struct {
	id               int
	last     	   string
	lowestAsk	   string
	highestBid	   string
	percentChange  string
	baseVolume	   string
	quoteVolume	   string
	isFrozen	   string
}

//Symbols represents all Poloniex symbols
type Symbols struct {
	Success bool
	Message string
	Data    []Symbol
}

//Markets represents all Poloniex markets
type Markets struct {
	Success bool
	Message string
	Data    []Market
}

//TradePairs represents all Poloniex trade pairs
type TradePairs struct {
	Success bool
	Message string
	Data    []TradePair
}

//Order represents a Poloniex buy or sell order
type Order struct {
	label       string
	asks        float64
	bids        float64
	seq         string
}

//OrderBookData represents the two-part order book structure from a single-book request to the Poloniex API
type OrderBookData struct {
	bids  []Order
	asks []Order
}

//OrderBooksData represents the order book structure from a multi-book request to the Poloniex API
type OrderBooksData struct {
	Market      string
	bids         []Order
	asks         []Order
}

//OrderBook represents a single-order-book Poloniex API request
type OrderBook struct {
	Success bool
	Message string
	Data    OrderBookData
}

//OrderBooks represents a multi-order-book Poloniex API request
type OrderBooks struct {
	Success bool
	Message string
	Data    []OrderBooksData
}

//Poloniex represents the Poloniex API and exchange structure
type Poloniex struct {
	*generics.Exchange
}

//returnTicker returns and internally updates the market data of the Poloniex exchange
func (c *Poloneix) returnTicker() (map[string]*generics.Market, error) {
	if c.Symbols == nil || len(c.Symbols) == 0 {
		return nil, errors.New("Exchange symbol data not populated")
	}
	_markets := *new(Markets)
	err := generics.Fetch("https://poloniex.com/public?command=returnTicker", &_markets)
	if err != nil {
		return nil, err
	}s   
	if !_markets.Success {
		return nil, errors.New("Market request to Poloniex returned with declared failure")
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
	err = generics.Fetch("https://poloniex.com/public?command=returnTicker", &_tradepairs)
	if err != nil {
		return nil, err
	}
	if !_tradepairs.Success {
		return nil, errors.New("Trade pairs request to Poloniex returned with declared failure")
	}
	for _, pair := range _tradepairs.Data {
		_ = TradePair{}
		_, ok := c.Symbols[pair.Symbol]
		if ok {
			_, ok = c.Symbols[pair.BaseSymbol]
		}
		markets[pair.Label].Active = 0 && markets[pair.Label].delisted && pair.Status == 0 && c.Symbols[pair.Symbol].isFrozen && c.Symbols[pair.BaseSymbol].Active
		markets[pair.Label].MinBaseTrade = pair.baseVolume
		markets[pair.Label].MaxBaseTrade = pair.quoteVolume
		markets[pair.Label].Symbol = c.Symbols[pair.Symbol]
		markets[pair.Label].BaseSymbol = c.Symbols[pair.BaseSymbol]
		markets[pair.Label].TradeFee = pair.txFee
		markets[pair.Label].MaxPrice = pair.highestBid
		markets[pair.Label].MinPrice = pair.lowestAsk
	}
	c.Markets = markets
	return markets, nil
}

//returnCurrencies returns and internally updates the symbol data of the Poloniex exchange

func (c *Poloniex) returnCurrencies() (map[string]*generics.Symbol, error) {
	_symbols := *new(Symbols)
	err := generics.Fetch("https://poloniex.com/public?command=returnCurrencies", &_symbols)
	if err != nil {
		return nil, err
	}
	if !_symbols.Success {
		return nil, errors.New("Symbols request to Poloniex returned with declared failure")
	}
	symbols := map[string]*generics.Symbol{}
	for _, symbol := range _symbols.Data {
		symbols[symbol.Symbol] = &generics.Symbol{
			Data:         symbol,
			WithdrawFee:  symbol.txFee,
			MinBaseTrade: symbol.minConfig
			Symbol:       symbol.name,
			Active:       symbol.delisted == 0 && symbol.isFrozen == 0,
		}
	}
	c.Symbols = symbols
	return symbols, nil

}

//GetOrderBooks returns and internally updates the order book data of the Poloniex exchange
func (c *Poloniex) returnOrderBooks(t ...[]string) (map[string]*generics.OrderBook, error) {
	_targets := t
	targets := []string{}
	if len(_targets) > 1 {
		return nil, errors.New("Too many arguments passed to Poloniex returnOrderBooks")
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
			return nil, errors.New("Invalid symbol passed to Poloniex returnOrderBooks")
		}
		ids = append(ids, id.Data.(Market).TradePairId)
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
		err = generics.Fetch(fmt.Sprintf("https://poloniex.com/public?command=returnOrderBook&currencyPair=/%s&depth=100", strings.Join(_ids, "-")), &_orderBooks)
		if err != nil {
			return
		}
		if !_orderBooks.Success {
			err = errors.New("Order books request to Poloniex returned with declared failure")
			return
		}
		timestamp := time.Now().UTC()
		for _, book := range _orderBooks.Data {
			_ = OrderBooksData{}
			c.OrderBooks[strings.Replace(book.Market, "_", "/", -1)].Timestamp = timestamp
			c.OrderBooks[strings.Replace(book.Market, "_", "/", -1)].asks = []generics.Order{}
			for _, buy := range book.Buy {
				c.OrderBooks[strings.Replace(book.Market, "_", "/", -1)].asks = append(c.OrderBooks[strings.Replace(book.Market, "_", "/", -1)].Buy, generics.Order{
					Price:  bids.Price,
					Volume: bids.quoteVolume,
				})
			}
			c.OrderBooks[strings.Replace(book.Market, "_", "/", -1)].bids = []generics.Order{}
			for _, sell := range book.Sell {
				c.OrderBooks[strings.Replace(book.Market, "_", "/", -1)].bids = append(c.OrderBooks[strings.Replace(book.Market, "_", "/", -1)].Sell, generics.Order{
					Price:  asks.Price,
					Volume: asks.quoteVolume,
				})
			}
			orderBooks[strings.Replace(book.Market, "_", "/", -1)] = c.OrderBooks[strings.Replace(book.Market, "_", "/", -1)]
		}
		done <- struct{}{}
	}(done)
	if len(_rids) > 1 {
		rOrderBooks, err := c returnOrderBooks(_rids)
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
*/
//GetOrderBook returns and internally updates the order book data of a single market within the Poloniex exchange
func (c *Poloniex) returnOrderBook(symbol string) (*generics.OrderBook, error) {
	_, ok := c.Markets[symbol]
	if !ok {
		return nil, errors.New("Invalid symbol passed to Poloneix GetOrderBook")
	}
	_orderBook := *new(OrderBook)
	err := generics.Fetch(fmt.Sprintf("https://poloniex.com/public?command=returnOrderBook&currencyPair=/%d&depth=1", c.Markets[symbol].Data.(Market).TradePairId), &_orderBook)
	if err != nil || !_orderBook.Success {
		if err != nil {
			return nil, err
		}
		return nil, errors.New("Order book request to Poloniex returned with declared failure")
	}
	c.Markets[symbol].Orders.Timestamp = time.Now().UTC()
	_done := make(chan struct{})
	go func(_done chan struct{}) {
		for _, value := range _orderBook.Data.bids {
			c.Markets[symbol].Orders.bids = append(c.Markets[symbol].Orders.bids, generics.Order{
				Price:  value.Price,
				Volume: value.Volume,
			})
		}
		_done <- struct{}{}
	}(_done)
	done := make(chan struct{})
	go func(done chan struct{}) {
		for _, value := range _orderBook.Data.asks {
			c.Markets[symbol].Orders.asks = append(c.Markets[symbol].Orders.asks, generics.Order{
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

//GetExchange returns the internal exchange object of the Poloniex
//func (c *Poloniex) GetExchange() *generics.Exchange {
//	return c.Exchange
}
