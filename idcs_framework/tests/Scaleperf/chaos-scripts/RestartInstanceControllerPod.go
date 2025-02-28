package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"os/exec"
	"strings"
	"time"
)

func restart_instance_controller(total_execution_time int64, min_restart_interval int, max_restart_interval int) {
	var execution_start_time = time.Now().UnixMilli()
	fmt.Println("Execution start time is :", execution_start_time)
	for (time.Now().UnixMilli() - execution_start_time) < (total_execution_time * 60 * 60 * 1000) {
		// Command for restarging the instance operator command
		restart := exec.Command("kubectl", "scale", "deployment/idcs-instance-operator", "--replicas=0", "-n", "idcs-system", "--kubeconfig=<<config file goes here>>")
		restart.Run()

		// sleep for certain interval
		time.Sleep(time.Duration(rand.Intn(max_restart_interval-min_restart_interval)+min_restart_interval) * time.Second)

		// Get the pod details after restart
		var output bytes.Buffer
		get_pod_cmd := exec.Command("kubectl", "get", "po", "-l", "app.kubernetes.io/name=instance-operator", "-n", "idcs-system", "--kubeconfig=<<config file goes here>>")
		get_pod_cmd.Stdout = &output
		error := get_pod_cmd.Run()
		if error != nil {
			fmt.Println("Execution of get pod command is not successful: ", error)
			break
		}

		// Validate the pod is up and running or not
		pod_status := strings.Split(output.String(), "\n")
		if (strings.Contains(pod_status[1], "Running")) && (strings.Contains(pod_status[1], "3/3")) {
			fmt.Println(pod_status[1])
			continue
		} else {
			fmt.Println("Pod didn't come up after restart!!!")
			break
		}
	}
	fmt.Println("Execution end time is :", time.Now().UnixMilli())
}

func main() {
	var min_restart_interval, max_restart_interval int
	var total_execution_time int64
	fmt.Print("Enter the total_execution_time in hours: ")
	fmt.Scan(&total_execution_time)
	fmt.Print("Enter the instance controller - Min restart interval in mins: ")
	fmt.Scan(&min_restart_interval)
	fmt.Print("Enter the instance controller - Max restart interval in mins: ")
	fmt.Scan(&max_restart_interval)
	restart_instance_controller(total_execution_time, min_restart_interval*60, max_restart_interval*60)
}
