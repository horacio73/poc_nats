package main

import (
	"context"
	"fmt"
	"gaudiumsoftware/nats/util"
	"log"
	"strconv"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/nats-io/nats.go/jetstream"
)

const (
	ccGaudium       string = "br.com.Gaudium"
	ccQryBateria    string = "bateria"
	ccQryLat        string = "lat"
	ccQryLng        string = "lng"
	ccQryTaxistaID  string = "taxista_id"
	ccQryCarregando string = "carregando"
	ccQryAcuracia   string = "acu"
	ccQryToken      string = "token"
	ccQryFluxo      string = "fluxo"
	ccQryVersao     string = "versao"
	ccQryVelocidade string = "velocidade"
	ccQryTempoPos   string = "tempo_pos"
	ccQryDataHora   string = "data_hora"
	ccQrySource     string = "source"
	ccQryTraceID    string = "trace_id"
)

// JetStreamType contém informações internas sobre o stream
type JetStreamType struct {
	js     jetstream.JetStream
	jsCFG  jetstream.StreamConfig
	stream jetstream.Stream
	ctx    context.Context
	cancel context.CancelFunc
}

var (
	chStreamProcess chan PosicaoType = make(chan PosicaoType)
	jstream         JetStreamType
)

// processStreamData recupera os parâmetros vindos do cliente e os persiste,
// tanto no banco de dados quanto no stream.
// ?bateria=63&lat=-22.9303644&lng=-43.3250495&taxista_id=5514&carregando=0&acu=14.232&
// token=cGp31HgeS&fluxo=%5BAC%5DH&vel=0.050177004&versao=A14.9.0&tempo_pos=1707896753000&
// data_hora=1708096753000&source=app&trace_id=7c78e02c0fdd4fe5844e75eee449acd9
func processStreamData(c *fiber.Ctx) error {

	msg := getQueryData(c)
	chStreamProcess <- msg

	c.SendString("ok")
	return nil
}

// initStream configura o stream no NATS
func initStream() {
	if jstream.js, err = jetstream.New(natsConn); err != nil {
		log.Fatal("initStreamService - Erro ao instanciar objeto stream")
	}

	var maxAge time.Duration

	if maxAge, err = time.ParseDuration(fmt.Sprint(natsConf.MaxAge) + "h"); err != nil {
		log.Println("maxage (TTL) sem parametrização, será definido como 'infinito'")
		maxAge = 0
	}

	jstream.jsCFG = jetstream.StreamConfig{
		Name:      natsConf.Event,
		Subjects:  natsConf.StreamSubjects,
		Storage:   jetstream.FileStorage,
		Retention: jetstream.LimitsPolicy,
		//		MaxConsumers:  -1,   // default
		//		MaxMsgs: -1, //default
		//		MaxBytes: -1, //default
		Discard: jetstream.DiscardNew,
		//		DiscardNewPerSubject,
		MaxAge:            maxAge,
		MaxMsgsPerSubject: -1,
		//		MaxMsgSize: -1  //default
		Replicas: int(natsConf.Replica),
		NoAck:    false,
		Template: "",
		//		Duplicates: 2 * time.Minute  //default,
		//		Placement,
		//		Mirror,
		//		Sources,
		Sealed:      false,
		DenyDelete:  false,
		DenyPurge:   false,
		AllowRollup: true,
		Compression: jetstream.NoCompression,
		FirstSeq:    1,
		//		SubjectTransform *SubjectTransformConfig
		//		RePublish *RePublish
		AllowDirect: true,
		//		MirrorDirect bool
		//		ConsumerLimits StreamConsumerLimits,
		//		Metadata map[string]string
	}

	jstream.ctx, jstream.cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer jstream.cancel()

	if jstream.stream, err = jstream.js.CreateOrUpdateStream(jstream.ctx, jstream.jsCFG); err != nil {
		log.Fatal("initStreamService - Erro ao criar stream")
	}
}

// Publicar mensagens aos subjects parametrizados no
// arquivo de configuração
func publishEvents(cEvent *event.Event) error {
	var (
		//		opts       jetstream.PublishOpt
		err        error = nil
		deadLetter error = nil
		ack        *jetstream.PubAck
		bMsg       []byte
	)

	jstream.ctx, jstream.cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer jstream.cancel()

	for _, pub := range natsConf.PubSubSubjects {
		cEvent.SetSubject(pub)
		if bMsg, err = cEvent.MarshalJSON(); err == nil {
			if ack, err = jstream.js.Publish(jstream.ctx, pub, bMsg); err != nil {
				log.Printf("Erro ao publicar:\n%v", err)
			} else {
				// debug
				if debugConf.Debug {
					log.Printf("%v \n\n", ack) //, opts)
					util.PrintStreamState(jstream.ctx, jstream.stream)
				}
			}

		} else {
			log.Printf("ERR - Dead Letter: %v /n", err)
			deadLetter = err
		}
	}

	return deadLetter
}

// pubStreamService é um serviço independente, assíncrono,
// que aguarda pela mensagem a ser publicada.
func pubStreamService() {
	// leitura do canal, comando blocante.
	for msg := range chStreamProcess {

		id := fmt.Sprint(postPosicaoDB(&msg))

		msg.Nome = getRandomName()
		go postTaxistaDB(&msg) //upsert na tabela taxista em paralelo

		// criar um repositório para empacotar a mensagem no formato "cloud event"
		cEvent := cloudevents.NewEvent()
		cEvent.SetID(id)
		cEvent.SetType(ccGaudium)
		cEvent.SetSource(msg.Source)
		cEvent.SetSpecVersion(event.CloudEventsVersionV1)
		cEvent.SetData(cloudevents.ApplicationJSON, &msg) // map[string]string{"hello": "world"})

		// publicar a mensagem recebida
		if err = publishEvents(&cEvent); err != nil {
			// em caso de erro, persistir na tabela "dead letter"
			go postDeadLetterDB(msg.TraceID, msg.Source)
		}

	}
}

// postPosicaoDB persiste dados na tabela de posição
// e retorna a chave primária
func postPosicaoDB(msg *PosicaoType) int64 {
	var id int64

	dao := PosicaoDAO{DB: dbConn}

	if id, err = dao.Insert(msg); err != nil {
		log.Fatal("publisher - postPosicaoDB - Erro ao gravar na tabela\n" + err.Error())
	}

	return id
}

// postTaxistaDB persiste dados na tabela do motorista
// incluindo ou alterando a linha existente
func postTaxistaDB(msg *PosicaoType) {
	dao := TaxistaDAO{DB: dbConn}
	msgTaxista := TaxistaType{
		ID:              msg.TaxistaID,
		Nome:            msg.Nome,
		Lat:             msg.Lat,
		Lng:             msg.Lng,
		TraceID:         msg.TraceID,
		DataHoraPosicao: msg.DataHoraPosicao,
	}
	if _, err := dao.Upsert(&msgTaxista); err != nil {
		log.Fatal("publisher - postTaxistaDB - Erro ao gravar na tabela\n" + err.Error())
	}
}

// postDeadLetter persiste a falha na gravação da stream no BD
func postDeadLetterDB(trace string, source string) {
	dao := DeadLetterDAO{DB: dbConn}
	deadLetter := DeadLetterType{
		TraceID: trace,
		Source:  source,
	}
	if err := dao.Insert(&deadLetter); err != nil {
		log.Fatal("publisher - postDeadLetterDB - Erro ao gravar na tabela")
	}
}

func getQueryData(c *fiber.Ctx) PosicaoType {
	msg := PosicaoType{}

	msg.Bateria, _ = strconv.Atoi(utils.CopyString(c.Query(ccQryBateria, "0")))
	msg.Acu, _ = strconv.ParseFloat(utils.CopyString(c.Query(ccQryAcuracia, "0")), 64)
	msg.Carregando, _ = strconv.Atoi(utils.CopyString(c.Query(ccQryCarregando, "0")))

	auxInt64, _ := strconv.ParseInt(utils.CopyString(c.Query(ccQryDataHora, "0")), 10, 64)
	msg.DataHoraChamada = time.UnixMilli(auxInt64)

	auxInt64, _ = strconv.ParseInt(utils.CopyString(c.Query(ccQryTempoPos, "0")), 10, 64)
	msg.DataHoraPosicao = time.UnixMilli(auxInt64)

	msg.Fluxo = utils.CopyString(c.Query(ccQryFluxo, "0"))
	msg.IP = c.IP()
	msg.Lat, _ = strconv.ParseFloat(utils.CopyString(c.Query(ccQryLat, "0")), 64)
	msg.Lng, _ = strconv.ParseFloat(utils.CopyString(c.Query(ccQryLng, "0")), 64)
	msg.Source = utils.CopyString(c.Query(ccQrySource, "0"))
	msg.Vel, _ = strconv.ParseFloat(utils.CopyString(c.Query(ccQryVelocidade, "0")), 64)
	msg.TaxistaID, _ = strconv.ParseInt(utils.CopyString(c.Query(ccQryTaxistaID, "0")), 10, 64)
	msg.Token = utils.CopyString(c.Query(ccQryToken, "0"))
	msg.TraceID = utils.CopyString(c.Query(ccQryTraceID, "0"))
	msg.Versao = utils.CopyString(c.Query(ccQryVersao, "0"))

	return msg
}
