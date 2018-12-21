package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Structs representing response of formsEndpoint
type formsResponse struct {
	Data struct {
		Forms struct {
			Forms []form
		}
	}
}

type form struct {
	FormDescriptionEn string `json:"form_description_en"`
	FormName          string `json:"form_name"`
}

// Structs representing response of formOfficesEndpoint
type formOfficesResponse struct {
	Data struct {
		FormOffices struct {
			Offices []struct {
				OfficeCode        string `json:"office_code"`
				OfficeDescription string `json:"office_description"`
			}
		} `json:"form_offices"`
	}
}

// Structs representing response of processingTimesEndpoint
type processingTimeResponse struct {
	Data struct {
		ProcessingTime struct {
			Range    []processingTimeUnit
			SubTypes []struct {
				FormType           string `json:"form_type"`
				Range              []processingTimeUnit
				ServiceRequestDate string `json:"service_request_date"`
				SubTypeInfo        string `json:"subtype_info_en"`
			}
		} `json:"processing_time"`
	}
}

type processingTimeUnit struct {
	Unit  string
	Value float32
}

const formsEndpoint = "https://egov.uscis.gov/processing-times/api/forms"
const formOfficesEndpoint = "https://egov.uscis.gov/processing-times/api/formoffices/%s"
const processingTimesEndpoint = "https://egov.uscis.gov/processing-times/api/processingtime/%s/%s"

var httpClient = &http.Client{Timeout: 10 * time.Second}

func GetAllForms() []form {
	resp, err := httpClient.Get(formsEndpoint)
	check(err)
	defer resp.Body.Close()

	var myForms formsResponse
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&myForms)
	check(err)

	return myForms.Data.Forms.Forms
}

func GetAllFormOffices(form string) formOfficesResponse {
	resp, err := httpClient.Get(fmt.Sprintf(formOfficesEndpoint, form))
	check(err)
	defer resp.Body.Close()

	var myFormOffices formOfficesResponse
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&myFormOffices)
	check(err)

	return myFormOffices
}

func GetProcessingTime(formName, officeName string) processingTimeResponse {
	resp, err := httpClient.Get(fmt.Sprintf(processingTimesEndpoint, formName, officeName))
	check(err)
	defer resp.Body.Close()

	var processingTimeResult processingTimeResponse
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&processingTimeResult)
	check(err)

	return processingTimeResult
}

func check(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func main() {
	allForms := GetAllForms()
	for _, formItem := range allForms {
		fmt.Printf("Processing form: %s\n", formItem.FormName)
		officesResult := GetAllFormOffices(formItem.FormName)
		for _, officesItem := range officesResult.Data.FormOffices.Offices {
			fmt.Printf("%s | %s\n", officesItem.OfficeCode, officesItem.OfficeDescription)
			resp := GetProcessingTime(formItem.FormName, officesItem.OfficeCode)
			for _, subType := range resp.Data.ProcessingTime.SubTypes {
				tRange := subType.Range
				fmt.Printf("\tResponse time: %.1f %s to %.1f %s\n",
					tRange[1].Value,
					tRange[1].Unit,
					tRange[0].Value,
					tRange[0].Unit)
				fmt.Printf("\tSubTypeInfo: %s\n", subType.SubTypeInfo)
				fmt.Printf("\tFormType: %s\n", subType.FormType)
				fmt.Printf("\tServiceDate: %s\n", subType.ServiceRequestDate)
				fmt.Println("-----------------------------------------------")

			}
		}
		fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	}
	fmt.Println("Done!")
}
