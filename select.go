package main

import (
	"bufio"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type City struct {
	coordN    float64
	coordW    float64
	temp      float64
	solarRad  float64
	optAng    float64
	optRad    float64
	avgEnergy float64
	instCost  float64
	companies []string
}
type Coordinates struct {
	Name  string
	Value float64
	Text  string
}

type House struct {
	Name  string
	Value float64
	Text  string
}

type PageVariables struct {
	PageTitle       string
	PageCoordinates []Coordinates
	PageHouseSize   []House
	Answer          string
	Value           float64
	Value2          float64
	Value3          float64
	Usage           float64
	Optimal         string
}

func main() {
	http.HandleFunc("/", DisplayCoordinates)
	http.HandleFunc("/selected", UserSelected)
	log.Fatal(http.ListenAndServe(getPort(), nil))
}

func getPort() string {
	p := os.Getenv("PORT")
	if p != "" {
		return ":" + p
	}
	return ":8080"
}

func DisplayCoordinates(w http.ResponseWriter, r *http.Request) {

	Title := "Solar Energy"
	MyCoordinates := []Coordinates{
		Coordinates{"coordinaten", 0, "North"},
		Coordinates{"coordinatew", 0, "West"},
	}
	MyHouse := []House{
		House{"housesize", 0, "Size"},
	}

	MyPageVariables := PageVariables{
		PageTitle:       Title,
		PageCoordinates: MyCoordinates,
		PageHouseSize:   MyHouse,
	}

	t, err := template.ParseFiles("select.html") //parse the html file homepage.html
	if err != nil {                              // if there is an error
		log.Print("template parsing error: ", err) // log it
	}

	err = t.Execute(w, MyPageVariables) //execute the template and pass it the HomePageVars struct to fill in the gaps
	if err != nil {                     // if there is an error
		log.Print("template executing error: ", err) //log it
	}

}

func UserSelected(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	cityData := MakeCityMap("energy.csv")
	northcoord, _ := strconv.ParseFloat(r.Form.Get("coordinaten"), 64)
	westcoord, _ := strconv.ParseFloat(r.Form.Get("coordinatew"), 64)
	closestcity := ClosestCity(cityData, northcoord, westcoord)
	houseSize, _ := strconv.ParseFloat(r.Form.Get("housesize"), 64)
	solarOutput := SolarOutput(closestcity, cityData, "horizontal", houseSize)
	solarOutput = float64(int(solarOutput*100)) / 100
	optAngle := OptAngle(cityData, closestcity)
	optEnergy := OptEnergy(cityData, closestcity, houseSize)
	optEnergy = float64(int(optEnergy*100)) / 100
	avgUsage := AverageEnergy(cityData, closestcity) * houseSize
	avgUsage = float64(int(avgUsage*100)) / 100
	recommendation := IsItOptimal(avgUsage, solarOutput)

	Title := "Your Home"
	MyPageVariables := PageVariables{
		PageTitle: Title,
		Answer:    closestcity,
		Value:     solarOutput,
		Value2:    optAngle,
		Value3:    optEnergy,
		Usage:     avgUsage,
		Optimal:   recommendation,
	}

	// generate page by passing page variables into template
	t, err := template.ParseFiles("select.html") //parse the html file homepage.html
	if err != nil {                              // if there is an error
		log.Print("template parsing error: ", err) // log it
	}

	err = t.Execute(w, MyPageVariables) //execute the template and pass it the HomePageVars struct to fill in the gaps
	if err != nil {                     // if there is an error
		log.Print("template executing error: ", err) //log it
	}
}

func ReadFile(filename string) []string {
	//reads file and makes a line for each file
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error: couldn't open the file")
		os.Exit(1)
	}
	lines := make([]string, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if scanner.Err() != nil {
		fmt.Println("Error: there was some kind of error during the file reading")
		os.Exit(1)
	}
	file.Close()
	return lines
}
func MakeCityMap(filename string) map[string]City {
	lines := ReadFile(filename)
	cityData := make(map[string]City)
	for i := 0; i < len(lines); i++ {
		var items []string = strings.Split(lines[i], ",")
		cityName := items[0]
		cityData[cityName] = MakeCity(cityData, items)
	}
	delete(cityData, "")
	keys := make([]string, 0)
	for key, _ := range cityData {
		keys = append(keys, key)
	}
	return cityData
}

func MakeCity(cityData map[string]City, items []string) City {
	var city City
	city.coordN, _ = strconv.ParseFloat(items[1], 64)
	city.coordW, _ = strconv.ParseFloat(items[2], 64)
	city.temp, _ = strconv.ParseFloat(items[3], 64)
	city.solarRad, _ = strconv.ParseFloat(items[4], 64)
	city.optAng, _ = strconv.ParseFloat(items[5], 64)
	city.optRad, _ = strconv.ParseFloat(items[6], 64)
	city.avgEnergy, _ = strconv.ParseFloat(items[7], 64)
	city.instCost, _ = strconv.ParseFloat(items[8], 64)
	var companyNames []string = strings.Split(items[9], ";")
	for i := range companyNames {
		city.companies = append(city.companies, companyNames[i])
	}
	return city
}

func ClosestCity(cityData map[string]City, userCoordN, userCoordW float64) string {
	distance := 10000.00
	closestCityName := ""
	var cityDistance float64
	for city, data := range cityData {
		cityDistance = math.Sqrt(math.Pow((userCoordN-data.coordN), 2) + math.Pow((userCoordW-data.coordW), 2))
		if cityDistance < distance {
			distance = cityDistance
			closestCityName = city
		}
	}
	return closestCityName
}

func SolarOutput(cityName string, cityData map[string]City, angleType string, houseSize float64) float64 {
	var radiation float64
	//241.5479 meters squared as solar panel area (average house size)
	//assume standard 15% efficiency
	//0.75 default performance ratio
	// E = Solar Panel Area * solar panel efficiency * radiation * performance ratio
	if angleType == "horizontal" {
		radiation = cityData[cityName].solarRad
	} else if angleType == "optimal" {
		radiation = cityData[cityName].optRad
	}
	houseSize *= 0.092903 //convert square feet to square meters
	energyOutput := houseSize * 15 * radiation * 0.75
	return energyOutput / 12
}

func OptEnergy(cityData map[string]City, cityName string, houseSize float64) float64 {
	optOutput := SolarOutput(cityName, cityData, "optimal", houseSize)
	return optOutput
}

func OptAngle(cityData map[string]City, cityName string) float64 {
	data := cityData[cityName]
	optAngle := data.optAng
	return optAngle
}

func AverageEnergy(cityData map[string]City, cityName string) float64 {
	data := cityData[cityName]
	averageEnergy := data.avgEnergy
	return averageEnergy / 2000
}

func IsItOptimal(avgUsage float64, solarOutput float64) string {
	solarOutput = solarOutput
	deficit := solarOutput - avgUsage
	if deficit <= 0 {
		return "is not recommended"
	} else {
		return "is recommended"
	}
}
