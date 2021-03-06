/**
 *  author:Abel
 *  email:abel.zhou@hotmail.com
 *  date:2019-06-04
 */
package sql

import (
	"database/sql"
	"github.com/AbelZhou/even/database"
	"math/rand"
	"time"
)

func NewMySQLPool(config *database.Config) *ConnPool {
	return NewPool(config, "even_mysql");
}

func NewPool(config *database.Config, driverName string) *ConnPool {
	//format database config
	configFormat(config)

	if config.Read == nil {
		config.Read[0] = config.Write
	}

	//load writer database connections
	var writerConn, err = sql.Open(driverName, config.Write.DSN)
	if err != nil {
		panic(err)
	}

	err = writerConn.Ping()
	if err != nil {
		panic(err)
	}
	writerConn.SetMaxOpenConns(config.Write.MaxActive)
	writerConn.SetMaxIdleConns(config.Write.MaxIdle)
	writerConn.SetConnMaxLifetime(time.Duration(config.Write.IdleTimeout) * time.Second)

	// load reader database connections.
	var readerConn []*sql.DB
	for _, readerConf := range config.Read {
		reader, err := sql.Open(driverName, readerConf.DSN)
		if err != nil {
			panic(err)
		}
		err = reader.Ping()
		if err != nil {
			panic(err)
		}
		reader.SetConnMaxLifetime(time.Duration(readerConf.IdleTimeout) * time.Second)
		reader.SetMaxIdleConns(readerConf.MaxIdle)
		reader.SetMaxOpenConns(readerConf.MaxActive)
		readerConn = append(readerConn, reader)
	}

	return &ConnPool{
		dbConfig: config,
		writer:   writerConn,
		reader:   readerConn,
	}
}

//Progress the database config.
func configFormat(dbConfig *database.Config) {
	if dbConfig.Write.MaxActive == 0 {
		dbConfig.Write.MaxActive = dbConfig.DefMaxActive
	}
	if dbConfig.Write.MaxIdle == 0 {
		dbConfig.Write.MaxIdle = dbConfig.DefMaxIdle
	}
	if dbConfig.Write.IdleTimeout == 0 {
		dbConfig.Write.IdleTimeout = dbConfig.DefIdleTimeout
	}

	for i := 0; i < len(dbConfig.Read); i++ {
		if dbConfig.Read[i].MaxActive == 0 {
			dbConfig.Read[i].MaxActive = dbConfig.DefMaxActive
		}
		if dbConfig.Read[i].MaxIdle == 0 {
			dbConfig.Read[i].MaxIdle = dbConfig.DefMaxIdle
		}
		if dbConfig.Read[i].IdleTimeout == 0 {
			dbConfig.Read[i].IdleTimeout = dbConfig.DefIdleTimeout
		}
	}
}

type ConnPool struct {
	dbConfig *database.Config
	writer   *sql.DB
	reader   []*sql.DB
}

func (pool *ConnPool) Master() *Conn {
	return &Conn{
		db:            pool.writer,
		inTransaction: false,
		isReader:false,
	}
}

func (pool *ConnPool) Slave() *Conn {
	return &Conn{
		db:            pool.reader[rand.Intn(len(pool.reader))],
		inTransaction: false,
		isReader:true,
	}
}
