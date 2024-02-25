// ***********************************************
// Serviço subscritor
// ***********************************************
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"gaudiumsoftware/nats/util"
	"log"
)

var db *sql.DB

// main é o programa principal
func main() {

	fmt.Println("Teste")

	// Carregar os dados do arquivo de configuração
	cfgPtr := flag.String("c", util.CCConfigPath, "the path to the config file")
	flag.Parse()

	if err := util.LoadConfig(*cfgPtr); err != nil {
		log.Fatal("Erro na leitura do arquivo de configuração.\n" + err.Error())
	}

}
