package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

type Data struct {
	Response []Response `json:"data"`
}

type Response struct {
	City               string              `json:"city"`
	IradianceResponses []IradianceResponse `json:"iradianceResponses"`
}

type IradianceResponse struct {
	Month string  `json:"month"`
	Ed    float64 `json:"ed"`
	Em    float64 `json:"em"`
	Hd    float64 `json:"hd"`
	Hm    float64 `json:"hm"`
	Sdm   float64 `json:"sdm"`
}

type City struct {
	name      string
	latitude  float64
	longitude float64
}

var cities []City

func initCities() {
	cities = []City{
		City{
			name:      "Paris",
			longitude: 2.3522,
			latitude:  48.566,
		},
		City{
			name:      "Lyon",
			longitude: 4.8357,
			latitude:  45.7640,
		},
		City{
			name:      "Marseille",
			longitude: 5.3698,
			latitude:  43.2965,
		},
		City{
			name:      "Bordeaux",
			longitude: 0.5792,
			latitude:  44.8378,
		},
		City{
			name:      "Lille",
			longitude: 3.0573,
			latitude:  50.6292,
		},
	}
}

func getCall(w http.ResponseWriter, r *http.Request) {
	var responses []Response
	ch := make(chan Response)
	for _, city := range cities {
		go MakeRequest(city, ch)
	}

	for range cities {
		resp := <-ch
		responses = append(responses, resp)
	}
	data := &Data{
		Response: responses,
	}
	responseFinal, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(responseFinal)
}

func parseBody(body string, cityName string) Response {
	temp := strings.Split(body, "\n")
	i := 10
	var iradianceResponses []IradianceResponse
	for i < 22 {
		temp2 := strings.Split(temp[i], "		")
		month := temp2[0]
		ed, _ := strconv.ParseFloat(temp2[1], 64)
		em, _ := strconv.ParseFloat(temp2[2], 64)
		hd, _ := strconv.ParseFloat(temp2[3], 64)
		hm, _ := strconv.ParseFloat(temp2[4], 64)
		sdmSize := len(temp2[5])
		sdmTemp := temp2[5]
		sdmTemp = sdmTemp[:sdmSize-1]
		sdm, _ := strconv.ParseFloat(sdmTemp, 64)

		iradianceResponses = append(iradianceResponses, IradianceResponse{
			Month: month,
			Ed:    ed,
			Em:    em,
			Hd:    hd,
			Hm:    hm,
			Sdm:   sdm,
		})
		i++
	}
	return Response{
		City:               cityName,
		IradianceResponses: iradianceResponses,
	}

}

func MakeRequest(city City, ch chan<- Response) {
	resp, _ := http.Get(fmt.Sprintf("http://re.jrc.ec.europa.eu/pvgis5/PVcalc.php?lat=%g&lon=%g&peakpower=3&loss=22&outputformat=csv", city.latitude, city.longitude))
	body, _ := ioutil.ReadAll(resp.Body)
	response := parseBody(string(body), city.name)
	ch <- response
}

func main() {
	initCities()
	router := mux.NewRouter()
	router.HandleFunc("/call", getCall).Methods("GET")
	http.ListenAndServe(":8000", router)
}
