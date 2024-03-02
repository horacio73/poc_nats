package util

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// ConnectDB tenta se conectar ao banco de dados e
// devolve o objeto de conexão se bem sucedido.
func ConnectDB(user string, pass string, endpoint string, schema string) *sql.DB {
	mysqlDSN := fmt.Sprintf("%s:%s@tcp(%s)/%s", user, pass, endpoint, schema)
	DbConn, ErrDB := sql.Open("mysql", mysqlDSN)

	if ErrDB != nil {
		log.Fatal(ErrDB.Error())
	} else {
		log.Printf("MySQL %s connected", endpoint)
	}

	if ErrDB = DbConn.Ping(); ErrDB != nil {
		log.Printf("Ping DB failed %s", endpoint)
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
	conn, err := nats.Connect("nats://" + natsDSN)

	if err != nil {
		log.Fatal(err.Error())
	} else {
		log.Printf("Nats %s connected", natsDSN)
	}

	return conn
}

// PrintStreamState exibe informações sobre a stream,
// usado principalmente nos testes.
func PrintStreamState(ctx context.Context, stream jetstream.Stream) {
	if info, err := stream.Info(ctx); err != nil {
		fmt.Println(err.Error())
	} else {
		b, _ := json.MarshalIndent(info.State, "", " ")
		fmt.Println("inspecting stream info")
		fmt.Println(string(b))
	}
}

// CloseNats encerra a conexão com o banco de dados
func CloseNats(conn *nats.Conn) {
	if conn != nil {
		conn.Drain()
	}
}
