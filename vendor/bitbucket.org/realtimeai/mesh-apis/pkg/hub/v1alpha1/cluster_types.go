/*
Copyright 2022.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ClusterSpec defines the desired state of Cluster
type ClusterSpec struct {
	// HubEndpoint is the Endpoint of the Hub
	HubEndpoint string `json:"hubEndpoint,omitempty"`
	//AccessToken is the Token to Access
	AccessToken string `json:"accessToken,omitempty"`
	// CA certificate to use for the cluster
	CACert string `json:"caCert,omitempty"`
	//NodeIP is the IP address of the Node
	NodeIP string `json:"nodeIP,omitempty"`
	//put in an object
	ClusterProperty ClusterProperty `json:"clusterProperty,omitempty"`
}

type ClusterProperty struct {
	//Telemetry contains Telemetry information
	Telemetry Telemetry `json:"telemetry,omitempty"`
	//GeoLocation contains information regarding Geographical Location of the Cluster
	GeoLocation GeoLocation `json:"geoLocation,omitempty"`
	//Monitoring contains the Kubernetes Monitoring Dashboard
	Monitoring Monitoring `json:"monitoring,omitempty"`
}

// Telemetry defines the field of ClusterSpec
type Telemetry struct {
	//Enabled is the enable status of the Telemetry
	Enabled bool `json:"enabled,omitempty"`
	//TelemetryProvider is the Telemetry Provider information
	TelemetryProvider string `json:"telemetryProvider,omitempty"`
	//Endpoint is the Telemetry Endpoint
	Endpoint string `json:"endpoint,omitempty"`
}

// GeoLocation defines the field of ClusterSpec
type GeoLocation struct {
	//CloudProvider is the cloud service provider
	CloudProvider string `json:"cloudProvider,omitempty"`
	//CloudRegion is the region of the cloud
	CloudRegion string `json:"cloudRegion,omitempty"`
}

// Monitoring defines the field of ClusterSpec
type Monitoring struct {
	//KubernetesDashboard contains the information regarding Kubernetes Monitoring Dashboard
	KubernetesDashboard KubernetesDashboard `json:"kubernetesDashboard,omitempty"`
}

// KubernetesDashboard defines the field of ClusterSpec
type KubernetesDashboard struct {
	//Enabled is the enable status of the KubernetesDashboard
	Enabled bool `json:"enabled,omitempty"`
	//AccessToken is the Access Token to access the KubernetesDashboard
	AccessToken string `json:"accessToken,omitempty"`
	//IngressPrefix is the prefix of ingress gateway for KubernetesDashboard
	IngressPrefix string `json:"ingressPrefix,omitempty"`
}

// ClusterStatus defines the observed state of Cluster
type ClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Cluster is the Schema for the clusters API
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterList contains a list of Cluster
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}