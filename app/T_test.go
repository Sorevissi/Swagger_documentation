package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockCallData struct {
	Response *Response
	Err      error
}

func (m *MockCallData) callData() (*Response, error) {
	return m.Response, m.Err
}

type MockSearchRequest struct {
	Query string `json:"query"`
	MockCallData
}

type MockGeocodeRequest struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
	MockCallData
}

func (sr *MockSearchRequest) callData() (*Response, error) {
	return sr.MockCallData.callData()
}

func TestSearchAddressHandler(t *testing.T) {
	mockSearchRequest := MockSearchRequest{Query: "Test Address"}

	mockSearchResponse := &Response{
		AddressesSearch: []*AddressSearch{
			{
				Source: "Mock Source",
				Result: "Mock Result",
				Metro: []struct {
					Distance float64 `json:"distance"`
					Line     string  `json:"line"`
					Name     string  `json:"name"`
				}{{Distance: 1.5, Line: "M1", Name: "Mock Metro"}},
			},
		},
	}

	mockSearchRequest.MockCallData.Response = mockSearchResponse
	mockSearchRequest.MockCallData.Err = nil

	requestBody, _ := json.Marshal(mockSearchRequest)
	request := httptest.NewRequest("POST", "/api/address/search", bytes.NewBuffer(requestBody))
	recorder := httptest.NewRecorder()

	SearchAddressHandler(recorder, request)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, status)
	}

	var responseJSON []*AddressSearch
	if err := json.Unmarshal(recorder.Body.Bytes(), &responseJSON); err != nil {
		t.Errorf("Error decoding response JSON: %v", err)
	}

	if len(responseJSON) != len(mockSearchResponse.AddressesSearch) {
		t.Errorf("Expected %d results, but got %d", len(mockSearchResponse.AddressesSearch), len(responseJSON))
	}
}

func TestGeocodeHandler(t *testing.T) {
	mockGeocodeRequest := MockGeocodeRequest{Lat: "12.34", Lon: "56.78"}

	mockGeocodeResponse := &Response{
		AddressesGeo: &AddressGeo{
			Suggestions: []struct {
				Value             string `json:"value"`
				UnrestrictedValue string `json:"unrestricted_value"`
				Data              struct {
					PostalCode string `json:"postal_code"`
					Country    string `json:"country"`
				} `json:"data"`
			}{
				{
					Value:             "Mock Value",
					UnrestrictedValue: "Mock Unrestricted Value",
					Data: struct {
						PostalCode string `json:"postal_code"`
						Country    string `json:"country"`
					}{
						PostalCode: "12345",
						Country:    "Mock Country",
					},
				},
			},
		},
	}

	mockGeocodeRequest.MockCallData.Response = mockGeocodeResponse
	mockGeocodeRequest.MockCallData.Err = nil

	requestBody, _ := json.Marshal(mockGeocodeRequest)
	request := httptest.NewRequest("POST", "/api/address/geocode", bytes.NewBuffer(requestBody))
	recorder := httptest.NewRecorder()

	GeocodeHandler(recorder, request)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, status)
	}

	var responseJSON *AddressGeo
	if err := json.Unmarshal(recorder.Body.Bytes(), &responseJSON); err != nil {
		t.Errorf("Error decoding response JSON: %v", err)
	}
}
