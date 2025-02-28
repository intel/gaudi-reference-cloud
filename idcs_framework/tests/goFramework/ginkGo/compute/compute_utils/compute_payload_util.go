package compute_utils

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"goFramework/framework/service_api/compute/frisby"

	"github.com/tidwall/gjson"
)

func EnrichInstancePayload(rawpayload string, instance_name string, instance_type string, machine_image string, sshkey string, vnet_name string) string {
	var enriched_payload = rawpayload
	enriched_payload = strings.Replace(enriched_payload, "<<instance-name>>", instance_name, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<instance-type>>", instance_type, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<machine-image>>", machine_image, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<ssh-public-key>>", sshkey, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<vnet-name>>", vnet_name, 1)
	return enriched_payload
}

func EnrichSSHKeyPayload(rawpayload string, sshkey_name string, ssh_key string) string {
	enriched_payload := rawpayload
	enriched_payload = strings.Replace(enriched_payload, "<<ssh-key-name>>", sshkey_name, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<ssh-user-public-key>>", ssh_key, 1)
	fmt.Println("Enriched SSH Key Payload: ", enriched_payload)
	return enriched_payload
}

func EnrichVnetPayload(rawpayload string, vnet_name string) string {
	enriched_payload := rawpayload
	enriched_payload = strings.Replace(enriched_payload, "<<vnet-name>>", vnet_name, 1)
	return enriched_payload
}

func EnrichInventoryData(raw_data string, proxy_ip string, proxy_user string, machine_ip string, pvt_key_path string) string {
	enriched_data := raw_data
	enriched_data = strings.ReplaceAll(enriched_data, "<<proxy_machine_ip>>", proxy_ip)
	enriched_data = strings.ReplaceAll(enriched_data, "<<proxy_machine_user>>", proxy_user)
	enriched_data = strings.ReplaceAll(enriched_data, "<<machine_ip>>", machine_ip)
	enriched_data = strings.ReplaceAll(enriched_data, "<<authorised_key_path>>", pvt_key_path)
	return enriched_data
}

func GetRandomString() string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, 6+2)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[2 : 6+2]
}

func CheckInstanceState(instance_endpoint, token, instance_id_created, expectedState string, done chan struct{}) func() bool {
	startTime := time.Now()
	return func() bool {
		_, get_response_byid_body := frisby.GetInstanceById(instance_endpoint, token, instance_id_created)
		// Expect(get_response_byid_status).To(Equal(200), "assertion failed on response code")
		instancePhase := gjson.Get(get_response_byid_body, "status.phase").String()
		log.Printf("instancePhase: ", instancePhase)
		if instancePhase != expectedState {
			log.Printf("Instance is not in " + expectedState + " state")
			return false
		} else if instancePhase == "Failed" {
			log.Printf("Instance is in Failed state")
			close(done)
			return false
		} else {
			log.Printf("Instance is in " + expectedState + " state")
			elapsedTime := time.Since(startTime)
			log.Printf("Time took for instance to get to "+expectedState+" state: ", elapsedTime)
			close(done)
			return true
		}
	}
}
