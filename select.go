package main

import (
  "net/http"
  "log"
  "html/template"
  "strconv"
)


type Coordinates struct {
	Name       string
	Value      float64
	Text       string
}

type PageVariables struct {
  PageTitle        string
  PageCoordinates []Coordinates
  Answer          []float64
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


func DisplayCoordinates(w http.ResponseWriter, r *http.Request){

   Title := "Solar Energy"
   MyCoordinates := []Coordinates{
     Coordinates{"coordinaten", 0, "North"},
     Coordinates{"coordinates", 0, "South"},
   }

  MyPageVariables := PageVariables{
    PageTitle: Title,
    PageCoordinates : MyCoordinates,
    }

   t, err := template.ParseFiles("select.html") //parse the html file homepage.html
   if err != nil { // if there is an error
     log.Print("template parsing error: ", err) // log it
   }

   err = t.Execute(w, MyPageVariables) //execute the template and pass it the HomePageVars struct to fill in the gaps
   if err != nil { // if there is an error
     log.Print("template executing error: ", err) //log it
   }

}

func UserSelected(w http.ResponseWriter, r *http.Request){
  r.ParseForm()

  northcoord,_ := strconv.ParseFloat(r.Form.Get("coordinaten"), 64)
  southcoord,_ := strconv.ParseFloat(r.Form.Get("coordinates"), 64)

  Title := "Your coordinates"
  MyPageVariables := PageVariables{
    PageTitle: Title,
    Answer: []float64{northcoord, southcoord},
    }

 // generate page by passing page variables into template
    t, err := template.ParseFiles("select.html") //parse the html file homepage.html
    if err != nil { // if there is an error
      log.Print("template parsing error: ", err) // log it
    }

    err = t.Execute(w, MyPageVariables) //execute the template and pass it the HomePageVars struct to fill in the gaps
    if err != nil { // if there is an error
      log.Print("template executing error: ", err) //log it
    }
}
