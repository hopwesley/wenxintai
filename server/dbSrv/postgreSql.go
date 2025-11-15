package dbSrv

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/hopwesley/wenxintai/server/comm"
)

var (
	_dbOnce = sync.Once{}

	_dbInstance *psDatabase = nil
)

func Instance() DbService {
	_dbOnce.Do(func() {
		_dbInstance = newPostgreSqlService()
	})

	return _dbInstance
}

type psDatabase struct {
	db *sql.DB
}

func newPostgreSqlService() *psDatabase {
	var db = &psDatabase{}
	return db
}

func (pdb *psDatabase) WithTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := pdb.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}
func (pdb *psDatabase) Init(cfg any) error {
	pbCfg, ok := cfg.(*PSDBConfig)
	if !ok {
		return comm.ErrType
	}

	db, err := pbCfg.connDB()
	if err != nil {
		return err
	}

	pdb.db = db
	return nil
}

func (pdb *psDatabase) Shutdown(ctx context.Context) error {
	if pdb.db == nil {
		return nil
	}
	err := pdb.db.Close()
	pdb.db = nil
	return err
}
