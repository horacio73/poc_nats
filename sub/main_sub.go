// ***********************************************
// Serviço subscritor
// ***********************************************
package main

import (
	"database/sql"
	"flag"
	"gaudiumsoftware/nats/util"
	"log"

	"github.com/nats-io/nats.go"
)

var (
	natsConf          util.NatsConfType
	dbConf            util.MysqlConfType
	debugConf         util.DebugType
	dbConn            *sql.DB
	natsConn          *nats.Conn
	err               error
	endpointAPIServer string
)

func loadConfig(cfgPtr *string) {
	if err := util.LoadConfig(*cfgPtr); err != nil {
		log.Fatal("Erro na leitura do arquivo de configuração.\n" + err.Error())
	}
}

// main é o programa principal
func main() {
	// Carregar os dados do arquivo de configuração
	cfgPtr := flag.String("c", util.CCConfigPath, "the path to the config file")
	flag.Parse()

	loadConfig(cfgPtr)

	dbConf = util.MysqlConfig
	natsConf = util.NatsConfig
	debugConf = util.DebugConfig

	// conectar com banco de dados
	dbConn = util.ConnectDB(util.MysqlConfig.Username, util.MysqlConfig.Password,
		util.MysqlConfig.Endpoint, util.MysqlConfig.Schema)

	defer util.CloseDB(dbConn)

	// conectar com NATS
	natsConn = util.ConnectNats(util.NatsConfig.Endpoint)
	defer util.CloseNats(natsConn)

	initStream()
	readStreamData()
}
