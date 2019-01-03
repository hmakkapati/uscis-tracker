package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/user"
	"path"
	"strings"
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
	FormName string `json:"form_name"`
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
const defaultErrorMessage = "Unknown error occurred"

var httpClient = &http.Client{Timeout: 10 * time.Second}

func checkFatal(err error, message string) {
	if err != nil {
		if message == "" {
			message = defaultErrorMessage
		}

		log.Printf("FATAL: %s ... aborting!", message)
		log.Fatalln(err)
	}
}

func check(err error, message string) {
	if err != nil {
		if message == "" {
			message = defaultErrorMessage
		}

		log.Printf("ERROR: %s ... skipping!", message)
		log.Println(err)
	}
}

func getAllForms() ([]form, error) {
	resp, err := httpClient.Get(formsEndpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		var myForms formsResponse
		decoder := json.NewDecoder(resp.Body)
		err = decoder.Decode(&myForms)
		if err != nil {
			return nil, err
		}

		return myForms.Data.Forms.Forms, nil
	}

	return nil, fmt.Errorf("Response code: %d", resp.StatusCode)
}

func getAllFormOffices(form string) (formOfficesResponse, error) {
	var myFormOffices formOfficesResponse

	resp, err := httpClient.Get(fmt.Sprintf(formOfficesEndpoint, form))
	if err != nil {
		return myFormOffices, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		decoder := json.NewDecoder(resp.Body)
		err = decoder.Decode(&myFormOffices)
		if err != nil {
			return myFormOffices, err
		}

		return myFormOffices, nil
	}

	return myFormOffices, fmt.Errorf("Response code: %d", resp.StatusCode)
}

func getProcessingTime(formName, officeName string) (processingTimeResponse, error) {
	var processingTimeResult processingTimeResponse

	resp, err := httpClient.Get(fmt.Sprintf(processingTimesEndpoint, formName, officeName))
	if err != nil {
		return processingTimeResult, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		decoder := json.NewDecoder(resp.Body)
		err = decoder.Decode(&processingTimeResult)
		if err != nil {
			return processingTimeResult, err
		}

		return processingTimeResult, nil
	}

	return processingTimeResult, fmt.Errorf("Response code: %d", resp.StatusCode)
}

func getOutputFile() string {
	currUser, _ := user.Current()
	currTime := time.Now()
	return path.Join(currUser.HomeDir, fmt.Sprintf("Processing-Times_%s", currTime.Format("Jan-02-2006_15-04-05")))
}

func getFormOfficeKey(formName, officeDesc string) string {
	return strings.ToLower(
		strings.TrimSpace(formName) + "|" + strings.TrimSpace(officeDesc))
}

func readConfiguration(configFile string) (map[string]bool, error) {
	configuration := make(map[string]bool)

	file, err := os.Open(configFile)
	if err != nil {
		return configuration, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	fmt.Println("* Reading configuration ...")
	for scanner.Scan() {
		line := scanner.Text()
		tokens := strings.Split(line, ",")
		if len(tokens) != 2 {
			fmt.Printf("WARNING: Invalid configuration `%s` .... skipping\n",
				line)
			continue
		}
		configKey := getFormOfficeKey(tokens[0], tokens[1])
		configuration[configKey] = true
	}

	return configuration, nil
}

func main() {
	configFile := flag.String("config", "", "Configuration file")
	outputFile := flag.String("output", getOutputFile(), "Output file")
	flag.Parse()

	configuration, err := readConfiguration(*configFile)
	if err != nil {
		if *configFile != "" {
			log.Printf("FATAL: Unable to read configuration from: %s\n", *configFile)
			log.Fatalln(err)
		}
	}

	file, err := os.Create(*outputFile)
	if err != nil {
		log.Printf("FATAL: Unable to create outout file: %s\n", *outputFile)
		log.Fatalln(err)
	}
	defer file.Close()

	currTime := time.Now()
	file.WriteString(fmt.Sprintf("Report generated at: %s\n", currTime.Format("01/02/2006 15:04:05")))
	file.WriteString("Form\tField Office/Service Center\tProcessing time range\tForm type\tCase inquiry date\n")
	file.Sync()

	allForms, err := getAllForms()
	if err != nil {
		log.Println("FATAL: Unable to fetch forms information from USCIS")
		log.Fatalln(err)
	}

	fmt.Println("* Fetching processing time data for:")
	for _, formItem := range allForms {
		officesResult, err := getAllFormOffices(formItem.FormName)
		if err != nil {
			log.Printf("WARNING: Unable to get offices information for form: %s ... skipping", formItem.FormName)
			log.Println(err)
			continue
		}

		for _, officesItem := range officesResult.Data.FormOffices.Offices {

			formOfficeKey := getFormOfficeKey(formItem.FormName, officesItem.OfficeDescription)
			_, keyExists := configuration[formOfficeKey]
			if len(configuration) != 0 && !keyExists {
				continue
			}

			fmt.Printf("%-8s | %-35s ...... ", formItem.FormName, officesItem.OfficeDescription)
			resp, err := getProcessingTime(formItem.FormName, officesItem.OfficeCode)
			if err != nil {
				fmt.Printf("Error\n")
				log.Printf("WARNING: Unable to get processing time information for %s | %s\n", formItem.FormName, officesItem.OfficeDescription)
				log.Println(err)
				file.WriteString(fmt.Sprintf("%s\t%s\tERROR\tERROR\tERROR",
					formItem.FormName,
					officesItem.OfficeDescription))
				file.Sync()
				continue
			}

			for _, subType := range resp.Data.ProcessingTime.SubTypes {
				tRange := subType.Range
				file.WriteString(fmt.Sprintf(
					"%s\t%s\t%.1f %s to %.1f %s\t%s\t%s\n",
					formItem.FormName,
					officesItem.OfficeDescription,
					tRange[1].Value,
					tRange[1].Unit,
					tRange[0].Value,
					tRange[0].Unit,
					strings.TrimSpace(subType.SubTypeInfo),
					subType.ServiceRequestDate))
				file.Sync()
			}
			fmt.Printf("Done\n")
		}
	}
	fmt.Println("Processing finishsed successfully!")
	fmt.Printf("Data is saved to file: %s\n", *outputFile)
}
