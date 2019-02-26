# go-challenge
Task(1A-2B-3D.pdf) - Generate repayment plan for the given loan amount, nominal rate and duration(in months)

The request and response formats are explained below. Please read the comments in code for better understanding. The application exposes an endpoint called /generate-plan and listens on port 8080.

## Generate Plan
The repayment plan is generated as a json and sent as a response to the client.
Calculations are done using the formulae mentioned in the "calculation basics" section in the given task document.

## Request Format
I used Postman to send the request and receive response.

Request format is similar to "payload:" request json as mentioned in the task document. Except "duration" all the other json values are parsed as strings. 
StartDate is accepted as RFC3339 format.
Duration represents the number of months and is parsed as int.

Request type: POST
Request body: application/json

```
{
	"loanAmount": "5000",
	"nominalRate": "5.0",
	"duration": 24,
	"startDate": "2018-01-01T00:00:01Z"
}
```

## Response Format
Response is a json object which contains an array of json objects. Except "date" all the numbers are floating point values. Keeping the task document in consideration, the floating point values were rounded of to 2 decimal points.mi

Note: In the task document all the response values in the array object were strings but I am using string only for the date and not for other values to avoid multiple conversion statements.

```
{
    "RepaymentPlan": [
        {
            "borrowerPaymentAmount": 219.36,
            "date": "2018-01-01T00:00:00Z",
            "initialOutstandingPrincipal": 5000,
            "interest": 20.83,
            "principal": 198.53,
            "remainingOutstandingPrincipal": 4801.47
        },
        ...
        {
            "borrowerPaymentAmount": 219.28,
            "date": "2019-12-01T00:00:00Z",
            "initialOutstandingPrincipal": 218.37,
            "interest": 0.91,
            "principal": 218.37,
            "remainingOutstandingPrincipal": 0
        }
    ]
}
```
