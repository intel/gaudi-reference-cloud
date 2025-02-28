package financials_utils

import (
	"encoding/json"
	"fmt"
	"goFramework/utils"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Payload interface {
	GeneratePayload(transactionId string, resourceId string, cloudAccountId string, timestamp string, properties map[string]string, serviceType string) Payload
}

type BaseMeteringPayload struct {
	TransactionId  string `json:"transactionId"`
	ResourceId     string `json:"resourceId"`
	CloudAccountId string `json:"cloudAccountId"`
	Timestamp      string `json:"timestamp"`
}

type BaseMaaSUsagePayload struct {
	CloudAccountId string `json:"cloudAccountId"`
	EndTime        string `json:"endTime"`
	Quantity       int    `json:"quantity"`
	Region         string `json:"region"`
	StartTime      string `json:"startTime"`
	Timestamp      string `json:"timestamp"`
	TransactionId  string `json:"transactionId"`
}

type ClusterMeteringPayload struct {
	BaseMeteringPayload
	Properties struct {
		AvailabilityZone    string `json:"availabilityZone"`
		ClusterId           string `json:"clusterId"`
		Deleted             string `json:"deleted"`
		FirstReadyTimestamp string `json:"firstReadyTimestamp"`
		ClusterType         string `json:"clusterType"`
		InstanceName        string `json:"instanceName"`
		InstanceType        string `json:"instanceType"`
		Region              string `json:"region"`
		RunningSeconds      string `json:"runningSeconds"`
		ServiceType         string `json:"serviceType"`
	} `json:"properties"`
}

type ComputeMeteringPayload struct {
	BaseMeteringPayload
	Properties struct {
		AvailabilityZone    string `json:"availabilityZone"`
		ClusterId           string `json:"clusterId"`
		Deleted             string `json:"deleted"`
		FirstReadyTimestamp string `json:"firstReadyTimestamp"`
		InstanceGroup       string `json:"instanceGroup"`
		InstanceGroupSize   string `json:"instanceGroupSize"`
		InstanceName        string `json:"instanceName"`
		InstanceType        string `json:"instanceType"`
		Region              string `json:"region"`
		RunningSeconds      string `json:"runningSeconds"`
		ServiceType         string `json:"serviceType"`
	} `json:"properties"`
}

type FileStorageMeteringPayload struct {
	BaseMeteringPayload
	Properties struct {
		TB                      string `json:"TB"`
		AvailabilityZone        string `json:"availabilityZone"`
		FilesystemConditionType string `json:"filesystemConditionType"`
		FilesystemName          string `json:"filesystemName"`
		Deleted                 string `json:"deleted"`
		FirstReadyTimestamp     string `json:"firstReadyTimestamp"`
		Hour                    string `json:"hour"`
		Region                  string `json:"region"`
		RunningSeconds          string `json:"runningSeconds"`
		ServiceType             string `json:"serviceType"`
	} `json:"properties"`
}

type MaaSUsagePayload struct {
	BaseMaaSUsagePayload
	Properties struct {
		ServiceType    string `json:"serviceType"`
		ProcessingType string `json:"processingType"`
	} `json:"properties"`
}

type ObjectStorageMeteringPayload struct {
	BaseMeteringPayload
	Properties struct {
		TB                  string `json:"TB"`
		AvailabilityZone    string `json:"availabilityZone"`
		BucketConditionType string `json:"bucketConditionType"`
		BucketName          string `json:"bucketName"`
		Deleted             string `json:"deleted"`
		FirstReadyTimestamp string `json:"firstReadyTimestamp"`
		Hour                string `json:"hour"`
		Region              string `json:"region"`
		RunningSeconds      string `json:"runningSeconds"`
		ServiceType         string `json:"serviceType"`
	} `json:"properties"`
}

func (m ComputeMeteringPayload) GeneratePayload(transactionId string, resourceId string, cloudAccountId string, timestamp string, properties map[string]string, serviceType string) Payload {
	m.CloudAccountId = cloudAccountId
	m.TransactionId = transactionId
	m.ResourceId = resourceId
	m.Timestamp = timestamp
	m.Properties.AvailabilityZone = properties["availabilityZone"]
	m.Properties.ClusterId = properties["clusterId"]
	m.Properties.FirstReadyTimestamp = properties["firstReadyTimestamp"]
	m.Properties.InstanceGroup = properties["instanceGroup"]
	m.Properties.InstanceGroupSize = properties["instanceGroupSize"]
	m.Properties.Deleted = properties["deleted"]
	m.Properties.InstanceName = properties["instanceName"]
	m.Properties.InstanceType = properties["instanceType"]
	m.Properties.Region = properties["region"]
	m.Properties.RunningSeconds = properties["runningSeconds"]
	m.Properties.ServiceType = serviceType
	return m
}

func (m ClusterMeteringPayload) GeneratePayload(transactionId string, resourceId string, cloudAccountId string, timestamp string, properties map[string]string, serviceType string) Payload {
	m.CloudAccountId = cloudAccountId
	m.TransactionId = transactionId
	m.ResourceId = resourceId
	m.Timestamp = timestamp
	m.Properties.AvailabilityZone = properties["availabilityZone"]
	m.Properties.ClusterId = properties["clusterId"]
	m.Properties.FirstReadyTimestamp = properties["firstReadyTimestamp"]
	m.Properties.ClusterType = properties["clusterType"]
	m.Properties.Deleted = properties["deleted"]
	m.Properties.InstanceName = properties["instanceName"]
	m.Properties.InstanceType = properties["instanceType"]
	m.Properties.Region = properties["region"]
	m.Properties.RunningSeconds = properties["runningSeconds"]
	m.Properties.ServiceType = serviceType
	return m
}

func (m FileStorageMeteringPayload) GeneratePayload(transactionId string, resourceId string, cloudAccountId string, timestamp string, properties map[string]string, serviceType string) Payload {
	m.CloudAccountId = cloudAccountId
	m.TransactionId = transactionId
	m.ResourceId = resourceId
	m.Timestamp = timestamp
	m.Properties.TB = properties["TB"]
	m.Properties.AvailabilityZone = properties["availabilityZone"]
	m.Properties.FilesystemName = properties["filesystemName"]
	m.Properties.FilesystemConditionType = properties["filesystemConditionType"]
	m.Properties.Deleted = properties["deleted"]
	m.Properties.FirstReadyTimestamp = properties["firstReadyTimestamp"]
	m.Properties.Hour = properties["hour"]
	m.Properties.Region = properties["region"]
	m.Properties.ServiceType = serviceType
	return m
}

func (m ObjectStorageMeteringPayload) GeneratePayload(transactionId string, resourceId string, cloudAccountId string, timestamp string, properties map[string]string, serviceType string) Payload {
	m.CloudAccountId = cloudAccountId
	m.TransactionId = transactionId
	m.ResourceId = resourceId
	m.Timestamp = timestamp
	m.Properties.TB = properties["TB"]
	m.Properties.AvailabilityZone = properties["availabilityZone"]
	m.Properties.BucketName = properties["bucketName"]
	m.Properties.BucketConditionType = properties["BucketConditionType"]
	m.Properties.Deleted = properties["deleted"]
	m.Properties.FirstReadyTimestamp = properties["firstReadyTimestamp"]
	m.Properties.Hour = properties["hour"]
	m.Properties.Region = properties["region"]
	m.Properties.ServiceType = serviceType
	return m
}

func (m MaaSUsagePayload) GeneratePayload(transactionId string, resourceId string, cloudAccountId string, timestamp string, properties map[string]string, serviceType string) Payload {
	// Convert the string to an integer
	quantity, err := strconv.Atoi(properties["quantity"])
	if err != nil {
		fmt.Println("Error converting quantity to int:", err)
	}
	current_time := time.Now().Add(-1 * time.Hour)
	// Subtract one day from current_time
	startTime := current_time.AddDate(0, 0, -1)
	// Define the custom layout for the desired format
	const layout = "2006-01-02T15:04:05.000Z"
	// Format startTime using the custom layout
	startTimeFormatted := startTime.Format(layout)
	endTimeFormatted := current_time.Format(layout)
	actual_time := time.Now()
	currentTimeFormatted := actual_time.Format(layout)

	fmt.Println("Current Time:", current_time.Format(time.RFC3339Nano))
	fmt.Println("Start Time (one day less):", startTimeFormatted)
	m.CloudAccountId = cloudAccountId
	m.TransactionId = transactionId
	m.Timestamp = currentTimeFormatted
	m.StartTime = startTimeFormatted
	m.EndTime = endTimeFormatted
	m.Quantity = quantity
	m.Region = properties["region"]
	m.Properties.ProcessingType = properties["processingType"]
	m.Properties.ServiceType = serviceType
	return m
}

func EnrichMmeteringSearchPayload(rawpayload string, resourceId string, cloudAccountId string) string {
	var enriched_payload = rawpayload
	enriched_payload = strings.Replace(enriched_payload, "<<resourceId>>", resourceId, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<cloudAccountId>>", cloudAccountId, 1)
	return enriched_payload
}

func EnrichMmeteringSearchPayloadCloudAcc(rawpayload string, cloudAccountId string) string {
	var enriched_payload = rawpayload
	enriched_payload = strings.Replace(enriched_payload, "<<cloudAccountId>>", cloudAccountId, 1)
	return enriched_payload
}

func EnrichRedeemCouponPayload1(rawpayload string, code string, cloudAccountId string) string {
	enriched_payload := rawpayload
	enriched_payload = strings.Replace(enriched_payload, "<<code>>", code, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<cloudAccountId>>", cloudAccountId, 1)
	return enriched_payload
}

func EnrichMeteringCreatePayload(rawpayload string, transactionId string, resourceId string, cloudAccountId string, timestamp string, instanceType string, instanceName string, runningSeconds string) string {
	var enriched_payload = rawpayload
	enriched_payload = strings.Replace(enriched_payload, "<<resourceId>>", resourceId, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<cloudAccountId>>", cloudAccountId, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<transactionId>>", transactionId, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<timestamp>>", timestamp, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<instanceType>>", instanceType, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<instanceName>>", instanceName, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<runningSeconds>>", runningSeconds, 1)
	return enriched_payload
}

func EnrichMeteringCreatePayloadUsages(rawpayload string, transactionId string, resourceId string, cloudAccountId string, timestamp string, instanceType string, instanceName string, runningSeconds string, serviceType string) string {
	var enriched_payload = rawpayload
	enriched_payload = strings.Replace(enriched_payload, "<<resourceId>>", resourceId, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<cloudAccountId>>", cloudAccountId, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<transactionId>>", transactionId, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<timestamp>>", timestamp, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<instanceType>>", instanceType, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<instanceName>>", instanceName, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<runningSeconds>>", runningSeconds, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<serviceType>>", serviceType, 1)
	return enriched_payload
}

func EnrichMeteringCreatePayloadUsagesSTaaS(rawpayload string, transactionId string, resourceId string, cloudAccountId string, timestamp string, instanceType string, instanceName string, runningSeconds string, serviceType string) string {
	var enriched_payload = rawpayload
	enriched_payload = strings.Replace(enriched_payload, "<<resourceId>>", resourceId, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<cloudAccountId>>", cloudAccountId, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<transactionId>>", transactionId, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<timestamp>>", timestamp, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<instanceType>>", instanceType, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<instanceName>>", instanceName, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<runningSeconds>>", runningSeconds, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<serviceType>>", serviceType, 1)
	return enriched_payload
}

func EnrichMeteringCreatePayloadUsagesIKS(rawpayload string, transactionId string, resourceId string, cloudAccountId string, timestamp string, instanceType string, instanceName string, runningSeconds string, serviceType string) string {
	var enriched_payload = rawpayload
	enriched_payload = strings.Replace(enriched_payload, "<<resourceId>>", resourceId, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<cloudAccountId>>", cloudAccountId, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<transactionId>>", transactionId, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<timestamp>>", timestamp, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<instanceType>>", instanceType, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<instanceName>>", instanceName, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<runningSeconds>>", runningSeconds, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<serviceType>>", serviceType, 1)
	return enriched_payload
}

func EnrichInvalidMeteringCreatePayload(rawpayload string, transactionId string, resourceId string, cloudAccountId string, timestamp string, instanceType string, instanceName string, runningSeconds string, availabilityZone string, clusterId string, deleted string, region string, serviceType string) string {
	var enriched_payload = rawpayload
	enriched_payload = strings.Replace(enriched_payload, "<<resourceId>>", resourceId, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<cloudAccountId>>", cloudAccountId, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<transactionId>>", transactionId, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<timestamp>>", timestamp, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<instanceType>>", instanceType, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<instanceName>>", instanceName, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<runningSeconds>>", runningSeconds, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<availabilityZone>>", availabilityZone, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<clusterId>>", clusterId, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<deleted>>", deleted, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<region>>", region, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<serviceType>>", serviceType, 1)
	return enriched_payload
}

func GenerateComputePropertiesMap(availabilityZone string, clusterId string, deleted string, firstReadyTimestamp string, instanceGroup string, instanceGroupSize string, instanceName string, instanceType string, region string, runningSeconds string) map[string]string {
	ComputePropertiesMap := make(map[string]string)
	ComputePropertiesMap["availabilityZone"] = availabilityZone
	ComputePropertiesMap["clusterId"] = clusterId
	ComputePropertiesMap["deleted"] = deleted
	ComputePropertiesMap["firstReadyTimestamp"] = firstReadyTimestamp
	ComputePropertiesMap["instanceGroup"] = instanceGroup
	ComputePropertiesMap["instanceGroupSize"] = instanceGroupSize
	ComputePropertiesMap["instanceName"] = instanceName
	ComputePropertiesMap["instanceType"] = instanceType
	ComputePropertiesMap["region"] = region
	ComputePropertiesMap["runningSeconds"] = runningSeconds
	return ComputePropertiesMap
}

func GenerateClusterPropertiesMap(availabilityZone string, clusterId string, deleted string, firstReadyTimestamp string, clusterType string, instanceName string, instanceType string, region string, runningSeconds string) map[string]string {
	ComputePropertiesMap := make(map[string]string)
	ComputePropertiesMap["availabilityZone"] = availabilityZone
	ComputePropertiesMap["clusterId"] = clusterId
	ComputePropertiesMap["deleted"] = deleted
	ComputePropertiesMap["clusterType"] = clusterType
	ComputePropertiesMap["firstReadyTimestamp"] = firstReadyTimestamp
	ComputePropertiesMap["instanceName"] = instanceName
	ComputePropertiesMap["instanceType"] = instanceType
	ComputePropertiesMap["region"] = region
	ComputePropertiesMap["runningSeconds"] = runningSeconds
	return ComputePropertiesMap
}

func GenerateFileStoragePropertiesMap(TB string, availabilityZone string, filesystemConditionType string, filesystemName string, deleted string, firstReadyTimestamp string, hour string, region string) map[string]string {
	FileStoragePropertiesMap := make(map[string]string)
	FileStoragePropertiesMap["TB"] = TB
	FileStoragePropertiesMap["availabilityZone"] = availabilityZone
	FileStoragePropertiesMap["filesystemConditionType"] = filesystemConditionType
	FileStoragePropertiesMap["filesystemName"] = filesystemName
	FileStoragePropertiesMap["deleted"] = deleted
	FileStoragePropertiesMap["firstReadyTimestamp"] = firstReadyTimestamp
	FileStoragePropertiesMap["hour"] = hour
	FileStoragePropertiesMap["region"] = region
	return FileStoragePropertiesMap
}

func GenerateObjectStoragePropertiesMap(TB string, availabilityZone string, bucketConditionType string, bucketName string, deleted string, firstReadyTimestamp string, hour string, region string) map[string]string {
	FileStoragePropertiesMap := make(map[string]string)
	FileStoragePropertiesMap["TB"] = TB
	FileStoragePropertiesMap["availabilityZone"] = availabilityZone
	FileStoragePropertiesMap["bucketConditionType"] = bucketConditionType
	FileStoragePropertiesMap["bucketName"] = bucketName
	FileStoragePropertiesMap["deleted"] = deleted
	FileStoragePropertiesMap["firstReadyTimestamp"] = firstReadyTimestamp
	FileStoragePropertiesMap["hour"] = hour
	FileStoragePropertiesMap["region"] = region
	return FileStoragePropertiesMap
}

func GenerateMaaSPropertiesMap(serviceType string, processingType string, region string, quantity string) map[string]string {
	MaaSPropertiesMap := make(map[string]string)
	MaaSPropertiesMap["serviceType"] = serviceType
	MaaSPropertiesMap["processingType"] = processingType
	MaaSPropertiesMap["region"] = region
	MaaSPropertiesMap["quantity"] = quantity
	return MaaSPropertiesMap
}

func ConvertStructToJson(p Payload) string {
	bytePayload, err := json.Marshal(p)
	if err != nil {
		fmt.Println(err)
		panic("Failed to convert the Struct to bytes")
	}
	return string(bytePayload)
}

func GenerateMetringPayload(serviceType string, productName string, cloud_account_id string, instanceGroupSize string, firstReadyTimeStamp string) Payload {
	properties := make(map[string]string)
	current_time := time.Now().Add(-1 * time.Hour)
	current_time_d := current_time.Format(time.RFC3339Nano)
	fmt.Println("Creation time to be set: ", current_time_d)
	instanceName := utils.GenerateString(10) + "-vacuum-tests"
	region := "us-staging-1"
	fmt.Println("Service Type", serviceType)
	switch {
	case productName == "iks-cluster":
		fmt.Println("Creating IKS Cluster product Payload...")
		var cluster_payload ClusterMeteringPayload
		properties = GenerateClusterPropertiesMap("us-staging-3a", "metal3-1", "false", firstReadyTimeStamp, "iks-cp-cluster", instanceName, productName, "us-staging-3", "3000")
		return cluster_payload.GeneratePayload(uuid.NewString(), uuid.NewString(), cloud_account_id, current_time_d, properties, serviceType)
	case productName == "sc-cluster":
		fmt.Println("Creating Super computing Cluster product Payload...")
		var cluster_payload ClusterMeteringPayload
		properties = GenerateClusterPropertiesMap("us-staging-3a", "metal3-1", "false", firstReadyTimeStamp, "sc-cp-cluster", instanceName, productName, "us-staging-3", "3000")
		return cluster_payload.GeneratePayload(uuid.NewString(), uuid.NewString(), cloud_account_id, current_time_d, properties, serviceType)
	case serviceType == "ComputeAsAService" && !strings.Contains(productName, "iks"):
		fmt.Println("Creating Compute Metering Payload...")
		var compute_payload ComputeMeteringPayload
		properties = GenerateComputePropertiesMap("us-staging-1a", "test", "false", firstReadyTimeStamp, instanceName, instanceGroupSize, instanceName, productName, region, "3000")
		return compute_payload.GeneratePayload(uuid.NewString(), uuid.NewString(), cloud_account_id, current_time_d, properties, serviceType)
	case serviceType == "FileStorageAsAService" && productName == "sc-storage-file":
		fmt.Println("Creating Supercomputing File Storage Metering Payload...")
		var staas_payload FileStorageMeteringPayload
		serviceType = "FileStorageAsAService-SC"
		properties = GenerateFileStoragePropertiesMap("4.547", "us-staging-1a", "Namespace Success", instanceName, "false", firstReadyTimeStamp, "20", "us-staging-3")
		return staas_payload.GeneratePayload(uuid.NewString(), uuid.NewString(), cloud_account_id, current_time_d, properties, serviceType)
	case serviceType == "FileStorageAsAService" && productName != "sc-storage-file":
		fmt.Println("Creating STaaS File Storage Metering Payload...")
		var staas_payload FileStorageMeteringPayload
		properties = GenerateFileStoragePropertiesMap("4.547", "us-staging-1a", "Namespace Success", instanceName, "false", firstReadyTimeStamp, "20", region)
		return staas_payload.GeneratePayload(uuid.NewString(), uuid.NewString(), cloud_account_id, current_time_d, properties, serviceType)
	case serviceType == "ObjectStorageAsAService" || serviceType == "ObjectStorageAsAService-SC":
		fmt.Println("Creating STaaS Object Storage Metering Payload...")
		var staas_payload ObjectStorageMeteringPayload
		properties = GenerateObjectStoragePropertiesMap("4.547", "us-staging-1a", "accepted", instanceName, "false", firstReadyTimeStamp, "20", region)
		return staas_payload.GeneratePayload(uuid.NewString(), uuid.NewString(), cloud_account_id, current_time_d, properties, serviceType)
	case serviceType == "ObjectStorageAsAService-SC":
		fmt.Println("Creating STaaS Object Storage Metering Payload...")
		var staas_payload ObjectStorageMeteringPayload
		properties = GenerateObjectStoragePropertiesMap("4.547", "us-staging-1a", "accepted", instanceName, "false", firstReadyTimeStamp, "20", "us-staging-3")
		return staas_payload.GeneratePayload(uuid.NewString(), uuid.NewString(), cloud_account_id, current_time_d, properties, serviceType)
	case serviceType == "KubernetesAsAService" && productName != "iks-cluster":
		fmt.Println("Creating IKS Metering Payload...")
		var compute_payload ComputeMeteringPayload
		properties = GenerateComputePropertiesMap("us-staging-1a", "harvester1", "false", firstReadyTimeStamp, instanceName, instanceGroupSize, instanceName, productName, region, "3000")
		return compute_payload.GeneratePayload(uuid.NewString(), uuid.NewString(), cloud_account_id, current_time_d, properties, serviceType)
	case serviceType == "SuperComputingAsAService" && productName != "sc-cluster":
		fmt.Println("Creating Supercomputing Metering Payload...")
		var compute_payload ComputeMeteringPayload
		properties = GenerateComputePropertiesMap("us-staging-3a", "metal3-1", "false", firstReadyTimeStamp, instanceName, instanceGroupSize, instanceName, productName, "us-staging-3", "3000")
		return compute_payload.GeneratePayload(uuid.NewString(), uuid.NewString(), cloud_account_id, current_time_d, properties, serviceType)
	case serviceType == "ModelAsAService":
		fmt.Println("Creating Model As A Service Usage Payload...")
		var maas_payload MaaSUsagePayload
		// Seed the random number generator and randomly set processingType
		processingType := "text"
		properties = GenerateMaaSPropertiesMap(serviceType, processingType, region, "10000000")
		return maas_payload.GeneratePayload(uuid.NewString(), uuid.NewString(), cloud_account_id, current_time_d, properties, serviceType)
	default:
		fmt.Println("Creating Compute Metering Payload DEFAULT......")
		var compute_payload ComputeMeteringPayload
		properties = GenerateComputePropertiesMap("us-staging-1a", "test", "false", firstReadyTimeStamp, "", instanceGroupSize, instanceName, productName, region, "3000")
		metering_paylaod := compute_payload.GeneratePayload(uuid.NewString(), uuid.NewString(), cloud_account_id, current_time_d, properties, serviceType)
		return metering_paylaod
	}
}
