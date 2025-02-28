package api_test

import (
	"goFramework/framework/service_api/training"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

var _ = Describe("TrainingClusterService", Ordered, Label("TrainingClusterService"), func() {
	It("should list no clusters for cloud account with no clusters", func() {
		status, clusters := training.ListClusters(baseApiUrl, premiumCloudAccount["token"], premiumCloudAccount["id"])
		Expect(status).To(Equal(200), "Failed to fetch the list of clusters")
		Expect(clusters).To(Equal(`{"clusters":[]}`), "Received cluster response other than empty array")
	})

	It("should list an existing cluster for the slurmaas management account", func() {
		status, clusters := training.ListClusters(baseApiUrl, adminToken, slurmaasManagementCloudAccountId)
		Expect(status).To(Equal(200), "Failed to fetch the list of clusters")
		clusterList := gjson.Get(clusters, "clusters").Array()
		Expect(len(clusterList)).To(Equal(1), "Received cluster response with length of not 1")
	})
})
