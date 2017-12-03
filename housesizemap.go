package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func DisplayHouseSize(w http.ResponseWriter, r *http.Request) {
	PageTitle := "Heat Map"

	MyHouse := []House{
		House{"housesizeinput", 0, "Size"},
	}

	PageVars := PageVariables{
		PageTitle:     PageTitle,
		PageHouseSize: MyHouse,
	}

	t, err := template.ParseFiles("housesizemap.html") //Parse the html file housesizemap.html
	if err != nil {                                    // if error
		log.Print("template parsing error:", err)
	}

	err = t.Execute(w, PageVars) //execute the template and pass it the PageVars struct to fill in the gaps
	if err != nil {              // if error
		log.Print("template executing error:", err) //log it
	}

}

func UserInteracts(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	red := "red"
	yellow := "yellow"
	green := "green"
	cityData := MakeCityMap("energy.csv")
	houseSize, _ := strconv.ParseFloat(r.Form.Get("housesizeinput"), 64)
	heatMap := MakeColorMarkers(cityData, houseSize)
	mapColors := MakeColors("energy.csv", cityData, houseSize)
	redList := MakeList(heatMap, red)
	yellowList := MakeList(heatMap, yellow)
	greenList := MakeList(heatMap, green)
	redPercentage := ColorPercent(heatMap, red)
	yellowPercentage := ColorPercent(heatMap, yellow)
	greenPercentage := ColorPercent(heatMap, green)
	Title := "House Size Map"

	PageVars := PageVariables{
		PageTitle:     Title,
		Map:           mapColors,
		RedList:       redList,
		YellowList:    yellowList,
		GreenList:     greenList,
		RedPercent:    redPercentage,
		YellowPercent: yellowPercentage,
		GreenPercent:  greenPercentage,
	}

	t, err := template.ParseFiles("housesizemap.html")
	if err != nil {
		log.Print("template parsing error: ", err)
	}

	err = t.Execute(w, PageVars)
	if err != nil {
		log.Print("template executing error: ", err)
	}
}

//Makes a map of color markers for each city based on chose house size and difference in output
func MakeColorMarkers(cityData map[string]City, houseSize float64) map[string]string {
	var output, avgEnergy float64
	var mapColor string
	colors := make(map[string]string)
	for cityName, _ := range cityData {
		output = SolarOutput(cityName, cityData, "horizontal", 15, houseSize)
		avgEnergy = AverageEnergy(cityData, cityName) * houseSize
		mapColor = MapColor(avgEnergy, output)
		colors[cityName] = mapColor
	}
	return colors
}

//computes differece in energy and chooses color
func MapColor(avgEnergy, energyOutput float64) string {
	energyDiff := energyOutput - avgEnergy
	var color string
	if energyDiff > 50 {
		color = "green"
	} else if energyDiff >= 0 {
		color = "yellow"
	} else {
		color = "red"
	}
	return color
}

//makes list of cities that have a certain color code
func MakeList(colors map[string]string, color string) []string {
	cityList := make([]string, 0)
	for cityName, mapColor := range colors {
		if mapColor == color {
			cityList = append(cityList, cityName)
		}
	}
	return cityList
}

//Computes the percentage of the cities that are difined as a certain color
func ColorPercent(colors map[string]string, color string) float64 {
	var colorCount int
	for _, mapColor := range colors {
		if mapColor == color {
			colorCount++
		}
	}
	return float64(colorCount) * (100 / 98)
	//since there are 98 cities
}

//Make an array of the city names.
func MakeCityArray(filename string) []string {
	lines := ReadFile(filename)
	cityArray := make([]string, 0)
	for i := 0; i < len(lines); i++ {
		var items []string = strings.Split(lines[i], ",")
		cityArray = append(cityArray, items[0])
	}
	return cityArray
}

//Make an array of colors based alphabetically.
func MakeColors(filename string, cityData map[string]City, houseSize float64) []string {
	var output, avgEnergy float64
	var mapColor string
	cityNames := MakeCityArray(filename)
	colors := make([]string, 0)
	for _, cityName := range cityNames {
		output = SolarOutput(cityName, cityData, "horizontal", 15, houseSize)
		avgEnergy = AverageEnergy(cityData, cityName) * houseSize
		mapColor = MapColor(avgEnergy, output)
		if mapColor == "red" {
			mapColor = "#FF0000"
		} else if mapColor == "yellow" {
			mapColor = "#FFFF00"
		} else {
			mapColor = "#008000"
		}
		colors = append(colors, mapColor)
	}
	return colors
}