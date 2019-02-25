package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"time"
)

type Payload struct {
	LoanAmount  string `json:"loanAmount"`
	NominalRate string `json:"nominalRate"`
	Duration    int    `json:"duration"`
	StartDate   string `json:"startDate"`
}

type RepaymentPlanResponse struct {
	BorrowerPaymentAmount         string `json:"borrowerPaymentAmount"`
	Date                          string `json:"date"`
	InitialOutstandingPrincipal   string `json:"initialOutstandingPrincipal"`
	Interest                      string `json:"interest"`
	Principal                     string `json:"principal"`
	RemainingOutstandingPrincipal string `json:"remainingOutstandingPrincipal"`
}

type RepaymentPlan struct {
	borrowerPaymentAmount         float64
	date                          string
	initialOutstandingPrincipal   float64
	interest                      float64
	principal                     float64
	remainingOutstandingPrincipal float64
}

func sendErrorResponse(rw http.ResponseWriter, err error, msg string, code int) {
	if err != nil {
		http.Error(rw, msg, code)
		return
	}
}

func generatePaymentPlan(rw http.ResponseWriter, p *Payload) {
	daysInMonth := 30
	daysInYear := 360
	numOfMonthsInYear := 12

	startDate, err := time.Parse(time.RFC3339, p.StartDate)
	sendErrorResponse(rw, err, "Could not parse startDate to RFC3339 format", 400)
	year, month, day := startDate.Date()
	date := time.Date(year, month, day, 00, 00, 00, 0, time.UTC)

	nominalRateCents, err := strconv.ParseFloat(p.NominalRate, 64)
	sendErrorResponse(rw, err, "Could not convert nominalRate from string to float64", 400)

	loanAmount, err := strconv.ParseFloat(p.LoanAmount, 64)
	sendErrorResponse(rw, err, "Could not convert loanAmount from string to float64", 400)

	if loanAmount > 0.0 && nominalRateCents > 0.0 && p.Duration > 0 {
		var rp RepaymentPlan
		rps := []RepaymentPlan{}

		nominalRateDollar := nominalRateCents / 100
		nominalRatePerYear := nominalRateDollar / float64(numOfMonthsInYear)
		annuity := (loanAmount * nominalRatePerYear) / (1 - math.Pow((1+nominalRatePerYear), -24))

		rp.initialOutstandingPrincipal = math.Ceil(loanAmount*100) / 100
		rp.interest = math.Ceil(((nominalRateDollar*float64(daysInMonth)*rp.initialOutstandingPrincipal)/float64(daysInYear))*100) / 100
		rp.borrowerPaymentAmount = math.Ceil(annuity*100) / 100
		rp.principal = math.Ceil((rp.borrowerPaymentAmount-rp.interest)*100) / 100
		rp.remainingOutstandingPrincipal = math.Ceil((rp.initialOutstandingPrincipal-rp.principal)*100) / 100
		rp.date = date.Format(time.RFC3339)
		rps = append(rps, rp)

		for i := 1; i < p.Duration; i++ {
			date = date.AddDate(0, 1, 0)
			rp.date = date.Format(time.RFC3339)

			rp.initialOutstandingPrincipal = math.Ceil(rps[i-1].remainingOutstandingPrincipal*100) / 100
			rp.interest = math.Ceil(((nominalRateDollar*float64(daysInMonth)*rp.initialOutstandingPrincipal)/float64(daysInYear))*100) / 100
			if i == p.Duration-1 { // TODO
				rp.borrowerPaymentAmount = math.Ceil((rp.initialOutstandingPrincipal+rp.interest)*100) / 100
				fmt.Println(p.Duration, rp.borrowerPaymentAmount)
			} else {
				rp.borrowerPaymentAmount = math.Ceil(rps[i-1].borrowerPaymentAmount*100) / 100
			}
			if rp.interest*100 > rp.initialOutstandingPrincipal {
				rp.principal = math.Ceil((rp.borrowerPaymentAmount-rp.initialOutstandingPrincipal)*100) / 100
			} else {
				rp.principal = math.Ceil((rp.borrowerPaymentAmount-rp.interest)*100) / 100
			}
			rp.remainingOutstandingPrincipal = math.Ceil((rp.initialOutstandingPrincipal-rp.principal)*100) / 100
			rps = append(rps, rp)

		}
		for i, val := range rps {
			fmt.Println("\n", i)
			fmt.Println("val: ", val)
			/*fmt.Println("BorrowerPaymentAmount: ", rpf.BorrowerPaymentAmount)
			fmt.Println("InitialOutstandingPrincipal: ", rpf.InitialOutstandingPrincipal)
			fmt.Println("Interest: ", rpf.Interest)
			fmt.Println("Principal: ", rpf.Principal)
			fmt.Println("RemainingOutstandingPrincipal: ", rpf.RemainingOutstandingPrincipal)*/
		}
	} else { // TODO
		http.Error(rw, "Invalid request", 400)
		return
	}
}

func generatePlanHandler(rw http.ResponseWriter, req *http.Request) {
	// Read body
	request, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		http.Error(rw, err.Error(), 500)
		return
	}

	// Unmarshal
	payload := Payload{}
	err = json.Unmarshal(request, &payload)
	if err != nil {
		http.Error(rw, err.Error(), 500)
		return
	}

	fmt.Println("payload", payload)

	//respMsg := generatePaymentPlan(&payload)
	generatePaymentPlan(rw, &payload)

	/* Write Response
	output, err := json.Marshal(respMsg)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Write(output)*/
}

func main() {
	http.HandleFunc("/generate-plan", generatePlanHandler)
	address := ":8080"
	err := http.ListenAndServe(address, nil)
	if err != nil {
		panic(err)
	}
}
