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

type Panel struct {
	efficiency float64
	watts      float64
	area       float64
	price      float64
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
	MyCity          string
	Output          float64
	OptAngle        float64
	OptOutput       float64
	Usage           float64
	Optimal         string
	InstCost        float64
	Companies       []string
	NumPanels       []int
	PanelCost       []int
	Recommendation  []string
	Map             []string
	RedList         []string
	YellowList      []string
	GreenList       []string
	RedPercent      float64
	YellowPercent   float64
	GreenPercent    float64
}

func main() {
	http.HandleFunc("/", DisplayCoordinates)
	http.HandleFunc("/selected", UserSelected)
	http.HandleFunc("/heatmap", DisplayHouseSize)
	http.HandleFunc("/displayheatmap", UserInteracts)
	log.Fatal(http.ListenAndServe(getPort(), nil))
}

func getPort() string {
	p := os.Getenv("PORT")
	if p != "" {
		return ":" + p
	}
	return ":8080"
}

//This is the function where it initially displays the page asking for the
//user's info such as house size and coordinates and then loads
//to the next page after this information is submitted.
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

//This is the main function where the user interacts with the web app.
//There are several different variables in use here to be able to interact
//with.
func UserSelected(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	cityData := MakeCityMap("energy.csv")
	solarPanels := MakeSolarMap("solar.csv")
	northcoord, _ := strconv.ParseFloat(r.Form.Get("coordinaten"), 64)
	westcoord, _ := strconv.ParseFloat(r.Form.Get("coordinatew"), 64)
	houseSize, _ := strconv.ParseFloat(r.Form.Get("housesize"), 64)
	closestcity := ClosestCity(cityData, northcoord, westcoord)
	solarOutput := SolarOutput(closestcity, cityData, "horizontal", 15, houseSize)
	solarOutput = float64(int(solarOutput*100)) / 100
	optAngle := OptAngle(cityData, closestcity)
	optEnergy := OptEnergy(cityData, closestcity, 15, houseSize)
	optEnergy = float64(int(optEnergy*100)) / 100
	avgUsage := AverageEnergy(cityData, closestcity) * houseSize
	avgUsage = float64(int(avgUsage*100)) / 100
	recommendation := IsItOptimal(avgUsage, solarOutput)
	companylist := Companies(closestcity, cityData)
	instCost := InstallationCost(cityData, closestcity)
	numPanels, panelCost := CalcCostBrand(solarOutput, houseSize, cityData, closestcity, solarPanels)
	preferences := Preferences(panelCost, solarPanels, closestcity, cityData, houseSize)

	Title := "Your Home"
	MyPageVariables := PageVariables{
		PageTitle:      Title,
		MyCity:         closestcity,
		Output:         solarOutput,
		OptAngle:       optAngle,
		OptOutput:      optEnergy,
		Usage:          avgUsage,
		Optimal:        recommendation,
		InstCost:       instCost,
		Companies:      companylist,
		NumPanels:      numPanels,
		PanelCost:      panelCost,
		Recommendation: preferences,
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

//Read in the file.
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

//Makes a map data structure of all of the City objects.
func MakeCityMap(filename string) map[string]City {
	lines := ReadFile(filename)
	cityData := make(map[string]City)
	for i := 0; i < len(lines); i++ {
		var items []string = strings.Split(lines[i], ",")
		cityName := items[0]
		cityData[cityName] = MakeCity(cityData, items)
	}
	delete(cityData, "")
	return cityData
}

//Creates a City object with its characteristics.
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

//Make the map data structure of all of the different solar panel brands.
func MakeSolarMap(filename string) map[string]Panel {
	lines := ReadFile(filename)
	solarPanels := make(map[string]Panel)
	for i := 0; i < len(lines); i++ {
		var items []string = strings.Split(lines[i], ",")
		panelName := items[0]
		solarPanels[panelName] = MakePanel(solarPanels, items)
	}
	keys := make([]string, 0)
	for key, _ := range solarPanels {
		keys = append(keys, key)
	}
	return solarPanels
}

//Make a Solar Panel object.
func MakePanel(solarPanels map[string]Panel, items []string) Panel {
	var panel Panel
	panel.efficiency, _ = strconv.ParseFloat(items[1], 64)
	panel.watts, _ = strconv.ParseFloat(items[2], 64)
	panel.area, _ = strconv.ParseFloat(items[3], 64)
	panel.price, _ = strconv.ParseFloat(items[4], 64)
	return panel
}

//Finds the city closest to their coordinates.
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

//Calculates expected generated energy from solar panels. (kwh per month)
func SolarOutput(cityName string, cityData map[string]City, angleType string, efficiency, houseSize float64) float64 {
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
	energyOutput := houseSize * efficiency * radiation * 0.75
	return energyOutput / 12
}

//Gives the potential optimal energy output for their home.
func OptEnergy(cityData map[string]City, cityName string, efficiency, houseSize float64) float64 {
	optOutput := SolarOutput(cityName, cityData, "optimal", efficiency, houseSize)
	return optOutput
}

//Output the optimal angle that they should use to get the optimal output.
func OptAngle(cityData map[string]City, cityName string) float64 {
	data := cityData[cityName]
	optAngle := data.optAng
	return optAngle
}

//Calculates average energy required at a house in their area. (kwh per month)
func AverageEnergy(cityData map[string]City, cityName string) float64 {
	data := cityData[cityName]
	averageEnergy := data.avgEnergy
	return averageEnergy / 2000
}

//Gives a recommendation based on energy produced from solar panels and energy requirement.
func IsItOptimal(avgUsage float64, solarOutput float64) string {
	solarOutput = solarOutput
	deficit := solarOutput - avgUsage
	if deficit <= 0 {
		return "is not recommended"
	} else if deficit < 50 && deficit > 0 {
		return "is recommended"
	} else {
		return "is highly recommended"
	}
}

//Calculates the installation cost.
func InstallationCost(cityData map[string]City, cityName string) float64 {
	return cityData[cityName].instCost * 5000
}

//Calculates the number of solar panels needed on their house.
func NumSolarPanels(energyOutput, houseSize float64, cityData map[string]City, panelName, cityName string, solarPanels map[string]Panel) int {
	houseSize *= 0.092903 //convert square feet to square meters
	oneSolarPanelOutput := (energyOutput * 12 / houseSize) * solarPanels[panelName].area
	numPanels := cityData[cityName].avgEnergy / oneSolarPanelOutput
	return int(numPanels)
}

//Calculates how much it would cost for user to get that brand of solar panels on their house.
func SolarPanelCost(energyOutput, houseSize float64, cityData map[string]City, panelName, cityName string, solarPanels map[string]Panel, numPanels int) float64 {
	houseSize *= 0.092903 //convert square feet to square meters
	cost := solarPanels[panelName].price * float64(numPanels)
	return cost + InstallationCost(cityData, cityName)
}

//Puts companies close to their city into a slice of strings.
func Companies(cityName string, cityData map[string]City) []string {
	data := cityData[cityName]
	companyNames := data.companies
	return companyNames
}

//Calculates the cost and number of panels required for each brand of solar panel.
//0: Suntech, 1: Samsung, 2: Kyocera, 3: Canadian Solar, 4: Grape Solar 390W, 5: Grape Solar 250
func CalcCostBrand(energyOutput, houseSize float64, cityData map[string]City, cityName string, solarPanels map[string]Panel) ([]int, []int) {
	NumPanels := make([]int, 6)
	PanelCosts := make([]int, 6)
	NumPanels[0] = NumSolarPanels(energyOutput, houseSize, cityData, "Suntech", cityName, solarPanels)
	NumPanels[1] = NumSolarPanels(energyOutput, houseSize, cityData, "Samsung", cityName, solarPanels)
	NumPanels[2] = NumSolarPanels(energyOutput, houseSize, cityData, "Kyocera", cityName, solarPanels)
	NumPanels[3] = NumSolarPanels(energyOutput, houseSize, cityData, "CanadianSolar", cityName, solarPanels)
	NumPanels[4] = NumSolarPanels(energyOutput, houseSize, cityData, "GrapeSolar390W", cityName, solarPanels)
	NumPanels[5] = NumSolarPanels(energyOutput, houseSize, cityData, "GrapeSolar250", cityName, solarPanels)
	PanelCosts[0] = int(SolarPanelCost(energyOutput, houseSize, cityData, "Suntech", cityName, solarPanels, NumPanels[0]))
	PanelCosts[1] = int(SolarPanelCost(energyOutput, houseSize, cityData, "Samsung", cityName, solarPanels, NumPanels[1]))
	PanelCosts[2] = int(SolarPanelCost(energyOutput, houseSize, cityData, "Kyocera", cityName, solarPanels, NumPanels[2]))
	PanelCosts[3] = int(SolarPanelCost(energyOutput, houseSize, cityData, "CanadianSolar", cityName, solarPanels, NumPanels[3]))
	PanelCosts[4] = int(SolarPanelCost(energyOutput, houseSize, cityData, "GrapeSolar390W", cityName, solarPanels, NumPanels[4]))
	PanelCosts[5] = int(SolarPanelCost(energyOutput, houseSize, cityData, "GrapeSolar250", cityName, solarPanels, NumPanels[5]))
	return NumPanels, PanelCosts
}

//Preferences in a slice, with 0: min cost, 1: max output, 2: max efficiency
func Preferences(panelCost []int, solarPanels map[string]Panel, cityName string, cityData map[string]City, houseSize float64) []string {
	minCostPanel := FindMinCostPanel(panelCost)
	efficiencyarray := MakeEfficiencyArray(solarPanels)
	mostEfficientPanel := FindMostEfficient(efficiencyarray)
	maxOutput := FindMaxOutput(efficiencyarray, cityName, cityData, houseSize)
	preferences := []string{minCostPanel, maxOutput, mostEfficientPanel}
	return preferences
}

//Converts index to panel brand name.
func idxToPanel(idx int) string {
	switch idx {
	case 0:
		return "Suntech"
	case 1:
		return "Samsung"
	case 2:
		return "Kyocera"
	case 3:
		return "CanadianSolar"
	case 4:
		return "GrapeSolar390W"
	case 5:
		return "GrapeSolar250"
	}
	return ""
}

//Puts panel brand efficiencies in an array according to the same indices as above.
func MakeEfficiencyArray(solarPanels map[string]Panel) []float64 {
	efficiencyArray := make([]float64, 6)
	efficiencyArray[0] = solarPanels["Suntech"].efficiency
	efficiencyArray[1] = solarPanels["Samsung"].efficiency
	efficiencyArray[2] = solarPanels["Kyocera"].efficiency
	efficiencyArray[3] = solarPanels["CanadianSolar"].efficiency
	efficiencyArray[4] = solarPanels["GrapeSolar390W"].efficiency
	efficiencyArray[5] = solarPanels["GrapeSolar250"].efficiency
	return efficiencyArray
}

//Gives the minimum cost panel option.
func FindMinCostPanel(panelCost []int) string {
	minCost := panelCost[0]
	minCostIDX := 0
	for idx, cost := range panelCost {
		if cost < minCost {
			minCost = cost
			minCostIDX = idx
		}
	}
	minCostPanel := idxToPanel(minCostIDX)
	return minCostPanel
}

//Finds the brand of solar panel with the highest efficiency.
func FindMostEfficient(efficiencyarray []float64) string {
	mostEfficient := efficiencyarray[0]
	mostEfficientIDX := 0
	for idx, efficiency := range efficiencyarray {
		if efficiency > mostEfficient {
			mostEfficient = efficiency
			mostEfficientIDX = idx
		}
	}
	mostEfficientPanel := idxToPanel(mostEfficientIDX)
	return mostEfficientPanel
}

//Finds the panel brand with the highest output of solar energy.
func FindMaxOutput(efficiencyarray []float64, cityName string, cityData map[string]City, houseSize float64) string {
	var maxOutput, output float64
	var maxIDX int
	for i := range efficiencyarray {
		output = SolarOutput(cityName, cityData, "horizontal", efficiencyarray[i], houseSize)
		if output > maxOutput {
			maxOutput = output
			maxIDX = i
		}
	}
	maxOutputPanel := idxToPanel(maxIDX)
	return maxOutputPanel
}
