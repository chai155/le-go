package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	daysInMonth       = 30
	daysInYear        = 360
	numOfMonthsInYear = 12
	endpoint          = "/generate-plan"
)

type Payload struct {
	LoanAmount  string `json:"loanAmount"`
	NominalRate string `json:"nominalRate"`
	Duration    int    `json:"duration"`
	StartDate   string `json:"startDate"`
}

type RepaymentPlan struct {
	BorrowerPaymentAmount         float64 `json:"borrowerPaymentAmount"`
	Date                          string  `json:"date"`
	InitialOutstandingPrincipal   float64 `json:"initialOutstandingPrincipal"`
	Interest                      float64 `json:"interest"`
	Principal                     float64 `json:"principal"`
	RemainingOutstandingPrincipal float64 `json:"remainingOutstandingPrincipal"`
}

type RepaymentPlanResponse struct {
	RepaymentPlan []RepaymentPlan
}

// round floating point values to 2 decimal points
func RoundFloat(x float64) float64 {
	f, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", x), 64)
	return f
}

// validate the request body
func (p *Payload) validateRequest() (float64, float64, time.Time, url.Values) {
	errors := url.Values{}

	if p.StartDate == "" {
		errors.Add("StartDate", "Required Field Missing")
	}
	if p.LoanAmount == "" {
		errors.Add("LoanAmount", "Required Field Missing")
	}
	if p.Duration <= 0 {
		errors.Add("Duration", "Required Field Missing")
	}
	if p.NominalRate == "" {
		errors.Add("NominalRate", "Required Field Missing")
	}

	if len(errors) > 0 {
		return 0.0, 0.0, time.Time{}, errors
	}

	startDate, err := time.Parse(time.RFC3339, p.StartDate)
	if err != nil {
		errors.Add("StartDate", "Could not parse startDate to RFC3339 format")
	}
	nominalRateCents, err := strconv.ParseFloat(p.NominalRate, 64)
	if err != nil {
		errors.Add("NominalRate", "Could not convert nominalRate from string to float64")
	}
	loanAmount, err := strconv.ParseFloat(p.LoanAmount, 64)
	if err != nil {
		errors.Add("LoanAmount", "Could not convert loanAmount from string to float64")
	}
	if loanAmount <= 0.0 && nominalRateCents < 0.0 && p.Duration <= 0 {
		errors.Add("LoanAmount, NominalRateCents, Duration", "Requests are negative numbers")
	}
	return loanAmount, nominalRateCents, startDate, errors
}

// generate a payment plan
func generatePaymentPlan(loanAmount, nominalRateCents float64, startDate time.Time, duration int) RepaymentPlanResponse {
	var rp RepaymentPlan
	var rps RepaymentPlanResponse

	year, month, day := startDate.Date()
	date := time.Date(year, month, day, 00, 00, 00, 0, time.UTC)
	nominalRateDollar := nominalRateCents / 100

	// use nominal rate per year to calculate annuity
	nominalRatePerYear := nominalRateDollar / float64(numOfMonthsInYear)
	annuity := (loanAmount * nominalRatePerYear) / (1 - math.Pow((1+nominalRatePerYear), -float64(duration)))

	// calculate the first month repayment plan
	rp.InitialOutstandingPrincipal = RoundFloat(loanAmount)
	rp.Interest = RoundFloat((nominalRateDollar * float64(daysInMonth) * rp.InitialOutstandingPrincipal) / float64(daysInYear))
	rp.BorrowerPaymentAmount = RoundFloat(annuity)
	rp.Principal = RoundFloat(rp.BorrowerPaymentAmount - rp.Interest)
	rp.RemainingOutstandingPrincipal = RoundFloat(rp.InitialOutstandingPrincipal - rp.Principal)
	rp.Date = date.Format(time.RFC3339)
	rps.RepaymentPlan = append(rps.RepaymentPlan, rp)

	// use the first month repayment plan to calculate the rest of the repayment plan
	for i := 1; i < duration; i++ {

		// increment the months and years
		date = date.AddDate(0, 1, 0)
		rp.Date = date.Format(time.RFC3339)

		rp.InitialOutstandingPrincipal = RoundFloat(rps.RepaymentPlan[i-1].RemainingOutstandingPrincipal)
		rp.Interest = RoundFloat((nominalRateDollar * float64(daysInMonth) * rp.InitialOutstandingPrincipal) / float64(daysInYear))

		// last month borrowerPaymentAmount
		if i == duration-1 { // TODO
			rp.BorrowerPaymentAmount = RoundFloat(rp.InitialOutstandingPrincipal + rp.Interest)
		} else {
			rp.BorrowerPaymentAmount = RoundFloat(rps.RepaymentPlan[i-1].BorrowerPaymentAmount)
		}

		if rp.Interest > rp.InitialOutstandingPrincipal {
			rp.Principal = RoundFloat(rp.BorrowerPaymentAmount - rp.InitialOutstandingPrincipal)
		} else {
			rp.Principal = RoundFloat(rp.BorrowerPaymentAmount - rp.Interest)
		}

		rp.RemainingOutstandingPrincipal = RoundFloat(rp.InitialOutstandingPrincipal - rp.Principal)
		rps.RepaymentPlan = append(rps.RepaymentPlan, rp)
	}

	return rps
}

func generatePlanHandler(rw http.ResponseWriter, req *http.Request) {
	// unmarshal the request body
	var payload Payload
	rw.Header().Set("content-type", "application/json")
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		http.Error(rw, "Could not unmarshal the request to JSON", http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	// validate request and generate the repayment plan
	loanAmount, nominalRateCents, startDate, err := payload.validateRequest()
	if len(err) > 0 {
		log.Println(err)
		err := map[string]interface{}{"Validation Errors": err}
		rw.WriteHeader(http.StatusBadRequest)
		if encError := json.NewEncoder(rw).Encode(err); encError != nil {
			http.Error(rw, "An error occurred!", http.StatusInternalServerError)
		}
		return
	}
	respMsg := generatePaymentPlan(loanAmount, nominalRateCents, startDate, payload.Duration)

	// Write Response
	if err := json.NewEncoder(rw).Encode(respMsg); err != nil {
		http.Error(rw, "Could not marshal the response json", http.StatusInternalServerError)
		return
	}
}

func main() {
	listenAddr := flag.String("http.addr", ":8080", "http listen address")
	flag.Parse()
	http.HandleFunc(endpoint, generatePlanHandler)
	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}
