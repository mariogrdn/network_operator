package main

import (
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/homedir"
	"fmt"
	"os/exec"
	"os"
	"strconv"
	"strings"
	"time"
	"context"
	"flag"
	"path/filepath"
)

var hysteresis time.Duration = 20000000000 // Time in nanoseconds. Default is 1m30s (Time_in_ns = time_in_min * 6000000000).


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
	//  Get the local kube config.
	fmt.Println("Connecting to Kubernetes Context")
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
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
			fmt.Println("Your internet connection may be down. Switching to local")
			if state == "local"{
				fmt.Println("Already using local instance")
			}else{

				if selectorPatcher(clientset,"local") == "Error" {
					fmt.Println("Error while switching instance")
					continue
				} else {
					fmt.Println("Switching executed successfully")
					state = "local"
					changeTime = time.Now()
				}
			}	
							
			time.Sleep(2000 * time.Millisecond)
			continue
		}
		
		strenght, err := strconv.Atoi(strings.TrimSuffix(getSigStrenght(), "\n"))
		if err != nil {
			fmt.Println("Your internet connection may be down. Switching to local")
			if state == "local"{
				fmt.Println("Already using local instance")
			}else{

				if selectorPatcher(clientset,"local") == "Error" {
					fmt.Println("Error while switching instance")
					continue
				} else {
					fmt.Println("Switching executed successfully")
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
				fmt.Println("Already using local instance")
			} else {
				fmt.Println("Switching to local instance")

				currentTime := time.Now()
				if currentTime.Sub(changeTime) <= hysteresis {
					fmt.Println("The last switching was too recent")
					continue
				}

				if selectorPatcher(clientset,"local") == "Error" {
					fmt.Println("Error while switching instance")
					continue
				} else {
					fmt.Println("Switching executed successfully")
					state = "local"
					changeTime = time.Now()
				}
			}
		} else {
			if state == "remote" {
				fmt.Println("Already using remote instance")
			} else {
				fmt.Println("Switching to remote instance")
				currentTime := time.Now()
				if currentTime.Sub(changeTime) <= hysteresis {
					fmt.Println("The last switching was too recent")
					continue
				}
				if selectorPatcher(clientset,"remote") == "Error" {
					fmt.Println("Error while switching instance")
					continue
				} else {
					fmt.Println("Switching executed successfully")
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


func selectorPatcher(clientSet *kubernetes.Clientset, selector string) string {

	payloadBytes := []byte("{\"spec\":{\"selector\":{\"name\":\"yolo-tiny-" + selector + "\"}}}")
	_, err := clientSet.
		CoreV1().
		Services("yolo").
		Patch(context.TODO(), "yolo-service", types.StrategicMergePatchType, payloadBytes, metav1.PatchOptions{})
		
	if err != nil{
		fmt.Println("Error while switching")
		fmt.Println(err)
		return "Error"
	}else{
		fmt.Println("Switch executed correctly")
		return "Ok"
	}
}


