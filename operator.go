package main

import (

	"fmt"
	"os/exec"
	"os"
	"bytes"
	"net/http"
	"strconv"
	"strings"
	"time"
	"context"
)

var hysteresis time.Duration = 30000000000 // Time in nanoseconds. Default is 1m30s (Time_in_ns = time_in_min * 6000000000).


var netCardName string = os.Getenv("NET_CARD") // WiFi Network Card name. Could be retrieved by means of "iwconfig" Linux tool. 

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
	var context = ""
	//  Get the local kube config.
	fmt.Printf("Connecting to Kubernetes Context %v\n", context)
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	// Creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	for {
		quality, err := strconv.Atoi(strings.TrimSuffix(getNetQuality(), "\n"))
		if err != nil {
			fmt.Printf("Your internet connection may be down. Switching to local\n")
			if state == "local"{
				fmt.Println("Already using local instance\n")
			}else{

				if selectorPatcher(clientset,"local") == "Error" {
					fmt.Printf("Error while switching instance")
					continue
				} else {
					fmt.Printf("Switching executed successfully")
					state = "local"
					changeTime = time.Now()
				}
			}	
							
			time.Sleep(2000 * time.Millisecond)
			continue
		}
		
		strenght, err := strconv.Atoi(strings.TrimSuffix(getSigStrenght(), "\n"))
		if err != nil {
			fmt.Printf("Your internet connection may be down. Switching to local\n")
			if state == "local"{
				fmt.Println("Already using local instance\n")
			}else{

				if selectorPatcher(clientset,"local") == "Error" {
					fmt.Printf("Error while switching instance")
					continue
				} else {
					fmt.Printf("Switching executed successfully")
					state = "local"
					changeTime = time.Now()
				}
			}

			time.Sleep(2000 * time.Millisecond)
			continue
		}
		
		fmt.Printf("Quality: %d/100\nSignal: %d dB\n", quality, strenght)

		if quality <= 40 || strenght <= -60 {
			if state == "local" {
				fmt.Printf("Already using local instance\n")
			} else {
				fmt.Printf("Switching to local instance\n")

				currentTime := time.Now()
				if currentTime.Sub(changeTime) <= hysteresis {
					fmt.Printf("The last switching was too recent\n")
					continue
				}

				if selectorPatcher(clientset,"local") == "Error" {
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
				if selectorPatcher(clientset,"remote") == "Error" {
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

	cmd = fmt.Sprintf("iwconfig %s | awk '{if ($1==\"Link\"){split($2,A,\"/\");print A[1]}}' | sed 's/Quality=//g' | grep -x -E '[0-9]+'", netCardName)

	
	out, err := exec.Command("sh", "-c", cmd).Output()
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

	cmd = fmt.Sprintf("iwconfig %s | awk '{if ($3==\"Signal\"){split($4,A, \" \");print A[1]}}' | sed 's/level=//g' | grep -x -E '\\-[0-9]+'", netCardName)
	
	out, err := exec.Command("sh", "-c", cmd).Output()
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
	baseURL := "http://localhost:8001/api/v1/namespaces/yolo/services/yolo-service"
	//http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	if selector == "local" || selector == "remote" {
		ymlString := []byte("{\"spec\":{\"selector\":{\"name\":\"yolo-tiny-" + selector + "\"}}}")
		fmt.Println(string(ymlString))
		req, _ := http.NewRequest(http.MethodPatch, baseURL, bytes.NewBuffer(ymlString))
		req.Header.Set("Content-Type", "application/strategic-merge-patch+json")
		resp, err := http.DefaultClient.Do(req)
		fmt.Println(resp)
		if err != nil {
			fmt.Println("Error while doing the HTTP Request")
			fmt.Println(err)
			return "Error"
		}

		fmt.Println(err)
		fmt.Println("Switch executed successfully\n")
		return "Ok"
	}

	return "Error"


}


