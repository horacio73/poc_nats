// Package util contém tratativas para leitura do
// arquivo de configuração contido em ./etc/
package util

import (
	"log"
	"strings"
	"time"

	"github.com/pelletier/go-toml"
)

// CCConfigPath define o caminho onde está o arquivo de configuração
const CCConfigPath = "./etc/nats.conf"

// DebugType indica se o flag de debug está ligado ou não
type DebugType struct {
	Debug bool
}

// NatsConfType tem a estrutura da configuração do NATS, pub e sub
type NatsConfType struct {
	Endpoint        string
	ConsumerName    string
	Event           string
	StreamSubjects  []string
	PubSubSubjects  []string
	MaxAge          int
	Replica         int
	SubDeliveryTime *time.Time
	AckWait         time.Duration
	MaxAckPending   int
	Timeout         time.Duration
	Batch           int
	DoubleAck       bool
	ErrorRate       int
	DelayRedelivery time.Duration
	MaxWaitingPulls int
}

// MysqlConfType tem a estrutura da configuração do banco de dados
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

// DebugConfig contém a parametrização referente ao flag de debug
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
	NatsConfig.ConsumerName = loadSafeStringNode(conf.Get("nats.consumer_name"))
	NatsConfig.StreamSubjects = loadSafeStringArrayNode(conf.GetArray("nats.stream_subjects"))
	NatsConfig.PubSubSubjects = loadSafeStringArrayNode(conf.GetArray("nats.pubsub_subjects"))
	NatsConfig.MaxAge = int(loadSafeInt64Node(conf.Get("nats.maxage")))
	NatsConfig.Replica = int(loadSafeInt64Node(conf.Get("nats.replica")))
	NatsConfig.Batch = int(loadSafeInt64Node(conf.Get("nats.batch")))
	NatsConfig.DoubleAck = loadSafeBoolNode(conf.Get("nats.double_ack"))
	NatsConfig.ErrorRate = int(loadSafeInt64Node(conf.Get("nats.error_rate")))
	NatsConfig.MaxAckPending = int(loadSafeInt64Node(conf.Get("nats.max_ack_pending")))
	NatsConfig.MaxWaitingPulls = int(loadSafeInt64Node(conf.Get("nats.max_waiting_pulls")))

	strtime := loadSafeStringNode(conf.Get("nats.sub_deliverytime"))
	if len(strtime) > 0 {
		if t, err := time.Parse(time.DateTime, strtime); err != nil {
			log.Println("Invalid sub_deliverytime, assuming 'all'")
			NatsConfig.SubDeliveryTime = nil
		} else {
			NatsConfig.SubDeliveryTime = &t
		}
	}

	strtime = loadSafeStringNode(conf.Get("nats.delay_redelivery"))
	if len(strtime) > 0 {
		if t, err := time.ParseDuration(strtime); err != nil {
			log.Println("Invalid redelivery delay time, assuming 20s")
			NatsConfig.DelayRedelivery = 20 * time.Second
		} else {
			NatsConfig.DelayRedelivery = t
		}
	}

	strtime = loadSafeStringNode(conf.Get("nats.timeout"))
	if len(strtime) > 0 {
		if t, err := time.ParseDuration(strtime); err != nil {
			log.Println("Invalid timeout, assuming 5s")
			NatsConfig.Timeout = 5 * time.Second
		} else {
			NatsConfig.Timeout = t
		}
	}

	strtime = loadSafeStringNode(conf.Get("nats.ack_wait"))
	if len(strtime) > 0 {
		if t, err := time.ParseDuration(strtime); err != nil {
			log.Println("Invalid ack_wait, assuming 30s")
			NatsConfig.AckWait = 30 * time.Second
		} else {
			NatsConfig.AckWait = t
		}
	}

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
