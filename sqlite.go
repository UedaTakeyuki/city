package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"database/sql"

	"github.com/UedaTakeyuki/erapse"
	qb "github.com/UedaTakeyuki/query"
	_ "github.com/mattn/go-sqlite3"
)

const tableName = `cities`

const sqlCreateTable = `CREATE TABLE IF NOT EXISTS %s (
	ID       INTEGER PRIMARY KEY, 
	Name     TEXT,
	State    TEXT,
	Country  TEXT,
	Lon      REAL,
	Lat      REAL
)`

const dbFileName = "city.sql3"

var SQLiteptr *sql.DB
var querybuilder qb.Query

func initializeSQL(dbfile string) (err error) {
	defer erapse.ShowErapsedTIme(time.Now())

	// if directory is not exist, create
	p := filepath.Dir(dbfile)
	// create it if not exist, refer https://stackoverflow.com/a/37932674/11073131
	if _, err = os.Stat(p); err != nil {
		if os.IsNotExist(err) {
			os.Mkdir(p, 0777)
		} else {
			return
		}
	}

	// open sqlite3 file
	log.Println("dbfile", dbfile)
	if SQLiteptr, err = sql.Open("sqlite3", dbfile); err != nil {
		return
	}

	// create table "cities" if not exist
	_, err = SQLiteptr.Exec(fmt.Sprintf(sqlCreateTable, tableName))

	return
}

func addToTable(city *CityType) (err error) {
	defer erapse.ShowErapsedTIme(time.Now())

	// make params
	params := []qb.Param{{Name: "ID", Value: city.ID},
		{Name: "Name", Value: city.Name},
		{Name: "State", Value: city.State},
		{Name: "Country", Value: city.Country},
		{Name: "Lon", Value: city.Coord.Lon},
		{Name: "Lat", Value: city.Coord.Lat}}

	_, err = SQLiteptr.Exec(querybuilder.SetTableName(tableName).ReplaceInto(params).QueryString())

	return
}

func createTransaction() (tx *sql.Tx, err error) {
	tx, err = SQLiteptr.Begin()
	return
}

func transactionAdd( /*tx *sql.Tx, */ stmt *sql.Stmt, city *CityType) (err error) {
	//	defer erapse.ShowErapsedTIme(time.Now())

	// make params
	/*
		params := []qb.Param{{Name: "ID", Value: city.ID},
			{Name: "Name", Value: city.Name},
			{Name: "State", Value: city.State},
			{Name: "Country", Value: city.Country},
			{Name: "Lon", Value: city.Coord.Lon},
			{Name: "Lat", Value: city.Coord.Lat}}
		queryStr := querybuilder.SetTableName(tableName).ReplaceInto(params).QueryString()
	*/
	if _, err = stmt.Exec(city.ID, city.Name, city.State, city.Country, city.Coord.Lon, city.Coord.Lat); err != nil {
		//		log.Println(queryStr)
		log.Println(err)
	}

	return
}

func transactionCommit(tx *sql.Tx) (err error) {
	err = tx.Commit()
	return
}

func query() (rows *sql.Rows, err error) {
	defer erapse.ShowErapsedTIme(time.Now())
	columns := []string{"Name", "State", "Country", "Lon", "Lat"}
	rows, err = SQLiteptr.Query(querybuilder.SetTableName(tableName).Select(columns).QueryString())

	return
}

func prepare(tx *sql.Tx) (stmt *sql.Stmt, err error) {
	defer erapse.ShowErapsedTIme(time.Now())

	queryStr := fmt.Sprintf(`REPLACE INTO %s (ID, Name, State, Country, Lon, Lat) VALUES(?, ?, ?, ?, ?, ?)`, tableName)
	stmt, err = tx.Prepare(queryStr)

	return
}
