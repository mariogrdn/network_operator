package main

import (
    "fmt"
	"os/exec"
	"time"
	"strconv"
	"strings"
)

func main(){

	var state string = "remote"
	fmt.Printf(remoteInstance())

	for {
		quality, err := strconv.Atoi(strings.TrimSuffix(getNetQuality(), "\n"))
		if err != nil{
			fmt.Printf("Error during parsing. Check your internet connection\n")
			time.Sleep(2000*time.Millisecond)
			continue
		}
		strenght, err := strconv.Atoi(strings.TrimSuffix(getSigStrenght(), "\n"))
		if err != nil{
			fmt.Printf("Error during parsing. Check your internet connection\n")
			time.Sleep(2000*time.Millisecond)
			continue
		}
		fmt.Printf("Quality: %d/100\nSignal: %d dB\n", quality, strenght)

		if(quality <= 40 || strenght <= -60){
			fmt.Printf("Switching to localInstance\n")
			if (state == "local"){
				fmt.Printf("Already using localInstance\n")
			}else{
				state = "local"
				fmt.Printf(localInstance())
			}
		}else{
			fmt.Printf("Switching to remoteInstance\n")
			if(state == "remote"){
				fmt.Printf("Already using remoteInstance\n")
			}else{
				state = "remote"
				fmt.Printf(remoteInstance())
			}
		}
		time.Sleep(1000 * time.Millisecond)
	}
	
}

func getNetQuality() string {
	cmd := "iwconfig | awk '{if ($1==\"Link\"){split($2,A,\"/\");print A[1]}}' | sed 's/Quality=//g' | grep -x -E '[0-9]+'"
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return fmt.Sprintf("Failed to execute command: %s", cmd)
	}
	return string(out)
}

func getSigStrenght() string {
	cmd := "iwconfig | awk '{if ($3==\"Signal\"){split($4,A, \" \");print A[1]}}' | sed 's/level=//g' | grep -x -E '\\-[0-9]+'"
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return fmt.Sprintf("Failed to execute command: %s", cmd)
	}
	return string(out)
}

func localInstance() string {
	return "Local instance\n"
}

func remoteInstance() string {
	return "Remote instance\n"
}
