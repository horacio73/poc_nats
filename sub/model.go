package main

import "time"

// PosicaoType contém os dados a serem publicados
// para o stream e consumidos dele.
type PosicaoType struct {
	ID              int64
	DataHoraChamada time.Time
	Bateria         int
	Lat             float64
	Lng             float64
	TaxistaID       int64
	Carregando      int
	Acu             float64
	Token           string
	Fluxo           string
	Vel             float64
	Versao          string
	DataHoraPosicao time.Time
	Source          string
	TraceID         string
	IP              string
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
