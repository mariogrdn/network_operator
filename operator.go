package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var hysteresis int64 = 90000000000 // Time in nanoseconds. Default is 1m30s (Time_in_ns = time_in_min * 6000000000).

				// Network card name. Could be retrieved by means of "iwconfig" Linux tool. 
var netCardName string = "" 	// An empty "netCardName" can be used in case the system has only one WiFi Network Card.
				// In case of multiple WiFi Network Cards, a name must be specified.


//**********************************************//
//					     	//
// Infinite cycle. Once every second, checks 	//
// the "netCardName" network card signal     	//
// status. By means of that data it decides  	//
// what service instance must be used.       	//
// Instance can be switched only once every  	//
// "hysteresis" nanoseconds.			//
//                                           	//
//**********************************************//

func main() { 
	
	changeTime := time.Time{}
	state := "remote"

	for {
		quality, err := strconv.Atoi(strings.TrimSuffix(getNetQuality(), "\n"))
		if err != nil {
			fmt.Printf("Error during parsing. Check your internet connection\n")
			time.Sleep(2000 * time.Millisecond)
			continue
		}
		strenght, err := strconv.Atoi(strings.TrimSuffix(getSigStrenght(), "\n"))
		if err != nil {
			fmt.Printf("Error during parsing. Check your internet connection\n")
			time.Sleep(2000 * time.Millisecond)
			continue
		}
		fmt.Printf("Quality: %d/100\nSignal: %d dB\n", quality, strenght)

		if quality <= 40 || strenght <= -60 {
			if state == "local" {
				fmt.Printf("Already using local instance\n")
			} else {
				fmt.Printf("Switching to local insatnce\n")

				currentTime := time.Now()
				if currentTime.Sub(changeTime) <= hysteresis {
					fmt.Printf("The last switching was too recent\n")
					continue
				}

				if selectorPatcher("local") == "Error" {
					fmt.Printf("Error while switching instance")
					continue
				} else {
					fmt.Printf("Switching executed successfully")
					state = "local"
					changeTime = time.Now()
				}
			}
		} else {
			if state == "remote" {
				fmt.Printf("Already using remote instance\n")
			} else {
				fmt.Printf("Switching to remote instance\n")
				currentTime := time.Now()
				if currentTime.Sub(changeTime) <= hysteresis {
					fmt.Printf("The last switching was too recent\n")
					continue
				}
				if selectorPatcher("remote") == "Error" {
					fmt.Printf("Error while switching instance")
					continue
				} else {
					fmt.Printf("Switching executed successfully")
					state = "remote"
					changeTime = time.Now()
				}
			}
		}
		time.Sleep(1000 * time.Millisecond)
	}

}

//**********************************************//
//						//
// Retrieves the "Link Quality" value.		//
// It returns either the data or an error.	//
//						//
//**********************************************//

func getNetQuality() string {
	var cmd string
	
	if (netCardName == ""){
		cmd = "iwconfig | awk '{if ($1==\"Link\"){split($2,A,\"/\");print A[1]}}' | sed 's/Quality=//g' | grep -x -E '[0-9]+'"
	}else{
		cmd = fmt.Sprintf("iwconfig %s | awk '{if ($1==\"Link\"){split($2,A,\"/\");print A[1]}}' | sed 's/Quality=//g' | grep -x -E '[0-9]+'", netCardName)
	}
	
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return fmt.Sprintf("Failed to execute command: %s", cmd)
	}
	return string(out)
}

//**********************************************//
//						//
// Retrieves the "Signal Strenght" value.	//
// It returns either the data or an error.	//
//						//
//**********************************************//
func getSigStrenght() string {
	var cmd string
	
	if (netCardName == ""){
		cmd = "iwconfig | awk '{if ($3==\"Signal\"){split($4,A, \" \");print A[1]}}' | sed 's/level=//g' | grep -x -E '\\-[0-9]+'"		
	}else{
		cmd = fmt.Sprintf("iwconfig %s | awk '{if ($3==\"Signal\"){split($4,A, \" \");print A[1]}}' | sed 's/level=//g' | grep -x -E '\\-[0-9]+'", netCardName)
	}
	
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return fmt.Sprintf("Failed to execute command: %s", cmd)
	}
	return string(out)
}

//**********************************************//
//						//
// Changes the current K8s service selector 	//
// to the one passed by the "selector" 		//
// parameter. It returns "Ok" if the request	//
// has been successful, "Error" if it has not	//
//						//
//**********************************************//

func selectorPatcher(selector string) string {
	baseURL := "http://127.0.0.1:8001/api/v1/namespaces/default/services/listener-service"
	if selector == "local" || selector == "remote" {
		ymlString := []byte("{\"spec\":{\"selector\":{\"app\":\"listener\",\"version\":\"" + selector + "\"}}}")
		fmt.Println(string(ymlString))
		req, _ := http.NewRequest(http.MethodPatch, baseURL, bytes.NewBuffer(ymlString))
		req.Header.Set("Content-Type", "application/strategic-merge-patch+json")
		_, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("Error while doing the HTTP Request")
			fmt.Println(err)
			return "Error"
		}
	}

	fmt.Println("Switch executed successfully")
	return "Ok"
}
