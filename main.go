package main

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/UedaTakeyuki/erapse"
)

type CityType struct {
	ID      int64          `json:"id"`
	Name    string         `json:"name"`
	State   string         `json:"state"`
	Country string         `json:"country"`
	Coord   CoordinateType `json:"coord"`
}

type CoordinateType struct {
	Lon float64 `json:"lon"`
	Lat float64 `json:"lat"`
}

// var cities []CityType
var cities = make([]*CityType, 0)

func main() {
	defer erapse.ShowErapsedTIme(time.Now())

	// set log
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)

	var lon, lat float64
	var err error

	log.Println("os.Args", os.Args)

	if lon, err = strconv.ParseFloat(os.Args[2], 64); err != nil {
		log.Println(err)
		return
	}
	if lat, err = strconv.ParseFloat(os.Args[1], 64); err != nil {
		log.Println(err)
		return
	}
	//	readJson()
	initializeSQL(dbFileName)
	var rows *sql.Rows
	//	if rows, err = query(); err != nil {
	if rows, err = getRowsOfClosingCities(lat, lon); err != nil {
		log.Println(err)
	} else {
		findNearestCityFromRows(rows, lat, lon)
	}
	//findNearestCity(lat, lon)
	// dumpBinaryFile()
	// dumpToDB()
}

func readJson() {
	defer erapse.ShowErapsedTIme(time.Now())

	if jsonfile, err := ioutil.ReadFile("./city.list.json"); err != nil {
		log.Println(err)
	} else {
		if err := json.Unmarshal(jsonfile, &cities); err != nil {
			log.Println(err)
		}
	}
	return
}

type nearestCityType struct {
	id      int64
	name    string
	lat     float64
	lon     float64
	sqrDist float64
}

func findNearestCity(lat float64, lon float64) {
	defer erapse.ShowErapsedTIme(time.Now())
	//	lat = 35.596
	//	lon = 139.610
	log.Println("len(cities)", len(cities))

	var nearestCity nearestCityType
	nearestCity.id = cities[0].ID
	nearestCity.name = cities[0].Name
	nearestCity.lat = cities[0].Coord.Lat
	nearestCity.lon = cities[0].Coord.Lon
	nearestCity.sqrDist = dist(lat, cities[0].Coord.Lat, lon, cities[0].Coord.Lon)

	for _, candidateCity := range cities {
		distance := dist(lat, candidateCity.Coord.Lat, lon, candidateCity.Coord.Lon)
		if distance < nearestCity.sqrDist {
			nearestCity.id = candidateCity.ID
			nearestCity.name = candidateCity.Name
			nearestCity.lat = candidateCity.Coord.Lat
			nearestCity.lon = candidateCity.Coord.Lon
			nearestCity.sqrDist = distance
		}
	}

	log.Println("nearest city", nearestCity.name)
	log.Println("id", nearestCity.id)
	log.Println("lat", nearestCity.lat)
	log.Println("lon", nearestCity.lon)
}

func dist(lat0 float64, lat1 float64, lon0 float64, lon1 float64) (distance float64) {
	//	defer erapse.ShowErapsedTIme(time.Now())

	distance = math.Pow((lat0-lat1), 2) + math.Pow((lon0-lon1), 2)
	return
}

func getJson() {
	return
}

/*
func dumpBinaryFile() {
	defer erapse.ShowErapsedTIme(time.Now())

	path := "city.list.bin"

	f, err := os.Create(path)
	if err != nil {
		log.Println(err)
	}
	for _, candidateCity := range cities {
		//		log.Println(*candidateCity)
		err = binary.Write(f, binary.LittleEndian, candidateCity)
		if err != nil {
			log.Println(err)
		}
	}
	err = f.Close()
	if err != nil {
		log.Println(err)
	}
}
*/

/*
func dumpToDB() {
	defer erapse.ShowErapsedTIme(time.Now())

	path := dbFileName

	if err := initializeSQL(path); err != nil {
		log.Println(err)
	} else {
		var tx *sql.Tx
		if tx, err = createTransaction(); err != nil {
			log.Println(err)
		} else {
			var stmt *sql.Stmt
			if stmt, err = prepare(tx); err != nil {
				log.Println(err)
			} else {
				for _, candidateCity := range cities {
					//		log.Println(*candidateCity)
					if err = transactionAdd(stmt, candidateCity); err != nil {
						log.Println(err)
					}
				}
				if err = transactionCommit(tx); err != nil {
					log.Println(err)
				}
			}
		}

	}
	query()
}
*/
