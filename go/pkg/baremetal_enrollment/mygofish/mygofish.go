// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package mygofish

//go:generate mockgen -destination ../mocks/mygofish.go -package mocks github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/mygofish GoFishManagerAccessor,GoFishClientAccessor,GoFishServiceAccessor,GoFishAccountServiceAccessor,GoFishComputerSystemAccessor,GoFishBootAccessor,GoFishUpdateServiceAccessor

import (
	"net/http"

	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
)

type GoFishManager interface {
	GoFishManagerAccessor
	GoFishClientAccessor
	GoFishServiceAccessor
	GoFishAccountServiceAccessor
	GoFishComputerSystemAccessor
	GoFishBootAccessor
}

type MyGoFishManager struct {
}

type GoFishManagerAccessor interface {
	Connect(config gofish.ClientConfig) (GoFishClientAccessor, error)
}

func (s *MyGoFishManager) Connect(config gofish.ClientConfig) (GoFishClientAccessor, error) {
	APIClient, err := gofish.Connect(config)
	myAPIClient := &MyGoFishClient{APIClient: APIClient}
	return myAPIClient, err
}

type GoFishClientAccessor interface {
	Get(url string) (*http.Response, error)
	Patch(url string, payload interface{}) (*http.Response, error)
	Post(url string, payload interface{}) (*http.Response, error)
	GetService() GoFishServiceAccessor
}

type MyGoFishClient struct {
	APIClient *gofish.APIClient
}

func (s *MyGoFishClient) Get(url string) (*http.Response, error) {
	return s.APIClient.Get(url)
}

func (s *MyGoFishClient) Patch(url string, payload interface{}) (*http.Response, error) {
	return s.APIClient.Patch(url, payload)
}

func (s *MyGoFishClient) Post(url string, payload interface{}) (*http.Response, error) {
	return s.APIClient.Post(url, payload)
}

func (s *MyGoFishClient) GetService() GoFishServiceAccessor {
	Service := s.APIClient.GetService()
	myService := &MyGoFishService{Service: Service}
	return myService
}

type GoFishServiceAccessor interface {
	AccountService() (GoFishAccountServiceAccessor, error)
	Systems() ([]GoFishComputerSystemAccessor, error)
	UpdateService() (GoFishUpdateServiceAccessor, error)
}

type MyGoFishService struct {
	Service *gofish.Service
}

func (s *MyGoFishService) AccountService() (GoFishAccountServiceAccessor, error) {
	AccountService, err := s.Service.AccountService()
	myAccountService := &MyGoFishAccountService{AccountService: AccountService}
	return myAccountService, err
}

func (s *MyGoFishService) Systems() ([]GoFishComputerSystemAccessor, error) {
	ComputerSystems, err := s.Service.Systems()

	myComputerSystems := make([]GoFishComputerSystemAccessor, len(ComputerSystems))

	for i, elem := range ComputerSystems {
		myElem := &MyGoFishComputerSystem{ComputerSystem: elem}
		myComputerSystems[i] = myElem
	}

	return myComputerSystems, err
}

func (s *MyGoFishService) UpdateService() (GoFishUpdateServiceAccessor, error) {
	UpdateService, err := s.Service.UpdateService()
	myUpdateService := &MyGoFishUpdateService{UpdateService: UpdateService}
	return myUpdateService, err
}

type GoFishComputerSystemAccessor interface {
	Boot() GoFishBootAccessor
	EthernetInterfaces() ([]*redfish.EthernetInterface, error)
	ID() string
	PowerState() redfish.PowerState
	Manufacturer() string
	Memory() ([]*redfish.Memory, error)
	Model() string
	Name() string
	NetworkInterfaces() ([]*redfish.NetworkInterface, error)
	ODataID() string
	PCIeDevices() ([]*redfish.PCIeDevice, error)
	Processors() ([]*redfish.Processor, error)
	Reset(resetType redfish.ResetType) error
	SetBoot(GoFishBootAccessor) error
}

type MyGoFishComputerSystem struct {
	ComputerSystem *redfish.ComputerSystem
}

func (s *MyGoFishComputerSystem) Boot() GoFishBootAccessor {
	Boot := s.ComputerSystem.Boot
	myBoot := &MyGoFishBoot{Boot: Boot}
	return myBoot
}

func (s *MyGoFishComputerSystem) EthernetInterfaces() ([]*redfish.EthernetInterface, error) {
	return s.ComputerSystem.EthernetInterfaces()
}

func (s *MyGoFishComputerSystem) ID() string {
	return s.ComputerSystem.ID
}

func (s *MyGoFishComputerSystem) PowerState() redfish.PowerState {
	return s.ComputerSystem.PowerState
}

func (s *MyGoFishComputerSystem) Manufacturer() string {
	return s.ComputerSystem.Manufacturer
}

func (s *MyGoFishComputerSystem) Model() string {
	return s.ComputerSystem.Model
}

func (s *MyGoFishComputerSystem) Memory() ([]*redfish.Memory, error) {
	return s.ComputerSystem.Memory()
}
func (s *MyGoFishComputerSystem) Name() string {
	return s.ComputerSystem.Name
}

func (s *MyGoFishComputerSystem) NetworkInterfaces() ([]*redfish.NetworkInterface, error) {
	return s.ComputerSystem.NetworkInterfaces()
}

func (s *MyGoFishComputerSystem) ODataID() string {
	return s.ComputerSystem.Entity.ODataID
}

func (s *MyGoFishComputerSystem) PCIeDevices() ([]*redfish.PCIeDevice, error) {
	return s.ComputerSystem.PCIeDevices()
}

func (s *MyGoFishComputerSystem) Processors() ([]*redfish.Processor, error) {
	return s.ComputerSystem.Processors()
}

func (s *MyGoFishComputerSystem) Reset(resetType redfish.ResetType) error {
	return s.ComputerSystem.Reset(resetType)
}

func (s *MyGoFishComputerSystem) SetBoot(newBoot GoFishBootAccessor) error {
	var myBoot redfish.Boot
	myBoot.BootOrder = newBoot.GetBoot().BootOrder
	return s.ComputerSystem.SetBoot(myBoot)
}

type GoFishBootAccessor interface {
	GetBoot() redfish.Boot
	BootOrder() []string
	SetBootOrder([]string) error
}

type MyGoFishBoot struct {
	Boot redfish.Boot
}

func (b *MyGoFishBoot) GetBoot() redfish.Boot {
	return b.Boot
}

func (b *MyGoFishBoot) BootOrder() []string {
	return b.Boot.BootOrder
}

func (b *MyGoFishBoot) SetBootOrder(newOrder []string) error {
	b.Boot.BootOrder = newOrder

	return nil
}

type GoFishAccountServiceAccessor interface {
	Accounts() ([]*redfish.ManagerAccount, error)
}

type MyGoFishAccountService struct {
	AccountService *redfish.AccountService
}

func (s *MyGoFishAccountService) Accounts() ([]*redfish.ManagerAccount, error) {
	return s.AccountService.Accounts()
}

type GoFishUpdateServiceAccessor interface {
	FirmwareInventories() ([]*redfish.SoftwareInventory, error)
}

type MyGoFishUpdateService struct {
	UpdateService *redfish.UpdateService
}

func (s *MyGoFishUpdateService) FirmwareInventories() ([]*redfish.SoftwareInventory, error) {
	return s.UpdateService.FirmwareInventories()
}
