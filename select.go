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

type PageVariables struct {
	PageTitle       string
	PageCoordinates []Coordinates
	Answer          string
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

	MyPageVariables := PageVariables{
		PageTitle:       Title,
		PageCoordinates: MyCoordinates,
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

	Title := "Your coordinates"
	MyPageVariables := PageVariables{
		PageTitle: Title,
		Answer:    closestcity,
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
