package action

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/databus23/helm-diff/v3/diff"
	"github.com/databus23/helm-diff/v3/manifest"
	"github.com/tidwall/sjson"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/resource"
	"sigs.k8s.io/yaml"
)

// UpdateResult records the manifests that are created, updated, and deleted.
// It is used to generate the diff output.
type UpdateResult struct {
	Created []*resource.Info
	Updated []UpdatedInfo
	Deleted []*resource.Info
}

// UpdatedInfo is to describe the updated resource.
type UpdatedInfo struct {
	From *resource.Info
	To   *resource.Info
}

// DiffUpdateResult returns a diff of the update result with `helm diff`.
// The caller can read the output from the io.Reader.
func DiffUpdateResult(result *UpdateResult) ([]byte, error) {
	if result == nil {
		return nil, errors.New("update result cannot be nil")
	}

	oldIndex := make(map[string]*manifest.MappingResult)
	newIndex := make(map[string]*manifest.MappingResult)
	for _, created := range result.Created {
		if created != nil {
			res, err := ResourceInfoToMappingResult(created)
			if err != nil {
				return nil, fmt.Errorf("failed to convert resource info to mapping result, %w", err)
			}
			newIndex[fmt.Sprintf("%s/%s", res.Kind, res.Name)] = res
		}
	}
	for _, deleted := range result.Deleted {
		if deleted != nil {
			res, err := ResourceInfoToMappingResult(deleted)
			if err != nil {
				return nil, fmt.Errorf("failed to convert resource info to mapping result, %w", err)
			}
			oldIndex[fmt.Sprintf("%s/%s", res.Kind, res.Name)] = res
		}
	}
	for _, updated := range result.Updated {
		if updated.From != nil {
			oldRes, err := ResourceInfoToMappingResult(updated.From)
			if err != nil {
				return nil, fmt.Errorf("failed to convert resource info to mapping result, %w", err)
			}
			oldIndex[fmt.Sprintf("%s/%s", oldRes.Kind, oldRes.Name)] = oldRes
		}
		if updated.To != nil {
			newRes, err := ResourceInfoToMappingResult(updated.To)
			if err != nil {
				return nil, fmt.Errorf("failed to convert resource info to mapping result, %w", err)
			}
			newIndex[fmt.Sprintf("%s/%s", newRes.Kind, newRes.Name)] = newRes
		}
	}

	diffBuffer := bytes.NewBuffer(nil)
	diff.Manifests(oldIndex, newIndex, &diff.Options{
		OutputContext:   6,
		StripTrailingCR: true,
		ShowSecrets:     false,
	}, diffBuffer)

	return diffBuffer.Bytes(), nil
}

// ResourceInfoToMappingResult converts the resource info to a mapping result for it can be diff easily.
// Attention: The Content field value in the MappingResult will be the yaml string.
// In v1.19.x it's like:
// object:
//   apiVersion: apps/v1
//   kind: Deployment
//   metadata:
//     name: deployname
// --------------------------------------------
// In v1.21.x  it's like:
// apiVersion: apps/v1
// kind: Deployment
// metadata:
//   name: deployname
func ResourceInfoToMappingResult(rc *resource.Info) (*manifest.MappingResult, error) {
	if rc == nil {
		return nil, nil
	}
	b, err := runtime.Encode(unstructured.UnstructuredJSONScheme, rc.Object)
	if err != nil {
		return nil, fmt.Errorf("failed to encode resource info, %w", err)
	}

	// Notice: I remove the `object.metadata.managedFields` field here because it's too dirty to review the diff content !!!
	b = removeDirtyDiffFields(b)
	if b, err = yaml.JSONToYAML(b); err != nil {
		return nil, fmt.Errorf("failed to parse json doc to yaml, %w", err)
	}

	return &manifest.MappingResult{
		Name:    rc.Name,
		Kind:    rc.Object.GetObjectKind().GroupVersionKind().Kind,
		Content: string(b),
	}, nil
}

func removeDirtyDiffFields(jsonBytes []byte) []byte {
	var err error
	var cleanContent = jsonBytes

	if cleanContent, err = sjson.DeleteBytes(cleanContent, "metadata.managedFields"); err != nil {
		return jsonBytes
	}
	if cleanContent, err = sjson.DeleteBytes(cleanContent, "status"); err != nil {
		return jsonBytes
	}
	return cleanContent
}
