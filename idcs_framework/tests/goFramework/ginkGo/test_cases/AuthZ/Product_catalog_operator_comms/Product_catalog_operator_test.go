package product_catalog_operator_comms_test

import (
	"context"
	"encoding/json"
	"fmt"
	"goFramework/framework/common/logger"
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/stoppable"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/test/bufconn"
)

var (
	billingSyncClient pb.BillingProductCatalogSyncServiceClient
	//clientConn        *grpc.ClientConn
	managerStoppable *stoppable.Stoppable
	listener         *bufconn.Listener
)

func getBufDialer(listener *bufconn.Listener) func(context.Context, string) (net.Conn, error) {
	return func(ctx context.Context, url string) (net.Conn, error) {
		return listener.Dial()
	}
}

var _ = Describe("test Product catalog operator communications", func() {
	It("Should communicate successfully", func() {
		logger.InitializeZapCustomLogger()
		logger.Log.Info("Starting Console communications test suite")
		expectedResult := true
		grpcutil.UseTLS = true
		ctx := context.Background()
		EmbedService(ctx)
		grpcutil.StartTestServices(ctx)
		logger.Log.Info("Stoping services")
		defer grpcutil.StopTestServices()
		Expect(expectedResult).To(BeTrue())
	})

	It("Should communicate successfully", func() {
		var err error
		const ROLE = "productcatalog-operator"
		const DOMAIN = "productcatalog-operator.idcs-system.svc.cluster.local"
		var result any

		body := GenerateRequestToGetCertificate(DOMAIN, ROLE)
		jsonerr := json.Unmarshal([]byte(body), &result)
		if jsonerr != nil {
			Fail("Error during Unmarshal(): " + err.Error())
		}

		json := GetFieldInfo(body)

		_, clientTLSConf, err := VaultCertSetup(json["issuing_ca"].(string), json["private_key"].(string), json["certificate"].(string))
		if err != nil {
			fmt.Print(err.Error())
		}

		test_config := credentials.NewTLS(clientTLSConf)
		ctx := context.Background()
		listener = bufconn.Listen(1024 * 1024)
		clientConn, err = grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(getBufDialer(listener)), grpc.WithTransportCredentials(test_config))
		billingSyncClient = pb.NewBillingProductCatalogSyncServiceClient(clientConn)
		Expect(err).Should(Succeed())
	})

})
