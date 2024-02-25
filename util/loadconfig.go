// Package util contém tratativas para leitura do
// arquivo de configuração contido em ./etc/
package util

import (
	"strings"

	"github.com/pelletier/go-toml"
)

// CCConfigPath define o caminho onde está o arquivo de configuração
const CCConfigPath = "/etc/nats_pub.conf"

type DebugType struct {
	Debug bool
}

type NatsConfType struct {
	Endpoint       string
	Event          string
	StreamSubjects []string
	PubSubSubjects []string
	Timeout        int
	MaxAge         int
	Replica        int
}

type MysqlConfType struct {
	Endpoint        string
	Schema          string
	Username        string
	Password        string
	PosicaoTable    string
	DeadLetterTable string
	TaxistaTable    string
}

// NatsConfig contém a parametrização referente ao nats
var NatsConfig NatsConfType

// MysqlConfig contém a parametrização referente ao banco de dados
var MysqlConfig MysqlConfType

var DebugConfig DebugType

// LoadConfig interpreta todo o arquivo de configuração e grava os dados
// nas estruturas em memória.  Serão usados ao longo do programa.
func LoadConfig(configFileName string) error {
	conf, err := toml.LoadFile(configFileName)
	if err != nil {
		return err
	}

	DebugConfig.Debug = loadSafeBoolNode(conf.Get("debug.debug"))

	NatsConfig.Endpoint = loadSafeStringNode(conf.Get("nats.endpoint"))
	NatsConfig.Event = loadSafeStringNode(conf.Get("nats.event"))
	NatsConfig.StreamSubjects = loadSafeStringArrayNode(conf.GetArray("nats.stream_subjects"))
	NatsConfig.PubSubSubjects = loadSafeStringArrayNode(conf.GetArray("nats.pubsub_subjects"))
	NatsConfig.Timeout = int(loadSafeInt64Node(conf.Get("nats.timeout")))
	NatsConfig.MaxAge = int(loadSafeInt64Node(conf.Get("nats.maxage")))
	NatsConfig.Replica = int(loadSafeInt64Node(conf.Get("nats.replica")))

	MysqlConfig.Endpoint = loadSafeStringNode(conf.Get("mysql.endpoint"))
	MysqlConfig.Schema = loadSafeStringNode(conf.Get("mysql.schema"))
	MysqlConfig.Username = loadSafeStringNode(conf.Get("mysql.username"))
	MysqlConfig.Password = loadSafeStringNode(conf.Get("mysql.password"))
	MysqlConfig.PosicaoTable = loadSafeStringNode(conf.Get("mysql.tbl_posicao"))
	MysqlConfig.DeadLetterTable = loadSafeStringNode(conf.Get("mysql.tbl_deadletter"))
	MysqlConfig.TaxistaTable = loadSafeStringNode(conf.Get("mysql.tbl_taxista"))

	return nil
}

func loadSafeStringArrayNode(node any) []string {
	if node != nil {
		if arr, ok := node.([]string); ok {
			return arr
		}
	}
	return []string{}
}

func loadSafeStringNode(node any) string {
	if node != nil {
		if str, ok := node.(string); ok {
			return strings.TrimSpace(str)
		}
	}
	return ""
}

func loadSafeBoolNode(node any, defaultValue ...bool) bool {
	if node != nil {
		if boolean, ok := node.(bool); ok {
			return boolean
		}
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return false
}

func loadSafeInt64Node(node any) int64 {
	if node != nil {
		if integer, ok := node.(int64); ok {
			return integer
		}
	}
	return 0
}
