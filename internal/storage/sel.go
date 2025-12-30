package storage

import (
	"context"
	"fmt"

	"github.com/PowerDNS/lmdb-go/lmdb"

	"github.com/luiz-simples/keyp.git/internal/domain"
)

func (client *Client) sel(ctx context.Context) (lmdb.DBI, error) {
	db, _ := ctx.Value(domain.DB).(uint8)
	var err error = nil

	if !client.hasDB(db) {
		err = client.openDB(db)
	}

	return client.dbi[db], err
}

func (client *Client) hasDB(db uint8) bool {
	client.mtx.RLock()
	defer client.mtx.RUnlock()
	_, hasDB := client.dbi[db]
	return hasDB
}

func (client *Client) openDB(db uint8) error {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	return client.env.Update(func(txn *lmdb.Txn) error {
		name := fmt.Sprintf("db_%d", db)
		dbi, err := txn.OpenDBI(name, lmdb.Create)
		client.dbi[db] = dbi
		return err
	})
}
