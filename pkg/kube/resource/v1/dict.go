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

package v1

import (
	"fmt"
	"strings"
)

var DefaultResourceDictionary = EmbeddedResourceDictionary()

const (
	LowerCertificate                  = "certificate"
	LowerCertificateRequest           = "certificaterequest"
	LowerChallenge                    = "challenge"
	LowerCiliumEndpoint               = "ciliumendpoint"
	LowerCiliumNetworkPolicy          = "ciliumnetworkpolicy"
	LowerClusterRole                  = "clusterrole"
	LowerClusterRoleBinding           = "clusterrolebinding"
	LowerConfigMap                    = "configmap"
	LowerCronJob                      = "cronjob"
	LowerDaemonSet                    = "daemonset"
	LowerDeployment                   = "deployment"
	LowerDNSChaos                     = "dnschaos"
	LowerHoriziontalPodAutoscaler     = "horizontalpodautoscaler"
	LowerHTTPChaos                    = "httpchaos"
	LowerIngress                      = "ingress"
	LowerIOChaos                      = "iochaos"
	LowerIssuer                       = "issuer"
	LowerJob                          = "job"
	LowerJVMChaos                     = "jvmchaos"
	LowerKongConsumer                 = "kongconsumer"
	LowerKongIngress                  = "kongingress"
	LowerKongPlugin                   = "kongplugin"
	LowerNamespace                    = "namespace"
	LowerNATGateway                   = "natgateway"
	LowerNetworkChaos                 = "networkchaos"
	LowerNetworkPolicy                = "networkpolicy"
	LowerMutatingWebhookConfiguration = "mutatingwebhookconfiguration"
	LowerOrder                        = "order"
	LowerPersistentVolumeClaim        = "persistentvolumeclaim"
	LowerPod                          = "pod"
	LowerPodChaos                     = "podchaos"
	LowerPodNetworkChaos              = "podnetworkchaos"
	LowerReplicaSet                   = "replicaset"
	LowerResourceQuota                = "resourcequota"
	LowerRole                         = "role"
	LowerRoleBinding                  = "rolebinding"
	LowerSchedule                     = "schedule"
	LowerSealedSecret                 = "sealedsecret"
	LowerSecret                       = "secret"
	LowerService                      = "service"
	LowerServiceAccount               = "serviceaccount"
	LowerStatefulSet                  = "statefulset"
	LowerStressChaos                  = "stresschaos"
	LowerTimeChaos                    = "timechaos"
	LowerUAPDaemon                    = "uapdaemon"
	LowerUAPDeployment                = "uapdeployment"
	LowerUAPService                   = "uapservice"
	LowerWorkflow                     = "workflow"
	LowerWorkloadPool                 = "workloadpool"
)

type ResourceDictionary interface {
	// KindToResource parse the kind to resource.
	KindToResource(kind string) (string, error)
}

func EmbeddedResourceDictionary() ResourceDictionary {
	return newResourceDictionary()
}

type resourceDictonary struct {
	dict map[string]string
}

func newResourceDictionary() *resourceDictonary {
	return &resourceDictonary{
		dict: map[string]string{
			// TODO: complete me
			LowerCertificate:              "certificate",
			LowerCertificateRequest:       "certificaterequests",
			LowerChallenge:                "challenges",
			LowerCiliumEndpoint:           "ciliumendpoints",
			LowerCiliumNetworkPolicy:      "ciliumnetworkpolicies",
			LowerClusterRole:              "clusterroles",
			LowerClusterRoleBinding:       "clusterrolebindings",
			LowerConfigMap:                "configmaps",
			LowerCronJob:                  "cronjobs",
			LowerDaemonSet:                "daemonsets",
			LowerDeployment:               "deployments",
			LowerDNSChaos:                 "dnschaos",
			LowerHoriziontalPodAutoscaler: "horizontalpodautoscalers",
			LowerHTTPChaos:                "httpchaos",
			LowerIngress:                  "ingresses",
			LowerIOChaos:                  "iochaos",
			LowerIssuer:                   "issuers",
			LowerJob:                      "jobs",
			LowerJVMChaos:                 "jvmchaos",
			LowerKongConsumer:             "kongconsumers",
			LowerKongIngress:              "kongingresses",
			LowerKongPlugin:               "kongplugins",
			LowerNamespace:                "namespaces",
			LowerNATGateway:               "natgateways",
			LowerNetworkChaos:             "networkchaos",
			LowerNetworkPolicy:            "networkpolicies",
			MutatingWebhookConfiguration:  "mutatingwebhookconfigurations",
			LowerOrder:                    "orders",
			LowerPersistentVolumeClaim:    "persistentvolumeclaims",
			LowerPod:                      "pods",
			LowerPodChaos:                 "podchaos",
			LowerPodNetworkChaos:          "podnetworkchaos",
			LowerReplicaSet:               "replicasets",
			LowerResourceQuota:            "resourcequotas",
			LowerRole:                     "roles",
			LowerRoleBinding:              "rolebindings",
			LowerSchedule:                 "schedules",
			LowerSealedSecret:             "sealedsecrets",
			LowerSecret:                   "secrets",
			LowerService:                  "services",
			LowerServiceAccount:           "serviceaccounts",
			LowerStatefulSet:              "statefulsets",
			LowerStressChaos:              "stresschaos",
			LowerTimeChaos:                "timechaos",
			LowerUAPDaemon:                "uapdaemons",
			LowerUAPDeployment:            "uapdeployments",
			LowerUAPService:               "uapservices",
			LowerWorkflow:                 "workflows",
			LowerWorkloadPool:             "workloadpools",
		},
	}
}

func (d resourceDictonary) KindToResource(kind string) (string, error) {
	resource, ok := d.dict[strings.ToLower(kind)]
	if !ok {
		return "", fmt.Errorf("no resource was found for kind `%s`", kind)
	}
	return resource, nil
}
