package controllers

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/tests/mock"
)

func TestBMHControllerManager(t *testing.T) {
	fmt.Printf("starting TestBMHControllerManager... \n")
	ctx := context.Background()
	mockRecorder := mock.FakeEventRecorder{}
	bmhConf := BMHControllerConf{
		BMHControllerCreationRetryInterval: 5 * time.Second,
		EventRecorder:                      mockRecorder,
	}

	allBeGood := false

	newFn := func(ctx context.Context, bmhConf BMHControllerConf, name string, kubeconfig string) (BMHControllerIF, error) {
		c := &TestBMHController{
			Name: kubeconfig,
		}

		if !allBeGood && strings.Contains(kubeconfig, "bad") {
			return nil, fmt.Errorf("unable to LoadKubeConfigFile.  Kubeconfig servers: %s", name)
		}
		return c, nil
	}

	bmhManager := NewBMHControllerManager(bmhConf, newFn, genName, "good1;good2;bad1")
	bmhManager.CreateAllControllers(ctx)
	bmhManager.StartAllControllers(ctx)

	var runnableControllers map[string]BMHControllerIF
	var failedControllers map[string]struct{}
	fmt.Printf("--- step 1 \n")
	notifyCh := make(chan struct{})
	go func() {
		for {
			runnableControllers = bmhManager.GetRunnableControllers()
			failedControllers = bmhManager.GetFailedControllers()
			fmt.Printf("runnableControllers: %+v \n", runnableControllers)
			fmt.Printf("failedControllers: %+v \n", failedControllers)
			if len(runnableControllers) == 2 && len(failedControllers) == 1 {
				notifyCh <- struct{}{}
			} else {
				time.Sleep(1 * time.Second)
				continue
			}
		}
	}()

FOR1:
	for {
		select {
		case <-time.After(30 * time.Second):
			t.Fatalf("timeout waiting for the expected step 1 result")
		case <-notifyCh:
			break FOR1
		}
	}

	fmt.Printf("--- step 2 \n")
	notifyCh2 := make(chan struct{})
	go func() {
		for {
			runnableControllers = bmhManager.GetRunnableControllers()
			failedControllers = bmhManager.GetFailedControllers()
			fmt.Printf("runnableControllers: %+v \n", runnableControllers)
			fmt.Printf("failedControllers: %+v \n", failedControllers)
			if len(runnableControllers) == 3 && len(failedControllers) == 0 {
				notifyCh2 <- struct{}{}
			} else {
				time.Sleep(1 * time.Second)
				continue
			}
		}
	}()

	// make all controllers creatable.
	allBeGood = true

FOR2:
	for {
		select {
		case <-time.After(30 * time.Second):
			t.Fatalf("timeout waiting for the expected step 2 result")
		case <-notifyCh2:
			break FOR2
		}
	}
}

func TestBMHControllerManagerRunFailedThenRecover(t *testing.T) {
	fmt.Printf("starting TestBMHControllerManagerRunFailedThenRecover... \n")
	ctx := context.Background()

	bmhConf := BMHControllerConf{
		BMHControllerCreationRetryInterval: 5 * time.Second,
	}
	runFuncShouldFail := true

	newFn := func(ctx context.Context, bmhConf BMHControllerConf, name string, kubeconfig string) (BMHControllerIF, error) {
		c := &TestBMHController{
			Name:              kubeconfig,
			RunFuncShouldFail: runFuncShouldFail,
		}

		return c, nil
	}

	bmhManager := NewBMHControllerManager(bmhConf, newFn, genName, "server1;server2;server3")
	bmhManager.CreateAllControllers(ctx)
	bmhManager.StartAllControllers(ctx)

	var runnableControllers map[string]BMHControllerIF
	var failedControllers map[string]struct{}
	fmt.Printf("--- step 1 \n")
	notifyCh := make(chan struct{})
	go func() {
		for {
			runnableControllers = bmhManager.GetRunnableControllers()
			failedControllers = bmhManager.GetFailedControllers()
			fmt.Printf("runnableControllers: %+v \n", runnableControllers)
			fmt.Printf("failedControllers: %+v \n", failedControllers)
			if len(runnableControllers) == 0 && len(failedControllers) == 3 {
				notifyCh <- struct{}{}
			} else {
				time.Sleep(1 * time.Second)
				continue
			}
		}
	}()

FOR1:
	for {
		select {
		case <-time.After(30 * time.Second):
			t.Fatalf("timeout waiting for the expected step 1 result")
		case <-notifyCh:
			break FOR1
		}
	}

	fmt.Printf("--- step 2 \n")
	runFuncShouldFail = false
	go func() {
		for {
			runnableControllers = bmhManager.GetRunnableControllers()
			failedControllers = bmhManager.GetFailedControllers()
			fmt.Printf("runnableControllers: %+v \n", runnableControllers)
			fmt.Printf("failedControllers: %+v \n", failedControllers)
			if len(runnableControllers) == 3 && len(failedControllers) == 0 {
				notifyCh <- struct{}{}
			} else {
				time.Sleep(1 * time.Second)
				continue
			}
		}
	}()
FOR2:
	for {
		select {
		case <-time.After(30 * time.Second):
			t.Fatalf("timeout waiting for the expected step 2 result")
		case <-notifyCh:
			break FOR2
		}
	}
}

func NewTestBMHController(ctx context.Context, bmhConf BMHControllerConf, name string, kubeconfig string) (BMHControllerIF, error) {
	c := &TestBMHController{
		Name: kubeconfig,
	}
	if strings.Contains(kubeconfig, "bad") {
		return nil, fmt.Errorf("unable to LoadKubeConfigFile.  Kubeconfig servers: %s", name)
	}
	return c, nil
}

func NewTestBMHControllerAllGood(ctx context.Context, bmhConf BMHControllerConf, name string, kubeconfig string) (BMHControllerIF, error) {
	c := &TestBMHController{
		Name: kubeconfig,
	}
	return c, nil
}

func NewTestBMHControllerAllRunFailed(ctx context.Context, bmhConf BMHControllerConf, name string, kubeconfig string) (BMHControllerIF, error) {
	c := &TestBMHController{
		Name:              kubeconfig,
		RunFuncShouldFail: true,
	}
	return c, nil
}

type TestBMHController struct {
	Name              string
	RunFuncShouldFail bool
}

func (b *TestBMHController) Run(ctx context.Context) error {
	if b.RunFuncShouldFail {
		return fmt.Errorf("failed")
	}
	return nil
}

func genName(input string) string {
	return input
}
