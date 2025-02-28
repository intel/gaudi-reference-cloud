// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package kubectl

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/utils"
	"github.com/jonboulle/clockwork"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	kubectlApply "k8s.io/kubectl/pkg/cmd/apply"
	kubectlUtil "k8s.io/kubectl/pkg/util"
)

const (
	ApplyOperation   = "apply"
	ReplaceOperation = "replace"
)

type AddonProvider struct {
	RestConfig *rest.Config
	ClientSet  *kubernetes.Clientset
	S3Addon    S3AddonConfig
}

func NewAddonProvider(restConfig *rest.Config, s3Info S3AddonConfig) (*AddonProvider, error) {
	kubernetesClient, err := utils.GetKubernetesClientFromConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return &AddonProvider{RestConfig: restConfig, ClientSet: kubernetesClient, S3Addon: s3Info}, nil
}

func (p *AddonProvider) Put(ctx context.Context, addon *privatecloudv1alpha1.Addon) error {
	manifests, err := getManifests(addon, p.S3Addon)
	if err != nil {
		return err
	}

	groupResources, err := restmapper.GetAPIGroupResources(p.ClientSet.Discovery())
	if err != nil {
		return err
	}

	restMapper := restmapper.NewDiscoveryRESTMapper(groupResources)

	operation := apply
	if strings.Split(string(addon.Spec.Type), "-")[1] == ReplaceOperation {
		operation = replace
	}

	for _, manifest := range manifests {
		if err := operation(&manifest, p.RestConfig, restMapper); err != nil {
			return err
		}
	}

	return nil
}

func (p *AddonProvider) Get(ctx context.Context, name string, namespace string) (*privatecloudv1alpha1.AddonStatus, error) {
	return nil, nil
}

func (p *AddonProvider) Delete(ctx context.Context, addon *privatecloudv1alpha1.Addon) error {
	manifests, err := getManifests(addon, p.S3Addon)
	if err != nil {
		return err
	}

	groupResources, err := restmapper.GetAPIGroupResources(p.ClientSet.Discovery())
	if err != nil {
		return err
	}

	restMapper := restmapper.NewDiscoveryRESTMapper(groupResources)

	for _, manifest := range manifests {
		if err := delete(&manifest, p.RestConfig, restMapper); err != nil {
			return err
		}
	}

	return nil
}

func getManifests(addon *privatecloudv1alpha1.Addon, s3Info S3AddonConfig) ([]unstructured.Unstructured, error) {
	manifests := make([]unstructured.Unstructured, 0)

	var artifact []byte
	var err error
	if strings.HasPrefix(addon.Spec.Artifact, "s3://") {
		artifact, err = getArtifactFromS3(addon.Spec.Artifact, s3Info)
	} else {
		artifact, err = getRemoteArtifact(addon.Spec.Artifact)
	}

	if err != nil {
		return manifests, err
	}

	manifestTemplate, err := getTemplate(addon.Name, string(artifact))
	if err != nil {
		return manifests, err
	}

	manifestConfig, err := getTemplateConfig(addon.Spec.ClusterName, addon.Name, addon.Spec.Args)
	if err != nil {
		return manifests, err
	}

	manifestBuffer := &bytes.Buffer{}
	if err := manifestTemplate.Execute(manifestBuffer, manifestConfig); err != nil {
		return manifests, err
	}

	manifests, err = splitManifest(manifestBuffer.Bytes())
	if err != nil {
		return manifests, err
	}

	return manifests, nil
}

func getHelperAndInfo(u *unstructured.Unstructured, restConfig *rest.Config, restMapper meta.RESTMapper) (*resource.Helper, *resource.Info, error) {
	restMapping, err := restMapper.RESTMapping(u.GroupVersionKind().GroupKind(), u.GroupVersionKind().Version)
	if err != nil {
		return nil, nil, err
	}

	restClient, err := newRestClient(restConfig, restMapping.GroupVersionKind.GroupVersion())
	if err != nil {
		return nil, nil, err
	}

	helper := resource.NewHelper(restClient, restMapping)

	info := &resource.Info{
		Client:          restClient,
		Mapping:         restMapping,
		Namespace:       u.GetNamespace(),
		Name:            u.GetName(),
		Source:          "",
		Object:          u,
		ResourceVersion: restMapping.Resource.Version,
	}

	return helper, info, nil
}

func delete(u *unstructured.Unstructured, restConfig *rest.Config, restMapper meta.RESTMapper) error {
	helper, info, err := getHelperAndInfo(u, restConfig, restMapper)
	if err != nil {
		return err
	}

	// If not found, no need to delete it.
	if err := info.Get(); err != nil {
		if !k8serrors.IsNotFound(err) {
			return err
		}
		return nil
	}

	_, err = helper.Delete(info.Namespace, info.Name)
	if err != nil {
		return err
	}

	return nil
}

func replace(u *unstructured.Unstructured, restConfig *rest.Config, restMapper meta.RESTMapper) error {
	helper, info, err := getHelperAndInfo(u, restConfig, restMapper)
	if err != nil {
		return err
	}

	if err := kubectlUtil.CreateOrUpdateAnnotation(false, info.Object, unstructured.UnstructuredJSONScheme); err != nil {
		return err
	}

	// If not found, let's create it
	if err := info.Get(); err != nil {
		if !k8serrors.IsNotFound(err) {
			return err
		}

		_, err := helper.Create(info.Namespace, true, info.Object)
		if err != nil {
			return err
		}
	} else {
		// If found, let's replace it
		_, err := helper.Replace(info.Namespace, info.Name, true, info.Object)
		if err != nil {
			return err
		}
	}

	return nil
}

func apply(u *unstructured.Unstructured, restConfig *rest.Config, restMapper meta.RESTMapper) error {
	helper, info, err := getHelperAndInfo(u, restConfig, restMapper)
	if err != nil {
		return err
	}

	patcher, err := newPatcher(info, helper)
	if err != nil {
		return err
	}

	modified, err := kubectlUtil.GetModifiedConfiguration(info.Object, true, unstructured.UnstructuredJSONScheme)
	if err != nil {
		return err
	}

	// If not found, let's create it
	if err := info.Get(); err != nil {
		if !k8serrors.IsNotFound(err) {
			return err
		}

		if err := kubectlUtil.CreateApplyAnnotation(info.Object, unstructured.UnstructuredJSONScheme); err != nil {
			return err
		}

		obj, err := helper.Create(info.Namespace, true, info.Object)
		if err != nil {
			return err
		}

		if err := info.Refresh(obj, true); err != nil {
			return err
		}
	}

	var errOut io.Writer
	_, _, err = patcher.Patch(info.Object, modified, info.Source, info.Namespace, info.Name, errOut)
	if err != nil {
		return err
	}

	return nil
}

func newPatcher(info *resource.Info, helper *resource.Helper) (*kubectlApply.Patcher, error) {
	return &kubectlApply.Patcher{
		Mapping:           info.Mapping,
		Helper:            helper,
		Overwrite:         true,
		BackOff:           clockwork.NewRealClock(),
		Force:             false,
		Timeout:           time.Duration(0),
		GracePeriod:       -1,
		Retries:           0,
		CascadingStrategy: v1.DeletePropagationBackground,
	}, nil
}

func newRestClient(restConfig *rest.Config, gv schema.GroupVersion) (rest.Interface, error) {
	restConfig.ContentConfig = resource.UnstructuredPlusDefaultContentConfig()
	restConfig.GroupVersion = &gv

	if len(gv.Group) == 0 {
		restConfig.APIPath = "/api"
	} else {
		restConfig.APIPath = "/apis"
	}

	return rest.RESTClientFor(restConfig)
}

func getRemoteArtifact(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return make([]byte, 0), err
	}
	defer resp.Body.Close()

	manifest, err := io.ReadAll(resp.Body)
	if err != nil {
		return make([]byte, 0), err
	}

	return manifest, nil
}

func getArtifactFromS3(path string, s3AddonConfig S3AddonConfig) ([]byte, error) {
	splitter := strings.Split(path, "//")
	objectPath := s3AddonConfig.S3Path + splitter[len(splitter)-1]

	minioClient, err := minio.New(s3AddonConfig.URL, &minio.Options{
		Creds:  credentials.NewStaticV4(s3AddonConfig.AccessKey, s3AddonConfig.SecretKey, ""),
		Secure: s3AddonConfig.UseSSL,
	})
	if err != nil {
		return make([]byte, 0), err
	}

	reader, err := minioClient.GetObject(context.Background(), s3AddonConfig.BucketName, objectPath, minio.GetObjectOptions{})
	if err != nil {
		return make([]byte, 0), err
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return make([]byte, 0), err
	}
	return data, nil
}

// splitManifest splits the manifest if it contains multiple docs in it (yaml separated by ---).
// Then it creates a list of unstructured objects based on the individual docs that were
// split.
//
// An unstructured's Object field is just a map[string]interface{}, before creating it, we can
// do more validations just like kubectl does for encoding / decoding the object and get an error
// sooner. But for now we will put the object as it is and let the kubernetes api-server to return
// an error in case object is invalid.
func splitManifest(manifest []byte) ([]unstructured.Unstructured, error) {
	manifests := make([]unstructured.Unstructured, 0)
	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(manifest), len(manifest))

	for {
		var object map[string]interface{}

		err := decoder.Decode(&object)
		if err == io.EOF {
			break
		}

		if err != nil {
			return manifests, err
		}

		if object != nil {
			manifests = append(manifests, unstructured.Unstructured{
				Object: object,
			})
		}
	}

	return manifests, nil
}
