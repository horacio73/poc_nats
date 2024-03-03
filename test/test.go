package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	ccURL      string        = "http://127.0.0.1/poc/nats/go/posicao"
	ccQtd      int64         = 120
	ccInterval time.Duration = 60 * time.Second
	ccSample   int64         = 10
	ccVersao   string        = "1.00"
	ccSource   string        = "test"
)

var tokens = []string{
	"cGp31HgeSPCM",
	"hsgaSDhqhsa9",
	"ofija7372HUs",
	"948jiSHDHs23",
	"75ffdfHDH643",
	"1greSDfsd34t",
	"y4pgerdfgSfd",
	"d543Fdst343g",
	"Sher543Gdsds",
	"KpdHSvnBSsa4",
}

func main() {

	// Carregar parâmetros da linha de comando
	qtdPtr := flag.Int64("q", ccQtd, "quantidade de hits")
	intervalPtr := flag.String("i", "", "intervalo de tempo até finalizar no formato <número><unid.medida>, por exemplo, 5m (cinco minutos)")
	urlPtr := flag.String("u", ccURL, "endpoint de acesso, por exemplo, http://127.0.0.1/poc/nats/posicao")
	flag.Parse()

	var err error
	var qtd int64 = *qtdPtr
	var uri string = *urlPtr
	var interval time.Duration = ccInterval
	var arquivo *os.File

	if *intervalPtr != "" {
		if interval, err = time.ParseDuration(*intervalPtr); err != nil {
			log.Fatalln("Formato do intervalo inválido, precisa ser <número><unid.medida>, por exemplo, 5m (cinco minutos)", err)
		}
	}

	ini := time.Now()
	if arquivo, err = createLogFile(ini); err != nil {
		log.Fatal("Erro ao criar o arquivo de log\n", err.Error())
	}
	defer arquivo.Close()

	nextSample := ccSample
	ini = time.Now()
	wait := time.Duration(interval.Abs().Microseconds()/int64(qtd)) * time.Microsecond

	log.Println(" - POC NATS - Iniciando teste em ", ini.Format(time.DateTime))
	fmt.Println(fmt.Sprintf("Expectativa de %d hits em %s\n", qtd, interval.String()))

	var i int64
	for i = 1; i <= qtd; i = i + 1 {
		// Fazendo a solicitação HTTP GET
		fullURL := fillURL(uri, i)
		resposta, err := http.Get(fullURL)
		if err != nil || resposta.StatusCode != 200 {
			log.Fatalln("Erro ao fazer a solicitação HTTP:", err)
		}

		go writeLogFile(arquivo, fullURL, i)

		// Exibindo o resultados parciais
		if (i * 100 / qtd) > nextSample {
			fmt.Printf("%d solicitações - %d%s - %s\n", i, nextSample, "%", time.Now().Sub(ini).String())
			nextSample = nextSample + ccSample
			// recalcular throttle
			wait = recalcThrottle(qtd-i, ini, interval)
		}
		if wait > 0 {
			time.Sleep(wait)
		}
	}

	fmt.Printf("%d solicitações - %s \n", i-1, "100%\n")

	fim := time.Now()
	intervalo := getIntervalo(ini, fim)
	log.Println(" - POC NATS - Finalizado em ", intervalo)
}

func createLogFile(ini time.Time) (*os.File, error) {
	if arquivo, err := os.Create("log-" + fmt.Sprint(ini.Unix()) + ".txt"); err != nil {
		return nil, err
	} else {
		return arquivo, nil
	}
}

func writeLogFile(arquivo *os.File, fullURL string, i int64) {
	str := fmt.Sprintf("%d - %s\n", i, fullURL)
	if _, err := arquivo.WriteString(str); err != nil {
		log.Fatal("Erro ao escrever no arquivo:\n", err)
	}
}

func recalcThrottle(qtdRestante int64, ini time.Time, intervalo time.Duration) time.Duration {
	agora := time.Now()
	if tempoGasto := agora.Sub(ini); tempoGasto > intervalo {
		// intervalo ultrapassado, vamos retirar o sleep para acelerar ao máximo o processamento.
		return 0
	}

	fim := ini.Add(intervalo)
	return time.Duration(fim.Sub(agora).Abs().Microseconds()/int64(qtdRestante)) * time.Microsecond
}

func getIntervalo(ini time.Time, fim time.Time) string {
	diferenca := fim.Sub(ini)

	horas := int(diferenca.Hours())
	minutos := int(diferenca.Minutes()) % 60
	segundos := int(diferenca.Seconds()) % 60

	return fmt.Sprintf("%dh %dm %ds", horas, minutos, segundos)
}

/*
http://127.0.0.1/poc/nats/go/posicao?bateria=32&lat=-22.9303644&lng=-43.3250495&taxista_id=541&
carregando=0&acu=14.232&token=cGp31HgeSPCM&fluxo=%5BAC%5DH&vel=0.050177004&versao=A14.9.0&
tempo_pos=1707896753000&data_hora=1708096753000&source=app&trace_id=7c78e02c0fdd4fe5844e75eee449acd9
*/
func fillURL(uri string, traceID int64) string {
	agora := time.Now().UnixMilli()

	valores := url.Values{}
	valores.Set("bateria", fmt.Sprint(rand.Intn(90)+10))

	valores.Set("lat", fmt.Sprint((-23330.5345345-rand.Float64()*999)/1000.0))
	valores.Set("lng", fmt.Sprint((-43425.9452128-rand.Float64()*999)/1000.0))
	valores.Set("taxista_id", fmt.Sprint(rand.Intn(10)+98)) //dez taxistas diferentes, no máximo
	valores.Set("carregando", fmt.Sprint(rand.Intn(2)))
	valores.Set("acu", fmt.Sprint(rand.Float32()*20+3))
	valores.Set("token", tokens[rand.Intn(10)])
	valores.Set("fluxo", "[AC]H")
	valores.Set("vel", fmt.Sprint(rand.Float32()*70+3))
	valores.Set("versao", ccVersao)
	valores.Set("tempo_pos", fmt.Sprint(agora-rand.Int63n(999999)+5122))
	valores.Set("data_hora", fmt.Sprint(agora))
	valores.Set("source", ccSource)
	valores.Set("trace_id", fmt.Sprint(traceID))

	return uri + "?" + valores.Encode()
}
