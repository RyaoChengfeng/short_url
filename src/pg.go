package main

import (
	"context"
	"github.com/go-pg/pg/v10/orm"
	"time"

	"github.com/go-pg/pg/extra/pgdebug"
	"github.com/go-pg/pg/v10"
	"src/config"
	. "src/util"
)

var (
	pgDB *pg.DB
)

type Url struct {
	ID       int64  `pg:"id,pk" json:"id"`
	LongUrl  string `pg:"long_url,use_zero" json:"long_url"`
	ShortUrl string `pg:"short_url,use_zero" json:"short_url"`
}

type dbTrait struct {
	tx *pg.Tx
}

func initPgDB() {
	options := &pg.Options{
		Addr:     config.C.PgDB.Host + ":" + config.C.PgDB.Port,
		User:     config.C.PgDB.User,
		Password: config.C.PgDB.Password,
		Database: config.C.PgDB.DB,
	}
	Logger.Debug(options)

	d := pg.Connect(options)

	err := d.Ping(context.Background())
	if err != nil {
		panic(err)
	}

	if config.C.Debug {
		d.AddQueryHook(pgdebug.DebugHook{
			Verbose: true,
		})
	}

	pgDB = d

	err = createPostgresSchema()
	if err != nil {
		panic(err)
	}

	Logger.Printf("PostgreSQL server connected")
}

func getDBTx(ctx context.Context) dbTrait {
	tx, err := pgDB.BeginContext(ctx)
	if err != nil {
		panic(err)
	}

	return dbTrait{
		tx: tx,
	}
}

func (m *model) Close() {
	if !m.abort {
		_ = m.tx.Commit()
	}
}

func (m *model) Abort() {
	_ = m.tx.Rollback()
	m.abort = true
}

type Model interface {
	// 关闭数据库连接
	Close()
	// 终止操作，用于如事务的取消
	Abort()
	RetrieveUrlByShorter(shorter string) (Url, error)
	CreateUrl(url Url) (orm.Result,int64, error)
	UpdateUrl(id int64,short_url string) (orm.Result, error)
	RetrieveUrlByLongUrl(longUrl string) (Url, error)
}

type model struct {
	dbTrait
	ctx   context.Context
	abort bool
}

func GetModel() Model {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	if config.C.Debug {
		ctx = context.Background()
	}

	ret := &model{
		dbTrait: getDBTx(ctx),
		ctx:     ctx,
		abort:   false,
	}

	return ret
}

func createPostgresSchema() error {
	models := []interface{}{
		(*Url)(nil),
	}

	for _, model := range models {
		err := pgDB.Model(model).CreateTable(&orm.CreateTableOptions{
			Temp:          false,
			IfNotExists:   true,
			FKConstraints: true,
		})
		if err != nil {
			return err
		}
	}
	//pgDB.Model((*Series)(nil)).Exec(`ALTER TABLE series add column if not exists tsv tsvector`)
	return nil
}

func (m *model) RetrieveUrlByLongUrl(longUrl string) (Url, error) {
	var url Url
	err := m.tx.Model(&url).Where("long_url = ?", longUrl).Select()
	if err != nil {
		return Url{}, err
	}
	return url, nil
}

func (m *model) RetrieveUrlByShorter(shorter string) (Url, error) {
	var url Url
	err := m.tx.Model(&url).Where("short_url = ?", shorter).Select()
	if err != nil {
		return Url{}, err
	}
	return url, nil
}

func (m *model) CreateUrl(url Url) (orm.Result,int64, error) {
	rst, err := m.tx.Model(&url).Insert()
	if err != nil {
		return nil,0, err
	}
	return rst,url.ID, nil
}

func (m *model) UpdateUrl(id int64, shortUrl string) (orm.Result, error) {
	rst,err:=m.tx.Model((*Url)(nil)).Set("short_url = ?", shortUrl).Where("id = ?",id).Update()
	if err != nil {
		return nil, err
	}
	return rst, nil
}