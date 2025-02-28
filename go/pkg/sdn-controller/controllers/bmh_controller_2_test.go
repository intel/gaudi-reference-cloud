package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	baremetalv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/metal3.io/v1alpha1"
	idcnetworkv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	"io/ioutil"
)

var testdataPath = "bmh-exports-testdata/" // Must end with a trailing /

var _ = Describe("bmhController", func() {

	It("findSwitchPortsFromBMH loads correct BMHs from bmh-exports", Label("bmhs"), func() {
		//func TestFindSwitchPortsFromBMH(t *testing.T) {
		ctx := context.Background()

		bmhController := &BMHController{}
		bmhController.conf = idcnetworkv1alpha1.SDNControllerConfig{
			ControllerConfig: idcnetworkv1alpha1.ControllerConfig{
				AllowedCountAccInterfaces: "0,24",
			},
		}

		files, err := ioutil.ReadDir(testdataPath)
		Expect(err).NotTo(HaveOccurred())

		for _, file := range files {
			if file.IsDir() {
				continue
			}
			fmt.Printf("reading %v... \n", file.Name())

			// Import all of the BMH from exported.yaml
			yamlData, err := ioutil.ReadFile(testdataPath + file.Name())
			Expect(err).NotTo(HaveOccurred())

			// Convert YAML to JSON because BMH object has `"json"` annotations, not yaml ones.
			jsonData, err := yaml.YAMLToJSON(yamlData)
			Expect(err).NotTo(HaveOccurred())

			// Unmarshal JSON into BMH structs
			var bmhs baremetalv1alpha1.BareMetalHostList
			err = json.Unmarshal(jsonData, &bmhs)
			Expect(err).NotTo(HaveOccurred())

			fmt.Printf("found %v BMHs in file. \n", len(bmhs.Items))

			for i, bmh := range bmhs.Items {
				fmt.Printf("finding switchports from file %v BMH #%v (%v)... \n", file.Name(), i, bmh.Name)
				frontendSwitchports, accSwitchports, storageSwitchports, err := bmhController.findSwitchPortsFromBMH(ctx, &bmh)
				if strings.Contains(file.Name(), "bad") {
					Expect(err).To(HaveOccurred())
				} else {
					Expect(err).NotTo(HaveOccurred())
					Expect(frontendSwitchports.SwitchFQDN).ToNot(BeNil())

					Expect(accSwitchports).To(Or(HaveLen(0), HaveLen(24)))
					Expect(storageSwitchports).To(Or(HaveLen(0), HaveLen(1)))
				}
			}

		}
	})

	It("findSwitchPortsFromBMH does not import NetworkNodes from BMH in the middle of enrollment", Label("bmhs"), func() {
		//func TestFindSwitchPortsFromBMH(t *testing.T) {
		ctx := context.Background()

		bmhController := &BMHController{}
		bmhController.conf = idcnetworkv1alpha1.SDNControllerConfig{
			ControllerConfig: idcnetworkv1alpha1.ControllerConfig{
				AllowedCountAccInterfaces: "0,24",
			},
		}

		files, err := ioutil.ReadDir(testdataPath + "bmh-during-enrollment/")
		Expect(err).NotTo(HaveOccurred())

		for _, file := range files {
			if file.IsDir() {
				continue
			}
			fmt.Printf("reading %v... \n", file.Name())

			// Import all of the BMH from exported.yaml
			yamlData, err := ioutil.ReadFile(testdataPath + "bmh-during-enrollment/" + file.Name())
			Expect(err).NotTo(HaveOccurred())

			// Convert YAML to JSON because BMH object has `"json"` annotations, not yaml ones.
			jsonData, err := yaml.YAMLToJSON(yamlData)
			Expect(err).NotTo(HaveOccurred())

			// Unmarshal JSON into BMH structs
			var bmhs baremetalv1alpha1.BareMetalHostList
			err = json.Unmarshal(jsonData, &bmhs)
			Expect(err).NotTo(HaveOccurred())

			fmt.Printf("found %v BMHs in file. \n", len(bmhs.Items))

			for i, bmh := range bmhs.Items {
				fmt.Printf("finding switchports from file %v BMH #%v (%v)... \n", file.Name(), i, bmh.Name)
				frontendSwitchports, accSwitchports, storageSwitchports, err := bmhController.findSwitchPortsFromBMH(ctx, &bmh)
				if strings.Contains(file.Name(), "bad") {
					Expect(err).To(HaveOccurred())
				} else {
					Expect(err).NotTo(HaveOccurred())
					Expect(frontendSwitchports.SwitchFQDN).ToNot(BeNil())

					Expect(accSwitchports).To(Or(HaveLen(0), HaveLen(24)))
					Expect(storageSwitchports).To(Or(HaveLen(0), HaveLen(1)))
				}
			}

		}
	})
})
