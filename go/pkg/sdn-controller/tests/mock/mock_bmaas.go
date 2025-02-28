package mock

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/client"
)

type MockBM struct {
	SDNClient *client.SDNClient
}

func New(sdnClient *client.SDNClient) *MockBM {
	return &MockBM{
		SDNClient: sdnClient,
	}
}

// ------------------------------------------
// non-Gaudi layer 2 (ie, SPR)
// ------------------------------------------
func (bm *MockBM) ReserveMultipleBMHWithVXX(ctx context.Context, hostNames []string, vlanID int64) error {
	var wg sync.WaitGroup
	wg.Add(len(hostNames))
	errCh := make(chan error, len(hostNames))
	for _, hostName := range hostNames {
		go func(hn string) {
			defer wg.Done()
			fmt.Printf("reserving BMH [%v], vlan: [%v] \n", hn, vlanID)
			err := bm.ReserveSingleBMHWithVXX(ctx, hn, vlanID)
			if err != nil {
				errCh <- err
			}
		}(hostName)
	}

	done := make(chan struct{})

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		fmt.Println("All operations completed")
	case <-time.After(60 * time.Second):
		fmt.Println("Timeout occurred before all operations were completed")
	case err := <-errCh:
		return fmt.Errorf("ReserveSingleBMHWithVXX failed %s \n", err.Error())
	}

	return nil
}

func (bm *MockBM) ReserveSingleBMHWithVXX(ctx context.Context, hostName string, vlanID int64) error {
	// TODO: update the BMH with consumerRef
	// TODO: update the other BMH fields if needed
	// call SDN to update NN
	err := bm.SDNClient.UpdateNetworkNodeConfig(ctx, client.NetworkNodeConfUpdateRequest{
		NetworkNodeName:    hostName,
		FrontEndFabricVlan: vlanID,
	})
	if err != nil {
		fmt.Printf("MockBM UpdateNetworkNodeConfig error: %v \n", err)
	}
	// sleep for 5 seconds, simulating the requeue behavior
	time.Sleep(5 * time.Second)

	var result client.NetworkNodeConfStatusCheckResponse
	for result.Status != client.UpdateCompleted {
		result, err = bm.SDNClient.CheckNetworkNodeStatus(ctx, client.NetworkNodeConfStatusCheckRequest{
			NetworkNodeName:           hostName,
			DesiredFrontEndFabricVlan: vlanID,
		})
		if err != nil {
			fmt.Printf("MockBM CheckNetworkNodeConfigStatus error: %v \n", err)
			return err
		}
		// sleep for 5 seconds, simulating the requeue behavior
		time.Sleep(5 * time.Second)
	}

	// TODO: update any BMH state field
	// ...
	return nil
}

// ------------------------------------------
// Gaudi Layer 2
// ------------------------------------------

func (bm *MockBM) ReserveMultipleBMHWithVVX(ctx context.Context, hostNames []string, feVlanID int64, accVlanID int64) error {
	var wg sync.WaitGroup
	wg.Add(len(hostNames))
	errCh := make(chan error, len(hostNames))
	for _, hostName := range hostNames {
		go func(h string) {
			err := bm.ReserveSingleBMHWithVVX(ctx, h, feVlanID, accVlanID)
			if err != nil {
				errCh <- err
			}
			defer wg.Done()
		}(hostName)
	}
	wg.Wait()
	if len(errCh) > 0 {
		// return one of them is ok for testing
		return <-errCh
	}
	return nil
}

func (bm *MockBM) ReserveSingleBMHWithVVX(ctx context.Context, hostName string, feVlanID int64, accVlanID int64) error {
	// update the BMH with consumerRef
	// update the other BMH fields if needed
	// call SDN to update NN
	err := bm.SDNClient.UpdateNetworkNodeConfig(ctx, client.NetworkNodeConfUpdateRequest{
		NetworkNodeName:       hostName,
		FrontEndFabricVlan:    feVlanID,
		AcceleratorFabricVlan: accVlanID,
	})
	if err != nil {
		fmt.Printf("MockBM UpdateNetworkNodeConfig error: %v \n", err)
	}
	// sleep for 5 seconds, simulating the requeue behavior
	time.Sleep(5 * time.Second)

	var result client.NetworkNodeConfStatusCheckResponse
	for result.Status != client.UpdateCompleted {
		result, err = bm.SDNClient.CheckNetworkNodeStatus(ctx, client.NetworkNodeConfStatusCheckRequest{
			NetworkNodeName:              hostName,
			DesiredFrontEndFabricVlan:    feVlanID,
			DesiredAcceleratorFabricVlan: accVlanID,
		})
		if err != nil {
			fmt.Printf("MockBM CheckNetworkNodeConfigStatus error: %v \n", err)
			return err
		}
		// sleep for 5 seconds, simulating the requeue behavior
		time.Sleep(5 * time.Second)
	}

	// update any BMH state field
	// ...
	return nil
}

// ------------------------------------------
// Gaudi Layer 3 (BGP)
// ------------------------------------------
func (bm *MockBM) ReserveMultipleBMHWithXBX(ctx context.Context, hostNames []string, BGPCommunityID int64) error {
	var wg sync.WaitGroup
	wg.Add(len(hostNames))
	errCh := make(chan error, len(hostNames))
	for _, hostName := range hostNames {
		go func(h string) {
			err := bm.ReserveSingleBMHWithXBX(ctx, h, BGPCommunityID)
			if err != nil {
				errCh <- err
			}
			defer wg.Done()
		}(hostName)
	}
	wg.Wait()
	if len(errCh) > 0 {
		// return one of them is ok for testing
		return <-errCh
	}
	return nil
}

func (bm *MockBM) ReserveSingleBMHWithXBX(ctx context.Context, hostName string, BGPCommunityID int64) error {
	err := bm.SDNClient.UpdateNetworkNodeConfig(ctx, client.NetworkNodeConfUpdateRequest{
		NetworkNodeName: hostName,
		// FrontEndFabricVlan:              vlan,
		AcceleratorFabricBGPCommunityID: BGPCommunityID,
	})
	if err != nil {
		fmt.Printf("MockBM UpdateNetworkNodeConfig error: %v \n", err)
		return err
	}

	time.Sleep(5 * time.Second)

	var result client.NetworkNodeConfStatusCheckResponse
	for result.Status != client.UpdateCompleted {
		result, err = bm.SDNClient.CheckNetworkNodeStatus(ctx, client.NetworkNodeConfStatusCheckRequest{
			NetworkNodeName: hostName,
			// DesiredFrontEndFabricVlan:              vlan,
			DesiredAcceleratorFabricBGPCommunityID: BGPCommunityID,
		})
		if err != nil {
			fmt.Printf("MockBM CheckNetworkNodeConfigStatus error: %v \n", err)
			return err
		}
		time.Sleep(5 * time.Second)
	}
	return nil
}
