package storage

import (
	"errors"
	"os"
	"sync"

	"github.com/PowerDNS/lmdb-go/lmdb"
)

var (
	ErrKeyNotFound = errors.New("key not found")
)

const (
	dirPerm      = 0755
	filePerm     = 0644
	maxDatabases = 100
	mapSizeBytes = 4 << 30
	noFlags      = 0
	performFlags = lmdb.WriteMap | lmdb.NoMetaSync | lmdb.NoSync | lmdb.MapAsync | lmdb.NoReadahead
)

type (
	TTL struct {
		Expire uint32
		Cancel func()
	}

	Client struct {
		env *lmdb.Env
		dbi map[uint8]lmdb.DBI
		ttl map[uint8]map[string]*TTL
		mtx sync.RWMutex
	}
)

func NewClient(dataDir string) (*Client, error) {
	err := os.MkdirAll(dataDir, dirPerm)

	if hasError(err) {
		return nil, err
	}

	env, err := lmdb.NewEnv()

	if noError(err) {
		err = env.SetMaxDBs(maxDatabases + 1)
	}

	if noError(err) {
		err = env.SetMapSize(mapSizeBytes)
	}

	if noError(err) {
		err = env.SetMaxReaders(128)
	}

	if noError(err) {
		err = env.Open(dataDir, performFlags, filePerm)
	}

	if hasError(err) {
		env.Close()
		return nil, err
	}

	storage := &Client{env: env}
	storage.dbi = make(map[uint8]lmdb.DBI)
	storage.ttl = make(map[uint8]map[string]*TTL)

	return storage, nil
}
