// ***********************************************
// Serviço publicador
// ***********************************************
package main

import (
	"database/sql"
	"flag"
	"gaudiumsoftware/nats/util"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/nats-io/nats.go"
	"github.com/pelletier/go-toml"
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

	conf, _ := toml.LoadFile(*cfgPtr)

	var ok bool
	if endpointAPIServer, ok = conf.Get("api-server.endpoint").(string); ok {
		endpointAPIServer = strings.TrimSpace(endpointAPIServer)
	} else {
		log.Fatal("Erro na leitura do arquivo de configuração, 'api-server.endpoint' não encontrado.\n" + err.Error())
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
	go pubStreamService()

	// https://docs.gofiber.io/api/fiber
	wsFiber := fiber.New(fiber.Config{
		Prefork:                 false,
		CaseSensitive:           true,
		StrictRouting:           false,
		Immutable:               false,
		UnescapePath:            false,
		ETag:                    false,
		BodyLimit:               4 * 1024 * 1024, //default body bytes limits.
		Concurrency:             256 * 1024,      //default concurrent conections
		ServerHeader:            "",
		ProxyHeader:             "",         //default
		DisableKeepalive:        false,      //default
		EnableTrustedProxyCheck: false,      //default
		EnablePrintRoutes:       false,      //default
		TrustedProxies:          []string{}, //default
		AppName:                 "POC NATS - API SERVER",
	})

	// Endpoint para receber os dados a serem persistidos no stream
	wsFiber.Get(endpointAPIServer, processStreamData)

	if errListen := wsFiber.Listen(":80"); errListen != nil {
		log.Println("A porta 80 já está em uso ou não há portas disponíveis no servidor.")
	}

}
