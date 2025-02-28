// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/openapi"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var _ = Describe("GRPC server test", func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	cloudAccountId1 := cloudaccount.MustNewId()
	name1 := "name1-" + uuid.New().String()
	pubKey1 := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC38LFb8lQmcT6KiDuvPu3N9XPvhE5ShbDtxcNtc1AqdsV7MRH7uxYIXDVd0tM80dgEKwyi3IzbNNILGWUkxhV9A3bEnVqPNG7Up6rHdo72uwK1koY3KIlu6BzBBB8QpDvGWUwP4DU84zwC4UJxHEtFL7qnUHKzfNAq8a1WCvABtUW3eaB1SjKzGfNWYR7X8/JwZUCUtTCAFy7gaFrNL7XXDNUzfBzXvlKiyOSbSxWXa51f66sPmPdPtLyveo+3PeruTTvWpCPXJW5zOmeDwbBx899pbc72f1U/KJt6fdwxMSDXdSbARC7ONhD2MRoHjdbbl5QkZbLxdtm/jq393vNCxSP6s8/RDg3+Xp/u0LMjW78JqjKMkKnwWIrSlABQEihM7AlsEKLHDXMredUhT9uXgdu/XOn5Q7zNAkHhOr10p8DpD2HqNkVfcmBf3HhWx6HWS6i7teJiinuAtlVvRD7Xaw8xZM1wue0V/lNF74dE9NKLxSBBhOAvsojr7kZorQLRoULWz8nMBSNoomrUdVVB+UGFdiYS07exYvHemCEXMDxRXl/72Cfkv+DCtkgACjbI2qk8h+kgzsAe9DUmOfxFO1JSDdOevU7eDGMEIDn5REbdsXztgjjdiokANwDTWyTWyDRJMoUhtK0TKM21PHxoEnW32AcHH7LeeHTJiI9sFQ== user1@example.org"

	It("Create should succeed", func() {
		clientConn, err := grpc.Dial(fmt.Sprintf("localhost:%d", grpcListenPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
		Expect(err).NotTo(HaveOccurred())
		client := pb.NewSshPublicKeyServiceClient(clientConn)
		sshPublicKey, err := client.Create(context.Background(), &pb.SshPublicKeyCreateRequest{
			Metadata: &pb.ResourceMetadataCreate{
				CloudAccountId: cloudAccountId1,
				Name:           name1,
			},
			Spec: &pb.SshPublicKeySpec{
				SshPublicKey: pubKey1,
			},
		})
		Expect(err).NotTo(HaveOccurred())
		log.Info("Create should succeed", "sshPublicKey", sshPublicKey)
	})
})

var _ = Describe("GRPC-REST server with net/http client", func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	cloudAccountId1 := cloudaccount.MustNewId()
	name1 := "name1-" + uuid.New().String()
	pubKey1 := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC38LFb8lQmcT6KiDuvPu3N9XPvhE5ShbDtxcNtc1AqdsV7MRH7uxYIXDVd0tM80dgEKwyi3IzbNNILGWUkxhV9A3bEnVqPNG7Up6rHdo72uwK1koY3KIlu6BzBBB8QpDvGWUwP4DU84zwC4UJxHEtFL7qnUHKzfNAq8a1WCvABtUW3eaB1SjKzGfNWYR7X8/JwZUCUtTCAFy7gaFrNL7XXDNUzfBzXvlKiyOSbSxWXa51f66sPmPdPtLyveo+3PeruTTvWpCPXJW5zOmeDwbBx899pbc72f1U/KJt6fdwxMSDXdSbARC7ONhD2MRoHjdbbl5QkZbLxdtm/jq393vNCxSP6s8/RDg3+Xp/u0LMjW78JqjKMkKnwWIrSlABQEihM7AlsEKLHDXMredUhT9uXgdu/XOn5Q7zNAkHhOr10p8DpD2HqNkVfcmBf3HhWx6HWS6i7teJiinuAtlVvRD7Xaw8xZM1wue0V/lNF74dE9NKLxSBBhOAvsojr7kZorQLRoULWz8nMBSNoomrUdVVB+UGFdiYS07exYvHemCEXMDxRXl/72Cfkv+DCtkgACjbI2qk8h+kgzsAe9DUmOfxFO1JSDdOevU7eDGMEIDn5REbdsXztgjjdiokANwDTWyTWyDRJMoUhtK0TKM21PHxoEnW32AcHH7LeeHTJiI9sFQ== user1@example.org"
	marshaler := &runtime.JSONPb{}

	It("GET /readyz should return ok", func() {
		url := fmt.Sprintf("http://localhost:%d/readyz", restListenPort)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		Expect(err).NotTo(HaveOccurred())
		httpClient := http.Client{}
		Eventually(func(g Gomega) {
			response, err := httpClient.Do(req)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(response.StatusCode).Should(Equal(http.StatusOK))
			defer response.Body.Close()
			bodyBytes, err := io.ReadAll(response.Body)
			g.Expect(err).NotTo(HaveOccurred())
			bodyString := string(bodyBytes)
			g.Expect(bodyString).Should(Equal("ok"))
		}, time.Millisecond*1000, time.Millisecond*500).Should(Succeed())
	})

	It("GET /v1/ping should succeed", func() {
		url := fmt.Sprintf("http://localhost:%d/v1/ping", restListenPort)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		Expect(err).NotTo(HaveOccurred())
		httpClient := http.Client{}
		Eventually(func(g Gomega) {
			response, err := httpClient.Do(req)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(response.StatusCode).Should(Equal(http.StatusOK))
			defer response.Body.Close()
			bodyBytes, err := io.ReadAll(response.Body)
			g.Expect(err).NotTo(HaveOccurred())
			bodyString := string(bodyBytes)
			g.Expect(bodyString).Should(Equal("{}"))
		}, time.Millisecond*1000, time.Millisecond*500).Should(Succeed())
	})

	It("POST sshpublickeys should succeed and return the submitted values", func() {
		url := fmt.Sprintf("http://localhost:%d/v1/cloudaccounts/%s/sshpublickeys", restListenPort, cloudAccountId1)
		reqPb := &pb.SshPublicKeyCreateRequest{
			Metadata: &pb.ResourceMetadataCreate{
				Name: name1,
			},
			Spec: &pb.SshPublicKeySpec{
				SshPublicKey: pubKey1,
			},
		}
		log.Info("POST sshpublickeys", "reqPb", reqPb)
		reqJson, err := marshaler.Marshal(reqPb)
		Expect(err).NotTo(HaveOccurred())
		log.Info("POST sshpublickeys", "reqJson", string(reqJson))

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqJson))
		Expect(err).NotTo(HaveOccurred())
		httpClient := http.Client{}
		Eventually(func(g Gomega) {
			response, err := httpClient.Do(req)
			g.Expect(err).Should(Succeed())
			g.Expect(response.StatusCode).Should(Equal(http.StatusOK))
			defer response.Body.Close()
			bodyBytes, err := io.ReadAll(response.Body)
			g.Expect(err).Should(Succeed())
			bodyString := string(bodyBytes)
			log.Info("POST sshpublickeys", "bodyString", bodyString)
			var respPb pb.SshPublicKey
			g.Expect(marshaler.Unmarshal(bodyBytes, &respPb)).Should(Succeed())
			log.Info("POST sshpublickeys", "respPb", &respPb)
			g.Expect(respPb.Metadata.CloudAccountId).Should(Equal(cloudAccountId1))
			g.Expect(respPb.Metadata.ResourceId).ShouldNot(BeEmpty())
			g.Expect(respPb.Metadata.Name).Should(Equal(reqPb.Metadata.Name))
			g.Expect(respPb.Spec.SshPublicKey).Should(Equal(reqPb.Spec.SshPublicKey))
		}, time.Millisecond*1000, time.Millisecond*500).Should(Succeed())
	})

	It("GET sshpublickeys/name/{name} should succeed", func() {
		url := fmt.Sprintf("http://localhost:%d/v1/cloudaccounts/%s/sshpublickeys/name/%s", restListenPort, cloudAccountId1, name1)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		Expect(err).NotTo(HaveOccurred())
		httpClient := http.Client{}
		Eventually(func(g Gomega) {
			response, err := httpClient.Do(req)
			g.Expect(err).Should(Succeed())
			g.Expect(response.StatusCode).Should(Equal(http.StatusOK))
			defer response.Body.Close()
			bodyBytes, err := io.ReadAll(response.Body)
			g.Expect(err).Should(Succeed())
			bodyString := string(bodyBytes)
			log.Info("GET sshpublickeys", "bodyString", bodyString)
			var respPb pb.SshPublicKey
			g.Expect(marshaler.Unmarshal(bodyBytes, &respPb)).Should(Succeed())
			log.Info("GET sshpublickeys", "respPb", &respPb)
			g.Expect(respPb.Metadata.ResourceId).ShouldNot(BeEmpty())
			g.Expect(respPb.Metadata.CloudAccountId).Should(Equal(cloudAccountId1))
			g.Expect(respPb.Metadata.Name).Should(Equal(name1))
			g.Expect(respPb.Spec.SshPublicKey).Should(Equal(pubKey1))
		}, time.Millisecond*1000, time.Millisecond*500).Should(Succeed())
	})
})

var _ = Describe("GRPC-REST server with OpenAPI client", func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	var api *openapi.SshPublicKeyServiceApiService

	BeforeEach(func() {
		clearDatabase(ctx)
		api = openApiClient.SshPublicKeyServiceApi
	})

	createSshPublicKey := func(cloudAccountId string, name string, pubkey string) (*openapi.ProtoSshPublicKey, error) {
		By("SshPublicKeyServiceCreate")
		resp, _, err := api.SshPublicKeyServiceCreate(ctx, cloudAccountId).Body(
			openapi.SshPublicKeyServiceCreateRequest{
				Metadata: &openapi.SshPublicKeyServiceCreateRequestMetadata{
					Name: &name,
				},
				Spec: &openapi.ProtoSshPublicKeySpec{
					SshPublicKey: &pubkey,
				},
			}).Execute()
		return resp, err
	}

	It("Create, get, delete without name should succeed", func() {
		cloudAccountId1 := cloudaccount.MustNewId()
		pubKey1 := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC38LFb8lQmcT6KiDuvPu3N9XPvhE5ShbDtxcNtc1AqdsV7MRH7uxYIXDVd0tM80dgEKwyi3IzbNNILGWUkxhV9A3bEnVqPNG7Up6rHdo72uwK1koY3KIlu6BzBBB8QpDvGWUwP4DU84zwC4UJxHEtFL7qnUHKzfNAq8a1WCvABtUW3eaB1SjKzGfNWYR7X8/JwZUCUtTCAFy7gaFrNL7XXDNUzfBzXvlKiyOSbSxWXa51f66sPmPdPtLyveo+3PeruTTvWpCPXJW5zOmeDwbBx899pbc72f1U/KJt6fdwxMSDXdSbARC7ONhD2MRoHjdbbl5QkZbLxdtm/jq393vNCxSP6s8/RDg3+Xp/u0LMjW78JqjKMkKnwWIrSlABQEihM7AlsEKLHDXMredUhT9uXgdu/XOn5Q7zNAkHhOr10p8DpD2HqNkVfcmBf3HhWx6HWS6i7teJiinuAtlVvRD7Xaw8xZM1wue0V/lNF74dE9NKLxSBBhOAvsojr7kZorQLRoULWz8nMBSNoomrUdVVB+UGFdiYS07exYvHemCEXMDxRXl/72Cfkv+DCtkgACjbI2qk8h+kgzsAe9DUmOfxFO1JSDdOevU7eDGMEIDn5REbdsXztgjjdiokANwDTWyTWyDRJMoUhtK0TKM21PHxoEnW32AcHH7LeeHTJiI9sFQ== user1@example.org"
		resp, err := createSshPublicKey(cloudAccountId1, "", pubKey1)

		Expect(err).Should(Succeed())
		Expect(*resp.Metadata.CloudAccountId).Should(Equal(cloudAccountId1))
		Expect(*resp.Metadata.ResourceId).ShouldNot(BeEmpty())
		Expect(*resp.Metadata.Name).Should(Equal(*resp.Metadata.ResourceId))
		Expect(*resp.Spec.SshPublicKey).Should(Equal(pubKey1))
		resourceId := *resp.Metadata.ResourceId

		By("SshPublicKeyServiceGet")
		getResp, _, err := api.SshPublicKeyServiceGet(ctx, cloudAccountId1, resourceId).Execute()
		Expect(err).Should(Succeed())
		Expect(*getResp.Metadata.CloudAccountId).Should(Equal(cloudAccountId1))
		Expect(*getResp.Metadata.ResourceId).Should(Equal(resourceId))
		Expect(*getResp.Metadata.Name).Should(Equal(resourceId))
		Expect(*getResp.Spec.SshPublicKey).Should(Equal(pubKey1))

		By("SshPublicKeyServiceDelete")
		_, _, err = api.SshPublicKeyServiceDelete(ctx, cloudAccountId1, resourceId).Execute()
		Expect(err).Should(Succeed())

		By("SshPublicKeyServiceGet should fail")
		_, _, err = api.SshPublicKeyServiceGet(ctx, cloudAccountId1, resourceId).Execute()
		Expect(err).ShouldNot(Succeed())
	})

	It("Ensure correct error code is returned if ssh key already exists", func() {
		cloudAccountId1 := cloudaccount.MustNewId()
		pubKey1 := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC38LFb8lQmcT6KiDuvPu3N9XPvhE5ShbDtxcNtc1AqdsV7MRH7uxYIXDVd0tM80dgEKwyi3IzbNNILGWUkxhV9A3bEnVqPNG7Up6rHdo72uwK1koY3KIlu6BzBBB8QpDvGWUwP4DU84zwC4UJxHEtFL7qnUHKzfNAq8a1WCvABtUW3eaB1SjKzGfNWYR7X8/JwZUCUtTCAFy7gaFrNL7XXDNUzfBzXvlKiyOSbSxWXa51f66sPmPdPtLyveo+3PeruTTvWpCPXJW5zOmeDwbBx899pbc72f1U/KJt6fdwxMSDXdSbARC7ONhD2MRoHjdbbl5QkZbLxdtm/jq393vNCxSP6s8/RDg3+Xp/u0LMjW78JqjKMkKnwWIrSlABQEihM7AlsEKLHDXMredUhT9uXgdu/XOn5Q7zNAkHhOr10p8DpD2HqNkVfcmBf3HhWx6HWS6i7teJiinuAtlVvRD7Xaw8xZM1wue0V/lNF74dE9NKLxSBBhOAvsojr7kZorQLRoULWz8nMBSNoomrUdVVB+UGFdiYS07exYvHemCEXMDxRXl/72Cfkv+DCtkgACjbI2qk8h+kgzsAe9DUmOfxFO1JSDdOevU7eDGMEIDn5REbdsXztgjjdiokANwDTWyTWyDRJMoUhtK0TKM21PHxoEnW32AcHH7LeeHTJiI9sFQ== user1@example.org"
		_, err := createSshPublicKey(cloudAccountId1, "test", pubKey1)
		// Creation should be successful.
		Expect(err).ShouldNot(HaveOccurred())
		// Attempt to create an ssh key with name
		_, err = createSshPublicKey(cloudAccountId1, "test", pubKey1)
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).Should(Equal("409 Conflict "))
	})

	It("Create, get, delete with name should succeed", func() {
		cloudAccountId1 := cloudaccount.MustNewId()
		name1 := "name1-" + uuid.New().String()
		pubKey1 := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC38LFb8lQmcT6KiDuvPu3N9XPvhE5ShbDtxcNtc1AqdsV7MRH7uxYIXDVd0tM80dgEKwyi3IzbNNILGWUkxhV9A3bEnVqPNG7Up6rHdo72uwK1koY3KIlu6BzBBB8QpDvGWUwP4DU84zwC4UJxHEtFL7qnUHKzfNAq8a1WCvABtUW3eaB1SjKzGfNWYR7X8/JwZUCUtTCAFy7gaFrNL7XXDNUzfBzXvlKiyOSbSxWXa51f66sPmPdPtLyveo+3PeruTTvWpCPXJW5zOmeDwbBx899pbc72f1U/KJt6fdwxMSDXdSbARC7ONhD2MRoHjdbbl5QkZbLxdtm/jq393vNCxSP6s8/RDg3+Xp/u0LMjW78JqjKMkKnwWIrSlABQEihM7AlsEKLHDXMredUhT9uXgdu/XOn5Q7zNAkHhOr10p8DpD2HqNkVfcmBf3HhWx6HWS6i7teJiinuAtlVvRD7Xaw8xZM1wue0V/lNF74dE9NKLxSBBhOAvsojr7kZorQLRoULWz8nMBSNoomrUdVVB+UGFdiYS07exYvHemCEXMDxRXl/72Cfkv+DCtkgACjbI2qk8h+kgzsAe9DUmOfxFO1JSDdOevU7eDGMEIDn5REbdsXztgjjdiokANwDTWyTWyDRJMoUhtK0TKM21PHxoEnW32AcHH7LeeHTJiI9sFQ== user1@example.org"
		resp, err := createSshPublicKey(cloudAccountId1, name1, pubKey1)
		Expect(err).Should(Succeed())
		Expect(*resp.Metadata.CloudAccountId).Should(Equal(cloudAccountId1))
		Expect(*resp.Metadata.ResourceId).ShouldNot(BeEmpty())
		Expect(*resp.Metadata.Name).Should(Equal(name1))
		Expect(*resp.Spec.SshPublicKey).Should(Equal(pubKey1))
		resourceId := *resp.Metadata.ResourceId

		By("SshPublicKeyServiceGet by name")
		now := time.Now()
		pastLimit := now.Add(-5 * time.Minute)
		futureLimit := now.Add(5 * time.Minute)

		getResp, _, err := api.SshPublicKeyServiceGet(ctx, cloudAccountId1, resourceId).Execute()
		Expect(err).Should(Succeed())
		Expect(*getResp.Metadata.CloudAccountId).Should(Equal(cloudAccountId1))
		Expect(*getResp.Metadata.ResourceId).Should(Equal(resourceId))
		Expect(*getResp.Metadata.Name).Should(Equal(name1))
		Expect(*getResp.Spec.SshPublicKey).Should(Equal(pubKey1))
		Expect(getResp.Metadata.CreationTimestamp.After(pastLimit)).To(BeTrue())
		Expect(getResp.Metadata.CreationTimestamp.Before(futureLimit)).To(BeTrue())

		By("SshPublicKeyServiceGet2 by resourceId")
		getResp, _, err = api.SshPublicKeyServiceGet2(ctx, cloudAccountId1, name1).Execute()
		Expect(err).Should(Succeed())
		Expect(*getResp.Metadata.CloudAccountId).Should(Equal(cloudAccountId1))
		Expect(*getResp.Metadata.ResourceId).Should(Equal(resourceId))
		Expect(*getResp.Metadata.Name).Should(Equal(name1))
		Expect(*getResp.Spec.SshPublicKey).Should(Equal(pubKey1))

		By("SshPublicKeyServiceDelete2 by name")
		_, _, err = api.SshPublicKeyServiceDelete2(ctx, cloudAccountId1, name1).Execute()
		Expect(err).Should(Succeed())

		By("SshPublicKeyServiceGet2 by name should fail")
		_, _, err = api.SshPublicKeyServiceGet2(ctx, cloudAccountId1, name1).Execute()
		Expect(err).ShouldNot(Succeed())
	})

	It("Deleting non-existing SSH public key by name should return 404 Not Found", func() {
		cloudAccountId1 := cloudaccount.MustNewId()
		name1 := "name1-" + uuid.New().String()

		_, resp, err := api.SshPublicKeyServiceDelete2(ctx, cloudAccountId1, name1).Execute()
		Expect(err).ShouldNot(Succeed())
		Expect(resp.StatusCode).Should(Equal(http.StatusNotFound))
	})

	It("SshPublicKeyServiceSearch (non-streaming) should succeed", func() {
		cloudAccountId1 := cloudaccount.MustNewId()
		pubKey1 := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC38LFb8lQmcT6KiDuvPu3N9XPvhE5ShbDtxcNtc1AqdsV7MRH7uxYIXDVd0tM80dgEKwyi3IzbNNILGWUkxhV9A3bEnVqPNG7Up6rHdo72uwK1koY3KIlu6BzBBB8QpDvGWUwP4DU84zwC4UJxHEtFL7qnUHKzfNAq8a1WCvABtUW3eaB1SjKzGfNWYR7X8/JwZUCUtTCAFy7gaFrNL7XXDNUzfBzXvlKiyOSbSxWXa51f66sPmPdPtLyveo+3PeruTTvWpCPXJW5zOmeDwbBx899pbc72f1U/KJt6fdwxMSDXdSbARC7ONhD2MRoHjdbbl5QkZbLxdtm/jq393vNCxSP6s8/RDg3+Xp/u0LMjW78JqjKMkKnwWIrSlABQEihM7AlsEKLHDXMredUhT9uXgdu/XOn5Q7zNAkHhOr10p8DpD2HqNkVfcmBf3HhWx6HWS6i7teJiinuAtlVvRD7Xaw8xZM1wue0V/lNF74dE9NKLxSBBhOAvsojr7kZorQLRoULWz8nMBSNoomrUdVVB+UGFdiYS07exYvHemCEXMDxRXl/72Cfkv+DCtkgACjbI2qk8h+kgzsAe9DUmOfxFO1JSDdOevU7eDGMEIDn5REbdsXztgjjdiokANwDTWyTWyDRJMoUhtK0TKM21PHxoEnW32AcHH7LeeHTJiI9sFQ== user1@example.org"

		By("SshPublicKeyServiceCreate")
		numRows := 10
		for i := 0; i < numRows; i++ {
			_, err := createSshPublicKey(cloudAccountId1, "", pubKey1)
			Expect(err).Should(Succeed())
		}

		By("SshPublicKeyServiceSearch")
		resp, _, err := api.SshPublicKeyServiceSearch(ctx, cloudAccountId1).Execute()
		log.Info("Search", "resp", resp)
		Expect(err).Should(Succeed())
		Expect(len(resp.Items)).Should(Equal(numRows))
	})

	It("SshPublicKey with a missing character should fail", func() {
		cloudAccountId1 := cloudaccount.MustNewId()
		pubKey1 := "ssh-r AAAAB3NzaC1yc2EAAAADAQABAAACAQC38LFb8lQmcT6KiDuvPu3N9XPvhE5ShbDtxcNtc1AqdsV7MRH7uxYIXDVd0tM80dgEKwyi3IzbNNILGWUkxhV9A3bEnVqPNG7Up6rHdo72uwK1koY3KIlu6BzBBB8QpDvGWUwP4DU84zwC4UJxHEtFL7qnUHKzfNAq8a1WCvABtUW3eaB1SjKzGfNWYR7X8/JwZUCUtTCAFy7gaFrNL7XXDNUzfBzXvlKiyOSbSxWXa51f66sPmPdPtLyveo+3PeruTTvWpCPXJW5zOmeDwbBx899pbc72f1U/KJt6fdwxMSDXdSbARC7ONhD2MRoHjdbbl5QkZbLxdtm/jq393vNCxSP6s8/RDg3+Xp/u0LMjW78JqjKMkKnwWIrSlABQEihM7AlsEKLHDXMredUhT9uXgdu/XOn5Q7zNAkHhOr10p8DpD2HqNkVfcmBf3HhWx6HWS6i7teJiinuAtlVvRD7Xaw8xZM1wue0V/lNF74dE9NKLxSBBhOAvsojr7kZorQLRoULWz8nMBSNoomrUdVVB+UGFdiYS07exYvHemCEXMDxRXl/72Cfkv+DCtkgACjbI2qk8h+kgzsAe9DUmOfxFO1JSDdOevU7eDGMEIDn5REbdsXztgjjdiokANwDTWyTWyDRJMoUhtK0TKM21PHxoEnW32AcHH7LeeHTJiI9sFQ== user1@example.org"

		_, err := createSshPublicKey(cloudAccountId1, "", pubKey1)
		By("SshPublicKeyServiceCreate")
		Expect(err).ShouldNot(Succeed())
	})

	It("Empty SshPublicKey should fail", func() {
		cloudAccountId1 := cloudaccount.MustNewId()
		pubKey1 := ""

		_, err := createSshPublicKey(cloudAccountId1, "", pubKey1)
		By("SshPublicKeyServiceCreate")
		Expect(err).ShouldNot(Succeed())
	})

	It("Empty SshPublicKey with valid ssh algorithm should fail", func() {
		cloudAccountId1 := cloudaccount.MustNewId()
		pubKey1 := "ssh-rsa"

		_, err := createSshPublicKey(cloudAccountId1, "", pubKey1)
		By("SshPublicKeyServiceCreate")
		Expect(err).ShouldNot(Succeed())
	})

	It("Create SshPublicKey should succeed when valid sshpublickey name contains -, ., @ and lower alphanumeric characters", func() {
		cloudAccountId1 := cloudaccount.MustNewId()
		pubKey1 := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC38LFb8lQmcT6KiDuvPu3N9XPvhE5ShbDtxcNtc1AqdsV7MRH7uxYIXDVd0tM80dgEKwyi3IzbNNILGWUkxhV9A3bEnVqPNG7Up6rHdo72uwK1koY3KIlu6BzBBB8QpDvGWUwP4DU84zwC4UJxHEtFL7qnUHKzfNAq8a1WCvABtUW3eaB1SjKzGfNWYR7X8/JwZUCUtTCAFy7gaFrNL7XXDNUzfBzXvlKiyOSbSxWXa51f66sPmPdPtLyveo+3PeruTTvWpCPXJW5zOmeDwbBx899pbc72f1U/KJt6fdwxMSDXdSbARC7ONhD2MRoHjdbbl5QkZbLxdtm/jq393vNCxSP6s8/RDg3+Xp/u0LMjW78JqjKMkKnwWIrSlABQEihM7AlsEKLHDXMredUhT9uXgdu/XOn5Q7zNAkHhOr10p8DpD2HqNkVfcmBf3HhWx6HWS6i7teJiinuAtlVvRD7Xaw8xZM1wue0V/lNF74dE9NKLxSBBhOAvsojr7kZorQLRoULWz8nMBSNoomrUdVVB+UGFdiYS07exYvHemCEXMDxRXl/72Cfkv+DCtkgACjbI2qk8h+kgzsAe9DUmOfxFO1JSDdOevU7eDGMEIDn5REbdsXztgjjdiokANwDTWyTWyDRJMoUhtK0TKM21PHxoEnW32AcHH7LeeHTJiI9sFQ== user1@example.org"
		name1 := "name@host"
		_, err := createSshPublicKey(cloudAccountId1, name1, pubKey1)
		Expect(err).Should(Succeed())

		name2 := "name.host.com"
		_, err = createSshPublicKey(cloudAccountId1, name2, pubKey1)
		Expect(err).Should(Succeed())

		name3 := "name3-host-com"
		_, err = createSshPublicKey(cloudAccountId1, name3, pubKey1)
		Expect(err).Should(Succeed())

		name4 := "name4@domain.com"
		_, err = createSshPublicKey(cloudAccountId1, name4, pubKey1)
		Expect(err).Should(Succeed())

		name5 := "name5-" + uuid.New().String()
		_, err = createSshPublicKey(cloudAccountId1, name5, pubKey1)
		Expect(err).Should(Succeed())
	})

	It("Create SshPublicKey should fail when sshpublickey name is invalid", func() {
		cloudAccountId1 := cloudaccount.MustNewId()
		pubKey1 := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC38LFb8lQmcT6KiDuvPu3N9XPvhE5ShbDtxcNtc1AqdsV7MRH7uxYIXDVd0tM80dgEKwyi3IzbNNILGWUkxhV9A3bEnVqPNG7Up6rHdo72uwK1koY3KIlu6BzBBB8QpDvGWUwP4DU84zwC4UJxHEtFL7qnUHKzfNAq8a1WCvABtUW3eaB1SjKzGfNWYR7X8/JwZUCUtTCAFy7gaFrNL7XXDNUzfBzXvlKiyOSbSxWXa51f66sPmPdPtLyveo+3PeruTTvWpCPXJW5zOmeDwbBx899pbc72f1U/KJt6fdwxMSDXdSbARC7ONhD2MRoHjdbbl5QkZbLxdtm/jq393vNCxSP6s8/RDg3+Xp/u0LMjW78JqjKMkKnwWIrSlABQEihM7AlsEKLHDXMredUhT9uXgdu/XOn5Q7zNAkHhOr10p8DpD2HqNkVfcmBf3HhWx6HWS6i7teJiinuAtlVvRD7Xaw8xZM1wue0V/lNF74dE9NKLxSBBhOAvsojr7kZorQLRoULWz8nMBSNoomrUdVVB+UGFdiYS07exYvHemCEXMDxRXl/72Cfkv+DCtkgACjbI2qk8h+kgzsAe9DUmOfxFO1JSDdOevU7eDGMEIDn5REbdsXztgjjdiokANwDTWyTWyDRJMoUhtK0TKM21PHxoEnW32AcHH7LeeHTJiI9sFQ== user1@example.org"
		name1 := "name1$host"
		_, err := createSshPublicKey(cloudAccountId1, name1, pubKey1)
		Expect(err).ShouldNot(Succeed())

		name2 := "name1@Host"
		_, err = createSshPublicKey(cloudAccountId1, name2, pubKey1)
		Expect(err).ShouldNot(Succeed())

		name3 := "test/user@domain.com"
		_, err = createSshPublicKey(cloudAccountId1, name3, pubKey1)
		Expect(err).ShouldNot(Succeed())
	})

	It("Create SshPublicKey should fail when sshpublickey name exceeds 63 characters", func() {
		cloudAccountId1 := cloudaccount.MustNewId()
		pubKey1 := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC38LFb8lQmcT6KiDuvPu3N9XPvhE5ShbDtxcNtc1AqdsV7MRH7uxYIXDVd0tM80dgEKwyi3IzbNNILGWUkxhV9A3bEnVqPNG7Up6rHdo72uwK1koY3KIlu6BzBBB8QpDvGWUwP4DU84zwC4UJxHEtFL7qnUHKzfNAq8a1WCvABtUW3eaB1SjKzGfNWYR7X8/JwZUCUtTCAFy7gaFrNL7XXDNUzfBzXvlKiyOSbSxWXa51f66sPmPdPtLyveo+3PeruTTvWpCPXJW5zOmeDwbBx899pbc72f1U/KJt6fdwxMSDXdSbARC7ONhD2MRoHjdbbl5QkZbLxdtm/jq393vNCxSP6s8/RDg3+Xp/u0LMjW78JqjKMkKnwWIrSlABQEihM7AlsEKLHDXMredUhT9uXgdu/XOn5Q7zNAkHhOr10p8DpD2HqNkVfcmBf3HhWx6HWS6i7teJiinuAtlVvRD7Xaw8xZM1wue0V/lNF74dE9NKLxSBBhOAvsojr7kZorQLRoULWz8nMBSNoomrUdVVB+UGFdiYS07exYvHemCEXMDxRXl/72Cfkv+DCtkgACjbI2qk8h+kgzsAe9DUmOfxFO1JSDdOevU7eDGMEIDn5REbdsXztgjjdiokANwDTWyTWyDRJMoUhtK0TKM21PHxoEnW32AcHH7LeeHTJiI9sFQ== user1@example.org"
		name1 := "ssh-public-key-name-exceeding-63-characters-functional@testing.com"
		_, err := createSshPublicKey(cloudAccountId1, name1, pubKey1)
		Expect(err).ShouldNot(Succeed())
	})

	It("Create SshPublicKey should fail when cloudaccountid is invalid", func() {
		cloudAccountId1 := "12345678901245"
		name1 := "name1-" + uuid.New().String()
		pubKey1 := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC38LFb8lQmcT6KiDuvPu3N9XPvhE5ShbDtxcNtc1AqdsV7MRH7uxYIXDVd0tM80dgEKwyi3IzbNNILGWUkxhV9A3bEnVqPNG7Up6rHdo72uwK1koY3KIlu6BzBBB8QpDvGWUwP4DU84zwC4UJxHEtFL7qnUHKzfNAq8a1WCvABtUW3eaB1SjKzGfNWYR7X8/JwZUCUtTCAFy7gaFrNL7XXDNUzfBzXvlKiyOSbSxWXa51f66sPmPdPtLyveo+3PeruTTvWpCPXJW5zOmeDwbBx899pbc72f1U/KJt6fdwxMSDXdSbARC7ONhD2MRoHjdbbl5QkZbLxdtm/jq393vNCxSP6s8/RDg3+Xp/u0LMjW78JqjKMkKnwWIrSlABQEihM7AlsEKLHDXMredUhT9uXgdu/XOn5Q7zNAkHhOr10p8DpD2HqNkVfcmBf3HhWx6HWS6i7teJiinuAtlVvRD7Xaw8xZM1wue0V/lNF74dE9NKLxSBBhOAvsojr7kZorQLRoULWz8nMBSNoomrUdVVB+UGFdiYS07exYvHemCEXMDxRXl/72Cfkv+DCtkgACjbI2qk8h+kgzsAe9DUmOfxFO1JSDdOevU7eDGMEIDn5REbdsXztgjjdiokANwDTWyTWyDRJMoUhtK0TKM21PHxoEnW32AcHH7LeeHTJiI9sFQ== user1@example.org"
		_, err := createSshPublicKey(cloudAccountId1, name1, pubKey1)
		Expect(err).ShouldNot(Succeed())

		cloudAccountId1 = "123456789"
		_, err = createSshPublicKey(cloudAccountId1, name1, pubKey1)
		Expect(err).ShouldNot(Succeed())

		cloudAccountId1 = "12345678test"
		_, err = createSshPublicKey(cloudAccountId1, name1, pubKey1)
		Expect(err).ShouldNot(Succeed())
	})

	It("Create SshPublicKey should fail when key length is too short", func() {
		cloudAccountId1 := cloudaccount.MustNewId()
		name1 := "name1-" + uuid.New().String()
		// Here `pubKey1` size is 1024 bits
		pubKey1 := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAgQC258wRRUY0lwxVE5qiQAyM3Wi5rAGMBn6ZkHy1TPShEhwJD5u9t0sfeg3wWN6YkBXXsfOawc+bCsMEeTCfyHwCRaLgLQNm46jFF4sHFijPhWfQCOOyfEyXG91zvd/CY0nS6hiDS4NPOhBaperZ2TM2vuOTNs3A6WZB3HFyvvexjw== user1@example.org"
		_, err := createSshPublicKey(cloudAccountId1, name1, pubKey1)
		Expect(err).ShouldNot(Succeed())
	})
})
