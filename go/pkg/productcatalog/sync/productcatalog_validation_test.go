// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package sync

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"unicode/utf8"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
)

type Family struct {
	Name        string `yaml:"name"`
	ID          string `yaml:"id"`
	Description string `yaml:"description"`
}

type Rate struct {
	AccountType string `yaml:"accountType"`
	Unit        string `yaml:"unit"`
	UsageExpr   string `yaml:"usageExpr"`
	Rate        string `yaml:"rate"`
}

// The Spec struct is utilized for both items (products and vendors, families are part of vendors).
// Any values that are not applicable to the respective item will remain empty and will not be utilized.
type Spec struct {
	ID          string   `yaml:"id"`
	Description string   `yaml:"description"`
	VendorID    string   `yaml:"vendorId"`  // Only for products.
	FamilyID    string   `yaml:"familyId"`  // Only for products.
	PCQ         string   `yaml:"pcq"`       // Only for products.
	ECCN        string   `yaml:"eccn"`      // Only for products.
	MatchExpr   string   `yaml:"matchExpr"` // Only for products.
	Rates       []Rate   `yaml:"rates"`     // Only for products.
	Families    []Family `yaml:"families"`  // Only for vendors.
}

type Metadata struct {
	Name string `yaml:"name"`
}

type File struct {
	Kind     string   `yaml:"kind"`
	Metadata Metadata `yaml:"metadata"`
	Spec     Spec     `yaml:"spec"`
}

// Could be products or vendors.
type Item struct {
	Type  string
	Files map[string]File
}

type Environment struct {
	Name  string
	Items map[string]Item
}

const (
	// Folders names.
	dev      = "dev"
	staging  = "staging"
	prod     = "prod"
	products = "products"
	vendors  = "vendors"

	// Shared patterns that validates presence and format of item values.
	uuidPattern        = "^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$"
	customIdPattern    = "^[0-9a-zA-Z]{8}-[0-9a-zA-Z]{4}-[0-9a-zA-Z]{4}-[0-9a-zA-Z]{4}-[0-9a-zA-Z]{12}$"
	namePattern        = "^[a-z0-9]+([-.][a-z0-9]+)*$"
	descriptionPattern = "^[\\w\\s\\pP\\pS]+$"
)

func readYAMLFiles(dir string) (map[string]File, error) {
	results, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	files := make(map[string]File)
	for _, result := range results {
		if result.IsDir() {
			continue
		}
		filePath := filepath.Join(dir, result.Name())
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		var file File
		err = yaml.Unmarshal(data, &file)
		if err != nil {
			return nil, err
		}
		files[result.Name()] = file
	}

	return files, nil
}

func familiesToMap(families []Family) map[string]Family {
	familyMap := make(map[string]Family)
	for _, family := range families {
		familyMap[family.ID] = family
	}
	return familyMap
}

func compareFamilies(families1, families2 []Family, env1, env2, filename string) {
	if len(families1) != len(families2) {
		fmt.Fprintf(GinkgoWriter, "*** Warning: Mismatch in number of families in %s: %s vs %s\n", filename, env1, env2)
	}
	for _, family1 := range families1 {
		found := false
		for _, family2 := range families2 {
			if family1.ID == family2.ID || family1.Name == family2.Name {
				found = true
				Expect(family1.ID).To(Equal(family2.ID), fmt.Sprintf("Mismatch in family ID in %s: %s ID = %s, %s ID = %s", filename, env1, family1.ID, env2, family2.ID))
				Expect(family1.Name).To(Equal(family2.Name), fmt.Sprintf("Mismatch in family Name in %s: %s Name = %s, %s Name = %s", filename, env1, family1.Name, env2, family2.Name))
				break
			}
		}
		if !found {
			fmt.Fprintf(GinkgoWriter, "*** Warning: Family with ID %s and Name %s found in %s but not in %s\n", family1.ID, family1.Name, env1, env2)
		}
	}
}

func compareItems(item1, item2 Item, env1, env2 string) {
	for filename, file1 := range item1.Files {
		if file2, exists := item2.Files[filename]; exists {
			Expect(file1.Spec.ID).To(Equal(file2.Spec.ID), fmt.Sprintf("Mismatch in %s: %s ID = %s, %s ID = %s", filename, env1, file1.Spec.ID, env2, file2.Spec.ID))
			Expect(file1.Metadata.Name).To(Equal(file2.Metadata.Name), fmt.Sprintf("Mismatch in %s: %s Name = %s, %s Name = %s", filename, env1, file1.Metadata.Name, env2, file2.Metadata.Name))
			if item1.Type == vendors {
				fmt.Printf("Comparing families..\n")
				compareFamilies(file1.Spec.Families, file2.Spec.Families, env1, env2, filename)
			}
		} else {
			fmt.Fprintf(GinkgoWriter, "*** Warning: File %s exists in %s but not in %s\n", filename, env1, env2)
		}
	}
	for filename, _ := range item2.Files {
		if _, exists := item1.Files[filename]; !exists {
			fmt.Fprintf(GinkgoWriter, "*** Warning: File %s exists in %s but not in %s\n", filename, env2, env1)
		}
	}
}

func compareEnvs(env1, env2 Environment) {
	for _, item1 := range env1.Items {
		item2 := env2.Items[item1.Type]
		fmt.Printf("Comparing %s..\n", item1.Type)
		compareItems(item1, item2, env1.Name, env2.Name)
	}
}

func checkUniqueIDsAndNames(env Environment) {
	for _, item := range env.Items {
		fmt.Printf("Comparing %s in %s..\n", item.Type, env.Name)
		ids := make(map[string]bool)
		names := make(map[string]bool)
		for filename, file := range item.Files {
			if ids[file.Spec.ID] {
				Fail(fmt.Sprintf("Duplicate ID found in %s: %s ID = %s", item.Type, filename, file.Spec.ID))
			}
			ids[file.Spec.ID] = true

			if names[file.Metadata.Name] {
				Fail(fmt.Sprintf("Duplicate Name found in %s: %s Name = %s", item.Type, filename, file.Metadata.Name))
			}
			names[file.Metadata.Name] = true

			if item.Type == vendors {
				fmt.Printf("Comparing families in %s..\n", env.Name)
				familyIDs := make(map[string]bool)
				familyNames := make(map[string]bool)
				for _, family := range file.Spec.Families {
					if familyIDs[family.ID] {
						Fail(fmt.Sprintf("Duplicate Family ID found in %s: %s Family ID = %s", item.Type, filename, family.ID))
					}
					familyIDs[family.ID] = true

					if familyNames[family.Name] {
						Fail(fmt.Sprintf("Duplicate Family Name found in %s: %s Family Name = %s", item.Type, filename, family.Name))
					}
					familyNames[family.Name] = true
				}
			}
		}
	}
}

func checkVendorValues(item Item) {
	for filename, file := range item.Files {
		validationMessage := fmt.Sprintf("Validation failed for filename: %s\n", filename)
		// Validate vendor values.
		Expect(file.Metadata.Name).To(MatchRegexp(namePattern), validationMessage)
		Expect(file.Spec).NotTo(BeNil(), validationMessage)
		Expect(file.Spec.ID).To(MatchRegexp(uuidPattern), validationMessage)
		Expect(file.Spec.Description).To(MatchRegexp(descriptionPattern), validationMessage)
		Expect(len(file.Spec.Families)).To(BeNumerically(">=", 1), validationMessage)
		// Validate families values.
		for _, family := range file.Spec.Families {
			Expect(family.Name).To(MatchRegexp(namePattern), validationMessage)
			Expect(family.ID).To(MatchRegexp(uuidPattern), validationMessage)
			Expect(family.Description).To(MatchRegexp(descriptionPattern), validationMessage)
		}
	}
}

func checkProductValues(item Item) {
	for filename, file := range item.Files {
		validationMessage := fmt.Sprintf("Validation failed for filename: %s\n", filename)
		// Validate products values.
		Expect(file.Metadata.Name).To(MatchRegexp(namePattern), validationMessage)
		Expect(file.Spec).NotTo(BeNil(), validationMessage)
		Expect(file.Spec.ID).To(MatchRegexp(customIdPattern), validationMessage)
		Expect(file.Spec.VendorID).To(MatchRegexp(uuidPattern), validationMessage)
		Expect(file.Spec.FamilyID).To(MatchRegexp(uuidPattern), validationMessage)
		Expect(file.Spec.Description).To(MatchRegexp(descriptionPattern), validationMessage)
		Expect(utf8.RuneCountInString(file.Spec.Description)).To(BeNumerically("<=", 99), validationMessage)
		Expect(len(file.Spec.MatchExpr)).To(BeNumerically(">=", 1), validationMessage)
		Expect(len(file.Spec.Rates)).To(BeNumerically("==", 4), validationMessage)
		// Validate rates values.
		for _, rate := range file.Spec.Rates {
			Expect(rate.AccountType).To(MatchRegexp("^(standard|premium|enterprise|intel)$"), validationMessage)
			Expect(rate.Unit).To(MatchRegexp("^(dollarsPerMinute|dollarsPerTBPerHour|dollarsPerInference|dollarsPerMillionTokens)$"), validationMessage)
			Expect(rate.Rate).To(MatchRegexp("^[0-9]+(\\.[0-9]+)?$"), validationMessage)
			Expect(rate.UsageExpr).To(MatchRegexp("^(time – previous\\.time)$"), validationMessage)
		}
	}
}

func checkValues(env Environment) {
	for _, item := range env.Items {
		fmt.Printf("Checking %s in %s..\n", item.Type, env.Name)
		if item.Type == vendors {
			checkVendorValues(item)
		}
		if item.Type == products {
			checkProductValues(item)
		}
	}
}

var _ = Describe("Product Catalog Validation", Ordered, func() {
	var environments = []string{dev, staging, prod}
	var items = []string{vendors, products}
	var catalog = make(map[string]Environment)

	BeforeAll(func() {
		productCatalogPath := os.Getenv("PRODUCTCATALOG_PATH")
		Expect(productCatalogPath).NotTo(BeEmpty())

		for _, env := range environments {
			var groups = make(map[string]Item)
			for _, item := range items {
				dir := filepath.Join(productCatalogPath, env, item)
				files, err := readYAMLFiles(dir)
				Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Error reading %s %s: %v", env, item, err))
				groups[item] = Item{Type: item, Files: files}
			}
			catalog[env] = Environment{Name: env, Items: groups}
		}
	})

	Context("Cross-env validation of items ids and names", Ordered, func() {

		It(fmt.Sprintf("ids and names of all items (vendors, families and products) in %s and %s must match", dev, staging), func() {
			compareEnvs(catalog[dev], catalog[staging])
		})

		It(fmt.Sprintf("ids and names of all items (vendors, families and products) in %s and %s must match", staging, prod), func() {
			compareEnvs(catalog[staging], catalog[prod])
		})
	})

	Context("Unique ids and names check per env", Ordered, func() {

		It(fmt.Sprintf("all items (vendors, families and products) in %s should have unique ids and names", dev), func() {
			checkUniqueIDsAndNames(catalog[dev])
		})

		It(fmt.Sprintf("all items (vendors, families and products) in %s should have unique ids and names", staging), func() {
			checkUniqueIDsAndNames(catalog[staging])
		})

		It(fmt.Sprintf("all items (vendors, families and products) in %s should have unique ids and names", prod), func() {
			checkUniqueIDsAndNames(catalog[prod])
		})
	})

	Context("Values presence and format validation per env", Ordered, func() {

		It(fmt.Sprintf("all items (vendors, families and products) in %s should have valid values", dev), func() {
			checkValues(catalog[dev])
		})

		It(fmt.Sprintf("all items (vendors, families and products) in %s should have valid values", staging), func() {
			checkValues(catalog[staging])
		})

		It(fmt.Sprintf("all items (vendors, families and products) in %s should have valid values", prod), func() {
			checkValues(catalog[prod])
		})
	})
})
