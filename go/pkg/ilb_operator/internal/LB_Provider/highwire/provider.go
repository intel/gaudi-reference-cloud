// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package highwire

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	ilbv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/ilb_operator/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"golang.org/x/exp/slices"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// provider specific end points
const (
	GET_TOKEN       string = "login"
	POOL            string = "ltm/pools"
	POOLS           string = "ltm/pools/"
	VIRTUAL_SERVER  string = "ltm/virtualServers"
	VIRTUAL_SERVERS string = "ltm/virtualServers/"
	MEMBERS         string = "/members"
	API_TOKEN       string = "?apiToken="
)

type stdMessage string

const (
	//Failures
	MarshallingFailed           stdMessage = "failed to marshal the data"
	ReadResponseFailed          stdMessage = "failed to read the response from http request"
	UnmarshallingFailed         stdMessage = "failed to unmarshal the data"
	PoolCreationFailed          stdMessage = "failed to create the pool"
	VirtualServerFailed         stdMessage = "failed to create the virtual server"
	VirtualServerToPoolFailed   stdMessage = "failed to link Virtual Server to desired Pool"
	PoolDeletionFailed          stdMessage = "failed to delete the desired Pool"
	VirtualServerDeletionFailed stdMessage = "failed to delete the desired virtual server"
	PoolUpdateFailed            stdMessage = "failed to update the pool"
	TokenRequestFailed          stdMessage = "failed to get the token"
	//success
	PoolCreationSucceeded          stdMessage = "successfully created the desired pool"
	VirtualServerCreationSucceeded stdMessage = "successfully created the desired virtual server"
	VirtualServerToPoolSucceeded   stdMessage = "successfully linked virtual server to the pool"
	PoolUpdateSucceeded            stdMessage = "successfully updates the desired pool"
)

// provider specific schemas to marshal payloads
type member struct {
	Name            string `json:"name,omitempty"`
	IP              string `json:"ip,omitempty"`
	Port            string `json:"port,omitempty"`
	ConnectionLimit int    `json:"connectionLimit,omitempty"`
	PriorityGroup   int    `json:"priorityGroup,omitempty"`
	Ratio           int    `json:"ratio,omitempty"`
	AdminState      string `json:"adminState,omitempty"`
	MonitorStatus   string `json:"monitorStatus,omitempty"`
}

type pool struct {
	ID                int      `json:"id,omitempty"`
	Environment       int      `json:"environment,omitempty"`
	UserGroup         int      `json:"userGroup,omitempty"`
	Name              string   `json:"name,omitempty"`
	Description       string   `json:"description,omitempty"`
	LoadBalancingMode string   `json:"loadBalancingMode,omitempty"`
	Monitor           string   `json:"monitor,omitempty"`
	MinActiveMembers  int      `json:"minActiveMembers,omitempty"`
	Members           []member `json:"members,omitempty"`
}

// schema used to marshal and unmarshall input/output with external ILB provider [ In this case F5/Highwire]
type virtualServer struct {
	ID          int    `json:"id,omitempty"`
	IP          string `json:"ip,omitempty"`
	Environment int    `json:"environment,omitempty"`
	UserGroup   int    `json:"userGroup,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	IPType      string `json:"ipType,omitempty"`
	Port        int    `json:"port,omitempty"`
	Pool        int    `json:"pool,omitempty"`
	Persist     string `json:"persist,omitempty"`
	IPProtocol  string `json:"ipProtocol,omitempty"`
}

type virtualServerList struct {
	VirtualServers []virtualServer `json:"virtualServers"`
	Code           int             `json:"code"`
}

type poolList struct {
	Pools []pool `json:"pools"`
	Code  int    `json:"code"`
}

// provider specific configurations
type HighWireProvider struct {
	BaseURL         string
	UserName        string
	Domain          string
	Secret          string
	token           string
	tokenExpiry     time.Time
	mu              sync.Mutex
	HighwireTimeout time.Duration
}

type OperatorMessage struct {
	ErrorCode int32  `json:"errorCode"`
	Message   string `json:"message"`
}

func NewHighWireProvider(bu string, d string, u string, s string, highwireTimeout time.Duration) (*HighWireProvider, error) {
	return &HighWireProvider{
		BaseURL:         bu,
		Domain:          d,
		UserName:        u,
		Secret:          s,
		HighwireTimeout: highwireTimeout,
	}, nil
}

func (hw *HighWireProvider) GetStatus(ilb *ilbv1alpha1.Ilb) error {
	ilb.Status.Name = ilb.Name
	ilb.Status.State = ilbv1alpha1.PENDING
	ilb.Status.Message = operatorMessageString(0, "Provisioning load balancer")

	if err := hw.getToken(); err != nil {
		return err
	}

	// Get virtual server using its name.
	url := hw.BaseURL + VIRTUAL_SERVER + API_TOKEN + hw.token + "&environment=" + strconv.Itoa(ilb.Spec.VIP.Environment)
	resp, err := make2HTTPRequest("GET", url, []byte{}, hw.HighwireTimeout)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		msg, reason := processStatusCode(resp.StatusCode, body)
		return fmt.Errorf("get virtual server status code: %d with message: %s", resp.StatusCode, msg+reason)
	}

	var virtualServerList virtualServerList
	err = json.Unmarshal(body, &virtualServerList)
	if err != nil {
		return err
	}

	var virtualServerPoolID int
	var virtualServerFound bool
	for _, virtualServer := range virtualServerList.VirtualServers {
		virtualServerName := strings.Split(virtualServer.Name, "/")
		if virtualServerName[len(virtualServerName)-1] == "lbauto-"+ilb.Spec.VIP.Name {
			virtualServerFound = true
			ilb.Status.Vip = virtualServer.IP
			ilb.Status.VipID = virtualServer.ID
			ilb.Status.Conditions.VIPCreated = true
			virtualServerPoolID = virtualServer.Pool

			break
		}
	}

	// In case there is data inconsistency between the controller in highwire - ILB was set to ready on CR but missing in highwire (which is not expected)
	// we need to set the state to ERROR and update the message.
	if !virtualServerFound && ilb.Status.Conditions.VIPCreated {
		ilb.Status.State = ilbv1alpha1.ERROR
		ilb.Status.Message = operatorMessageString(404, "Virtual server was created but missing in provider")
		ilb.Status.Vip = ""
		ilb.Status.VipID = 0
		virtualServerPoolID = 0

		return nil
	}

	// Get pool using its name.
	if len(ilb.Spec.Pool.Name) != 0 {
		url = hw.BaseURL + POOL + API_TOKEN + hw.token + "&environment=" + strconv.Itoa(ilb.Spec.VIP.Environment)
		resp, err = make2HTTPRequest("GET", url, []byte{}, hw.HighwireTimeout)
		if err != nil {
			return err
		}

		defer resp.Body.Close()
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK {
			msg, reason := processStatusCode(resp.StatusCode, body)
			return fmt.Errorf("get pool status code: %d with message: %s", resp.StatusCode, msg+reason)
		}

		var poolList poolList
		err = json.Unmarshal(body, &poolList)
		if err != nil {
			return err
		}

		var poolFound bool
		for _, pool := range poolList.Pools {
			poolName := strings.Split(pool.Name, "/")
			if poolName[len(poolName)-1] == "lbauto-"+ilb.Spec.Pool.Name {
				poolFound = true
				ilb.Status.PoolID = pool.ID
				ilb.Status.Conditions.PoolCreated = true
				break
			}
		}

		// In case there is data inconsistency between the controller and highwire - pool was set to created on CR but missing in highwire (which is not expected)
		// we need to set the state to ERROR and update the message.
		if !poolFound && ilb.Status.Conditions.PoolCreated {
			ilb.Status.State = ilbv1alpha1.ERROR
			ilb.Status.Message = operatorMessageString(404, "Pool was created but missing in provider")
			ilb.Status.PoolID = 0
			ilb.Status.Conditions.PoolCreated = false
		}
	}

	// Compare virtual server pool ID and pool ID to ensure
	// right pool is linked.
	if ilb.Status.Conditions.PoolCreated && virtualServerPoolID == ilb.Status.PoolID {
		ilb.Status.Conditions.VIPPoolLinked = true
	}

	if ilb.Status.Conditions.PoolCreated && ilb.Status.Conditions.VIPCreated && ilb.Status.Conditions.VIPPoolLinked {
		ilb.Status.State = ilbv1alpha1.READY
		ilb.Status.Message = operatorMessageString(0, "Load balancer ready")
	}

	return nil
}

func (hw *HighWireProvider) CreatePool(v *ilbv1alpha1.Ilb) error {
	if err := hw.getToken(); err != nil {
		return err
	}

	url := hw.BaseURL + POOL + API_TOKEN + hw.token

	var members []member
	//Get the members collection ready.
	//Future Todo : Validate IP address against given set of subnets (attached as annotation or spec) and make sure it is in the range
	//This will avoid adding ANY IP address to the pool.
	for _, val := range v.Spec.Pool.Members {
		//var m member
		var m member
		m.IP = val.IP
		m.Port = strconv.Itoa(v.Spec.Pool.Port)
		m.AdminState = val.AdminState
		m.ConnectionLimit = val.ConnectionLimit
		//m.MonitorStatus = val.MonitorStatus
		m.PriorityGroup = val.PriorityGroup
		m.Ratio = val.Ratio
		members = append(members, m)
	}

	//get the new pool payload ready
	newPool := pool{
		Environment:       v.Spec.Pool.Environment,       //11,
		UserGroup:         v.Spec.Pool.UserGroup,         //1545,
		Name:              v.Spec.Pool.Name,              //"iks-controller-pool-443",
		Description:       v.Spec.Pool.Description,       //"This is created by a controller",
		LoadBalancingMode: v.Spec.Pool.LoadBalancingMode, //"least-connections-member",
		Monitor:           v.Spec.Pool.Monitor,           //"i_tcp",
		MinActiveMembers:  v.Spec.Pool.MinActiveMembers,  //1,
		Members:           members,
	}

	log.Log.Info("createPool", logkeys.NewPool, newPool)

	//prepare the request
	requestParam, err := json.Marshal(newPool)
	if err != nil {
		log.Log.Error(err, string(MarshallingFailed))
		return err
	}

	//create the pool now
	resp, err := make2HTTPRequest("POST", url, requestParam, hw.HighwireTimeout)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Log.Error(err, string(ReadResponseFailed))
		}
		log.Log.Info("CreatePool Response Body", logkeys.NewPool, newPool, logkeys.Response, string(body))
		msg, reasons := processStatusCode(resp.StatusCode, body)
		v.Status.Message = operatorMessageString(int32(resp.StatusCode), msg+reasons) //get the reasons field in the status struct and refactor this
		log.Log.Error(err, string(PoolCreationFailed), logkeys.ResponseStatusCode, resp.StatusCode, logkeys.Issues, msg+reasons)
		return fmt.Errorf("failed creating pool, response code: %d, message: %s", resp.StatusCode, msg+reasons)
	}

	return nil
}

func (hw *HighWireProvider) CreateVirtualServer(v *ilbv1alpha1.Ilb) error {
	if err := hw.getToken(); err != nil {
		return err
	}

	url := hw.BaseURL + VIRTUAL_SERVER + API_TOKEN + hw.token

	newVS := virtualServer{
		Environment: v.Spec.VIP.Environment, //11,
		UserGroup:   v.Spec.VIP.UserGroup,   //1545,
		Name:        v.Spec.VIP.Name,        //"iks-controller-pool-443",
		Description: v.Spec.VIP.Description, //"This is created using controller",
		IPType:      v.Spec.VIP.IPType,      //"private",
		Port:        v.Spec.VIP.Port,
		Pool:        v.Status.PoolID,
		Persist:     v.Spec.VIP.Persist,    //"i_client_ip_5min",
		IPProtocol:  v.Spec.VIP.IPProtocol, //"tcp",
	}

	log.Log.Info("Virtual Server Request", logkeys.VirtualServer, newVS)

	requestParam, err := json.Marshal(newVS)
	if err != nil {
		log.Log.Error(err, string(MarshallingFailed))
		return err //convert this into return code constants
	}

	//Get the Pool members
	resp, err := make2HTTPRequest("POST", url, requestParam, hw.HighwireTimeout)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Log.Error(err, string(ReadResponseFailed))
		}
		log.Log.Info("CreateVirtualServer Response Body", logkeys.VirtualServer, newVS, logkeys.Response, string(body))
		msg, reasons := processStatusCode(resp.StatusCode, body)
		v.Status.Message = operatorMessageString(int32(resp.StatusCode), msg+reasons) //get the reasons field in the status struct and refactor this
		log.Log.Error(err, string(VirtualServerFailed), logkeys.ResponseStatusCode, resp.StatusCode, logkeys.Issues, msg+reasons)
		return fmt.Errorf("failed creating virtual server, response code: %d, message: %s", resp.StatusCode, msg+reasons)
	}

	return nil
}

func (hw *HighWireProvider) LinkVSToPool(v *ilbv1alpha1.Ilb) error {
	//get the token
	if err := hw.getToken(); err != nil {
		return err
	}
	//url := "https://internal-placeholder.com/v1/ltm/virtualServers/" + strconv.Itoa(v.Status.VipID) + "?apiToken=" + token
	url := hw.BaseURL + VIRTUAL_SERVERS + strconv.Itoa(v.Status.VipID) + API_TOKEN + hw.token

	vs := virtualServer{
		Pool: v.Status.PoolID,
	}

	log.Log.Info("Link Virtual Server To Pool", logkeys.VipId, v.Status.VipID, logkeys.PoolId, v.Status.PoolID)
	//fmt.Printf("The virtual server to pool link request is :VIP ID - %v, Pool ID - %v", v.Status.VipID, v.Status.PoolID)

	requestParam, err := json.Marshal(vs)

	if err != nil {
		log.Log.Error(err, string(MarshallingFailed))
		return err
	}

	resp, err := make2HTTPRequest("PUT", url, requestParam, hw.HighwireTimeout)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Log.Error(err, string(ReadResponseFailed))
		}
		log.Log.Info("LinkVSToPool Response Body", logkeys.VipId, v.Status.VipID, logkeys.PoolId, v.Status.PoolID, logkeys.Response, string(body))
		msg, reasons := processStatusCode(resp.StatusCode, body)                      //passing the empty byte array as we do not read body here. test for sanity here
		v.Status.Message = operatorMessageString(int32(resp.StatusCode), msg+reasons) //get the reasons field in the status struct and refactor this
		log.Log.Error(err, string(VirtualServerFailed), logkeys.ResponseStatusCode, resp.StatusCode, logkeys.Issues, msg+reasons)
		return fmt.Errorf("failed linking pool to virtual server, response code: %d, message: %s", resp.StatusCode, msg+reasons)
	}

	return nil
}

func (hw *HighWireProvider) ProcessFinalizers(v *ilbv1alpha1.Ilb) error {
	//delete the Pool, associated VIP
	//vipURL := "https://internal-placeholder.com/v1/ltm/virtualServers/" + strconv.Itoa(v.Status.VipID) + "?apiToken=" + token
	//poolURL := "https://internal-placeholder.com/v1/ltm/pools/" + strconv.Itoa(v.Status.PoolID) + "?apiToken=" + token

	//get the token
	if err := hw.getToken(); err != nil {
		return err
	}
	vipURL := hw.BaseURL + VIRTUAL_SERVERS + strconv.Itoa(v.Status.VipID) + API_TOKEN + hw.token
	poolURL := hw.BaseURL + POOLS + strconv.Itoa(v.Status.PoolID) + API_TOKEN + hw.token

	var req []byte
	var vipRemoved, poolRemoved = false, false

	log.Log.Info("Processing finalizers..")

	if v.Status.State == ilbv1alpha1.TERMINATED {
		v.Status.Message = operatorMessageString(0, "Deleting load balancer")
		return nil // we already have taken the action. Return no errors
	}

	//Check if VIP is even created. We may be here when things are false, In that case, we dont need to clean up and return true
	if v.Status.VipID != 0 {
		log.Log.Info("Delete vip URL", logkeys.VipURL, vipURL)

		resp, err := make2HTTPRequest("DELETE", vipURL, req, hw.HighwireTimeout)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// trying to delete VIP regardless to previous condition, if not found - it is already deleted
		if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusNotFound {
			log.Log.Info("Successfully deleted the VIP", logkeys.VipName, v.Status.Vip, logkeys.VipId, v.Status.VipID)
			vipRemoved = true
		} else {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Log.Error(err, string(ReadResponseFailed))
			}
			log.Log.Info("Delete vip URL Failed", logkeys.VipName, v.Status.Vip, logkeys.Response, string(body))
			log.Log.Error(err, string(VirtualServerDeletionFailed), logkeys.ResponseStatusCode, resp.StatusCode)
		}
	} else {
		//assume it is already removed -- NOT created is removed
		vipRemoved = true
	}

	if v.Status.PoolID != 0 {
		log.Log.Info("Delete Pool URL", logkeys.PoolURL, poolURL)

		resp, err := make2HTTPRequest("DELETE", poolURL, req, hw.HighwireTimeout)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// trying to delete Pool regardless to previous condition, if not found - it is already deleted
		if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusNotFound {
			log.Log.Info("Successfully deleted the Pool", logkeys.PoolId, v.Status.PoolID)
			poolRemoved = true
		} else {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Log.Error(err, string(ReadResponseFailed))
			}
			log.Log.Info("Delete Pool URL Failed", logkeys.PoolId, v.Status.PoolID, logkeys.Response, string(body))
			log.Log.Error(err, string(PoolDeletionFailed), logkeys.ResponseStatusCode, resp.StatusCode)
		}
	} else {
		//assume it is already removed -- NOT created is removed
		poolRemoved = true
	}

	if !(vipRemoved && poolRemoved) {
		return fmt.Errorf(string(VirtualServerDeletionFailed) + "and/or" + string(PoolDeletionFailed))
	}

	log.Log.Info("Processed finalizers..")
	return nil // successful
}

func (hw *HighWireProvider) ObserveCurrentAndReconcile(v *ilbv1alpha1.Ilb) error {
	//get the token
	if err := hw.getToken(); err != nil {
		return err
	}

	pid := strconv.Itoa(v.Status.PoolID)
	//url := "https://internal-placeholder.com/v1/ltm/pools/" + pid + "/members?apiToken=" + token
	url := hw.BaseURL + POOLS + pid + MEMBERS + API_TOKEN + hw.token
	var req []byte

	//assume no mismatch with current members and desired members
	mismatch := false

	//Get the current pool members
	resp, err := make2HTTPRequest("GET", url, req, hw.HighwireTimeout)
	if err != nil {
		return err
	}

	//defer body close. Very important.
	defer resp.Body.Close()

	var p pool
	//var p pool
	body, err := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusOK {

		if err != nil {
			log.Log.Error(err, string(ReadResponseFailed))
			return err
		}
		err = json.Unmarshal(body, &p)
		if err != nil {
			log.Log.Error(err, string(UnmarshallingFailed))
			return err
		}
	} else {
		//failed to read the response body
		log.Log.Info("ObserveCurrentAndReconcile Response Body", logkeys.ResponseStatusCode, string(body))
		msg, reasons := processStatusCode(resp.StatusCode, body)
		v.Status.Message = operatorMessageString(int32(resp.StatusCode), msg+reasons) //get the reasons field in the status struct and refactor this
		log.Log.Error(err, string(VirtualServerFailed), logkeys.ResponseStatusCode, resp.StatusCode, logkeys.Issues, msg+reasons)
		return err
	}

	//get the desired pool members
	m := v.Spec.Pool.Members

	var desiredMembers []string
	for _, val := range m {
		desiredMembers = append(desiredMembers, val.IP)
	}

	log.Log.Info("Desired Members", logkeys.Members, desiredMembers)

	var currentMembers []string
	for _, member := range p.Members {
		currentMembers = append(currentMembers, member.IP)
	}

	log.Log.Info("Current Members", logkeys.Members, currentMembers)

	//handle case where pool is empty (valid condition)
	if len(currentMembers) == 0 && len(desiredMembers) != 0 {
		mismatch = true
	} else if len(currentMembers) != 0 && len(desiredMembers) == 0 {
		mismatch = true
	} else if len(currentMembers) != len(desiredMembers) {
		mismatch = true
	} else {
		//check if there is any mismatch Current Vs Desired = unDesired
		for _, ip := range desiredMembers {
			if slices.Index(currentMembers, ip) < 0 { //mismatch
				log.Log.Info("Mismatch Found", logkeys.Member, ip)
				//we break on first mismatch and replace all the elements instead of selectively updating for now
				mismatch = true
				break //get out of for loop on the first mismatch since we are replacing all new with existing.
			}
		}
	}

	if mismatch {
		//v.Status.Conditions.PoolCreated = false // We know we have a descrepancy
		v.Status.State = ilbv1alpha1.PENDING
		if err := hw.updatePoolMembers(pid, strconv.Itoa(v.Spec.Pool.Port), desiredMembers, v); err != nil {
			return err
		}
	}
	//update successful at this point. [This is the only place where we are updating status down stream. See if we can pull up in the controller]
	v.Status.State = ilbv1alpha1.READY
	return nil // check this
}

func (hw *HighWireProvider) updatePoolMembers(pid string, port string, nodes []string, v *ilbv1alpha1.Ilb) error {
	//Logic 2 : Get the desired ones (thats what we need) and replace all the current.
	//url = "https://internal-placeholder.com/v1/ltm/pools/" + pid + "?apiToken=" + token
	url := hw.BaseURL + POOLS + pid + API_TOKEN + hw.token

	//prepare request parameter
	//var members []member
	var members = make([]member, 0)

	for _, val := range nodes {
		//var m member
		var m member
		m.IP = val
		m.Port = port // strconv.Itoa(v.Spec.Pool.Port)
		members = append(members, m)
	}

	newPool := pool{
		Description:       v.Spec.Pool.Description,       //"This is created by a controller",
		LoadBalancingMode: v.Spec.Pool.LoadBalancingMode, //"least-connections-member",
		Monitor:           v.Spec.Pool.Monitor,           //"i_tcp",
		MinActiveMembers:  v.Spec.Pool.MinActiveMembers,  //1,
		Members:           members,
	}

	log.Log.Info("ObserveCurrentAndReconcile", logkeys.NewPool, newPool)
	requestParam, err := json.Marshal(newPool)

	if err != nil {
		log.Log.Error(err, string(MarshallingFailed))
		return err //convert this into return code constants
	}

	resp, err := make2HTTPRequest("PUT", url, requestParam, hw.HighwireTimeout)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Log.Error(err, string(ReadResponseFailed))
		}
		log.Log.Info("UpdatePoolMembers Response Body", logkeys.NewPool, newPool, logkeys.Response, string(body))
		msg, reasons := processStatusCode(resp.StatusCode, body)
		log.Log.Error(err, string(PoolUpdateFailed), logkeys.PoolURL, url, logkeys.Members, members, logkeys.Issues, msg+reasons)
		return fmt.Errorf(string(PoolUpdateFailed))
	}

	log.Log.Info("PoolUpdateSucceeded", logkeys.Pool, newPool)
	return nil
}

func make2HTTPRequest(method string, ep string, req []byte, highwireTimeout time.Duration) (*http.Response, error) {
	if highwireTimeout == 0 {
		highwireTimeout = time.Second * 15
	}

	client := &http.Client{
		Timeout: highwireTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
		},
	}

	//Set up a new request and execute it
	request, err := http.NewRequest(method, ep, bytes.NewBuffer(req))
	if err != nil {
		return nil, err
	}

	request.Header.Set("content-type", "application/json")
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (hw *HighWireProvider) getToken() error {
	// Making sure once process can check and update the token at a time
	hw.mu.Lock()
	defer hw.mu.Unlock()

	// Check if token is valid
	if hw.isTokenValid() {
		log.Log.Info("GetTokenResponse, Token is Valid. It will expire in duration", logkeys.Duration, time.Until(hw.tokenExpiry))
		return nil
	}

	//url := "https://internal-placeholder.com/v1/login"
	url := hw.BaseURL + GET_TOKEN
	requestParam, err := json.Marshal(map[string]string{
		"domain":   hw.Domain,
		"username": hw.UserName,
		"password": hw.Secret,
	})

	if err != nil {
		log.Log.Error(err, "Marshallling error")
		return err
	}

	resp, err := make2HTTPRequest("POST", url, requestParam, hw.HighwireTimeout)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// get the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		log.Log.Info("Get Token Response Body", logkeys.Response, string(body))
		var b []byte
		msg, reasons := processStatusCode(resp.StatusCode, b) //passing the empty byte array as we do not read body here. test for sanity here
		//v.Status.Message = updateStatusMessage(msg+reasons, int32(resp.StatusCode)) //get the reasons field in the status struct and refactor this
		log.Log.Error(err, string(TokenRequestFailed), logkeys.ResponseStatusCode, resp.StatusCode, logkeys.Issues, msg+reasons)
		return fmt.Errorf(string(TokenRequestFailed))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if s, ok := result["apiToken"].(string); ok {
		hw.token = s
		if expiry, ok := result["sessionExpires"].(float64); ok {
			hw.tokenExpiry = time.Now().Add(time.Duration(expiry) * time.Second)
		} else {
			log.Log.Info("GetExpirationTime: Could not find expiration time in response", logkeys.Expiry, expiry)
		}
		return nil
	}
	return fmt.Errorf("could not assert the api token as string value")
}

// isTokenValid function will check if current token is valid and returns true if it is, false otherwise
func (hw *HighWireProvider) isTokenValid() bool {
	// Check if token is valid and not about to expire within buffer time of 10 minutes
	return time.Now().Before(hw.tokenExpiry.Add(-10 * time.Minute))

}

func processStatusCode(c int, b []byte) (string, string) {

	// provider specific error payload
	type message struct {
		message string
	}

	//this is how errors will be returned by the provider
	type issues struct {
		Messages []message `json:"messages"`
	}

	//Set of messages associaed with the status codes in case of issues from this provider
	const (
		BAD_REQUEST  string = "Bad Request: Invalid parameters or incorrect values for env type"
		UNAUTHORIZED string = "Unauthorized: Missing, Expired, or Invalid apiToken"
		FORBIDDEN    string = "Forbidden: User/User Group does not have access to given object"
		NOT_FOUND    string = "Not Found: Object or supporting object not found"
		CONFLICT     string = "Conflict: Object already exists with name/ip/... or object is in use by another object"
		ERROR        string = "Error Occured"
	)

	//Get the error messages and return with status codes
	var ret issues
	var m, reasons string
	err := json.Unmarshal(b, &ret)
	if err != nil {
		log.Log.Error(err, string(UnmarshallingFailed))
	}
	//get all the issues appended. We have seen one message only.
	for _, msg := range ret.Messages {
		reasons += msg.message
	}

	//See what code we have received and assign the message
	switch c {
	case 400:
		m = BAD_REQUEST
	case 401:
		m = UNAUTHORIZED
	case 403:
		m = FORBIDDEN
	case 404:
		m = NOT_FOUND
	case 409:
		m = CONFLICT
	default:
		m = ERROR
	}

	return m, reasons
}

// operatorMessageString creates a json string out of OperatorMessage struct. This is the format expected by
// UI.
func operatorMessageString(errCode int32, message string) string {
	msg, err := json.Marshal(OperatorMessage{
		ErrorCode: errCode,
		Message:   message,
	})

	if err != nil {
		log.Log.Error(err, "marshal message")
	}

	return string(msg)
}
