// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package ddi

import (
	"context"
	"encoding/json"
	"errors"

	"bytes"
	"io/ioutil"
	"net/http"

	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/mocks"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MenAndMice tests")
}

var Any = gomock.Any()

var _ = Describe("Men and mice", func() {
	const (
		myUsername              = "someone"
		myPassword              = "something"
		menAndMiceUrl           = "https://1.1.1.1"
		menAndMiceServerAddress = "localhost"
		rack                    = "myrack"
		rangeType               = "BMC"
		deviceName              = "device-1"
		matchingMac             = "00:01:02:03"
		differentMac            = "00:01:02:04"
	)

	var (
		ctx                   context.Context
		menAndMice            DDI
		rangeAvailable        *Range
		mockCtrl              *gomock.Controller
		dhcpReservation       *DhcpReservation
		dhcpLeasesResults     *DhcpLeasesResult
		dhcpLease             DhcpLease
		dhcpReservationResult *DhcpReservationsResult
	)

	BeforeEach(func() {
		ctx = context.Background()
		menAndMice, _ = NewMenAndMice(ctx, myUsername, myPassword, menAndMiceUrl, menAndMiceServerAddress)
		mockCtrl = gomock.NewController(GinkgoT())
		dhcpLease = DhcpLease{
			Name:         deviceName,
			Mac:          matchingMac,
			Address:      "1.1.1.1",
			Lease:        "somedata",
			State:        "active",
			DhcpScopeRef: "someref",
		}
		dhcpLeasesResults = &DhcpLeasesResult{
			Result: struct {
				DhcpLeases   []DhcpLease "json:\"dhcpLeases\""
				TotalResults int         "json:\"totalResults\""
			}{
				DhcpLeases: []DhcpLease{
					dhcpLease,
				},
			},
		}
		rangeAvailable = &Range{
			Name: rack,
			From: "100.83.0.1",
			Ref:  "100.83.0.1",
			CustomProperties: struct {
				Title       string "json:\"Title\""
				Description string "json:\"Description\""
				Rack        string "json:\"Rack\""
				Type        string "json:\"Type\""
			}{
				Type:        rangeType,
				Title:       rack,
				Description: rack,
				Rack:        rack,
			},
			DhcpScopes: []struct {
				Ref     string "json:\"ref\""
				ObjType string "json:\"objType\""
				Name    string "json:\"name\""
			}{
				{
					Ref:     "somescope",
					ObjType: "someObject",
					Name:    "scope",
				},
			},
		}
		dhcpReservation = &DhcpReservation{
			Name: deviceName,
			Ref:  "someref",
			Addresses: []string{
				"1.1.1.2",
			},
			ClientIdentifier:  matchingMac,
			ReservationMethod: "HardwareAddress",
		}
		dhcpReservationResult = &DhcpReservationsResult{
			Result: struct {
				DhcpReservations []DhcpReservation "json:\"dhcpReservations\""
				TotalResults     int               "json:\"totalResults\""
			}{
				DhcpReservations: []DhcpReservation{
					*dhcpReservation,
				},
				TotalResults: 1,
			},
		}

	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("GetRangeByName", func() {
		It("Expect GetRangeByName to return a valid range", func() {
			response := RangesResult{
				Result: struct {
					Ranges       []Range "json:\"ranges\""
					TotalResults int     "json:\"totalResults\""
				}{
					Ranges: []Range{
						*rangeAvailable,
					},
					TotalResults: 1,
				},
			}
			menAndMice.SetClient(&mocks.HttpClientMock{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					responseBody, _ := json.Marshal(response)
					return &http.Response{
						StatusCode: 200,
						Body:       ioutil.NopCloser(bytes.NewReader(responseBody)),
					}, nil
				},
			})
			ipRange, err := menAndMice.GetRangeByName(ctx, rack, rangeType)
			Expect(err).NotTo(HaveOccurred())
			Expect(ipRange.Name).To(Equal(rack))
			Expect(ipRange.CustomProperties.Type).To(Equal(rangeType))
		})

		It("Expect GetRangeByName to return an error", func() {
			response := RangesResult{
				Result: struct {
					Ranges       []Range "json:\"ranges\""
					TotalResults int     "json:\"totalResults\""
				}{},
			}
			menAndMice.SetClient(&mocks.HttpClientMock{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					responseBody, _ := json.Marshal(response)
					return &http.Response{
						StatusCode: 200,
						Body:       ioutil.NopCloser(bytes.NewReader(responseBody)),
					}, nil
				},
			})
			_, err := menAndMice.GetRangeByName(ctx, rack, rangeType)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("GetAvailableIp", func() {
		It("Expect GetAvailableIp to return an IP Address", func() {

			response := IpAddressResult{
				Result: struct {
					Address string "json:\"address\""
				}{
					Address: "100.83.0.4",
				},
			}
			menAndMice.SetClient(&mocks.HttpClientMock{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					responseBody, _ := json.Marshal(response)
					return &http.Response{
						StatusCode: 200,
						Body:       ioutil.NopCloser(bytes.NewReader(responseBody)),
					}, nil
				},
			})
			result, _ := menAndMice.GetAvailableIp(ctx, rangeAvailable)
			Expect(result).To(Equal("100.83.0.4"))

		})
	})

	Describe("UpdateDhcpReservationByMacAddress", func() {
		It("Expect UpdateDhcpReservationByMacAddress to return an error", func() {
			err := menAndMice.UpdateDhcpReservationOptions(ctx, rangeAvailable, deviceName, "ipxe.efi", "nextServer", "00:01:02:03", "snponly.efi")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("DeleteDhcpReservation", func() {
		It("Expect DeleteDhcpReservation to return an error", func() {
			menAndMice.SetClient(&mocks.HttpClientMock{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					responseBody, _ := json.Marshal(Any)
					return &http.Response{
						StatusCode: 503,
						Body:       ioutil.NopCloser(bytes.NewReader(responseBody)),
					}, errors.New("some errors")
				},
			})
			_, err := menAndMice.DeleteDhcpReservation(ctx, deviceName)
			Expect(err).To(HaveOccurred())
		})
		It("Expect DeleteDhcpReservation succeed", func() {
			menAndMice.SetClient(&mocks.HttpClientMock{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					responseBody, _ := json.Marshal(Any)
					return &http.Response{
						StatusCode: 200,
						Body:       ioutil.NopCloser(bytes.NewReader(responseBody)),
					}, nil
				},
			})
			statusCode, _ := menAndMice.DeleteDhcpReservation(ctx, deviceName)
			Expect(statusCode).To(Equal(200))
		})
	})

	Describe("DeleteDhcpLease", func() {
		It("Expect DeleteDhcpReservation to return an error", func() {
			menAndMice.SetClient(&mocks.HttpClientMock{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					responseBody, _ := json.Marshal(Any)
					return &http.Response{
						StatusCode: 200,
						Body:       ioutil.NopCloser(bytes.NewReader(responseBody)),
					}, errors.New("some errors")
				},
			})
			_, err := menAndMice.DeleteDhcpLease(ctx, "1.2.3.4", "someName")
			Expect(err).To(HaveOccurred())
		})
		It("Expect DeleteDhcpLease fail", func() {
			menAndMice.SetClient(&mocks.HttpClientMock{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					responseBody, _ := json.Marshal(Any)
					return &http.Response{
						StatusCode: 500,
						Body:       ioutil.NopCloser(bytes.NewReader(responseBody)),
					}, nil
				},
			})
			statusCode, _ := menAndMice.DeleteDhcpLease(ctx, "1.2.3.4", "someName")
			Expect(statusCode).To(Equal(500))
		})
	})

	Describe("GetDhcpReservationsByMacAddress", func() {
		It("Expect GetDhcpReservationsByMacAddress to succeed and return a DHCP reservation", func() {
			menAndMice.SetClient(&mocks.HttpClientMock{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					responseBody, _ := json.Marshal(dhcpReservationResult)
					return &http.Response{
						StatusCode: 200,
						Body:       ioutil.NopCloser(bytes.NewReader(responseBody)),
					}, nil
				},
			})
			reservation, _ := menAndMice.GetDhcpReservationsByMacAddress(ctx, "somedhcpScope", matchingMac)
			Expect(reservation).To(Equal(dhcpReservation))
		})
		It("Expect GetDhcpReservationsByMacAddress to fail if it can't find any reservation", func() {
			menAndMice.SetClient(&mocks.HttpClientMock{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					responseBody, _ := json.Marshal(dhcpReservationResult)
					return &http.Response{
						StatusCode: 200,
						Body:       ioutil.NopCloser(bytes.NewReader(responseBody)),
					}, nil
				},
			})
			_, err := menAndMice.GetDhcpReservationsByMacAddress(ctx, "somedhcpScope", differentMac)
			Expect(err).To(HaveOccurred())
		})
		It("Expect GetDhcpReservationsByMacAddress to fail if no DHCP reservations found", func() {
			menAndMice.SetClient(&mocks.HttpClientMock{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					dhcpReservationResult.Result.DhcpReservations = []DhcpReservation{}
					responseBody, _ := json.Marshal(dhcpReservationResult)
					return &http.Response{
						StatusCode: 200,
						Body:       ioutil.NopCloser(bytes.NewReader(responseBody)),
					}, nil
				},
			})
			_, err := menAndMice.GetDhcpReservationsByMacAddress(ctx, "somedhcpScope", matchingMac)
			Expect(err).To(HaveOccurred())
		})
		It("Expect GetDhcpReservationsByMacAddress to fail if http request  to DDI return error", func() {
			menAndMice.SetClient(&mocks.HttpClientMock{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					dhcpReservationResult.Result.DhcpReservations = []DhcpReservation{}
					responseBody, _ := json.Marshal(dhcpReservationResult)
					return &http.Response{
						StatusCode: 200,
						Body:       ioutil.NopCloser(bytes.NewReader(responseBody)),
					}, errors.New("some error")
				},
			})
			_, err := menAndMice.GetDhcpReservationsByMacAddress(ctx, "somedhcpScope", matchingMac)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("GetDhcpLeasesByScope", func() {
		It("Expect GetDhcpLeasesByScope to fail if http request  to DDI return error", func() {
			menAndMice.SetClient(&mocks.HttpClientMock{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					responseBody, _ := json.Marshal(Any)
					return &http.Response{
						StatusCode: 200,
						Body:       ioutil.NopCloser(bytes.NewReader(responseBody)),
					}, errors.New("some error")
				},
			})
			_, err := menAndMice.GetDhcpLeasesByScope(ctx, "somedhcpScope", matchingMac)
			Expect(err).To(HaveOccurred())
		})

		It("Expect GetDhcpLeasesByScope to succeeed if match found", func() {
			menAndMice.SetClient(&mocks.HttpClientMock{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					responseBody, _ := json.Marshal(dhcpLeasesResults)
					return &http.Response{
						StatusCode: 200,
						Body:       ioutil.NopCloser(bytes.NewReader(responseBody)),
					}, nil
				},
			})
			result, _ := menAndMice.GetDhcpLeasesByScope(ctx, "somedhcpScope", matchingMac)
			Expect(result).To(Equal(&dhcpLease))
		})

		It("Expect GetDhcpLeasesByScope to fail if no leases found", func() {
			dhcpLeasesResults.Result.DhcpLeases = []DhcpLease{}
			menAndMice.SetClient(&mocks.HttpClientMock{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					responseBody, _ := json.Marshal(dhcpLeasesResults)
					return &http.Response{
						StatusCode: 200,
						Body:       ioutil.NopCloser(bytes.NewReader(responseBody)),
					}, nil
				},
			})
			_, err := menAndMice.GetDhcpLeasesByScope(ctx, "somedhcpScope", matchingMac)
			Expect(err).To(HaveOccurred())
		})
	})

})
