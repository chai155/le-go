package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

const (
	localhost      = "localhost:8080"
	invalidRequest = "InvalidRequest"
	invalidURL     = "InvalidURL"
	randomURL      = "randomURL"
	forbiddenTest  = "403Test"
)

func Test_generatePlanHandler(t *testing.T) {
	payload := Payload{LoanAmount: "1000", NominalRate: "5.0", Duration: 4, StartDate: "2018-01-01T00:00:01Z"}
	requestByte, _ := json.Marshal(payload)
	expected := RepaymentPlanResponse{[]RepaymentPlan{{BorrowerPaymentAmount: 252.61, Date: "2018-01-01T00:00:00Z", InitialOutstandingPrincipal: 1000, Interest: 4.17, Principal: 248.44, RemainingOutstandingPrincipal: 751.56},
		{BorrowerPaymentAmount: 252.61, Date: "2018-02-01T00:00:00Z", InitialOutstandingPrincipal: 751.56, Interest: 3.13, Principal: 249.48, RemainingOutstandingPrincipal: 502.08},
		{BorrowerPaymentAmount: 252.61, Date: "2018-03-01T00:00:00Z", InitialOutstandingPrincipal: 502.08, Interest: 2.09, Principal: 250.52, RemainingOutstandingPrincipal: 251.56},
		{BorrowerPaymentAmount: 252.61, Date: "2018-04-01T00:00:00Z", InitialOutstandingPrincipal: 251.56, Interest: 1.05, Principal: 251.56, RemainingOutstandingPrincipal: 0}}}
	expectedByte, _ := json.Marshal(expected)

	req, err := http.NewRequest("POST", "/generate-plan", bytes.NewReader(requestByte))
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	rec := httptest.NewRecorder()
	generatePlanHandler(rec, req)
	res := rec.Result()
	defer res.Body.Close()

	// check the status code
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status OK; got %v", res.Status)
	}

	var rpResp RepaymentPlanResponse
	if err = json.NewDecoder(res.Body).Decode(&rpResp); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}

	resbody, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatalf("could not read the response body: %v", err)
	}

	if reflect.DeepEqual(resbody, expectedByte) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			resbody, expectedByte)
	}
}
