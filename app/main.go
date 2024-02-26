package main

import (
	_ "GeoAPI/app/docs"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/swaggo/http-swagger"
	"io/ioutil"
	"log"
	"net/http"
)

type Call interface {
	callData() (*Response, error)
}

// @name SearchRequest
// @description SearchRequest represents the request body for address search
type SearchRequest struct {
	// The search query
	// @jsonQuery query
	// @example Example search query: "123 Main St"
	Query string `json:"query"`
}

// @name GeocodeRequest
// @description GeocodeRequest represents the request body for address geocoding.
type GeocodeRequest struct {
	// The latitude
	// @jsonQuery query
	// @example Example latitude: "40.7128"
	Lat string `json:"lat"`
	// The longitude
	// @jsonQuery query
	// @example Example longitude: "-74.0060"
	Lon string `json:"lon"`
}

// @name SearchResponse
// @description SearchResponse represents the response body for address search.
type SearchResponse struct {
	// The list of addresses matching the search query
	// @jsonField addressesSearch
	// @example Example response:
	// [{"source": "Dadata", "result": "123 Main St", "metro": []}]
	AddressesSearch []*AddressSearch `json:"addressesSearch"`
}

// @name GeocodeResponse
// @description GeocodeResponse represents the response body for address geocoding.
type GeocodeResponse struct {
	// The geocode result for the given coordinates
	// @jsonField addressesGeo
	// @example Example response:
	// {"suggestions": [{"value": "123 Main St", "unrestricted_value": "123 Main St, City", "data": {"postal_code": "12345", "country": "USA"}}]}
	AddressesGeo *AddressGeo `json:"addressesGeo"`
}

func (sr *SearchRequest) callData() (*Response, error) {
	url := "https://cleaner.dadata.ru/api/v1/clean/address"
	requestData := []string{sr.Query}

	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Token 9a84b6e525fb548e7170b77175e9e15af84a30ac")
	req.Header.Add("X-Secret", "6ecfe8510311d14daf5de31de9a5af4ceeb5b0d5")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var searchResponse Response
	if err := json.Unmarshal(body, &searchResponse.AddressesSearch); err != nil {
		return nil, err
	}
	return &searchResponse, nil
}

func (gr *GeocodeRequest) callData() (*Response, error) {
	url := "http://suggestions.dadata.ru/suggestions/api/4_1/rs/geolocate/address"

	requestBody, err := json.Marshal(gr)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Token 9a84b6e525fb548e7170b77175e9e15af84a30ac")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var geocodeResponse Response
	if err := json.Unmarshal(body, &geocodeResponse.AddressesGeo); err != nil {
		return nil, err
	}

	log.Println(geocodeResponse.AddressesGeo)
	return &geocodeResponse, nil
}

// @name AddressGeo
// @description AddressGeo represents the geocode result for an address.
type AddressGeo struct {
	// The list of suggestions for the geocode result
	// @jsonField suggestions
	// @example Example suggestions: [{"value": "123 Main St", "unrestricted_value": "123 Main St, City", "data": {"postal_code": "12345", "country": "USA"}}]
	Suggestions []struct {
		Value             string `json:"value"`
		UnrestrictedValue string `json:"unrestricted_value"`
		Data              struct {
			PostalCode string `json:"postal_code"`
			Country    string `json:"country"`
		} `json:"data"`
	} `json:"suggestions"`
}

// @name AddressSearch
// @description AddressSearch represents the search result for an address.
type AddressSearch struct {
	// The source of the address information
	// @jsonField source
	// @example Dadata
	Source string `json:"source"`
	// The result address
	// @jsonField result
	// @example 123 Main St
	Result string `json:"result"`
	// The list of metro stations nearby
	// @jsonField metro
	// @example Example metro stations: [{"distance": 1.5, "line": "A", "name": "Station 1"}]
	Metro []struct {
		Distance float64 `json:"distance"`
		Line     string  `json:"line"`
		Name     string  `json:"name"`
	} `json:"metro"`
}

type Response struct {
	AddressesGeo    *AddressGeo      `json:"addressesGeo"`
	AddressesSearch []*AddressSearch `json:"addressesSearch"`
}

// @title My API
// @version 1.0
// @description This is a sample API for address searching and geocoding using Dadata API.
// @termsOfService http://localhost:8080/swagger/index.html
// @BasePath /api
func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/api/address/search", SearchAddressHandler)

	r.Post("/api/address/geocode", GeocodeHandler)

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	port := 8080
	fmt.Printf("Server is running on :%d...\n", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), r)
}

// @Summary Search for an address
// @ID searchAddress
// @Accept  json
// @Produce  json
// @Param request body SearchRequest true "Search Request"
// @Success 200 {object} SearchResponse
// @Failure 400 "Invalid request format"
// @Failure 500 "Dadata API error"
// @Router /address/search [post]
func SearchAddressHandler(w http.ResponseWriter, r *http.Request) {
	var searchRequest SearchRequest
	err := json.NewDecoder(r.Body).Decode(&searchRequest)
	if err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	response, err := searchRequest.callData()
	if err != nil {
		http.Error(w, "Dadata API error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response.AddressesSearch)
}

// @Summary Geocode an address
// @ID geocodeAddress
// @Accept  json
// @Produce  json
// @Param request body GeocodeRequest true "Geocode Request"
// @Success 200 {object} GeocodeResponse
// @Failure 400 "Invalid request format"
// @Failure 500 "Dadata API error"
// @Router /address/geocode [post]
func GeocodeHandler(w http.ResponseWriter, r *http.Request) {
	var geocodeRequest GeocodeRequest
	err := json.NewDecoder(r.Body).Decode(&geocodeRequest)
	if err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	response, err := geocodeRequest.callData()
	if err != nil {
		http.Error(w, "Dadata API error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response.AddressesGeo)
}
