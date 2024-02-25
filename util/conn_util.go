package util

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/nats-io/nats.go"
)

// ConnectDB tenta se conectar ao banco de dados e
// devolve o objeto de conexão se bem sucedido.
func ConnectDB(mysqlDSN string) *sql.DB {
	DbConn, ErrDB := sql.Open("mysql", mysqlDSN)

	if ErrDB != nil {
		log.Fatal(ErrDB.Error())
	} else {
		log.Printf("MySQL %s connected", mysqlDSN)
	}

	if ErrDB = DbConn.Ping(); ErrDB != nil {
		log.Printf("Ping DB failed %s", mysqlDSN)
	}

	return DbConn
}

// CloseDB encerra a conexão com o banco de dados
func CloseDB(db *sql.DB) {
	if db != nil {
		_ = db.Close()
	}
}

// ConnectNats tenta se conectar à plataforma de
// streaming e devolve o objeto de conexão se bem sucedido.
func ConnectNats(natsDSN string) *nats.Conn {
	conn, err := nats.Connect(natsDSN)

	if err != nil {
		log.Fatal(err.Error())
	} else {
		log.Printf("Nats %s connected", natsDSN)
	}

	return conn
}

// CloseNats encerra a conexão com o banco de dados
func CloseNats(conn *nats.Conn) {
	if conn != nil {
		conn.Drain()
	}
}
