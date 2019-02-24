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

type RepaymentPlan struct {
	BorrowerPaymentAmount         string `json:"borrowerPaymentAmount"`
	Date                          string `json:"date"`
	InitialOutstandingPrincipal   string `json:"initialOutstandingPrincipal"`
	Interest                      string `json:"interest"`
	Principal                     string `json:"principal"`
	RemainingOutstandingPrincipal string `json:"remainingOutstandingPrincipal"`
}

type RepaymentPlanFloat struct {
	BorrowerPaymentAmount         float64
	Date                          string
	InitialOutstandingPrincipal   float64
	Interest                      float64
	Principal                     float64
	RemainingOutstandingPrincipal float64
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
	sd := time.Date(year, month, day, 00, 00, 00, 0, time.UTC)

	nominalRateCents, err := strconv.ParseFloat(p.NominalRate, 64)
	sendErrorResponse(rw, err, "Could not convert nominalRate from string to float64", 400)

	loanAmount, err := strconv.ParseFloat(p.LoanAmount, 64)
	sendErrorResponse(rw, err, "Could not convert loanAmount from string to float64", 400)

	if loanAmount > 0.0 && nominalRateCents > 0.0 && p.Duration > 0 {
		var rpf RepaymentPlanFloat
		rpfs := []RepaymentPlanFloat{}

		nominalRateDollar := nominalRateCents / 100
		nominalRatePerYear := nominalRateDollar / float64(numOfMonthsInYear)
		annuity := (loanAmount * nominalRatePerYear) / (1 - math.Pow((1+nominalRatePerYear), -24))

		/*var rp RepaymentPlan
		rps := []RepaymentPlan{}
		rp.InitialOutstandingPrincipal = loanAmount
		rp.Interest = (nominalRate * float64(daysInMonth) * loanAmount) / float64(daysInYear)
		rp.BorrowerPaymentAmount = annuity
		rp.Principal = rp.BorrowerPaymentAmount - rp.Interest
		rp.RemainingOutstandingPrincipal = rp.InitialOutstandingPrincipal - rp.Principal
		rps = append(rps, rp)*/

		rpf.InitialOutstandingPrincipal = math.Ceil(loanAmount*100) / 100
		rpf.Interest = math.Ceil(((nominalRateDollar*float64(daysInMonth)*rpf.InitialOutstandingPrincipal)/float64(daysInYear))*100) / 100
		rpf.BorrowerPaymentAmount = math.Ceil(annuity*100) / 100
		rpf.Principal = math.Ceil((rpf.BorrowerPaymentAmount-rpf.Interest)*100) / 100
		rpf.RemainingOutstandingPrincipal = math.Ceil((rpf.InitialOutstandingPrincipal-rpf.Principal)*100) / 100
		rpf.Date = sd.Format(time.RFC3339)
		rpfs = append(rpfs, rpf)

		for i := 1; i < p.Duration; i++ {

			prevDate := rpfs[i-1].Date
			y, m, d := prevDate.Date()
			if int(month) <= 12 {
				modifiedStartDate := sd.AddDate(0, 1, 0)
				rpf.Date = modifiedStartDate.Format(time.RFC3339)
			} else {

			}

			rpf.InitialOutstandingPrincipal = math.Ceil(rpfs[i-1].RemainingOutstandingPrincipal*100) / 100
			rpf.Interest = math.Ceil(((nominalRateDollar*float64(daysInMonth)*rpf.InitialOutstandingPrincipal)/float64(daysInYear))*100) / 100
			if i == p.Duration-1 { // TODO
				rpf.BorrowerPaymentAmount = math.Ceil((rpf.InitialOutstandingPrincipal+rpf.Interest)*100) / 100
				fmt.Println(p.Duration, rpf.BorrowerPaymentAmount)
			} else {
				rpf.BorrowerPaymentAmount = math.Ceil(rpfs[i-1].BorrowerPaymentAmount*100) / 100
			}
			if rpf.Interest*100 > rpf.InitialOutstandingPrincipal {
				rpf.Principal = math.Ceil((rpf.BorrowerPaymentAmount-rpf.InitialOutstandingPrincipal)*100) / 100
			} else {
				rpf.Principal = math.Ceil((rpf.BorrowerPaymentAmount-rpf.Interest)*100) / 100
			}
			rpf.RemainingOutstandingPrincipal = math.Ceil((rpf.InitialOutstandingPrincipal-rpf.Principal)*100) / 100
			rpfs = append(rpfs, rpf)

		}
		/*for i, val := range rpfs {
			fmt.Println("\n", i)
			fmt.Println("val: ", val)
			fmt.Println("BorrowerPaymentAmount: ", rpf.BorrowerPaymentAmount)
			fmt.Println("InitialOutstandingPrincipal: ", rpf.InitialOutstandingPrincipal)
			fmt.Println("Interest: ", rpf.Interest)
			fmt.Println("Principal: ", rpf.Principal)
			fmt.Println("RemainingOutstandingPrincipal: ", rpf.RemainingOutstandingPrincipal)
		}*/
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
