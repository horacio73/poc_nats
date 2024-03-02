package main

import "time"

// PosicaoType contém os dados a serem publicados
// para o stream e consumidos dele.
type PosicaoType struct {
	DataHoraChamada time.Time `json:"data_hora"`
	Bateria         int       `json:"bateria"`
	Lat             float64   `json:"lat"`
	Lng             float64   `json:"lng"`
	TaxistaID       int64     `json:"taxista_id"`
	Nome            string    `json:"nome"`
	Carregando      int       `json:"carregando"`
	Acu             float64   `json:"acu"`
	Token           string    `json:"token"`
	Fluxo           string    `json:"fluxo"`
	Vel             float64   `json:"vel"`
	Versao          string    `json:"versao"`
	DataHoraPosicao time.Time `json:"tempo_pos"`
	Source          string    `json:"source"`
	TraceID         string    `json:"trace_id"`
	IP              string    `json:"ip"`
}

// TaxistaType contém dados de identificação do motorista
type TaxistaType struct {
	ID              int64
	Nome            string
	Lat             float64
	Lng             float64
	TraceID         string
	DataHoraPosicao time.Time
}

// DeadLetterType contém dados de identificação do evento cuja
// publicação foi rejeitada pelo broker
type DeadLetterType struct {
	ID      int64
	Source  string
	TraceID string
}
