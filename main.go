package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"

	"github.com/joho/godotenv"
)

type ViaCEPResponse struct {
	CEP         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	UF          string `json:"uf"`
	IBGE        string `json:"ibge"`
	Erro        bool   `json:"erro"`
}

type WeatherAPIResponse struct {
	Location struct {
		Name      string  `json:"name"`
		Region    string  `json:"region"`
		Country   string  `json:"country"`
		Lat       float64 `json:"lat"`
		Lon       float64 `json:"lon"`
		TimeZone  string  `json:"tz_id"`
		LocalTime string  `json:"localtime"`
	} `json:"location"`
	Current struct {
		TempC float64 `json:"temp_c"`
	} `json:"current"`
}

type TemperatureResponse struct {
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func main() {
	godotenv.Load()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.HandleFunc("/weather/", handleWeatherRequest)
	http.HandleFunc("/health", healthCheckHandler)
	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{'status':'healthy'}"))
}

func handleWeatherRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	cep := r.URL.Path[len("/weather/"):]
	if !isValidCEP(cep) {
		respondWithError(w, http.StatusUnprocessableEntity, "invalid zipcode")
		return
	}
	location, err := getLocationByCEP(cep)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "can not find zipcode")
		return
	}
	temp, err := getTemperatureByLocation(location.Localidade)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error fetching temperature data")
		return
	}
	tempC := temp.Current.TempC
	tempF := tempC*1.8 + 32
	tempK := tempC + 273.15
	response := TemperatureResponse{
		TempC: tempC,
		TempF: tempF,
		TempK: tempK,
	}
	jsonResponse, _ := json.Marshal(response)
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func isValidCEP(cep string) bool {
	re := regexp.MustCompile(`[^0-9]`)
	cleanCEP := re.ReplaceAllString(cep, "")
	return len(cleanCEP) == 8
}

func getLocationByCEP(cep string) (*ViaCEPResponse, error) {
	re := regexp.MustCompile(`[^0-9]`)
	cleanCEP := re.ReplaceAllString(cep, "")
	url := fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cleanCEP)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var location ViaCEPResponse
	if err := json.Unmarshal(body, &location); err != nil {
		return nil, err
	}
	if location.Erro || location.Localidade == "" {
		return nil, fmt.Errorf("CEP not found")
	}
	return &location, nil
}

func getTemperatureByLocation(city string) (*WeatherAPIResponse, error) {
	apiKey := os.Getenv("WEATHER_API_KEY")
	if apiKey == "" {
		apiKey = "sua_chave_api"
	}
	paramEncoded := url.QueryEscape(city)
	url := fmt.Sprintf("http://api.weatherapi.com/v1/current.json?key=%s&q=%s&aqi=no", "cfbabc7298d04b5896f154926252104", paramEncoded)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("weather API returned status code %d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var weather WeatherAPIResponse
	if err := json.Unmarshal(body, &weather); err != nil {
		return nil, err
	}
	return &weather, nil
}

func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	errorResponse := ErrorResponse{Message: message}
	jsonResponse, _ := json.Marshal(errorResponse)
	w.WriteHeader(statusCode)
	w.Write(jsonResponse)
}
