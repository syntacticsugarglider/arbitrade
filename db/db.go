package db

import (
	"arbitrade/exchanges/generics"
	"errors"
	"strconv"
	"strings"

	"golang.org/x/net/context"

	"github.com/dgraph-io/dgraph/client"
	"github.com/dgraph-io/dgraph/protos/api"
	"google.golang.org/grpc"
)

//OrderType represents an enumerated order buy/sell type
type OrderType int

const (
	//Buy represents the type of a buy order
	Buy OrderType = iota
	//Sell represents the type of a sell order
	Sell
)

//OrderNode represents a market node within the dgraph instance
type OrderNode struct {
	UID    string     `json:"uid,omitempty"`
	Price  float64    `json:"price,omitempty"`
	Volume float64    `json:"volume,omitempty"`
	Market MarketNode `json:"market,omitempty"`
	Type   OrderType  `json:"type,omitempty"`
}

//MarketNode represents a market node within the dgraph instance
type MarketNode struct {
	UID          string       `json:"uid,omitempty,omitempty"`
	MinTrade     float64      `json:"min_trade,omitempty"`
	MinBaseTrade float64      `json:"min_base_trade,omitempty"`
	MaxTrade     float64      `json:"max_trade,omitempty"`
	MaxBaseTrade float64      `json:"max_base_trade,omitempty"`
	Targets      []MarketNode `json:"targets,omitempty"`
	TradeFee     float64      `json:"trade_fee,omitempty"`
	BuyOrders    []OrderNode  `json:"buy_orders,omitempty"`
	MinPrice     float64      `json:"min_price,omitempty"`
	MaxPrice     float64      `json:"max_price,omitempty"`
	Symbol       string       `json:"symbol,omitempty"`
	Name         string       `json:"name,omitempty"`
	BaseSymbol   string       `json:"base_symbol,omitempty"`
	BaseName     string       `json:"base_name,omitempty"`
	Exchange     ExchangeNode `json:"exchange,omitempty"`
	SellOrders   []OrderNode  `json:"sell_orders,omitempty"`
}

//ExchangeNode represents an exchange node within the dgraph instance
type ExchangeNode struct {
	UID     string       `json:"uid,omitempty"`
	Name    string       `json:"name,omitempty"`
	Markets []MarketNode `json:"markets,omitempty"`
}

//Manager handles performing queries to the exchange graph and keeping the data up to date
type Manager struct {
	exchanges  []generics.ExchangeAPI
	database   *client.Dgraph
	isRunning  bool
	open       bool
	connection *grpc.ClientConn
	context    context.Context
}

//ManagerAPI represents the operations that may be performed on a Manager
type ManagerAPI interface {
	Start()
	Stop()
	Close()
	Mutate(interface{}) (api.Assigned, error)
}

//Start initializes the update loop of a Manager
func (m *Manager) Start() {
	if !m.isRunning {
		m.isRunning = true
		defer (func() {
			for _, exchange := range m.exchanges {
				go (func(e generics.ExchangeAPI) {
					for m.isRunning {
						_, err := e.GetOrderBooks()
						if err != nil {
							continue
						}
					}
				})(exchange)
			}
		})()
	}
}

//Stop stops the update loop of a Manager
func (m *Manager) Stop() {
	m.isRunning = false
}

//Close stops the update loop and closes the database connection of a Manager
func (m *Manager) Close() {
	m.isRunning = false
	m.connection.Close()
	m.database = nil
}

//NewManager creates, clears, structures, and returns a new database manager
func NewManager(exs []generics.ExchangeAPI) (ManagerAPI, error) {
	if len(exs) < 1 {
		return nil, errors.New("No exchanges passed to initalize manager")
	}
	for _, exchange := range exs {
		if (len(exchange.GetExchange().Symbols) < 1) || (len(exchange.GetExchange().Markets) < 1) {
			return nil, errors.New("One or more exchanges are uninitialized")
		}
	}
	manager := new(Manager)
	d, err := grpc.Dial("localhost:9080", grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	manager.connection = d
	manager.context = context.Background()
	manager.database = client.NewDgraphClient(api.NewDgraphClient(d))
	manager.exchanges = exs
	err = manager.database.Alter(context.Background(), &api.Operation{DropAll: true})
	if err != nil {
		return nil, err
	}
	err = manager.database.Alter(manager.context, &api.Operation{
		Schema: `
			name: string @index(exact) .
			symbol: string @index(exact) .
		`,
	})
	if err != nil {
		return nil, err
	}
	data := []ExchangeNode{}
	for _, _ex := range manager.exchanges {
		ex := _ex.GetExchange()
		_data := ExchangeNode{
			Name:    ex.Name,
			Markets: []MarketNode{},
		}
		data = append(data, _data)
	}
	assigned, err := manager.Mutate(data)
	if err != nil {
		return nil, err
	}
	update := []ExchangeNode{}
	for key, uid := range assigned.Uids {
		index, err := strconv.Atoi(strings.Split(key, "-")[1])
		if err != nil {
			return nil, errors.New("Invalid key in returned dgraph data")
		}
		data := (func() []MarketNode {
			nodes := []MarketNode{}
			for _, market := range manager.exchanges[index].GetExchange().Markets {
				if market.Active {
					nodes = append(nodes, MarketNode{
						MinTrade:     market.MinTrade,
						MinBaseTrade: market.MinBaseTrade,
						MaxTrade:     market.MaxTrade,
						MaxBaseTrade: market.MaxBaseTrade,
						Targets:      []MarketNode{},
						TradeFee:     market.TradeFee,
						BuyOrders:    []OrderNode{},
						MinPrice:     market.MinPrice,
						MaxPrice:     market.MaxPrice,
						Symbol:       market.Symbol.Symbol,
						Name:         manager.exchanges[index].GetExchange().Symbols[market.Symbol.Symbol].Name,
						BaseSymbol:   market.BaseSymbol.Symbol,
						BaseName:     manager.exchanges[index].GetExchange().Symbols[market.BaseSymbol.Symbol].Name,
						Exchange: ExchangeNode{
							UID: uid,
						},
						SellOrders: []OrderNode{},
					})
				}
			}
			return nodes
		})()
		update = append(update, ExchangeNode{
			UID:     uid,
			Markets: data,
		})
	}
	assigned, err = manager.Mutate(update)
	if err != nil {
		return nil, err
	}
	return manager, nil
}
