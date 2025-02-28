package validator

import (
	"bufio"
	"catalog-validator/pkg/common"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	ext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/validation"
	"k8s.io/client-go/kubernetes/scheme"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	apimachyaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/kube-openapi/pkg/validation/validate"
)

func ValidateResourceYaml(args []string, src string) {

	validators := map[schema.GroupVersionKind]*validate.SchemaValidator{}

	if err := fs.WalkDir(os.DirFS(src), ".", func(filePath string, dirPath fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if dirPath.IsDir() {
			return nil
		}

		fileHandle, err := os.Open(filepath.Join(src, filePath))
		if err != nil {
			fmt.Print("Error encountered: ", err)
			os.Exit(1)
		}
		defer fileHandle.Close()

		yamlReader := apimachyaml.NewYAMLReader(bufio.NewReader(fileHandle))
		for {
			yamlContent, err := yamlReader.Read()
			if err != nil && err != io.EOF {
				return err
			}
			if err == io.EOF {
				break
			}
			if len(yamlContent) == 0 {
				continue
			}

			resourceDefinition := &common.CustomResourceDefinitionExtV1{}
			if err := yaml.Unmarshal(yamlContent, resourceDefinition); err != nil {
				return err
			}

			k8sResourceDefinition := &ext.CustomResourceDefinition{}
			if err := common.Convert_v1_CustomResourceDefinition_To_apiextensions_CustomResourceDefinition(resourceDefinition, k8sResourceDefinition, nil); err != nil {
				return err
			}

			for _, ver := range k8sResourceDefinition.Spec.Versions {
				var sv *validate.SchemaValidator
				var err error
				sv, _, err = validation.NewSchemaValidator(ver.Schema)

				if err != nil {
					return err
				}

				if k8sResourceDefinition.Spec.Validation != nil {
					sv, _, err = validation.NewSchemaValidator(k8sResourceDefinition.Spec.Validation)
					if err != nil {
						return err
					}
				}

				validators[schema.GroupVersionKind{
					Group:   k8sResourceDefinition.Spec.Group,
					Version: ver.Name,
					Kind:    k8sResourceDefinition.Spec.Names.Kind,
				}] = sv
			}

		}
		return nil
	}); err != nil {
		fmt.Print("Error encountered: ", err)
		os.Exit(1)
	}

	for _, yamlFilePath := range args {

		fmt.Println("===== Validating file: ", yamlFilePath, "=========")

		yamlContent := common.GetYamlContent(yamlFilePath)
		if yamlContent == nil {
			fmt.Println("Please ensure input file has valid contents!")
			continue
		}

		obj := &unstructured.Unstructured{}
		_, gvk, err := scheme.Codecs.UniversalDeserializer().Decode(yamlContent, nil, obj)
		if err != nil {
			fmt.Println("Couldn't parse unstructured yaml into structure: ", yamlFilePath)
		}

		if gvk == nil {
			fmt.Println("Couldn't determine GroupVersionKind for the input yaml: ", yamlFilePath)
		}

		v, ok := validators[obj.GetObjectKind().GroupVersionKind()]
		if !ok {
			fmt.Println("Could not find validator for: " + obj.GetObjectKind().GroupVersionKind().String())
			fmt.Println("Please ensure input file is of valid kind!")
			continue
		}

		re := v.Validate(obj)

		if re.Errors == nil {
			fmt.Println("Validated, OK!")
		}

		for i, e := range re.Errors {
			fmt.Printf("Validation Error %d (%s)(%s): %s\n", i, obj.GroupVersionKind().String(), obj.GetName(), e.Error())
		}

	}
}
