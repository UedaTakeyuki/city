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

const SquaredDistance = "(Lat - %f)*(Lat - %f) + (Lon - %f)*(Lon - %f)"

const dbFileName = "city.sql3"

var SQLiteptr *sql.DB
var querybuilder qb.Query

func squaredDistance(lat float64, lon float64) (result string) {
	result = fmt.Sprintf(SquaredDistance, lat, lat, lon, lon)
	return
}

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
	columns := []string{"ID", "Name", "Lon", "Lat"}
	queryStr := querybuilder.SetTableName(tableName).Select(columns).QueryString()
	log.Println(queryStr)
	rows, err = SQLiteptr.Query(queryStr)

	return
}

func prepare(tx *sql.Tx) (stmt *sql.Stmt, err error) {
	defer erapse.ShowErapsedTIme(time.Now())

	queryStr := fmt.Sprintf(`REPLACE INTO %s (ID, Name, State, Country, Lon, Lat) VALUES(?, ?, ?, ?, ?, ?)`, tableName)
	stmt, err = tx.Prepare(queryStr)

	return
}

func findNearestCityFromRows(rows *sql.Rows, lat float64, lon float64) {
	defer erapse.ShowErapsedTIme(time.Now())

	var nearestCity nearestCityType
	rows.Next()
	if err := rows.Scan(
		&nearestCity.id,
		&nearestCity.name,
		&nearestCity.lat,
		&nearestCity.lon,
	); err != nil {
		log.Println(err)
		return
	}
	nearestCity.sqrDist = dist(lat, nearestCity.lat, lon, nearestCity.lon)

	var candidateCity nearestCityType
	for rows.Next() {
		if err := rows.Scan(
			&candidateCity.id,
			&candidateCity.name,
			&candidateCity.lon,
			&candidateCity.lat,
		); err != nil {
			log.Println(err)
		} else {
			distance := dist(lat, candidateCity.lat, lon, candidateCity.lon)
			if distance < nearestCity.sqrDist {
				nearestCity.id = candidateCity.id
				nearestCity.name = candidateCity.name
				nearestCity.lat = candidateCity.lat
				nearestCity.lon = candidateCity.lon
				nearestCity.sqrDist = distance
			}
		}
	}
	log.Println("nearest city", nearestCity.name)
	log.Println("id", nearestCity.id)
	log.Println("lat", nearestCity.lat)
	log.Println("lon", nearestCity.lon)
}

func getRowsOfClosingCities(lat float64, lon float64) (rows *sql.Rows, err error) {
	defer erapse.ShowErapsedTIme(time.Now())
	var closerThan float64
	closerThan = 0.1

	for {
		log.Println("closerThan", closerThan)
		if cityNumbers := getNumberOfClosingCities(closerThan, lat, lon); cityNumbers != 0 {
			log.Println("Cities", cityNumbers)
			rows, err = getClosingCitiesRows(closerThan, lat, lon)
			return
		}
		closerThan = closerThan * 10.0
	}
}

func getNumberOfClosingCities(closerThan float64, lat float64, lon float64) (number int) {
	defer erapse.ShowErapsedTIme(time.Now())

	columns := []string{"count(*)"}
	queryStr := querybuilder.SetTableName(tableName).Select(columns).Where(qb.LessThan(squaredDistance(lat, lon), closerThan)).QueryString()
	log.Println(queryStr)
	rows, err := SQLiteptr.Query(queryStr)
	if err != nil {
		log.Println(err)
	} else {
		rows.Next()
		if err = rows.Scan(&number); err != nil {
			log.Println(err)
		}
	}
	return
}

func getClosingCitiesRows(closerThan float64, lat float64, lon float64) (rows *sql.Rows, err error) {
	defer erapse.ShowErapsedTIme(time.Now())

	columns := []string{"ID", "Name", "Lon", "Lat"}
	queryStr := querybuilder.SetTableName(tableName).Select(columns).Where(qb.LessThan(squaredDistance(lat, lon), closerThan)).QueryString()
	log.Println(queryStr)
	rows, err = SQLiteptr.Query(queryStr)

	return
}
