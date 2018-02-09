package db

import (
	"encoding/json"
	"errors"

	"github.com/dgraph-io/dgraph/protos/api"
)

//Mutate mutates the data in the manager with the provided structure
func (m *Manager) Mutate(data interface{}) (api.Assigned, error) {
	switch data.(type) {
	case []ExchangeNode:
	case []OrderNode:
	case []MarketNode:
	default:
		return api.Assigned{}, errors.New("Invalid type passed to manager Mutate")
	}
	tx := m.database.NewTxn()
	defer tx.Discard(m.context)
	mu := &api.Mutation{
		CommitNow: true,
	}
	pb, err := json.Marshal(data)
	if err != nil {
		return api.Assigned{}, err
	}
	mu.SetJson = pb
	assigned, err := tx.Mutate(m.context, mu)
	if err != nil {
		return api.Assigned{}, err
	}
	return *assigned, nil
}
