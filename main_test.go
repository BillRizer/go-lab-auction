package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestIsValidCEP(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"12345678", true},
		{"12345-678", true},
		{"12.345-678", true},
		{"1234", false},
		{"abcde678", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isValidCEP(tt.input)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

type mockRoundTripper func(req *http.Request) *http.Response

func (m mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m(req), nil
}

func mockHTTPResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

func TestGetLocationByCEP_Success(t *testing.T) {
	expected := `{"cep":"01001-000","logradouro":"Praça da Sé","bairro":"Sé","localidade":"São Paulo","uf":"SP"}`
	http.DefaultClient.Transport = mockRoundTripper(func(req *http.Request) *http.Response {
		return mockHTTPResponse(http.StatusOK, expected)
	})
	defer func() { http.DefaultClient.Transport = nil }()

	location, err := getLocationByCEP("01001-000")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if location.Localidade != "São Paulo" {
		t.Errorf("expected São Paulo, got %s", location.Localidade)
	}
}

func TestGetLocationByCEP_Invalid(t *testing.T) {
	http.DefaultClient.Transport = mockRoundTripper(func(req *http.Request) *http.Response {
		return mockHTTPResponse(http.StatusOK, `{"erro": true}`)
	})
	defer func() { http.DefaultClient.Transport = nil }()

	_, err := getLocationByCEP("00000000")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestGetTemperatureByLocation_Success(t *testing.T) {
	os.Setenv("WEATHER_API_KEY", "fakekey")
	expected := `{"location":{"name":"São Paulo","region":"SP","country":"Brazil","lat":-23.55,"lon":-46.64,"tz_id":"America/Sao_Paulo","localtime":"2023-08-01 10:00"},"current":{"temp_c":25.0}}`

	http.DefaultClient.Transport = mockRoundTripper(func(req *http.Request) *http.Response {
		return mockHTTPResponse(http.StatusOK, expected)
	})
	defer func() { http.DefaultClient.Transport = nil }()

	data, err := getTemperatureByLocation("São Paulo")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if data.Current.TempC != 25.0 {
		t.Errorf("expected 25.0, got %f", data.Current.TempC)
	}
}

func TestHandleWeatherRequest_InvalidCEP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/weather/1234", nil)
	rr := httptest.NewRecorder()

	handleWeatherRequest(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected status 422, got %d", rr.Code)
	}
	var resp ErrorResponse
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp.Message != "invalid zipcode" {
		t.Errorf("unexpected error message: %s", resp.Message)
	}
}

func TestHandleWeatherRequest_Integration(t *testing.T) {
	os.Setenv("WEATHER_API_KEY", "fakekey")

	viaCEPResp := `{"cep":"01001-000","logradouro":"Praça da Sé","bairro":"Sé","localidade":"São Paulo","uf":"SP"}`
	weatherResp := `{"location":{"name":"São Paulo","region":"SP","country":"Brazil","lat":-23.55,"lon":-46.64,"tz_id":"America/Sao_Paulo","localtime":"2023-08-01 10:00"},"current":{"temp_c":25.0}}`

	called := 0
	http.DefaultClient.Transport = mockRoundTripper(func(req *http.Request) *http.Response {
		if strings.Contains(req.URL.Host, "viacep") {
			called++
			return mockHTTPResponse(http.StatusOK, viaCEPResp)
		}
		if strings.Contains(req.URL.Host, "weatherapi") {
			called++
			return mockHTTPResponse(http.StatusOK, weatherResp)
		}
		t.Fatalf("unexpected call to: %s", req.URL.String())
		return nil
	})
	defer func() { http.DefaultClient.Transport = nil }()

	req := httptest.NewRequest(http.MethodGet, "/weather/01001-000", nil)
	rr := httptest.NewRecorder()

	handleWeatherRequest(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var temp TemperatureResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &temp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if temp.TempC != 25.0 {
		t.Errorf("expected 25.0°C, got %f", temp.TempC)
	}
	if called != 2 {
		t.Errorf("expected 2 HTTP calls, got %d", called)
	}
}
