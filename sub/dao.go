package main

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const ccPosicaoTable = "sub_tst_nats_posicao_taxi"
const ccTaxistaTable = "sub_tst_nats_taxista"

// PosicaoDAO é o objeto de persistência da tabela pub_tst_nats_posicao_taxi
type PosicaoDAO struct {
	DB *sql.DB
}

// TaxistaDAO é o objeto de persistência da tabela pub_tst_nats_taxista
type TaxistaDAO struct {
	DB *sql.DB
}

// Upsert tentará fazer um update primeiro, se não conseguir,
// fará o insert.
func (dao *TaxistaDAO) Upsert(msg *TaxistaType) error {
	if msg == nil {
		return nil
	}

	var (
		stmtIns      *sql.Stmt
		errDB        error
		dbCmd        string
		res          sql.Result
		rowsAffected int64
	)

	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in TaxistaDAO.Upsert()", r)
		}
		if stmtIns != nil {
			stmtIns.Close()
		}
	}()

	tmDB := msg.DataHoraPosicao.Format("2006-01-02 15:04:05") //formato de timestamp esperado pelo MYSQL
	tmNow := time.Now().Format("2006-01-02 15:04:05.999")     //formato de timestamp esperado pelo MYSQL

	dbCmd = "UPDATE " + ccTaxistaTable + " SET lat=?,lng=?,data_hora_posicao=?,trace_id=?, modified_at=? " +
		"WHERE id=?"

	if stmtIns, errDB = dao.DB.Prepare(dbCmd); errDB != nil {
		return errDB
	}

	res, errDB = stmtIns.Exec(msg.Lat, msg.Lng, tmDB, msg.TraceID, tmNow, msg.ID)
	if errDB == nil {
		if rowsAffected, errDB = res.RowsAffected(); errDB != nil {
			return errDB
		}
	} else {
		return errDB
	}

	if rowsAffected == 0 {
		dbCmd = "INSERT INTO " + ccTaxistaTable + " (id,nome,lat,lng,data_hora_posicao,trace_id) " +
			"VALUES (?,?,?,?,?,?)"

		if stmtIns, errDB = dao.DB.Prepare(dbCmd); errDB != nil {
			return errDB
		}

		if res, errDB = stmtIns.Exec(msg.ID, msg.Nome, msg.Lat, msg.Lng, tmDB, msg.TraceID); errDB != nil {
			return errDB
		}
	}

	return nil
}

// Insert realiza um "insert" na tabela pub_tst_nats_posicao_taxi
func (dao *PosicaoDAO) Insert(msg *PosicaoType) error {
	if msg == nil {
		return nil
	}

	var stmtIns *sql.Stmt
	var errDB error
	var dbCmd string

	tmDB := msg.DataHoraChamada.Format("2006-01-02 15:04:05")    //formato de timestamp esperado pelo MYSQL
	tmPosDB := msg.DataHoraPosicao.Format("2006-01-02 15:04:05") //formato de timestamp esperado pelo MYSQL

	dbCmd = "INSERT INTO " + ccPosicaoTable + " (taxista_id,data_hora,lat,lng,trace_id," +
		"source_id,ip,token,velocidade_informada,tempo_posicao,acuracidade,fluxo,bateria,carregando," +
		"versao) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"

	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in PosicaoDAO.Insert()", r)
		}
		if stmtIns != nil {
			stmtIns.Close()
		}
	}()

	stmtIns, errDB = dao.DB.Prepare(dbCmd)

	if errDB == nil {
		_, errDB = stmtIns.Exec(msg.TaxistaID, tmDB, msg.Lat, msg.Lng, msg.TraceID,
			msg.Source, msg.IP, msg.Token, msg.Vel, tmPosDB, msg.Acu, msg.Fluxo,
			msg.Bateria, msg.Carregando, msg.Versao)
	}

	return errDB
}
