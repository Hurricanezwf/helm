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
	"encoding/json"
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

type UpdateManifest struct {
	Created []string          `json:"created"`
	Updated []UpdatedManifest `json:"updated"`
	Deleted []string          `json:"deleted"`
}

func (r *UpdateResult) Marshal() (string, error) {
	updates := UpdateManifest{}
	buf := bytes.NewBuffer(nil)

	for _, created := range r.Created {
		if created == nil {
			continue
		}
		buf.Reset()
		if err := unstructured.UnstructuredJSONScheme.Encode(created.Object, buf); err != nil {
			return "", fmt.Errorf("failed to encode `%s/%s`, %w", created.Namespace, created.Name, err)
		}
		updates.Created = append(updates.Created, buf.String())
	}
	for _, updated := range r.Updated {
		m, err := updated.ToManifest()
		if err != nil {
			return "", err
		}
		updates.Updated = append(updates.Updated, m)
	}
	for _, deleted := range r.Deleted {
		if deleted == nil {
			continue
		}
		buf.Reset()
		if err := unstructured.UnstructuredJSONScheme.Encode(deleted.Object, buf); err != nil {
			return "", fmt.Errorf("failed to encode `%s/%s`, %w", deleted.Namespace, deleted.Name, err)
		}
		updates.Deleted = append(updates.Deleted, buf.String())
	}

	b, err := json.Marshal(updates)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// UpdatedInfo is to describe the updated resource.
type UpdatedInfo struct {
	From *resource.Info
	To   *resource.Info
}

type UpdatedManifest struct {
	From string `json:"from"`
	To   string `json:"to"`
}

func (u *UpdatedInfo) ToManifest() (UpdatedManifest, error) {
	update := UpdatedManifest{}
	buf := bytes.NewBuffer(nil)

	if u.From != nil {
		buf.Reset()
		if err := unstructured.UnstructuredJSONScheme.Encode(u.From.Object, buf); err != nil {
			return update, fmt.Errorf("failed to encode `%s/%s`, %w", u.From.Namespace, u.From.Name, err)
		}
		update.From = buf.String()
	}
	if u.To != nil {
		buf.Reset()
		if err := unstructured.UnstructuredJSONScheme.Encode(u.To.Object, buf); err != nil {
			return update, fmt.Errorf("failed to encode `%s/%s`, %w", u.To.Namespace, u.To.Name, err)
		}
		update.To = buf.String()
	}
	return update, nil
}

// DiffUpdateResult returns a diff of the update result with `helm diff`.
// The caller can read the output from the io.Reader.
// @forceUpdate indicates if the new object replace the old object.
func DiffUpdateResult(result *UpdateResult, forceUpdate bool) ([]byte, error) {
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
			if forceUpdate {
				oldIndex[fmt.Sprintf("[FORCE UPDATE FROM OLD] %s/%s", oldRes.Kind, oldRes.Name)] = oldRes
			} else {
				oldIndex[fmt.Sprintf("%s/%s", oldRes.Kind, oldRes.Name)] = oldRes
			}
		}
		if updated.To != nil {
			newRes, err := ResourceInfoToMappingResult(updated.To)
			if err != nil {
				return nil, fmt.Errorf("failed to convert resource info to mapping result, %w", err)
			}
			if forceUpdate {
				newIndex[fmt.Sprintf("[FORCE UPDATE TO NEW] %s/%s", newRes.Kind, newRes.Name)] = newRes
			} else {
				newIndex[fmt.Sprintf("%s/%s", newRes.Kind, newRes.Name)] = newRes
			}
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
	b = RemoveDirtyDiffFields(b)
	if b, err = yaml.JSONToYAML(b); err != nil {
		return nil, fmt.Errorf("failed to parse json doc to yaml, %w", err)
	}

	return &manifest.MappingResult{
		Name:    rc.Name,
		Kind:    rc.Object.GetObjectKind().GroupVersionKind().Kind,
		Content: string(b),
	}, nil
}

func RemoveDirtyDiffFields(jsonBytes []byte) []byte {
	var err error
	var cleanContent = jsonBytes

	for _, path := range []string{
		"metadata.managedFields",
		"metadata.creationTimestamp",
		"metadata.resourceVersion",
		"metadata.selfLink",
		"metadata.uid",
		"spec.finalizers",
		"status",
	} {
		if cleanContent, err = sjson.DeleteBytes(cleanContent, path); err != nil {
			return jsonBytes
		}
	}
	return cleanContent
}
