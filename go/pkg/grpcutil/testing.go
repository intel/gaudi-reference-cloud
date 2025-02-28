// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package grpcutil

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"google.golang.org/grpc"
)

// Optional Hosts method for services that want to register with
// multiple hostnames
type ServiceMultiHost interface {
	Hosts() []string
}

func ServiceHosts[C Config](svc Service[C]) []string {
	if mh, ok := svc.(ServiceMultiHost); ok {
		return mh.Hosts()
	}
	return []string{svc.Name()}
}

type tTestService struct {
	name       string
	hosts      []string
	config     Config
	grpcServer *grpc.Server
	listener   net.Listener
	init       func(context.Context, Resolver, *grpc.Server) error
	done       func()
}

var (
	mutex          sync.Mutex
	services       []*tTestService
	doneChan       = make(chan struct{})
	servicesByName = map[string]*tTestService{}
)

func AddTestService[C Config](service Service[C], config C) {
	mutex.Lock()
	defer mutex.Unlock()
	name := service.Name()
	if _, found := servicesByName[name]; found {
		// already have this service
		return
	}

	services = append(services,
		&tTestService{
			config: config,
			name:   name,
			hosts:  ServiceHosts(service),
			// The init and done funcs store type information in
			// their closures. We can't directly put service into a slice
			// because all the elements of a slice must have the same
			// type
			init: func(ctx context.Context, resolver Resolver, grpcServer *grpc.Server) error {
				return service.Init(ctx, config, resolver, grpcServer)
			},
			done: func() {
				ServiceDone(service)
			},
		})
}

func StartTestServices(ctx context.Context) {
	resolver := &tTestResolver{}
	logger := log.FromContext(ctx).WithName("StartTestServices")

	// Set up the listeners and configure the resolver first.
	for _, svc := range services {
		listener, err := net.Listen("tcp", "localhost:")
		if err != nil {
			logger.Error(err, "list failed", "service", svc.name)
			os.Exit(1)
		}
		svc.listener = listener
		addr := listener.Addr().String()
		colon := strings.LastIndex(addr, ":")
		if colon == -1 {
			logger.Error(fmt.Errorf("no port in listener address %v", addr), "failed")
			os.Exit(1)
		}
		port, err := strconv.ParseInt(addr[colon+1:], 10, 32)
		if err != nil {
			logger.Error(err, "error parsing port in listeneraddress", "addr", addr)
			os.Exit(1)
		}
		for _, host := range svc.hosts {
			resolver.AddAddr(host, addr)
		}
		svc.config.SetListenPort(uint16(port))
		svc.grpcServer = grpc.NewServer()
	}

	// Now that listeners are created and the resolver has all the
	// hostnames, call the service init methods. The services can find
	// and connect to each other using the addresses stored in the resolver.
	for _, svc := range services {
		if err := svc.init(ctx, resolver, svc.grpcServer); err != nil {
			logger.Error(err, "init error", "service", svc.name)
			os.Exit(1)
		}
	}

	go func() {
		wg := sync.WaitGroup{}
		wg.Add(len(services))
		// Run each service in its own goroutine
		for _, svc := range services {
			svc := svc
			go func() {
				defer wg.Done()
				if err := svc.grpcServer.Serve(svc.listener); err != nil {
					logger := log.FromContext(ctx).WithName("RunTestServices")
					logger.Error(err, "error serving grpc", "service", svc.name)
				}
			}()
		}
		wg.Wait()
		for _, svc := range services {
			svc.done()
		}
		close(doneChan)
	}()
}

func StopTestServices() {
	for _, svc := range services {
		svc.grpcServer.GracefulStop()
	}
	<-doneChan
}

type tTestResolver struct {
	mutex sync.Mutex
	addrs map[string]string
}

func (tr *tTestResolver) Resolve(ctx context.Context, name string) (string, error) {
	tr.mutex.Lock()
	defer tr.mutex.Unlock()
	if tr.addrs == nil {
		return "", fmt.Errorf("%v not found", name)
	}
	addr := tr.addrs[name]
	if addr == "" {
		return "", fmt.Errorf("%v not found", name)
	}
	return addr, nil
}

func (tr *tTestResolver) AddAddr(name string, addr string) {
	tr.mutex.Lock()
	defer tr.mutex.Unlock()
	if tr.addrs == nil {
		tr.addrs = make(map[string]string)
	}
	tr.addrs[name] = addr
}
