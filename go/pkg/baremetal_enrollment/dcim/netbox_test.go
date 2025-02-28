// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package dcim

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/netbox-community/go-netbox/netbox/client/dcim"
	"github.com/netbox-community/go-netbox/netbox/client/virtualization"
	"github.com/netbox-community/go-netbox/netbox/models"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Netbox tests")
}

var _ = Describe("Netbox", func() {

	var (
		ctx        context.Context
		token      string
		netBox     *NetBox
		netBoxURL  *url.URL
		netBoxHost string
		rackName   string

		// interface vars
		macAddress             string
		invalidMacAddress      string
		bmcLabel               string
		interfaceType          *models.InterfaceType
		interfaceName          string
		interface1DeviceInfo   *models.NestedDevice
		interface1Model        *models.Interface
		interface1Response     *dcim.DcimInterfacesListOKBody
		interface1ResponseData []byte
		interface2DeviceInfo   *models.NestedDevice
		interface2Model        *models.Interface
		interface2Response     *dcim.DcimInterfacesListOKBody
		interface2ResponseData []byte
		interface3DeviceInfo   *models.NestedDevice
		interface3Model        *models.Interface
		interface3Response     *dcim.DcimInterfacesListOKBody
		interface3ResponseData []byte
		interface4DeviceInfo   *models.NestedDevice
		interface4Model        *models.Interface
		interface4Response     *dcim.DcimInterfacesListOKBody
		interface4ResponseData []byte

		// sites vars
		site1Region       *models.NestedRegion
		site1Model        *models.Site
		site1Response     *dcim.DcimSitesListOKBody
		site1ResponseData []byte
		site2Region       *models.NestedRegion
		site2Model        *models.Site
		site2Response     *dcim.DcimSitesListOKBody
		site2ResponseData []byte
		site3Region       *models.NestedRegion
		site3Model        *models.Site
		site3Response     *dcim.DcimSitesListOKBody
		site3ResponseData []byte

		// devices vars
		deviceModelName       string
		deviceRoleName        string
		deviceType            *models.NestedDeviceType
		device1Role           *models.NestedDeviceRole
		device1Rack           *models.NestedRack
		device1Site           *models.NestedSite
		device1Config         *models.DeviceWithConfigContext
		devices1Response      *dcim.DcimDevicesListOKBody
		devices1ResponseData  []byte
		devices5ResponseData  []byte
		device2Role           *models.NestedDeviceRole
		device2Rack           *models.NestedRack
		device2Site           *models.NestedSite
		device2Config         *models.DeviceWithConfigContext
		devices2Response      *dcim.DcimDevicesListOKBody
		devices2ResponseData  []byte
		devicePutResponseData []byte
		clusterType           *models.NestedClusterType

		site1Name string
		site2Name string
		site3Name string

		device1Name string
		device5Name string
		device2Name string
		device3Name string
		device4Name string

		count       int64
		emptyCount  int64
		excessCount int64

		clusterTypeName      string
		cluster1Name         string
		cluster2Name         string
		cluster3Name         string
		cluster4Name         string
		site4                *models.NestedSite
		cluster1             *models.Cluster
		cluster1Response     *virtualization.VirtualizationClustersListOKBody
		cluster1ResponseData []byte
		cluster2             *models.Cluster
		cluster2Response     *virtualization.VirtualizationClustersListOKBody
		cluster2ResponseData []byte
		cluster3             *models.Cluster
		cluster3Response     *virtualization.VirtualizationClustersListOKBody
		cluster3ResponseData []byte
		cluster4             *models.Cluster
		cluster4Response     *virtualization.VirtualizationClustersListOKBody
		cluster4ResponseData []byte

		// http test server
		server *httptest.Server
	)

	Describe("NetBox tests", Ordered, func() {
		BeforeAll(func() {

			interfaceName = "BMC"
			count = 1
			emptyCount = 0
			excessCount = 2
			macAddress = "001122334455"
			invalidMacAddress = "0011223344"
			bmcLabel = "http://1.1.1.1"
			deviceModelName = "device-type-1"
			deviceRoleName = "bmaas"
			rackName = "rack-1"

			// device names
			device1Name = "device-1"
			device5Name = "device-5"
			device2Name = "device-2"
			device3Name = "device-3"
			device4Name = "device-4"

			// clusters
			clusterTypeName = "gaudi"
			cluster1Name = "1"
			cluster2Name = "2"
			cluster3Name = "3"
			cluster4Name = "4"

			//interface-1 config
			interface1DeviceInfo = &models.NestedDevice{ID: 1, Name: &device1Name, Display: "device-1"}
			interface1Model = &models.Interface{ID: 1, Name: &interfaceName, Device: interface1DeviceInfo, Type: interfaceType, Label: bmcLabel, MacAddress: &macAddress}
			interface1Response = &dcim.DcimInterfacesListOKBody{Count: &count, Results: []*models.Interface{interface1Model}}
			interface1ResponseData, _ = json.Marshal(interface1Response)

			//interface-2 config
			interface2DeviceInfo = &models.NestedDevice{ID: 2, Name: &device2Name, Display: "device-2"}
			interface2Model = &models.Interface{ID: 2, Name: &interfaceName, Device: interface2DeviceInfo, Type: interfaceType, Label: "", MacAddress: nil}
			interface2Response = &dcim.DcimInterfacesListOKBody{Count: &count, Results: []*models.Interface{interface2Model}}
			interface2ResponseData, _ = json.Marshal(interface2Response)

			//interface-3 config
			// setting count to 0
			interface3DeviceInfo = &models.NestedDevice{ID: 3, Name: &device3Name, Display: "device-3"}
			interface3Model = &models.Interface{ID: 3, Name: &interfaceName, Device: interface3DeviceInfo, Type: interfaceType, Label: "", MacAddress: nil}
			interface3Response = &dcim.DcimInterfacesListOKBody{Count: &emptyCount, Results: []*models.Interface{interface3Model}}
			interface3ResponseData, _ = json.Marshal(interface3Response)

			//interface-4 config
			// setting mac address to invalid MAC address

			interface4DeviceInfo = &models.NestedDevice{ID: 4, Name: &device4Name, Display: "device-4"}
			interface4Model = &models.Interface{ID: 4, Name: &interfaceName, Device: interface4DeviceInfo, Type: interfaceType, Label: "", MacAddress: &invalidMacAddress}
			interface4Response = &dcim.DcimInterfacesListOKBody{Count: &count, Results: []*models.Interface{interface4Model}}
			interface4ResponseData, _ = json.Marshal(interface4Response)

			// site-1
			site1Name = "us-dev-1a"
			site1Region = &models.NestedRegion{Name: &site1Name, Slug: &site1Name, ID: 1}
			site1Model = &models.Site{Name: &site1Name, Slug: &site1Name, ID: 1, Region: site1Region}
			site1Response = &dcim.DcimSitesListOKBody{Count: &count, Results: []*models.Site{site1Model}}
			site1ResponseData, _ = json.Marshal(site1Response)

			// site-2
			// Empty region name
			site2Name = "us-dev-1b"
			site2Region = &models.NestedRegion{Name: nil, Slug: &site2Name, ID: 2}
			site2Model = &models.Site{Name: &site2Name, Slug: &site1Name, ID: 2, Region: site2Region}
			site2Response = &dcim.DcimSitesListOKBody{Count: &count, Results: []*models.Site{site2Model}}
			site2ResponseData, _ = json.Marshal(site2Response)

			// site-3
			// Count 0
			site3Name = "us-dev-1c"
			site3Region = &models.NestedRegion{Name: nil, Slug: &site3Name, ID: 3}
			site3Model = &models.Site{Name: &site3Name, Slug: &site1Name, ID: 3, Region: site3Region}
			site3Response = &dcim.DcimSitesListOKBody{Count: &emptyCount, Results: []*models.Site{site3Model}}
			site3ResponseData, _ = json.Marshal(site3Response)

			// devices-1
			deviceType = &models.NestedDeviceType{ID: 1, Model: &deviceModelName, Slug: &deviceModelName}
			device1Role = &models.NestedDeviceRole{ID: 1, Name: &deviceRoleName, Slug: &device1Name}
			device1Rack = &models.NestedRack{ID: 1, Name: &rackName}
			device1Site = &models.NestedSite{ID: 1, Name: &site1Name, Slug: &site1Name}
			device1Config = &models.DeviceWithConfigContext{ID: 1, DeviceRole: device1Role, DeviceType: deviceType, Rack: device1Rack, Site: device1Site, Name: &device1Name}
			devices1Response = &dcim.DcimDevicesListOKBody{Count: &count, Results: []*models.DeviceWithConfigContext{device1Config}}
			devices1ResponseData, _ = json.Marshal(devices1Response)
			// devices-5 ; device with prexisting custom fields
			deviceType = &models.NestedDeviceType{ID: 5, Model: &deviceModelName, Slug: &deviceModelName}
			device1Role = &models.NestedDeviceRole{ID: 5, Name: &deviceRoleName, Slug: &device5Name}
			device1Rack = &models.NestedRack{ID: 5, Name: &rackName}
			device1Site = &models.NestedSite{ID: 5, Name: &site1Name, Slug: &site1Name}
			device1Config = &models.DeviceWithConfigContext{ID: 5, DeviceRole: device1Role, DeviceType: deviceType, Rack: device1Rack, Site: device1Site,
				Name: &device5Name, CustomFields: &DeviceCustomFields{BMEnrollmentStatus: "enroll"}}
			devices1Response = &dcim.DcimDevicesListOKBody{Count: &count, Results: []*models.DeviceWithConfigContext{device1Config}}
			devices5ResponseData, _ = json.Marshal(devices1Response)

			// devices-2
			deviceType = &models.NestedDeviceType{ID: 2, Model: &deviceModelName, Slug: &deviceModelName}
			device2Role = &models.NestedDeviceRole{ID: 2, Name: &deviceRoleName, Slug: &device2Name}
			device2Rack = &models.NestedRack{ID: 2, Name: &rackName}
			device2Site = &models.NestedSite{ID: 2, Name: &site1Name, Slug: &site2Name}
			device2Config = &models.DeviceWithConfigContext{ID: 2, DeviceRole: device2Role, DeviceType: deviceType, Rack: device2Rack, Site: device2Site, Name: &device2Name}
			devices2Response = &dcim.DcimDevicesListOKBody{Count: &emptyCount, Results: []*models.DeviceWithConfigContext{device2Config}}
			devices2ResponseData, _ = json.Marshal(devices2Response)

			//device PUT response
			devicePutResponseData, _ = json.Marshal(device1Config)

			// Cluster1 data
			clusterType = &models.NestedClusterType{ID: 1, Name: &clusterTypeName, Slug: &clusterTypeName}
			site4 = &models.NestedSite{ID: 1, Name: &site1Name}
			cluster1 = &models.Cluster{ID: 1, Name: &cluster1Name, Type: clusterType, Site: site4, DeviceCount: 5}
			cluster1Response = &virtualization.VirtualizationClustersListOKBody{Count: &count, Results: []*models.Cluster{cluster1}}
			cluster1ResponseData, _ = json.Marshal(cluster1Response)

			// Cluster2 data - result count == 0
			cluster2 = &models.Cluster{ID: 2, Name: &cluster2Name, Type: clusterType, Site: site4, DeviceCount: 5}
			cluster2Response = &virtualization.VirtualizationClustersListOKBody{Count: &emptyCount, Results: []*models.Cluster{cluster2}}
			cluster2ResponseData, _ = json.Marshal(cluster2Response)

			// Cluster3 data - result count > 1
			cluster3 = &models.Cluster{ID: 3, Name: &cluster3Name, Type: clusterType, Site: site4, DeviceCount: 5}
			cluster3Response = &virtualization.VirtualizationClustersListOKBody{Count: &excessCount, Results: []*models.Cluster{cluster3}}
			cluster3ResponseData, _ = json.Marshal(cluster3Response)

			// Cluster4 data - device count == 1
			cluster4 = &models.Cluster{ID: 4, Name: &cluster4Name, Type: clusterType, Site: site4, DeviceCount: 0}
			cluster4Response = &virtualization.VirtualizationClustersListOKBody{Count: &count, Results: []*models.Cluster{cluster4}}
			cluster4ResponseData, _ = json.Marshal(cluster4Response)

			// http test server
			server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("Content-Type", "application/json")
				// interfaces response
				if r.URL.Path == "/api/dcim/interfaces/" {
					if (r.URL.Query().Get("device") == "device-1") && (r.URL.Query().Get("name") == "BMC") {
						w.WriteHeader(http.StatusOK)
						w.Write(interface1ResponseData)
					} else if (r.URL.Query().Get("device") == "device-5") && (r.URL.Query().Get("name") == "BMC") {
						w.WriteHeader(http.StatusOK)
						w.Write(interface1ResponseData)
					} else if (r.URL.Query().Get("device") == "device-2") && (r.URL.Query().Get("name") == "BMC") {
						w.WriteHeader(http.StatusOK)
						w.Write(interface2ResponseData)
					} else if (r.URL.Query().Get("device") == "device-3") && (r.URL.Query().Get("name") == "BMC") {
						w.WriteHeader(http.StatusOK)
						w.Write(interface3ResponseData)
					} else if (r.URL.Query().Get("device") == "device-4") && (r.URL.Query().Get("name") == "BMC") {
						w.WriteHeader(http.StatusOK)
						w.Write(interface4ResponseData)
					} else {
						w.WriteHeader(http.StatusNotFound)
					}
				} else if r.URL.Path == "/api/dcim/sites/" {
					if r.URL.Query().Get("name") == "us-dev-1a" {
						w.WriteHeader(http.StatusOK)
						w.Write(site1ResponseData)
					} else if r.URL.Query().Get("name") == "us-dev-1b" {
						w.WriteHeader(http.StatusOK)
						w.Write(site2ResponseData)
					} else if r.URL.Query().Get("name") == "us-dev-1c" {
						w.WriteHeader(http.StatusOK)
						w.Write(site3ResponseData)
					} else {
						w.WriteHeader(http.StatusNotFound)
					}
				} else if r.URL.Path == "/api/dcim/devices/1/" {
					w.WriteHeader(http.StatusOK)
					w.Write(devicePutResponseData)
					// deviceID 2 returns 404 when trying update device
				} else if r.URL.Path == "/api/dcim/devices/2/" {
					w.WriteHeader(http.StatusNotFound)
					w.Write(devicePutResponseData)
					// deviceID 3 returns 500 response code when trying update device
				} else if r.URL.Path == "/api/dcim/devices/3/" {
					w.WriteHeader(http.StatusInternalServerError)
				} else if r.URL.Path == "/api/dcim/devices/" {
					if r.URL.Query().Get("name") == "device-1" {
						w.WriteHeader(http.StatusOK)
						w.Write(devices1ResponseData)
					} else if r.URL.Query().Get("name") == "device-2" {
						w.WriteHeader(http.StatusOK)
						w.Write(devices2ResponseData)
					} else if r.URL.Query().Get("name") == "device-5" {
						w.WriteHeader(http.StatusOK)
						w.Write(devices5ResponseData)
					} else {
						w.WriteHeader(http.StatusNotFound)
					}
				} else if r.URL.Path == "/api/virtualization/clusters/" {
					if r.URL.Query().Get("name") == "1" {
						w.WriteHeader(http.StatusOK)
						w.Write(cluster1ResponseData)
					} else if r.URL.Query().Get("name") == "2" {
						w.WriteHeader(http.StatusOK)
						w.Write(cluster2ResponseData)
					} else if r.URL.Query().Get("name") == "3" {
						w.WriteHeader(http.StatusOK)
						w.Write(cluster3ResponseData)
					} else if r.URL.Query().Get("name") == "4" {
						w.WriteHeader(http.StatusOK)
						w.Write(cluster4ResponseData)
					} else {
						w.WriteHeader(http.StatusNotFound)
					}
				}
			}))

			netBoxURL, _ = url.Parse(server.URL)
			netBoxHost = netBoxURL.Hostname() + ":" + netBoxURL.Port()

			ctx = context.Background()
			token = "123456789"

			Expect(os.Setenv("NETBOX_HOST", netBoxHost)).To(Succeed())
			netBox, _ = NewNetBoxClient(ctx, token, false)

		})
		AfterAll(func() {
			defer server.Close()
		})

		It("Expect GetBMCURL to return a valid URL", func() {
			bmcURL, err := netBox.GetBMCURL(ctx, device1Name)
			fmt.Printf("bmcURL:%v error: %+v\n", bmcURL, err)
			Expect(bmcURL).To(Equal(bmcURL))
			Expect(err).NotTo(HaveOccurred())
		})
		It("Expect GetBMCURL to return an error because of empty URL", func() {
			bmcURL, err := netBox.GetBMCURL(ctx, device2Name)
			fmt.Printf("bmcURL:%v error: %+v\n", bmcURL, err)
			Expect(bmcURL).To(Equal(""))
			Expect(err).Should(HaveOccurred())
		})
		It("Expect GetBMCURL to return an error because of empty interface list", func() {
			bmcURL, err := netBox.GetBMCURL(ctx, device3Name)
			fmt.Printf("bmcURL:%v error: %+v\n", bmcURL, err)
			Expect(bmcURL).To(Equal(""))
			Expect(err).Should(HaveOccurred())
		})
		It("Expect GetBMCURL to return an error because get interface list query failed", func() {
			bmcURL, err := netBox.GetBMCURL(ctx, "InvalidDevice")
			fmt.Printf("bmcURL:%v error: %+v\n", bmcURL, err)
			Expect(bmcURL).To(Equal(""))
			Expect(err).Should(HaveOccurred())
		})
		It("Expect GetBMCMACAddress to return a valid MacAddress", func() {
			macAddr, err := netBox.GetBMCMACAddress(ctx, device1Name, interfaceName)
			fmt.Printf("macAddr:%v error: %+v\n", macAddr, err)
			Expect(macAddr).To(Equal(macAddress))
			Expect(err).NotTo(HaveOccurred())
		})
		It("Expect GetBMCMACAddress to return an error because of an invalid MacAddress", func() {
			macAddr, err := netBox.GetBMCMACAddress(ctx, device4Name, interfaceName)
			fmt.Printf("macAddr:%v error: %+v\n", macAddr, err)
			Expect(macAddr).To(Equal(""))
			Expect(err).Should(HaveOccurred())
		})
		It("Expect GetBMCMACAddress to return an error because of empty MacAddress", func() {
			macAddr, err := netBox.GetBMCMACAddress(ctx, device2Name, interfaceName)
			fmt.Printf("macAddr:%v error: %+v\n", macAddr, err)
			Expect(macAddr).To(Equal(""))
			Expect(err).Should(HaveOccurred())
		})
		It("Expect GetBMCMACAddress to return an error because of empty interface list", func() {
			macAddr, err := netBox.GetBMCMACAddress(ctx, device3Name, interfaceName)
			fmt.Printf("macAddr:%v error: %+v\n", macAddr, err)
			Expect(macAddr).To(Equal(""))
			Expect(err).Should(HaveOccurred())
		})
		It("Expect GetBMCMACAddress to return an error because get interface list query failed", func() {
			macAddr, err := netBox.GetBMCMACAddress(ctx, "InvalidDevice", interfaceName)
			fmt.Printf("macAddr:%v error: %+v\n", macAddr, err)
			Expect(macAddr).To(Equal(""))
			Expect(err).Should(HaveOccurred())
		})
		It("Expect GetRegionName to return a valid region", func() {
			region, err := netBox.GetDeviceRegionName(ctx, site1Name)
			fmt.Printf("region:%v error: %+v\n", region, err)
			Expect(region).To(Equal(site1Name))
			Expect(err).NotTo(HaveOccurred())
		})
		It("Expect GetRegionName to return an error because of an empty region", func() {
			region, err := netBox.GetDeviceRegionName(ctx, site2Name)
			fmt.Printf("region:%v error: %+v\n", region, err)
			Expect(region).To(Equal(""))
			Expect(err).Should(HaveOccurred())
		})
		It("Expect GetRegionName to return an error because of empty region list", func() {
			region, err := netBox.GetDeviceRegionName(ctx, site3Name)
			fmt.Printf("region:%v error: %+v\n", region, err)
			Expect(region).To(Equal(""))
			Expect(err).Should(HaveOccurred())
		})
		It("Expect GetRegionName to return an error because get interface list query failed", func() {
			region, err := netBox.GetDeviceRegionName(ctx, "InvalidSite")
			fmt.Printf("region:%v error: %+v\n", region, err)
			Expect(region).To(Equal(""))
			Expect(err).Should(HaveOccurred())
		})
		It("Expect GetClusterSize to return expected expected size", func() {
			clusterSize, err := netBox.GetClusterSize(ctx, cluster1Name, site1Name)
			fmt.Printf("clusterSize:%v error: %+v\n", clusterSize, err)
			Expect(clusterSize).To(Equal(int64(5)))
			Expect(err).NotTo(HaveOccurred())
		})
		It("Expect GetClusterSize to return error when using an invalid cluster name", func() {
			clusterSize, err := netBox.GetClusterSize(ctx, "test", site1Name)
			fmt.Printf("clusterSize:%v error: %+v\n", clusterSize, err)
			Expect(clusterSize).To(Equal(int64(0)))
			Expect(err).Should(HaveOccurred())
		})
		It("Expect GetClusterSize to return error when the cluster result count is 0", func() {
			clusterSize, err := netBox.GetClusterSize(ctx, "2", site1Name)
			fmt.Printf("clusterSize:%v error: %+v\n", clusterSize, err)
			Expect(clusterSize).To(Equal(int64(0)))
			Expect(err).Should(HaveOccurred())
		})
		It("Expect GetClusterSize to return error when the cluster result count is greater than 1", func() {
			clusterSize, err := netBox.GetClusterSize(ctx, "3", site1Name)
			fmt.Printf("clusterSize:%v error: %+v\n", clusterSize, err)
			Expect(clusterSize).To(Equal(int64(0)))
			Expect(err).Should(HaveOccurred())
		})
		It("Expect GetClusterSize to return error when device count is 0 in the cluster", func() {
			clusterSize, err := netBox.GetClusterSize(ctx, "4", site1Name)
			fmt.Printf("clusterSize:%v error: %+v\n", clusterSize, err)
			Expect(clusterSize).To(Equal(int64(0)))
			Expect(err).Should(HaveOccurred())
		})
		It("Expect UpdateDeviceStatus to return no errors", func() {
			err := netBox.UpdateDeviceCustomFields(ctx, device1Name, 1, &DeviceCustomFields{
				BMEnrollmentStatus:  BMEnrolled,
				BMEnrollmentComment: "Enrollment is complete",
			})
			fmt.Printf("error: %+v\n", err)
			Expect(err).NotTo(HaveOccurred())
		})
		It("Expect UpdateDeviceStatus to return no errors", func() {
			actualDeviceId, err := netBox.GetDeviceId(ctx, device5Name)
			Expect(err).NotTo(HaveOccurred())
			Expect(actualDeviceId).To(Equal(int64(5)))
			// update custom field with pre-existing custom field
			err = netBox.UpdateDeviceCustomFields(ctx, device5Name, 5, &DeviceCustomFields{
				BMEnrollmentComment: "Enrollment is complete",
			})
			fmt.Printf("error: %+v\n", err)
			Expect(err).NotTo(HaveOccurred())
		})
		It("Expect UpdateDeviceStatus to return an error as device list is empty", func() {
			err := netBox.UpdateDeviceCustomFields(ctx, device2Name, 1, &DeviceCustomFields{
				BMEnrollmentStatus:  BMDisenrollmentFailed,
				BMEnrollmentComment: "return count==0",
			})
			fmt.Printf("error: %+v\n", err)
			Expect(err).Should(HaveOccurred())
		})
		It("Expect UpdateDeviceStatus to return an error because get devices query failed", func() {
			err := netBox.UpdateDeviceCustomFields(ctx, "InvalidDevice", 1, &DeviceCustomFields{
				BMEnrollmentStatus:  BMDisenrollmentFailed,
				BMEnrollmentComment: "rgot 404",
			})
			fmt.Printf("error: %+v\n", err)
			Expect(err).Should(HaveOccurred())
		})
		It("Expect UpdateDeviceStatus to return an error as device update failed", func() {
			err := netBox.UpdateDeviceCustomFields(ctx, device1Name, 2, &DeviceCustomFields{
				BMEnrollmentStatus:  BMDisenrollmentFailed,
				BMEnrollmentComment: "got 404 with data",
			})
			fmt.Printf("error: %+v\n", err)
			Expect(err).Should(HaveOccurred())
		})
		It("Expect UpdateDeviceStatus to return an error as device update failed with internal server error", func() {
			err := netBox.UpdateDeviceCustomFields(ctx, device1Name, 3, &DeviceCustomFields{
				BMEnrollmentStatus:  BMDisenrollmentFailed,
				BMEnrollmentComment: "got 500 internal server error",
			})
			fmt.Printf("error: %+v\n", err)
			Expect(err).Should(HaveOccurred())
		})
		It("Expect NewNetBoxClient to return an error as netbox host not present", func() {
			Expect(os.Setenv("NETBOX_HOST", "")).To(Succeed())
			netBox, err := NewNetBoxClient(ctx, token, false)
			fmt.Printf("netbox: %+v\n error: %+v\n", netBox, err)
			Expect(err).Should(HaveOccurred())
		})
	})
})
