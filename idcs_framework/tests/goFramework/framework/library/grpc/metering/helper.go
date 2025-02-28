package metering

import (
	"time"
)

//Structs to store responses

type PostResponse struct {
}

type PatchResponse struct {
}

type SearchStruct struct {
	Error struct {
		Code    int `json:"code"`
		Details []struct {
			Type            string `json:"@type"`
			AdditionalProp1 string `json:"additionalProp1"`
			AdditionalProp2 string `json:"additionalProp2"`
			AdditionalProp3 string `json:"additionalProp3"`
		} `json:"details"`
		Message string `json:"message"`
	} `json:"error"`
	Result struct {
		CloudAccountID string `json:"cloudAccountId"`
		ID             string `json:"id"`
		Properties     struct {
			AdditionalProp1 string `json:"additionalProp1"`
			AdditionalProp2 string `json:"additionalProp2"`
			AdditionalProp3 string `json:"additionalProp3"`
		} `json:"properties"`
		Reported      bool      `json:"reported"`
		ResourceID    string    `json:"resourceId"`
		Timestamp     time.Time `json:"timestamp"`
		TransactionID string    `json:"transactionId"`
	} `json:"result"`
}

type PreviousRecordStruct struct {
	CloudAccountID string `json:"cloudAccountId"`
	ID             string `json:"id"`
	Properties     struct {
		AdditionalProp1 string `json:"additionalProp1"`
		AdditionalProp2 string `json:"additionalProp2"`
		AdditionalProp3 string `json:"additionalProp3"`
	} `json:"properties"`
	Reported      bool      `json:"reported"`
	ResourceID    string    `json:"resourceId"`
	Timestamp     time.Time `json:"timestamp"`
	TransactionID string    `json:"transactionId"`
}

type CreatePostStruct struct {
	TransactionId  string            `json:"transactionId,omitempty"`
	ResourceId     string            `json:"resourceId,omitempty"`
	CloudAccountId string            `json:"cloudAccountId,omitempty"`
	Timestamp      string            `json:"timestamp,omitempty"`
	Properties     map[string]string `json:"properties,omitempty"`
}

type UsageFilter struct {
	Id             *int64  `json:"id,omitempty"`
	TransactionId  *string `json:"transactionId,omitempty"`
	ResourceId     *string `json:"resourceId,omitempty"`
	CloudAccountId *string `json:"cloudAccountId,omitempty"`
	StartTime      *string `json:"startTime,omitempty"`
	EndTime        *string `json:"endTime,omitempty"`
	Reported       *bool   `json:"reported,omitempty"`
}

type UsagePrevious struct {
	Id         string `json:"id,omitempty"`
	ResourceId string `json:"resourceId,omitempty"`
}

type UsageUpdate struct {
	Id       []int64 `json:"id,omitempty"`
	Reported bool    `json:"reported,omitempty"`
}

var CREATE_ENDPOINT = "proto.MeteringService/Create"
var SEARCH_ENDPOINT = "proto.MeteringService/Search"
var UPDATE_ENDPOINT = "proto.MeteringService/Update"
var FINDPREVIOUS_ENDPOINT = "proto.MeteringService/FindPrevious"
var FINDPREVIOUS = "proto.MeteringService/FindPrevious"
