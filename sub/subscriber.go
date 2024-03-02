package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

// JetStreamType contém informações internas sobre o stream
type JetStreamType struct {
	js       jetstream.JetStream
	consCFG  jetstream.ConsumerConfig
	consumer jetstream.Consumer
	stream   jetstream.Stream
	ctx      context.Context
	cancel   context.CancelFunc
}

var (
	chConsumerProcess chan string = make(chan string)
	jstream           JetStreamType
)

// processStreamData contém o loop de leitura dos eventos
func readStreamData() {
	for {
		if msgs, err := jstream.consumer.Fetch(natsConf.Batch, jetstream.FetchMaxWait(natsConf.Timeout)); err == nil {
			for msg := range msgs.Messages() {
				if err = processStreamService(msg); err != nil {
					msg.NakWithDelay(time.Minute) //pede ao servidor para refazer a entrega daqui a 1 minuto
					if debugConf.Debug {
						fmt.Println(string(msg.Data()) + "\n" + err.Error() + "\n\n")
					}
				} else if natsConf.DoubleAck {
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					msg.DoubleAck(ctx)
					if cancel != nil {
						cancel()
					}
					if debugConf.Debug {
						fmt.Println("DOUBLE ACK - " + string(msg.Data()) + "\n\n")
					}
				} else {
					msg.Ack()
					if debugConf.Debug {
						fmt.Println("ACK - " + string(msg.Data()) + "\n\n")
					}
				}
			}
		}
	}
}

// Leitura do json e persistência
/* Formato esperado do json:  o campo "data" contém a informação em si.
   Os outros campos estão relacionados com a especificação "cloud event".
{"specversion":"1.0","id":"84","source":"app","type":"br.com.Gaudium",
 "subject":"Gaudium.posicao","datacontenttype":"application/json",
 "data":{"data_hora":"2024-02-16T12:19:13-03:00","bateria":63,"lat":-22.9303644,"lng":-43.3250495,
		"taxista_id":5514,"carregando":0,"acu":14.232,"token":"cGp31HgeSPCM","fluxo":"[AC]H",
		"vel":0,"versao":"A14.9.0","tempo_pos":"2024-02-14T04:45:53-03:00","ip":"127.0.0.1",
		"trace_id":"7c78e02c0fdd4fe5844e75eee449acd9"}
}
*/

// DataType recupera os dados de negócio do evento.
type DataType struct {
	Posicao PosicaoType `json:"data"`
}

func processStreamService(msg jetstream.Msg) error {
	data := DataType{}

	if err = json.Unmarshal(msg.Data(), &data); err != nil {
		fmt.Println(err.Error())
		return err
	}

	if err = postPosicaoDB(&data.Posicao); err != nil {
		return err
	}

	return postTaxistaDB(&data.Posicao)
}

// postPosicaoDB persiste dados na tabela de posição
// e retorna a chave primária
func postPosicaoDB(msg *PosicaoType) error {
	dao := PosicaoDAO{DB: dbConn}
	return dao.Insert(msg)
}

// postTaxistaDB persiste dados na tabela do motorista
// incluindo ou alterando a linha existente
func postTaxistaDB(msg *PosicaoType) error {
	dao := TaxistaDAO{DB: dbConn}
	msgTaxista := TaxistaType{
		ID:              msg.TaxistaID,
		Nome:            msg.Nome,
		Lat:             msg.Lat,
		Lng:             msg.Lng,
		TraceID:         msg.TraceID,
		DataHoraPosicao: msg.DataHoraPosicao,
	}

	return dao.Upsert(&msgTaxista)
}

// initStream configura o stream no NATS
func initStream() {
	if jstream.js, err = jetstream.New(natsConn); err != nil {
		log.Fatal("initStreamService - Erro ao instanciar objeto stream")
	}

	jstream.ctx, jstream.cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer jstream.cancel()

	if jstream.stream, err = jstream.js.Stream(jstream.ctx, natsConf.Event); err != nil {
		log.Fatal("initStreamService - Erro ao recuperar stream, execute o serviço publicador primeiro\n" + err.Error())
	}

	var deliveryPolicy jetstream.DeliverPolicy
	if natsConf.SubDeliveryTime == nil {
		deliveryPolicy = jetstream.DeliverAllPolicy
	} else {
		deliveryPolicy = jetstream.DeliverByStartTimePolicy
	}

	jstream.consCFG = jetstream.ConsumerConfig{
		Durable:       natsConf.ConsumerName,
		DeliverPolicy: deliveryPolicy,
		//	   		OptStartSeq        uint64
		OptStartTime: natsConf.SubDeliveryTime,
		AckPolicy:    jetstream.AckExplicitPolicy,
		AckWait:      natsConf.AckWait,
		MaxDeliver:   10, //dez "redeliveries", depois desiste
		//		BackOff            []time.Duration
		// 		FilterSubject:       string

		// NOTE: FilterSubjects requires nats-server v2.10.0+
		FilterSubjects: natsConf.PubSubSubjects,

		ReplayPolicy: jetstream.ReplayInstantPolicy,
		RateLimit:    0,
		//		SampleFrequency    string
		MaxWaiting:    1,
		MaxAckPending: natsConf.Batch,
		HeadersOnly:   false,
		//MaxRequestBatch: 100,
		//	MaxRequestExpires  time.Duration
		MaxRequestMaxBytes: 0, //int

		// Inactivity threshold.
		//InactiveThreshold: -1, // time.Duration

		// Generally inherited by parent stream and other markers, now can be configured directly.
		//		Replicas int
		// Force memory storage.
		//		MemoryStorage bool

		// Metadata is additional metadata for the Consumer.
		// Keys starting with `_nats` are reserved.
		// NOTE: Metadata requires nats-server v2.10.0+
		//		Metadata map[string]string

	}

	if jstream.consumer, err = jstream.js.CreateOrUpdateConsumer(jstream.ctx, natsConf.Event, jstream.consCFG); err != nil {
		log.Fatal("Consumer error \n" + err.Error())
	}
}
