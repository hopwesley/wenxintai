package dbSrv

import (
	"context"
	"database/sql"
	"errors"
	"sync"

	"github.com/hopwesley/wenxintai/server/comm"
	"github.com/rs/zerolog"
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
	db  *sql.DB
	log zerolog.Logger
}

func newPostgreSqlService() *psDatabase {
	var db = &psDatabase{
		log: comm.LogInst().With().Str("model", "PostgreSql").Logger()}
	return db
}

func (pdb *psDatabase) WithTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := pdb.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (pdb *psDatabase) Init(cfg any) error {
	pbCfg, ok := cfg.(*PSDBConfig)
	if !ok {
		return errors.New("invalid type convert")
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
