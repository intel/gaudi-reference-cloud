// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// Copyright © 2023 Intel Corporation
package bmc_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/bmc"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/mocks"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/mygofish"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stmcginnis/gofish/common"
	"github.com/stmcginnis/gofish/redfish"
)

func TestBMC(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "BMC Interface Suite")
}

var Any = gomock.Any()

var _ = Describe("Baseboard Management Controller", func() {
	const (
		testRegion  = "us-dev-1"
		bmcURL      = "https://127.0.0.1"
		bmcUsername = "bmcuser"
		bmcPassword = "bmcpass"
	)

	var (
		ctx context.Context

		mockCtrl      *gomock.Controller
		bmcManager    *mocks.MockBMCInterface
		gofishManager *mocks.MockGoFishManagerAccessor
	)

	BeforeEach(func() {
		ctx = context.Background()

		// create mock objects
		mockCtrl = gomock.NewController(GinkgoT())
		bmcManager = mocks.NewMockBMCInterface(mockCtrl)
		gofishManager = mocks.NewMockGoFishManagerAccessor(mockCtrl)
	})

	Describe("Testing BMC Interface ", func() {
		It("should initialize dependencies", func() {
			By("initializing interface")
			Expect(bmcManager).NotTo(BeNil())
			Expect(gofishManager).NotTo(BeNil())
		})

		It("should return BMC interface - virtual", func() {
			client := mocks.NewMockGoFishClientAccessor(mockCtrl)
			service := mocks.NewMockGoFishServiceAccessor(mockCtrl)
			client.EXPECT().GetService().AnyTimes().Return(service)
			gofishManager.EXPECT().Connect(Any).AnyTimes().Return(client, nil)

			systems := make([]mygofish.GoFishComputerSystemAccessor, 1)
			comSys := mocks.NewMockGoFishComputerSystemAccessor(mockCtrl)
			systems[0] = comSys

			service.EXPECT().Systems().AnyTimes().Return(systems, nil)
			comSys.EXPECT().ODataID().AnyTimes().Return("SushyODataID")
			comSys.EXPECT().Manufacturer().AnyTimes().Return(bmc.SushyEmulator)
			comSys.EXPECT().Model().AnyTimes().Return("sushy")
			comSys.EXPECT().PowerState().AnyTimes().Return(redfish.OnPowerState)

			bmc, err := bmc.New(
				gofishManager,
				&bmc.Config{
					URL:      bmcURL,
					Username: bmcUsername,
					Password: bmcPassword,
				})
			Expect(bmc).ToNot(BeNil())
			Expect(err).NotTo(HaveOccurred())

			Expect(bmc.IsVirtual()).To(BeTrue())

			err = bmc.SanitizeBMCBootOrder(ctx)
			Expect(err).NotTo(HaveOccurred())

			cnt, gpuType, err := bmc.GPUDiscovery(ctx)
			Expect(cnt).To(BeNumerically("==", 0))
			Expect(gpuType).To(BeComparableTo(""))
			Expect(err).NotTo(HaveOccurred())

			hbm, err := bmc.HBMDiscovery(ctx)
			Expect(hbm).To(BeComparableTo(""))
			Expect(err).NotTo(HaveOccurred())

			err = bmc.VerifyPlatformFirmwareResilience(ctx)
			Expect(err).NotTo(HaveOccurred())

			power, err := bmc.GetBMCPowerState(ctx)
			Expect(power).NotTo(BeComparableTo(""))
			Expect(err).NotTo(HaveOccurred())

			comSys.EXPECT().Reset(Any).AnyTimes().Return(nil)
			err = bmc.PowerOnBMC(ctx)
			Expect(err).NotTo(HaveOccurred())

			err = bmc.ConfigureNTP(ctx)
			Expect(err).NotTo(HaveOccurred())

			myStatus := common.Status{
				State: common.EnabledState,
			}
			myProc0 := &redfish.Processor{
				ProcessorType:     redfish.CPUProcessorType,
				Manufacturer:      "Intel",
				Status:            myStatus,
				TotalCores:        2,
				TotalEnabledCores: 2,
				TotalThreads:      4,
			}
			myProc1 := &redfish.Processor{
				ProcessorType:     redfish.CPUProcessorType,
				Manufacturer:      "Intel",
				Status:            myStatus,
				TotalCores:        2,
				TotalEnabledCores: 2,
				TotalThreads:      4,
			}
			myProcessors := make([]*redfish.Processor, 2)
			myProcessors[0] = myProc0
			myProcessors[1] = myProc1

			comSys.EXPECT().Processors().Return(myProcessors, nil)

			CPUInfo, err := bmc.GetHostCPU(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(CPUInfo).NotTo(BeNil())

			var id0, id1 redfish.ProcessorID
			id0.IdentificationRegisters = "123-45678"
			id1.IdentificationRegisters = "0x1234"
			myProc0.ProcessorID = id0
			myProc1.ProcessorID = id1
			comSys.EXPECT().Processors().Return(myProcessors, nil)

			CPUInfo, err = bmc.GetHostCPU(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(CPUInfo).NotTo(BeNil())

			myProc0.ProcessorType = redfish.AcceleratorProcessorType
			myProc1.Status.State = common.DisabledState
			comSys.EXPECT().Processors().Return(myProcessors, nil)

			CPUInfo, err = bmc.GetHostCPU(ctx)
			Expect(err).To(HaveOccurred())
			Expect(CPUInfo).To(BeNil())

			noProcessors := make([]*redfish.Processor, 1)
			comSys.EXPECT().Processors().Return(noProcessors, fmt.Errorf("NOPROC"))

			CPUInfo, err = bmc.GetHostCPU(ctx)
			Expect(err).To(HaveOccurred())
			Expect(CPUInfo).To(BeNil())

			noProcessors = make([]*redfish.Processor, 0)
			comSys.EXPECT().Processors().Return(noProcessors, nil)

			CPUInfo, err = bmc.GetHostCPU(ctx)
			Expect(err).To(HaveOccurred())
			Expect(CPUInfo).To(BeNil())
		})

		It("Virtial - Fails for Systems() error", func() {
			client := mocks.NewMockGoFishClientAccessor(mockCtrl)
			service := mocks.NewMockGoFishServiceAccessor(mockCtrl)
			client.EXPECT().GetService().AnyTimes().Return(service)
			gofishManager.EXPECT().Connect(Any).AnyTimes().Return(client, nil)

			systems := make([]mygofish.GoFishComputerSystemAccessor, 1)
			comSys := mocks.NewMockGoFishComputerSystemAccessor(mockCtrl)
			systems[0] = comSys

			service.EXPECT().Systems().Return(systems, nil)
			comSys.EXPECT().ODataID().AnyTimes().Return("SushyODataID")
			comSys.EXPECT().Manufacturer().AnyTimes().Return(bmc.SushyEmulator)
			comSys.EXPECT().Model().AnyTimes().Return("sushy")
			comSys.EXPECT().PowerState().AnyTimes().Return(redfish.OnPowerState)

			bmc, err := bmc.New(
				gofishManager,
				&bmc.Config{
					URL:      bmcURL,
					Username: bmcUsername,
					Password: bmcPassword,
				})
			Expect(bmc).ToNot(BeNil())
			Expect(err).NotTo(HaveOccurred())

			service.EXPECT().Systems().Return(systems, fmt.Errorf("No Systems"))
			CPUInfo, err := bmc.GetHostCPU(ctx)
			Expect(err).To(HaveOccurred())
			Expect(CPUInfo).To(BeNil())
		})

		It("Virtual - GetHostMACAddress", func() {
			client := mocks.NewMockGoFishClientAccessor(mockCtrl)
			service := mocks.NewMockGoFishServiceAccessor(mockCtrl)
			client.EXPECT().GetService().AnyTimes().Return(service)
			gofishManager.EXPECT().Connect(Any).AnyTimes().Return(client, nil)

			systems := make([]mygofish.GoFishComputerSystemAccessor, 1)
			comSys := mocks.NewMockGoFishComputerSystemAccessor(mockCtrl)
			systems[0] = comSys

			service.EXPECT().Systems().AnyTimes().Return(systems, nil)
			comSys.EXPECT().ODataID().AnyTimes().Return("SushyODataID")
			comSys.EXPECT().Manufacturer().AnyTimes().Return(bmc.SushyEmulator)
			comSys.EXPECT().Model().AnyTimes().Return("sushy")
			comSys.EXPECT().PowerState().AnyTimes().Return(redfish.OnPowerState)

			bmc, err := bmc.New(
				gofishManager,
				&bmc.Config{
					URL:      bmcURL,
					Username: bmcUsername,
					Password: bmcPassword,
				})
			Expect(bmc).ToNot(BeNil())
			Expect(err).NotTo(HaveOccurred())

			Expect(bmc.IsVirtual()).To(BeTrue())

			myStatus := common.Status{
				State:  common.EnabledState,
				Health: common.OKHealth,
			}
			myEth0 := &redfish.EthernetInterface{
				Status:     myStatus,
				LinkStatus: redfish.LinkUpLinkStatus,
				MACAddress: "00:00:ac:ed:00:00",
			}
			myEth1 := &redfish.EthernetInterface{
				Status:     myStatus,
				LinkStatus: redfish.LinkUpLinkStatus,
				MACAddress: "00:00:de:ad:00:00",
			}
			myEthers := make([]*redfish.EthernetInterface, 2)
			myEthers[0] = myEth0
			myEthers[1] = myEth1

			comSys.EXPECT().EthernetInterfaces().AnyTimes().Return(myEthers, nil)

			mac, err := bmc.GetHostMACAddress(ctx)
			Expect(mac).ToNot(BeNil())
			Expect(err).NotTo(HaveOccurred())

			err = bmc.SanitizeBMCBootOrder(ctx)
			Expect(err).NotTo(HaveOccurred())
		})

		It("BMC interface - virtual - Power Tests", func() {
			client := mocks.NewMockGoFishClientAccessor(mockCtrl)
			service := mocks.NewMockGoFishServiceAccessor(mockCtrl)
			client.EXPECT().GetService().AnyTimes().Return(service)
			gofishManager.EXPECT().Connect(Any).AnyTimes().Return(client, nil)

			systems := make([]mygofish.GoFishComputerSystemAccessor, 1)
			comSys := mocks.NewMockGoFishComputerSystemAccessor(mockCtrl)
			systems[0] = comSys
			comSys.EXPECT().ODataID().AnyTimes().Return("SushyODataID")
			comSys.EXPECT().Manufacturer().AnyTimes().Return(bmc.SushyEmulator)
			comSys.EXPECT().Model().AnyTimes().Return("sushy")

			service.EXPECT().Systems().AnyTimes().Return(systems, nil)

			bmc, err := bmc.New(
				gofishManager,
				&bmc.Config{
					URL:      bmcURL,
					Username: bmcUsername,
					Password: bmcPassword,
				})
			Expect(bmc).ToNot(BeNil())
			Expect(err).NotTo(HaveOccurred())

			Expect(bmc.IsVirtual()).To(BeTrue())

			comSys.EXPECT().PowerState().Return(redfish.OffPowerState)
			err = bmc.PowerOffBMC(ctx)
			Expect(err).NotTo(HaveOccurred())
		})

		It("BMC interface - virtual - Fail 1 Power Tests", func() {
			client := mocks.NewMockGoFishClientAccessor(mockCtrl)
			service := mocks.NewMockGoFishServiceAccessor(mockCtrl)
			client.EXPECT().GetService().AnyTimes().Return(service)
			gofishManager.EXPECT().Connect(Any).AnyTimes().Return(client, nil)

			systems := make([]mygofish.GoFishComputerSystemAccessor, 1)
			comSys := mocks.NewMockGoFishComputerSystemAccessor(mockCtrl)
			systems[0] = comSys

			service.EXPECT().Systems().Return(nil, nil)
			badBmc, err := bmc.New(
				gofishManager,
				&bmc.Config{
					URL:      bmcURL,
					Username: bmcUsername,
					Password: bmcPassword,
				})
			Expect(badBmc).To(BeNil())
			Expect(err).To(HaveOccurred())

			comSys.EXPECT().Manufacturer().AnyTimes().Return("")
			comSys.EXPECT().Model().AnyTimes().Return("sushy")

			service.EXPECT().Systems().AnyTimes().Return(systems, nil)
			badBmc, err = bmc.New(
				gofishManager,
				&bmc.Config{
					URL:      bmcURL,
					Username: bmcUsername,
					Password: bmcPassword,
				})
			Expect(badBmc).To(BeNil())
			Expect(err).To(HaveOccurred())
		})

		It("BMC interface - virtual - Fail 2 Power Tests", func() {
			client := mocks.NewMockGoFishClientAccessor(mockCtrl)
			service := mocks.NewMockGoFishServiceAccessor(mockCtrl)
			client.EXPECT().GetService().AnyTimes().Return(service)
			gofishManager.EXPECT().Connect(Any).AnyTimes().Return(client, nil)

			systems := make([]mygofish.GoFishComputerSystemAccessor, 1)
			comSys := mocks.NewMockGoFishComputerSystemAccessor(mockCtrl)
			systems[0] = comSys

			comSys.EXPECT().Model().AnyTimes().Return("")
			comSys.EXPECT().Manufacturer().AnyTimes().Return("Not Sushy")
			service.EXPECT().Systems().AnyTimes().Return(systems, nil)
			badBmc, err := bmc.New(
				gofishManager,
				&bmc.Config{
					URL:      bmcURL,
					Username: bmcUsername,
					Password: bmcPassword,
				})
			Expect(badBmc).To(BeNil())
			Expect(err).To(HaveOccurred())
		})

		It("BMC interface - virtual - Create Account", func() {
			client := mocks.NewMockGoFishClientAccessor(mockCtrl)
			service := mocks.NewMockGoFishServiceAccessor(mockCtrl)
			client.EXPECT().GetService().Return(service)

			systems := make([]mygofish.GoFishComputerSystemAccessor, 1)
			comSys := mocks.NewMockGoFishComputerSystemAccessor(mockCtrl)
			systems[0] = comSys

			service.EXPECT().Systems().AnyTimes().Return(systems, nil)
			comSys.EXPECT().ODataID().AnyTimes().Return("SushyODataID")
			comSys.EXPECT().Manufacturer().AnyTimes().Return(bmc.SushyEmulator)
			comSys.EXPECT().Model().AnyTimes().Return("sushy")
			comSys.EXPECT().PowerState().AnyTimes().Return(redfish.OnPowerState)

			gofishManager.EXPECT().Connect(Any).AnyTimes().Return(client, nil)

			// Going to have to mock this as well to capture the ManagerAccount.Update() call
			account := &redfish.ManagerAccount{
				UserName: "newuser",
				Password: "newpass",
			}
			accounts := make([]*redfish.ManagerAccount, 1)
			accounts[0] = account
			mockAcctSrvs := mocks.NewMockGoFishAccountServiceAccessor(mockCtrl)
			mockAcctSrvs.EXPECT().Accounts().Return(accounts, nil)

			service.EXPECT().AccountService().Return(mockAcctSrvs, nil)

			bmc, err := bmc.New(
				gofishManager,
				&bmc.Config{
					URL:      bmcURL,
					Username: bmcUsername,
					Password: bmcPassword,
				})
			Expect(bmc).ToNot(BeNil())
			Expect(err).NotTo(HaveOccurred())

			Expect(bmc.IsVirtual()).To(BeTrue())

			// client.EXPECT().GetService().Return(service)
			// err = bmc.UpdateAccount(ctx, "newuser", "upd_pass")
			// Expect(err).NotTo(HaveOccurred())

			// check for account service returns an error
			service.EXPECT().AccountService().Return(mockAcctSrvs, fmt.Errorf("account service error"))
			client.EXPECT().GetService().Return(service)
			err = bmc.UpdateAccount(ctx, "newuser", "upd_pass")
			Expect(err).To(HaveOccurred())

			// check for accounts returns an erorro
			service.EXPECT().AccountService().Return(mockAcctSrvs, nil)
			mockAcctSrvs.EXPECT().Accounts().Return(accounts, fmt.Errorf("Accounts error"))
			client.EXPECT().GetService().Return(service)
			err = bmc.UpdateAccount(ctx, "newuser", "upd_pass")
			Expect(err).To(HaveOccurred())

			// check for account not found
			client.EXPECT().GetService().Return(service)
			err = bmc.UpdateAccount(ctx, "other", "upd_pass")
			Expect(err).To(HaveOccurred())

			// check for GetService returns nil
			client.EXPECT().GetService().Return(nil)
			err = bmc.UpdateAccount(ctx, "newuser", "upd_pass")
			Expect(err).To(HaveOccurred())
		})

		It("should return BMC interface - denali pass", func() {
			client := mocks.NewMockGoFishClientAccessor(mockCtrl)
			service := mocks.NewMockGoFishServiceAccessor(mockCtrl)
			client.EXPECT().GetService().AnyTimes().Return(service)

			systems := make([]mygofish.GoFishComputerSystemAccessor, 1)
			comSys := mocks.NewMockGoFishComputerSystemAccessor(mockCtrl)
			systems[0] = comSys

			service.EXPECT().Systems().AnyTimes().Return(systems, nil)
			comSys.EXPECT().ODataID().AnyTimes().Return("DenaliODataID")
			comSys.EXPECT().Manufacturer().AnyTimes().Return(bmc.IntelCorporation)
			comSys.EXPECT().Model().AnyTimes().Return("D50DNP") // Denali pass
			comSys.EXPECT().PowerState().AnyTimes().Return(redfish.OnPowerState)

			gofishManager.EXPECT().Connect(Any).AnyTimes().Return(client, nil)
			bmc, err := bmc.New(
				gofishManager,
				&bmc.Config{
					URL:      bmcURL,
					Username: bmcUsername,
					Password: bmcPassword,
				})
			Expect(bmc).ToNot(BeNil())
			Expect(err).NotTo(HaveOccurred())

			Expect(bmc.IsVirtual()).To(BeFalse())

			addr, err := bmc.GetHostBMCAddress()
			Expect(addr).ToNot(BeNil())
			Expect(err).NotTo(HaveOccurred())

			response := &http.Response{StatusCode: 200}
			// {"NTP":{"ProtocolEnabled": true}}
			enablePayload := map[string]interface{}{
				"NTP": map[string]bool{
					"ProtocolEnabled": true},
			}
			client.EXPECT().Patch("/redfish/v1/Managers/bmc/NetworkProtocol", Any).Return(response, nil)
			client.EXPECT().Patch("/redfish/v1/Managers/bmc/NetworkProtocol", enablePayload).Return(response, nil)
			err = bmc.ConfigureNTP(ctx)
			Expect(err).NotTo(HaveOccurred())

			response = &http.Response{StatusCode: 500}
			client.EXPECT().Patch("/redfish/v1/Managers/bmc/NetworkProtocol", Any).Return(response, fmt.Errorf("set servers fails"))
			err = bmc.ConfigureNTP(ctx)
			Expect(err).To(HaveOccurred())

			response = &http.Response{StatusCode: 405}
			client.EXPECT().Patch("/redfish/v1/Managers/bmc/NetworkProtocol", Any).Return(response, nil)
			err = bmc.ConfigureNTP(ctx)
			Expect(err).To(HaveOccurred())

			response = &http.Response{StatusCode: 200}
			client.EXPECT().Patch("/redfish/v1/Managers/bmc/NetworkProtocol", Any).Return(response, nil)
			response = &http.Response{StatusCode: 500}
			client.EXPECT().Patch("/redfish/v1/Managers/bmc/NetworkProtocol", enablePayload).Return(response, fmt.Errorf("set servers fails"))
			err = bmc.ConfigureNTP(ctx)
			Expect(err).To(HaveOccurred())

			response = &http.Response{StatusCode: 200}
			client.EXPECT().Patch("/redfish/v1/Managers/bmc/NetworkProtocol", Any).Return(response, nil)
			response = &http.Response{StatusCode: 405}
			client.EXPECT().Patch("/redfish/v1/Managers/bmc/NetworkProtocol", enablePayload).Return(response, nil)
			err = bmc.ConfigureNTP(ctx)
			Expect(err).To(HaveOccurred())
		})

		It("Denali Pass - Boot Order", func() {
			client := mocks.NewMockGoFishClientAccessor(mockCtrl)
			service := mocks.NewMockGoFishServiceAccessor(mockCtrl)
			client.EXPECT().GetService().AnyTimes().Return(service)
			gofishManager.EXPECT().Connect(Any).AnyTimes().Return(client, nil)

			systems := make([]mygofish.GoFishComputerSystemAccessor, 1)
			comSys := mocks.NewMockGoFishComputerSystemAccessor(mockCtrl)
			systems[0] = comSys

			service.EXPECT().Systems().AnyTimes().Return(systems, nil)
			comSys.EXPECT().ODataID().AnyTimes().Return("DenaliODataID")
			comSys.EXPECT().Manufacturer().AnyTimes().Return(bmc.IntelCorporation)
			comSys.EXPECT().Model().AnyTimes().Return("D50DNP") // Denali pass
			comSys.EXPECT().PowerState().AnyTimes().Return(redfish.OnPowerState)

			myBmc, err := bmc.New(
				gofishManager,
				&bmc.Config{
					URL:      bmcURL,
					Username: bmcUsername,
					Password: bmcPassword,
				})
			Expect(myBmc).ToNot(BeNil())
			Expect(err).NotTo(HaveOccurred())

			Expect(myBmc.IsVirtual()).To(BeFalse())

			myBoot := mocks.NewMockGoFishBootAccessor(mockCtrl)
			comSys.EXPECT().Boot().AnyTimes().Return(myBoot)
			comSys.EXPECT().SetBoot(Any).AnyTimes().Return(nil)

			boot0 := fmt.Sprintf("%s%s", bmc.DenaliPassBootIntelRegex, bmc.DenaliBaseboardCard)
			boot1 := fmt.Sprintf("%s%s", bmc.DenaliPassBootRegex, bmc.DenaliRiserCard)
			boot2 := fmt.Sprintf("%s%s", bmc.DenaliPassBootIntelRegex, bmc.DenaliRiserCard)
			myBootOrder := []string{boot0, boot1, boot2}

			myBoot.EXPECT().BootOrder().AnyTimes().Return(myBootOrder)
			myBoot.EXPECT().SetBootOrder(Any).AnyTimes().Return(nil)
			err = myBmc.SanitizeBMCBootOrder(ctx)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Denali Pass - PFR", func() {
			client := mocks.NewMockGoFishClientAccessor(mockCtrl)
			service := mocks.NewMockGoFishServiceAccessor(mockCtrl)
			client.EXPECT().GetService().AnyTimes().Return(service)
			gofishManager.EXPECT().Connect(Any).AnyTimes().Return(client, nil)

			systems := make([]mygofish.GoFishComputerSystemAccessor, 1)
			comSys := mocks.NewMockGoFishComputerSystemAccessor(mockCtrl)
			systems[0] = comSys

			service.EXPECT().Systems().Return(systems, nil)
			comSys.EXPECT().ODataID().AnyTimes().Return("DenaliODataID")
			comSys.EXPECT().Manufacturer().AnyTimes().Return(bmc.IntelCorporation)
			comSys.EXPECT().Model().AnyTimes().Return("D50DNP") // Denali pass
			comSys.EXPECT().PowerState().AnyTimes().Return(redfish.OnPowerState)

			myBmc, err := bmc.New(
				gofishManager,
				&bmc.Config{
					URL:      bmcURL,
					Username: bmcUsername,
					Password: bmcPassword,
				})
			Expect(myBmc).ToNot(BeNil())
			Expect(err).NotTo(HaveOccurred())

			Expect(myBmc.IsVirtual()).To(BeFalse())

			// Create a sample response body
			var responseBody bmc.OpenBMCSystem

			responseBody.Oem.OpenBmc.FirmwareProvisioning.ProvisioningStatus = bmc.DenaliPFRAndLocked
			jsonResponse, err := json.Marshal(responseBody)
			Expect(err).NotTo(HaveOccurred())

			// Create a new HTTP response recorder
			recorder := httptest.NewRecorder()
			// Write the response body to the recorder
			recorder.Header().Set("Content-Type", "application/json")
			recorder.Write(jsonResponse)

			// Get the response from the recorder
			response := recorder.Result()
			// [DenaliODataID]
			client.EXPECT().Get("DenaliODataID").Return(response, nil)
			service.EXPECT().Systems().Return(systems, nil)
			err = myBmc.VerifyPlatformFirmwareResilience(ctx)
			Expect(err).NotTo(HaveOccurred())

			// Not Locked
			responseBody.Oem.OpenBmc.FirmwareProvisioning.ProvisioningStatus = bmc.DenaliPFRNotLocked
			recorder = httptest.NewRecorder()
			recorder.Header().Set("Content-Type", "application/json")
			jsonResponse, err = json.Marshal(responseBody)
			recorder.Write(jsonResponse)
			Expect(err).NotTo(HaveOccurred())

			// Get the response from the recorder
			response = recorder.Result()
			// [DenaliODataID]
			client.EXPECT().Get("DenaliODataID").Return(response, nil)
			service.EXPECT().Systems().Return(systems, nil)
			err = myBmc.VerifyPlatformFirmwareResilience(ctx)
			Expect(err).To(HaveOccurred())

			// Invalid Get() response
			recorder = httptest.NewRecorder()
			// Write the response body to the recorder
			recorder.Header().Set("Content-Type", "application/json")
			recorder.Body = nil
			// Get the response from the recorder
			response = recorder.Result()
			// [DenaliODataID]
			client.EXPECT().Get("DenaliODataID").Return(response, fmt.Errorf("Get fails"))
			service.EXPECT().Systems().Return(systems, nil)
			err = myBmc.VerifyPlatformFirmwareResilience(ctx)
			Expect(err).To(HaveOccurred())

			// Invalid BODY response
			recorder = httptest.NewRecorder()
			// Write the response body to the recorder
			recorder.Header().Set("Content-Type", "application/json")
			recorder.Body = nil
			// Get the response from the recorder
			response = recorder.Result()
			// [DenaliODataID]
			client.EXPECT().Get("DenaliODataID").Return(response, nil)
			service.EXPECT().Systems().Return(systems, nil)
			err = myBmc.VerifyPlatformFirmwareResilience(ctx)
			Expect(err).To(HaveOccurred())

			// Invalid JSON response
			recorder = httptest.NewRecorder()
			// Write the response body to the recorder
			recorder.Header().Set("Content-Type", "application/json")
			recorder.WriteString("fail")
			// Get the response from the recorder
			response = recorder.Result()
			// [DenaliODataID]
			client.EXPECT().Get("DenaliODataID").Return(response, nil)
			service.EXPECT().Systems().Return(systems, nil)
			err = myBmc.VerifyPlatformFirmwareResilience(ctx)
			Expect(err).To(HaveOccurred())

			// Fail case for Systems() error
			service.EXPECT().Systems().Return(systems, fmt.Errorf("No Systems"))
			err = myBmc.VerifyPlatformFirmwareResilience(ctx)
			Expect(err).To(HaveOccurred())

			// Fail case for Systems() is empty
			noSystems := make([]mygofish.GoFishComputerSystemAccessor, 0)
			service.EXPECT().Systems().Return(noSystems, nil)
			err = myBmc.VerifyPlatformFirmwareResilience(ctx)
			Expect(err).To(HaveOccurred())
		})

		It("Denali Pass - GPU Discovery", func() {
			client := mocks.NewMockGoFishClientAccessor(mockCtrl)
			service := mocks.NewMockGoFishServiceAccessor(mockCtrl)
			client.EXPECT().GetService().AnyTimes().Return(service)
			gofishManager.EXPECT().Connect(Any).AnyTimes().Return(client, nil)

			systems := make([]mygofish.GoFishComputerSystemAccessor, 1)
			comSys := mocks.NewMockGoFishComputerSystemAccessor(mockCtrl)
			systems[0] = comSys

			service.EXPECT().Systems().Return(systems, nil)
			comSys.EXPECT().ODataID().AnyTimes().Return("DenaliODataID")
			comSys.EXPECT().Manufacturer().AnyTimes().Return(bmc.IntelCorporation)
			comSys.EXPECT().Model().AnyTimes().Return("D50DNP") // Denali pass
			comSys.EXPECT().PowerState().AnyTimes().Return(redfish.OnPowerState)

			myBmc, err := bmc.New(
				gofishManager,
				&bmc.Config{
					URL:      bmcURL,
					Username: bmcUsername,
					Password: bmcPassword,
				})
			Expect(myBmc).ToNot(BeNil())
			Expect(err).NotTo(HaveOccurred())

			Expect(myBmc.IsVirtual()).To(BeFalse())

			myPci0 := &redfish.PCIeDevice{
				Description: "pci0",
			}
			myPcis := make([]*redfish.PCIeDevice, 1)
			myPcis[0] = myPci0

			// Create a sample response body
			var responseBody bmc.OpenBMCPcieFunction

			responseBody.VendorID = "0x8086"
			responseBody.DeviceID = "0x56c0"
			jsonResponse, err := json.Marshal(responseBody)
			Expect(err).NotTo(HaveOccurred())

			// Create a new HTTP response recorder
			recorder := httptest.NewRecorder()
			// Write the response body to the recorder
			recorder.Header().Set("Content-Type", "application/json")
			recorder.Write(jsonResponse)

			// Get the response from the recorder
			response := recorder.Result()
			// /PCIeFunctions/0
			client.EXPECT().Get("/PCIeFunctions/0").Return(response, nil)
			service.EXPECT().Systems().Return(systems, nil)
			comSys.EXPECT().PCIeDevices().Return(myPcis, nil)
			cnt, gpuType, err := myBmc.GPUDiscovery(ctx)
			Expect(cnt).To(BeNumerically("==", 1))
			Expect(gpuType).To(BeComparableTo("GPU-Flex-170"))
			Expect(err).NotTo(HaveOccurred())

			// Other GPU type
			responseBody.DeviceID = "0x0bda"
			jsonResponse, err = json.Marshal(responseBody)
			Expect(err).NotTo(HaveOccurred())
			recorder = httptest.NewRecorder()
			// Write the response body to the recorder
			recorder.Header().Set("Content-Type", "application/json")
			recorder.Write(jsonResponse)

			// Get the response from the recorder
			response = recorder.Result()
			// /PCIeFunctions/0
			client.EXPECT().Get("/PCIeFunctions/0").Return(response, nil)
			service.EXPECT().Systems().Return(systems, nil)
			comSys.EXPECT().PCIeDevices().Return(myPcis, nil)
			cnt, gpuType, err = myBmc.GPUDiscovery(ctx)
			Expect(cnt).To(BeNumerically("==", 1))
			Expect(gpuType).To(BeComparableTo("GPU-Max-1100"))
			Expect(err).NotTo(HaveOccurred())

			// PCIeDevices returns error
			// /PCIeFunctions/0
			service.EXPECT().Systems().Return(systems, nil)
			comSys.EXPECT().PCIeDevices().Return(myPcis, fmt.Errorf("PCI Devices Error"))
			cnt, gpuType, err = myBmc.GPUDiscovery(ctx)
			Expect(cnt).To(BeNumerically("==", 0))
			Expect(gpuType).To(BeComparableTo(""))
			Expect(err).To(HaveOccurred())

			// Invalid Get() response
			recorder = httptest.NewRecorder()
			// Write the response body to the recorder
			recorder.Header().Set("Content-Type", "application/json")
			recorder.Body = nil
			// Get the response from the recorder
			response = recorder.Result()
			client.EXPECT().Get("/PCIeFunctions/0").Return(response, fmt.Errorf("Get fails"))
			service.EXPECT().Systems().Return(systems, nil)
			comSys.EXPECT().PCIeDevices().Return(myPcis, nil)
			cnt, gpuType, err = myBmc.GPUDiscovery(ctx)
			Expect(cnt).To(BeNumerically("==", 0))
			Expect(gpuType).To(BeComparableTo(""))
			Expect(err).To(HaveOccurred())

			// Invalid BODY response
			recorder = httptest.NewRecorder()
			// Write the response body to the recorder
			recorder.Header().Set("Content-Type", "application/json")
			recorder.Body = nil
			// Get the response from the recorder
			response = recorder.Result()
			client.EXPECT().Get("/PCIeFunctions/0").Return(response, nil)
			service.EXPECT().Systems().Return(systems, nil)
			comSys.EXPECT().PCIeDevices().Return(myPcis, nil)
			cnt, gpuType, err = myBmc.GPUDiscovery(ctx)
			Expect(cnt).To(BeNumerically("==", 0))
			Expect(gpuType).To(BeComparableTo(""))
			Expect(err).To(HaveOccurred())

			// Invalid JSON response
			recorder = httptest.NewRecorder()
			// Write the response body to the recorder
			recorder.Header().Set("Content-Type", "application/json")
			recorder.WriteString("fail")
			// Get the response from the recorder
			response = recorder.Result()
			client.EXPECT().Get("/PCIeFunctions/0").Return(response, nil)
			service.EXPECT().Systems().Return(systems, nil)
			comSys.EXPECT().PCIeDevices().Return(myPcis, nil)
			cnt, gpuType, err = myBmc.GPUDiscovery(ctx)
			Expect(cnt).To(BeNumerically("==", 0))
			Expect(gpuType).To(BeComparableTo(""))
			Expect(err).To(HaveOccurred())

			// Fail case for Systems() error
			service.EXPECT().Systems().Return(systems, fmt.Errorf("No Systems"))
			cnt, gpuType, err = myBmc.GPUDiscovery(ctx)
			Expect(cnt).To(BeNumerically("==", 0))
			Expect(gpuType).To(BeComparableTo(""))
			Expect(err).To(HaveOccurred())

			// Fail case for Systems() is empty
			noSystems := make([]mygofish.GoFishComputerSystemAccessor, 0)
			service.EXPECT().Systems().Return(noSystems, nil)
			cnt, gpuType, err = myBmc.GPUDiscovery(ctx)
			Expect(cnt).To(BeNumerically("==", 0))
			Expect(gpuType).To(BeComparableTo(""))
			Expect(err).To(HaveOccurred())
		})

		It("Denali Pass - HBM Discovery", func() {
			client := mocks.NewMockGoFishClientAccessor(mockCtrl)
			service := mocks.NewMockGoFishServiceAccessor(mockCtrl)
			client.EXPECT().GetService().AnyTimes().Return(service)
			gofishManager.EXPECT().Connect(Any).AnyTimes().Return(client, nil)

			systems := make([]mygofish.GoFishComputerSystemAccessor, 1)
			comSys := mocks.NewMockGoFishComputerSystemAccessor(mockCtrl)
			systems[0] = comSys

			service.EXPECT().Systems().Return(systems, nil)
			comSys.EXPECT().ODataID().AnyTimes().Return("DenaliODataID")
			comSys.EXPECT().Manufacturer().AnyTimes().Return(bmc.IntelCorporation)
			comSys.EXPECT().Model().AnyTimes().Return("D50DNP") // Denali pass
			comSys.EXPECT().PowerState().AnyTimes().Return(redfish.OnPowerState)

			myBmc, err := bmc.New(
				gofishManager,
				&bmc.Config{
					URL:      bmcURL,
					Username: bmcUsername,
					Password: bmcPassword,
				})
			Expect(myBmc).ToNot(BeNil())
			Expect(err).NotTo(HaveOccurred())

			Expect(myBmc.IsVirtual()).To(BeFalse())

			myDimm0 := &redfish.Memory{
				MemoryDeviceType: redfish.MemoryDeviceType(bmc.DDR5),
			}
			myDimm1 := &redfish.Memory{
				MemoryDeviceType: redfish.MemoryDeviceType(bmc.HBM),
			}
			myDimms := make([]*redfish.Memory, 2)
			myDimms[0] = myDimm0
			myDimms[1] = myDimm1

			// Mixed - flat
			service.EXPECT().Systems().Return(systems, nil)
			comSys.EXPECT().Memory().Return(myDimms, nil)
			hbmMode, err := myBmc.HBMDiscovery(ctx)
			Expect(hbmMode).To(BeComparableTo(bmc.HBMFlat))
			Expect(err).NotTo(HaveOccurred())

			// DDR5 only, none
			service.EXPECT().Systems().Return(systems, nil)
			myDimms[0] = myDimm0
			myDimms[1] = myDimm0
			comSys.EXPECT().Memory().Return(myDimms, nil)
			hbmMode, err = myBmc.HBMDiscovery(ctx)
			Expect(hbmMode).To(BeComparableTo(bmc.HBMNone))
			Expect(err).NotTo(HaveOccurred())

			// HBM only, HBM
			service.EXPECT().Systems().Return(systems, nil)
			comSys.EXPECT().Memory().Return(myDimms, nil)
			myDimms[0] = myDimm1
			myDimms[1] = myDimm1
			hbmMode, err = myBmc.HBMDiscovery(ctx)
			Expect(hbmMode).To(BeComparableTo(bmc.HBMOnly))
			Expect(err).NotTo(HaveOccurred())

			// Memory returns an error
			service.EXPECT().Systems().Return(systems, nil)
			comSys.EXPECT().Memory().Return(myDimms, fmt.Errorf("Memory error"))
			hbmMode, err = myBmc.HBMDiscovery(ctx)
			Expect(hbmMode).To(BeComparableTo(bmc.HBMNone))
			Expect(err).To(HaveOccurred())

			// Fail case for Systems() error
			service.EXPECT().Systems().Return(systems, fmt.Errorf("No Systems"))
			hbmMode, err = myBmc.HBMDiscovery(ctx)
			Expect(hbmMode).To(BeComparableTo(bmc.HBMNone))
			Expect(err).To(HaveOccurred())

			// Fail case for Systems() is empty
			noSystems := make([]mygofish.GoFishComputerSystemAccessor, 0)
			service.EXPECT().Systems().Return(noSystems, nil)
			hbmMode, err = myBmc.HBMDiscovery(ctx)
			Expect(hbmMode).To(BeComparableTo(bmc.HBMNone))
			Expect(err).To(HaveOccurred())
		})

		It("Gaudi2 - GPU Discovery", func() {
			client := mocks.NewMockGoFishClientAccessor(mockCtrl)
			service := mocks.NewMockGoFishServiceAccessor(mockCtrl)
			client.EXPECT().GetService().AnyTimes().Return(service)
			gofishManager.EXPECT().Connect(Any).AnyTimes().Return(client, nil)

			systems := make([]mygofish.GoFishComputerSystemAccessor, 1)
			comSys := mocks.NewMockGoFishComputerSystemAccessor(mockCtrl)
			systems[0] = comSys

			service.EXPECT().Systems().Return(systems, nil)
			comSys.EXPECT().ODataID().AnyTimes().Return("Gaudi2ODataID")
			comSys.EXPECT().Manufacturer().AnyTimes().Return(bmc.Supermicro)
			comSys.EXPECT().Model().AnyTimes().Return("SYS-820GH-TNR2") // Gaudi2
			comSys.EXPECT().PowerState().AnyTimes().Return(redfish.OnPowerState)

			myBmc, err := bmc.New(
				gofishManager,
				&bmc.Config{
					URL:      bmcURL,
					Username: bmcUsername,
					Password: bmcPassword,
				})
			Expect(myBmc).ToNot(BeNil())
			Expect(err).NotTo(HaveOccurred())

			Expect(myBmc.IsVirtual()).To(BeFalse())
			Expect(myBmc.IsIntelPlatform()).To(BeFalse())

			myPcis := &bmc.GaudiPCIeDevices{
				Name: "PCIe Device Collection",
				Members: []struct {
					OdataID string "json:\"@odata.id\""
				}{
					{
						OdataID: "/redfish/v1/Chassis/1/PCIeDevices/GPU1",
					},
				},
			}

			mynoGPUpce := &bmc.GaudiPCIeDevices{
				Name: "PCIe Device Collection",
				Members: []struct {
					OdataID string "json:\"@odata.id\""
				}{},
			}

			myPcieFunctions := &bmc.GaudiPcieFunctionsMembers{
				Name: "PCIe Function Collection",
				Members: []struct {
					OdataID string "json:\"@odata.id\""
				}{
					{
						OdataID: "/redfish/v1/Chassis/1/PCIeDevices/GPU1/PCIeFunctions/1",
					},
				},
			}

			// Create a sample response body
			var responseBody bmc.OpenBMCPcieFunction

			responseBody.VendorID = "0x1da3"
			responseBody.DeviceID = "0x1020"
			jsonResponse, err := json.Marshal(responseBody)
			Expect(err).NotTo(HaveOccurred())

			// Create a new HTTP response recorder
			recorder := httptest.NewRecorder()
			// Write the response body to the recorder
			recorder.Header().Set("Content-Type", "application/json")
			recorder.Write(jsonResponse)

			// Get the response from the recorder
			response := recorder.Result()

			//getPCIE devices response
			myPcisJsonResponse, err := json.Marshal(myPcis)
			Expect(err).NotTo(HaveOccurred())

			// Create a new HTTP response recorder
			recorder = httptest.NewRecorder()
			// Write the response body to the recorder
			recorder.Header().Set("Content-Type", "application/json")
			recorder.Write(myPcisJsonResponse)
			myPcisResponse := recorder.Result()

			// Get pcie functions members
			myPcisFunctionsJsonResponse, err := json.Marshal(myPcieFunctions)
			Expect(err).NotTo(HaveOccurred())

			// Create a new HTTP response recorder
			recorder = httptest.NewRecorder()
			// Write the response body to the recorder
			recorder.Header().Set("Content-Type", "application/json")
			recorder.Write(myPcisFunctionsJsonResponse)
			myPcisFunctionsResponse := recorder.Result()

			client.EXPECT().Get("/redfish/v1/Chassis/1/PCIeDevices").Return(myPcisResponse, nil)
			client.EXPECT().Get(myPcis.Members[0].OdataID+"/PCIeFunctions").Return(myPcisFunctionsResponse, nil)
			client.EXPECT().Get(myPcieFunctions.Members[0].OdataID).Return(response, nil)

			cnt, gpuType, err := myBmc.GPUDiscovery(ctx)

			Expect(cnt).To(BeNumerically("==", 1))
			Expect(gpuType).To(BeComparableTo("HL-225"))
			Expect(err).NotTo(HaveOccurred())

			// Fail id chassis PCIE devices return an error
			client.EXPECT().Get("/redfish/v1/Chassis/1/PCIeDevices").Return(myPcisResponse, errors.New("error occurred"))
			_, _, err = myBmc.GPUDiscovery(ctx)
			Expect(err).To(HaveOccurred())

			// Create a new HTTP response recorder
			noGPUJson, _ := json.Marshal(mynoGPUpce)
			recorder = httptest.NewRecorder()
			// Write the response body to the recorder
			recorder.Header().Set("Content-Type", "application/json")
			recorder.Write(noGPUJson)
			// Get the response from the recorder
			noGPUJsonResponse := recorder.Result()
			// Return 0 gpu count if no GPU conneted
			client.EXPECT().Get("/redfish/v1/Chassis/1/PCIeDevices").Return(noGPUJsonResponse, nil)

			cnt, _, err = myBmc.GPUDiscovery(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(cnt).To(BeNumerically("==", 0))

			// Return failure if PCIE Functions return err

			// Create a new HTTP response recorder
			recorder = httptest.NewRecorder()
			// Write the response body to the recorder
			recorder.Header().Set("Content-Type", "application/json")
			recorder.Write(myPcisJsonResponse)
			myPcisResponse = recorder.Result()
			client.EXPECT().Get("/redfish/v1/Chassis/1/PCIeDevices").Return(myPcisResponse, nil)
			client.EXPECT().Get(myPcis.Members[0].OdataID+"/PCIeFunctions").Return(myPcisFunctionsResponse, errors.New("error occurred"))
			_, _, err = myBmc.GPUDiscovery(ctx)
			Expect(err).To(HaveOccurred())

			// Return failure if PCIE Functions 1 return error

			// Create a new HTTP response recorder
			recorder = httptest.NewRecorder()
			// Write the response body to the recorder
			recorder.Header().Set("Content-Type", "application/json")
			recorder.Write(myPcisJsonResponse)
			myPcisResponse = recorder.Result()

			// Create a new HTTP response recorder
			recorder = httptest.NewRecorder()
			// Write the response body to the recorder
			recorder.Header().Set("Content-Type", "application/json")
			recorder.Write(myPcisFunctionsJsonResponse)
			myPcisFunctionsResponse = recorder.Result()

			client.EXPECT().Get("/redfish/v1/Chassis/1/PCIeDevices").Return(myPcisResponse, nil)
			client.EXPECT().Get(myPcis.Members[0].OdataID+"/PCIeFunctions").Return(myPcisFunctionsResponse, nil)
			client.EXPECT().Get(myPcieFunctions.Members[0].OdataID).Return(response, errors.New("error occurred"))
			_, _, err = myBmc.GPUDiscovery(ctx)
			Expect(err).To(HaveOccurred())
		})
		// It("Denali Pass - GetHostMACAddress", func() {
		// 	client := mocks.NewMockGoFishClientAccessor(mockCtrl)
		// 	service := mocks.NewMockGoFishServiceAccessor(mockCtrl)
		// 	client.EXPECT().GetService().AnyTimes().Return(service)
		// 	gofishManager.EXPECT().Connect(Any).AnyTimes().Return(client, nil)

		// 	systems := make([]mygofish.GoFishComputerSystemAccessor, 1)
		// 	comSys := mocks.NewMockGoFishComputerSystemAccessor(mockCtrl)
		// 	systems[0] = comSys

		// 	service.EXPECT().Systems().AnyTimes().Return(systems, nil)
		// 	comSys.EXPECT().ODataID().AnyTimes().Return("DenaliODataID")
		// 	comSys.EXPECT().Manufacturer().AnyTimes().Return(bmc.IntelCorporation)
		// 	comSys.EXPECT().Model().AnyTimes().Return("D50DNP") // Denali pass
		// 	comSys.EXPECT().PowerState().AnyTimes().Return(redfish.OnPowerState)

		// 	bmc, err := bmc.New(
		// 		gofishManager,
		// 		&bmc.Config{
		// 			URL:      bmcURL,
		// 			Username: bmcUsername,
		// 			Password: bmcPassword,
		// 		})
		// 	Expect(bmc).ToNot(BeNil())
		// 	Expect(err).NotTo(HaveOccurred())

		// 	Expect(bmc.IsVirtual()).To(BeFalse())

		// 	myStatus := common.Status{
		// 		State:  common.EnabledState,
		// 		Health: common.OKHealth,
		// 	}
		// 	myEth0 := &redfish.EthernetInterface{
		// 		Status:     myStatus,
		// 		LinkStatus: redfish.LinkUpLinkStatus,
		// 		MACAddress: "00:00:ac:ed:00:00",
		// 	}
		// 	myEth1 := &redfish.EthernetInterface{
		// 		Status:     myStatus,
		// 		LinkStatus: redfish.LinkUpLinkStatus,
		// 	}
		// 	myEthers := make([]*redfish.EthernetInterface, 2)
		// 	myEthers[0] = myEth0
		// 	myEthers[1] = myEth1
		// 	comSys.EXPECT().EthernetInterfaces().AnyTimes().Return(myEthers, nil)

		// 	myNet0 := &redfish.NetworkInterface{
		// 		Status: myStatus,
		// 	}
		// 	myNet1 := &redfish.NetworkInterface{
		// 		Status: myStatus,
		// 	}
		// 	myNets := make([]*redfish.NetworkInterface, 2)
		// 	myNets[0] = myNet0
		// 	myNets[1] = myNet1
		// 	comSys.EXPECT().NetworkInterfaces().AnyTimes().Return(myNets, nil)

		// 	// TODO: High Effort
		// 	// Requires mocking system.NetworkInterfaces(), *redfish.NetworkInterface,
		// 	// nic.NetworkPorts(), GetClient().Get(),
		// 	mac, err := bmc.GetHostMACAddress(ctx)
		// 	Expect(mac).ToNot(BeNil())
		// 	Expect(err).NotTo(HaveOccurred())
		// })

		It("should return BMC interface - coyote pass", func() {
			client := mocks.NewMockGoFishClientAccessor(mockCtrl)
			service := mocks.NewMockGoFishServiceAccessor(mockCtrl)
			client.EXPECT().GetService().AnyTimes().Return(service)

			systems := make([]mygofish.GoFishComputerSystemAccessor, 1)
			comSys := mocks.NewMockGoFishComputerSystemAccessor(mockCtrl)
			systems[0] = comSys

			service.EXPECT().Systems().AnyTimes().Return(systems, nil)
			comSys.EXPECT().ODataID().AnyTimes().Return("CoyoteODataID")
			comSys.EXPECT().Manufacturer().AnyTimes().Return(bmc.IntelCorporation)
			comSys.EXPECT().Model().AnyTimes().Return("M50CYP") // Coyote pass
			comSys.EXPECT().PowerState().AnyTimes().Return(redfish.OnPowerState)

			gofishManager.EXPECT().Connect(Any).AnyTimes().Return(client, nil)
			bmc, err := bmc.New(
				gofishManager,
				&bmc.Config{
					URL:      bmcURL,
					Username: bmcUsername,
					Password: bmcPassword,
				})
			Expect(bmc).ToNot(BeNil())
			Expect(err).NotTo(HaveOccurred())

			Expect(bmc.IsVirtual()).To(BeFalse())

			addr, err := bmc.GetHostBMCAddress()
			Expect(addr).ToNot(BeNil())
			Expect(err).NotTo(HaveOccurred())

			// Not implemented, so no Patch calls expected
			err = bmc.ConfigureNTP(ctx)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Coyote Pass - GetHostMACAddress", func() {
			client := mocks.NewMockGoFishClientAccessor(mockCtrl)
			service := mocks.NewMockGoFishServiceAccessor(mockCtrl)
			client.EXPECT().GetService().AnyTimes().Return(service)
			gofishManager.EXPECT().Connect(Any).AnyTimes().Return(client, nil)

			systems := make([]mygofish.GoFishComputerSystemAccessor, 1)
			comSys := mocks.NewMockGoFishComputerSystemAccessor(mockCtrl)
			systems[0] = comSys

			service.EXPECT().Systems().AnyTimes().Return(systems, nil)
			comSys.EXPECT().ODataID().AnyTimes().Return("CoyoteODataID")
			comSys.EXPECT().Manufacturer().AnyTimes().Return(bmc.IntelCorporation)
			comSys.EXPECT().Model().AnyTimes().Return("M50CYP") // Coyote pass
			comSys.EXPECT().PowerState().AnyTimes().Return(redfish.OnPowerState)

			bmc, err := bmc.New(
				gofishManager,
				&bmc.Config{
					URL:      bmcURL,
					Username: bmcUsername,
					Password: bmcPassword,
				})
			Expect(bmc).ToNot(BeNil())
			Expect(err).NotTo(HaveOccurred())

			Expect(bmc.IsVirtual()).To(BeFalse())

			myStatus := common.Status{
				State:  common.EnabledState,
				Health: common.OKHealth,
			}
			myEth0 := &redfish.EthernetInterface{
				Status:     myStatus,
				LinkStatus: redfish.LinkUpLinkStatus,
				MACAddress: "00:00:ac:ed:00:00",
			}
			myEth1 := &redfish.EthernetInterface{
				Status:     myStatus,
				LinkStatus: redfish.LinkUpLinkStatus,
			}
			myEthers := make([]*redfish.EthernetInterface, 2)
			myEthers[0] = myEth0
			myEthers[1] = myEth1
			comSys.EXPECT().EthernetInterfaces().AnyTimes().Return(myEthers, nil)

			myNet0 := &redfish.NetworkInterface{
				Status: myStatus,
			}
			myNet1 := &redfish.NetworkInterface{
				Status: myStatus,
			}
			myNets := make([]*redfish.NetworkInterface, 2)
			myNets[0] = myNet0
			myNets[1] = myNet1
			comSys.EXPECT().NetworkInterfaces().AnyTimes().Return(myNets, nil)

			mac, err := bmc.GetHostMACAddress(ctx)
			Expect(mac).ToNot(BeNil())
			Expect(err).NotTo(HaveOccurred())
		})

		It("Coyote Pass - Boot Order", func() {
			client := mocks.NewMockGoFishClientAccessor(mockCtrl)
			service := mocks.NewMockGoFishServiceAccessor(mockCtrl)
			client.EXPECT().GetService().AnyTimes().Return(service)
			gofishManager.EXPECT().Connect(Any).AnyTimes().Return(client, nil)

			systems := make([]mygofish.GoFishComputerSystemAccessor, 1)
			comSys := mocks.NewMockGoFishComputerSystemAccessor(mockCtrl)
			systems[0] = comSys

			service.EXPECT().Systems().AnyTimes().Return(systems, nil)
			comSys.EXPECT().ODataID().AnyTimes().Return("CoyoteODataID")
			comSys.EXPECT().Manufacturer().AnyTimes().Return(bmc.IntelCorporation)
			comSys.EXPECT().Model().AnyTimes().Return("M50CYP") // Coyote pass
			comSys.EXPECT().PowerState().AnyTimes().Return(redfish.OnPowerState)

			myBmc, err := bmc.New(
				gofishManager,
				&bmc.Config{
					URL:      bmcURL,
					Username: bmcUsername,
					Password: bmcPassword,
				})
			Expect(myBmc).ToNot(BeNil())
			Expect(err).NotTo(HaveOccurred())

			Expect(myBmc.IsVirtual()).To(BeFalse())

			myBoot := mocks.NewMockGoFishBootAccessor(mockCtrl)
			comSys.EXPECT().Boot().AnyTimes().Return(myBoot)
			comSys.EXPECT().SetBoot(Any).AnyTimes().Return(nil)
			myBootOrder := []string{bmc.CoyoteBaseboardCard, bmc.CoyoteRiserCard}
			myBoot.EXPECT().BootOrder().AnyTimes().Return(myBootOrder)
			myBoot.EXPECT().SetBootOrder(Any).AnyTimes().Return(nil)
			err = myBmc.SanitizeBMCBootOrder(ctx)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Coyote Pass - PFR", func() {
			client := mocks.NewMockGoFishClientAccessor(mockCtrl)
			service := mocks.NewMockGoFishServiceAccessor(mockCtrl)
			client.EXPECT().GetService().AnyTimes().Return(service)
			gofishManager.EXPECT().Connect(Any).AnyTimes().Return(client, nil)

			systems := make([]mygofish.GoFishComputerSystemAccessor, 1)
			comSys := mocks.NewMockGoFishComputerSystemAccessor(mockCtrl)
			systems[0] = comSys

			service.EXPECT().Systems().Return(systems, nil)
			comSys.EXPECT().ODataID().AnyTimes().Return("CoyoteODataID")
			comSys.EXPECT().Manufacturer().AnyTimes().Return(bmc.IntelCorporation)
			comSys.EXPECT().Model().AnyTimes().Return("M50CYP") // Coyote pass
			comSys.EXPECT().PowerState().AnyTimes().Return(redfish.OnPowerState)

			myBmc, err := bmc.New(
				gofishManager,
				&bmc.Config{
					URL:      bmcURL,
					Username: bmcUsername,
					Password: bmcPassword,
				})
			Expect(myBmc).ToNot(BeNil())
			Expect(err).NotTo(HaveOccurred())

			Expect(myBmc.IsVirtual()).To(BeFalse())

			// Create a sample response body
			var responseBody bmc.CoyotePFR
			responseBody.Oem.FirmwareProvisioning.Provisioned = "true"
			responseBody.Oem.FirmwareProvisioning.Locked = "true"
			jsonResponse, err := json.Marshal(responseBody)
			Expect(err).NotTo(HaveOccurred())

			// Create a new HTTP response recorder
			recorder := httptest.NewRecorder()
			// Write the response body to the recorder
			recorder.Header().Set("Content-Type", "application/json")
			recorder.Write(jsonResponse)

			// Get the response from the recorder
			response := recorder.Result()
			// [CoyoteODataID]
			client.EXPECT().Get("CoyoteODataID").Return(response, nil)
			service.EXPECT().Systems().Return(systems, nil)
			err = myBmc.VerifyPlatformFirmwareResilience(ctx)
			Expect(err).NotTo(HaveOccurred())

			// Not Locked
			responseBody.Oem.FirmwareProvisioning.Locked = "false"
			recorder = httptest.NewRecorder()
			recorder.Header().Set("Content-Type", "application/json")
			jsonResponse, err = json.Marshal(responseBody)
			recorder.Write(jsonResponse)
			Expect(err).NotTo(HaveOccurred())

			// Get the response from the recorder
			response = recorder.Result()
			// [CoyoteODataID]
			client.EXPECT().Get("CoyoteODataID").Return(response, nil)
			service.EXPECT().Systems().Return(systems, nil)
			err = myBmc.VerifyPlatformFirmwareResilience(ctx)
			Expect(err).To(HaveOccurred())

			// Invalid Get() response
			recorder = httptest.NewRecorder()
			// Write the response body to the recorder
			recorder.Header().Set("Content-Type", "application/json")
			recorder.Body = nil
			// Get the response from the recorder
			response = recorder.Result()
			// [CoyoteODataID]
			client.EXPECT().Get("CoyoteODataID").Return(response, fmt.Errorf("Get fails"))
			service.EXPECT().Systems().Return(systems, nil)
			err = myBmc.VerifyPlatformFirmwareResilience(ctx)
			Expect(err).To(HaveOccurred())

			// Invalid BODY response
			recorder = httptest.NewRecorder()
			// Write the response body to the recorder
			recorder.Header().Set("Content-Type", "application/json")
			recorder.Body = nil
			// Get the response from the recorder
			response = recorder.Result()
			// [CoyoteODataID]
			client.EXPECT().Get("CoyoteODataID").Return(response, nil)
			service.EXPECT().Systems().Return(systems, nil)
			err = myBmc.VerifyPlatformFirmwareResilience(ctx)
			Expect(err).To(HaveOccurred())

			// Invalid JSON response
			recorder = httptest.NewRecorder()
			// Write the response body to the recorder
			recorder.Header().Set("Content-Type", "application/json")
			recorder.WriteString("fail")
			// Get the response from the recorder
			response = recorder.Result()
			// [CoyoteODataID]
			client.EXPECT().Get("CoyoteODataID").Return(response, nil)
			service.EXPECT().Systems().Return(systems, nil)
			err = myBmc.VerifyPlatformFirmwareResilience(ctx)
			Expect(err).To(HaveOccurred())

			// Fail case for Systems() error
			service.EXPECT().Systems().Return(systems, fmt.Errorf("No Systems"))
			err = myBmc.VerifyPlatformFirmwareResilience(ctx)
			Expect(err).To(HaveOccurred())

			// Fail case for Systems() is empty
			noSystems := make([]mygofish.GoFishComputerSystemAccessor, 0)
			service.EXPECT().Systems().Return(noSystems, nil)
			err = myBmc.VerifyPlatformFirmwareResilience(ctx)
			Expect(err).To(HaveOccurred())
		})

		It("Coyote Pass - GPU Discovery", func() {
			client := mocks.NewMockGoFishClientAccessor(mockCtrl)
			service := mocks.NewMockGoFishServiceAccessor(mockCtrl)
			client.EXPECT().GetService().AnyTimes().Return(service)
			gofishManager.EXPECT().Connect(Any).AnyTimes().Return(client, nil)

			systems := make([]mygofish.GoFishComputerSystemAccessor, 1)
			comSys := mocks.NewMockGoFishComputerSystemAccessor(mockCtrl)
			systems[0] = comSys

			service.EXPECT().Systems().Return(systems, nil)
			comSys.EXPECT().ODataID().AnyTimes().Return("CoyoteODataID")
			comSys.EXPECT().Manufacturer().AnyTimes().Return(bmc.IntelCorporation)
			comSys.EXPECT().Model().AnyTimes().Return("M50CYP") // Coyote pass
			comSys.EXPECT().PowerState().AnyTimes().Return(redfish.OnPowerState)

			myBmc, err := bmc.New(
				gofishManager,
				&bmc.Config{
					URL:      bmcURL,
					Username: bmcUsername,
					Password: bmcPassword,
				})
			Expect(myBmc).ToNot(BeNil())
			Expect(err).NotTo(HaveOccurred())

			Expect(myBmc.IsVirtual()).To(BeFalse())

			// Create a sample response body
			var responseBody bmc.CoyoteOEM

			responseBody.Oem.IntelRackScale.PciDevices = []struct {
				VendorID string `json:"VendorId"`
				DeviceID string `json:"DeviceId"`
			}{
				{VendorID: "0x8086", DeviceID: "0x56c0"},
				{VendorID: "0x8086", DeviceID: "0x56c0"},
			}

			jsonResponse, err := json.Marshal(responseBody)
			Expect(err).NotTo(HaveOccurred())

			// Create a new HTTP response recorder
			recorder := httptest.NewRecorder()
			// Write the response body to the recorder
			recorder.Header().Set("Content-Type", "application/json")
			recorder.Write(jsonResponse)

			// Get the response from the recorder
			response := recorder.Result()
			// /PCIeFunctions/0
			client.EXPECT().Get("CoyoteODataID").Return(response, nil)
			service.EXPECT().Systems().Return(systems, nil)
			cnt, gpuType, err := myBmc.GPUDiscovery(ctx)
			Expect(cnt).To(BeNumerically("==", 2))
			Expect(gpuType).To(BeComparableTo("GPU-Flex-170"))
			Expect(err).NotTo(HaveOccurred())

			// Other GPU type
			responseBody.Oem.IntelRackScale.PciDevices = []struct {
				VendorID string `json:"VendorId"`
				DeviceID string `json:"DeviceId"`
			}{
				{VendorID: "0x8086", DeviceID: "0x0bda"},
				{VendorID: "0x8086", DeviceID: "0x0bda"},
			}
			jsonResponse, err = json.Marshal(responseBody)
			Expect(err).NotTo(HaveOccurred())
			recorder = httptest.NewRecorder()
			// Write the response body to the recorder
			recorder.Header().Set("Content-Type", "application/json")
			recorder.Write(jsonResponse)

			// Get the response from the recorder
			response = recorder.Result()
			// /PCIeFunctions/0
			client.EXPECT().Get("CoyoteODataID").Return(response, nil)
			service.EXPECT().Systems().Return(systems, nil)
			cnt, gpuType, err = myBmc.GPUDiscovery(ctx)
			Expect(cnt).To(BeNumerically("==", 2))
			Expect(gpuType).To(BeComparableTo("GPU-Max-1100"))
			Expect(err).NotTo(HaveOccurred())

			// Invalid Get() response
			recorder = httptest.NewRecorder()
			// Write the response body to the recorder
			recorder.Header().Set("Content-Type", "application/json")
			recorder.Body = nil
			// Get the response from the recorder
			response = recorder.Result()
			client.EXPECT().Get("CoyoteODataID").Return(response, fmt.Errorf("Get fails"))
			service.EXPECT().Systems().Return(systems, nil)
			cnt, gpuType, err = myBmc.GPUDiscovery(ctx)
			Expect(cnt).To(BeNumerically("==", 0))
			Expect(gpuType).To(BeComparableTo(""))
			Expect(err).To(HaveOccurred())

			// Invalid BODY response
			recorder = httptest.NewRecorder()
			// Write the response body to the recorder
			recorder.Header().Set("Content-Type", "application/json")
			recorder.Body = nil
			// Get the response from the recorder
			response = recorder.Result()
			client.EXPECT().Get("CoyoteODataID").Return(response, nil)
			service.EXPECT().Systems().Return(systems, nil)
			cnt, gpuType, err = myBmc.GPUDiscovery(ctx)
			Expect(cnt).To(BeNumerically("==", 0))
			Expect(gpuType).To(BeComparableTo(""))
			Expect(err).To(HaveOccurred())

			// Invalid JSON response
			recorder = httptest.NewRecorder()
			// Write the response body to the recorder
			recorder.Header().Set("Content-Type", "application/json")
			recorder.WriteString("fail")
			// Get the response from the recorder
			response = recorder.Result()
			client.EXPECT().Get("CoyoteODataID").Return(response, nil)
			service.EXPECT().Systems().Return(systems, nil)
			cnt, gpuType, err = myBmc.GPUDiscovery(ctx)
			Expect(cnt).To(BeNumerically("==", 0))
			Expect(gpuType).To(BeComparableTo(""))
			Expect(err).To(HaveOccurred())

			// Fail case for Systems() error
			service.EXPECT().Systems().Return(systems, fmt.Errorf("No Systems"))
			cnt, gpuType, err = myBmc.GPUDiscovery(ctx)
			Expect(cnt).To(BeNumerically("==", 0))
			Expect(gpuType).To(BeComparableTo(""))
			Expect(err).To(HaveOccurred())

			// Fail case for Systems() is empty
			noSystems := make([]mygofish.GoFishComputerSystemAccessor, 0)
			service.EXPECT().Systems().Return(noSystems, nil)
			cnt, gpuType, err = myBmc.GPUDiscovery(ctx)
			Expect(cnt).To(BeNumerically("==", 0))
			Expect(gpuType).To(BeComparableTo(""))
			Expect(err).To(HaveOccurred())
		})

		It("should fail BMC interface: bmc service", func() {
			client := mocks.NewMockGoFishClientAccessor(mockCtrl)
			gofishManager.EXPECT().Connect(Any).AnyTimes().Return(client, fmt.Errorf("failed to connect to BMC service"))
			bmc, err := bmc.New(
				gofishManager,
				&bmc.Config{
					URL:      bmcURL,
					Username: bmcUsername,
					Password: bmcPassword,
				})
			Expect(bmc).To(BeNil())
			Expect(err).To(HaveOccurred())
		})

		It("should return BMC interface - system error", func() {
			client := mocks.NewMockGoFishClientAccessor(mockCtrl)
			service := mocks.NewMockGoFishServiceAccessor(mockCtrl)
			client.EXPECT().GetService().AnyTimes().Return(service)

			systems := make([]mygofish.GoFishComputerSystemAccessor, 1)
			comSys := mocks.NewMockGoFishComputerSystemAccessor(mockCtrl)
			systems[0] = comSys

			service.EXPECT().Systems().AnyTimes().Return(systems, nil)
			comSys.EXPECT().ODataID().AnyTimes().Return("DenaliODataID")
			comSys.EXPECT().Manufacturer().AnyTimes().Return(bmc.IntelCorporation)
			comSys.EXPECT().Model().AnyTimes().Return("D50DNP") // Denali pass
			comSys.EXPECT().PowerState().AnyTimes().Return(redfish.OnPowerState)

			gofishManager.EXPECT().Connect(Any).AnyTimes().Return(client, nil)
			bmc, err := bmc.New(
				gofishManager,
				&bmc.Config{
					URL:      bmcURL,
					Username: bmcUsername,
					Password: bmcPassword,
				})
			Expect(bmc).ToNot(BeNil())
			Expect(err).NotTo(HaveOccurred())

			response := &http.Response{StatusCode: 200}
			client.EXPECT().Post("/redfish/v1/AccountService/Accounts", Any).Return(response, nil)
			err = bmc.CreateAccount(ctx, "newuser", "newpass")
			Expect(err).NotTo(HaveOccurred())

			response = &http.Response{StatusCode: 500}
			client.EXPECT().Post("/redfish/v1/AccountService/Accounts", Any).Return(response, fmt.Errorf("failed to add"))
			err = bmc.CreateAccount(ctx, "newuser", "newpass")
			Expect(err).To(HaveOccurred())

			response = &http.Response{StatusCode: 405}
			client.EXPECT().Post("/redfish/v1/AccountService/Accounts", Any).Return(response, nil)
			err = bmc.CreateAccount(ctx, "newuser", "newpass")
			Expect(err).To(HaveOccurred())
		})
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})
})
