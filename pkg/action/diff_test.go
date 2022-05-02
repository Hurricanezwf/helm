/*
Copyright Hurricanezwf.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package action

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/utils/encoding"

	"github.com/tidwall/gjson"
	// Import to initialize client auth plugins.
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/util/homedir"
)

func TestResourceInfoToMappingResult(t *testing.T) {
	manifest := `{
    "apiVersion": "apps/v1",
    "kind": "Deployment",
    "metadata": {
        "name": "deployname"
    }
}`

	resourceInfoList, err := resource.NewBuilder(mustNewRestGetter(t)).
		Flatten().
		RequireObject(true).
		Unstructured().
		Stream(bytes.NewBufferString(manifest), "testcase").
		Do().Infos()

	if err != nil {
		t.Fatalf("failed to build resource info list, %v", err)
	}

	mappingResult, err := ResourceInfoToMappingResult(resourceInfoList[0])
	if err != nil {
		t.Fatalf("failed to convert resource info to mapping result, %v", err)
	}
	if mappingResult.Name != "deployname" {
		t.Fatalf("unexpected name %s, expect deployname", mappingResult.Name)
	}
	if mappingResult.Kind != "Deployment" {
		t.Fatalf("unexpected kind %s, expect Deployment", mappingResult.Kind)
	}

	jsonChunk, err := encoding.YAMLStreamToJSONChunk(mappingResult.Content)
	if err != nil {
		t.Fatalf("failed to convert yaml stream to json chunk, %v", err)
	}
	j := gjson.Parse(jsonChunk[0])
	if v := j.Get("apiVersion").String(); v != "apps/v1" {
		t.Fatalf("unexpected apiVersion `%s` in content, expect apps/v1", v)
	}
	if v := j.Get("kind").String(); v != "Deployment" {
		t.Fatalf("unexpected kind `%s` in content, expect Deployment", v)
	}
	if v := j.Get("metadata.name").String(); v != "deployname" {
		t.Fatalf("unexpected metadata.name `%s` in content, expect deployname", v)
	}
}

func TestDiffUpdateResult(t *testing.T) {
	builder := resource.NewBuilder(mustNewRestGetter(t)).Flatten().RequireObject(true).Unstructured()

	createdOne := `{
    "apiVersion": "v1",
    "kind": "Secret",
    "metadata": {
        "name": "create_secret"
    }
}`

	updateFrom := `{
    "apiVersion": "apps/v1",
    "kind": "Deployment",
    "metadata": {
        "name": "update_deploy",
        "labels": {
            "managedby": "demeter",
            "appsuite": "xxx"
        }
    }
}`

	updateTo := `{
    "apiVersion": "apps/v1",
    "kind": "Deployment",
    "metadata": {
        "name": "update_deploy",
        "labels": {
            "managedby": "demeter",
            "appsuite": "zzz"
        }
    }
}`

	deletedOne := `{
    "apiVersion": "apps/v1",
    "kind": "Deployment",
    "metadata": {
        "name": "delete_deploy"
    }
}`

	createdResourceList, err := builder.Stream(bytes.NewBufferString(createdOne), "create_one").Do().Infos()
	if err != nil {
		t.Fatalf("failed to build resource info list, %v", err)
	}
	updatedFromResourceList, err := builder.Stream(bytes.NewBufferString(updateFrom), "update_from").Do().Infos()
	if err != nil {
		t.Fatalf("failed to build resource info list, %v", err)
	}
	updatedToResourceList, err := builder.Stream(bytes.NewBufferString(updateTo), "update_to").Do().Infos()
	if err != nil {
		t.Fatalf("failed to build resource info list, %v", err)
	}
	deletedResourceList, err := builder.Stream(bytes.NewBufferString(deletedOne), "delete_one").Do().Infos()
	if err != nil {
		t.Fatalf("failed to build resource info list, %v", err)
	}

	updateResult := &UpdateResult{
		Created: createdResourceList,
		Updated: []UpdatedInfo{
			{
				From: updatedFromResourceList[0],
				To:   updatedToResourceList[0],
			},
		},
		Deleted: deletedResourceList,
	}

	b, err := DiffUpdateResult(updateResult, false)
	if err != nil {
		t.Fatalf("failed to diff update result, %v", err)
	}
	os.Stdout.Write(b)
}

func mustNewRestGetter(t *testing.T) genericclioptions.RESTClientGetter {
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config.k116")

	settings := cli.New()
	settings.KubeConfig = kubeconfig

	return settings.RESTClientGetter()
}
