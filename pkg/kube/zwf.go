// Copyright 2022 Wenfeng Zhou (zwf1094646850@gmail.com)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kube

import (
	"context"
	"fmt"

	resourcev1 "helm.sh/helm/v3/pkg/kube/resource/v1"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// QueryResourceWithJSONDoc 给定 json doc 查询指定资源;
func QueryResourceWithJSONDoc(ctx context.Context, jsonDoc string, dynamicKubeClient dynamic.Interface) (*unstructured.Unstructured, error) {
	j := gjson.Parse(jsonDoc)
	apiVersion := j.Get("apiVersion").String()
	kind := j.Get("kind").String()
	name := j.Get("metadata.name").String()
	namespace := j.Get("metadata.namespace").String()

	if apiVersion == "" {
		return nil, errors.New("apiVersion cannot be empty in manifest")
	}
	if kind == "" {
		return nil, errors.New("kind cannot be empty in manifest")
	}
	if name == "" {
		return nil, errors.New("resource name cannot be empty in manifest")
	}

	gv, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid apiVersion value `%s`", apiVersion)
	}

	resource, err := resourcev1.DefaultResourceDictionary.KindToResource(kind)
	if err != nil {
		return nil, err
	}

	obj, err := dynamicKubeClient.Resource(gv.WithResource(resource)).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if obj == nil {
		return nil, ErrNotFound // for compatibility
	}
	return obj, nil
}
