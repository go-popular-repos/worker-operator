package controllers

import (
	//	"context"
	//	"errors"
	//	"os"
	//	"strconv"
	//	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"

	meshv1beta1 "bitbucket.org/realtimeai/kubeslice-operator/api/v1beta1"
)

// LabelsForSliceGwDeployment returns the labels for creating slice gw deployment
func labelsForSliceGwDeployment(name string, slice string) map[string]string {
	return map[string]string{"networkservicemesh.io/app": name, "avesha.io/pod-type": "slicegateway", "avesha.io/slice": slice}
}

func (r *SliceGwReconciler) deploymentForGatewayServer(g *meshv1beta1.SliceGateway) *appsv1.Deployment {
	ls := labelsForSliceGwDeployment(g.Name, g.Spec.SliceName)

	var replicas int32 = 1
	var vpnSecretDefaultMode int32 = 420
	var privileged bool = true

	sidecarImg := "nexus.dev.aveshalabs.io/mesh-netops:1.0.0"
	sidecarPullPolicy := corev1.PullAlways
	vpnImg := "nexus.dev.aveshalabs.io/avesha/openvpn-server.ubuntu.18.04:1.0.0"
	vpnPullPolicy := corev1.PullAlways

	if len(gwSidecarImage) != 0 {
		sidecarImg = gwSidecarImage
	}

	if len(gwSidecarImagePullPolicy) != 0 {
		sidecarPullPolicy = corev1.PullPolicy(gwSidecarImagePullPolicy)
	}

	if len(openVpnServerImage) != 0 {
		vpnImg = openVpnServerImage
	}

	if len(openVpnServerPullPolicy) != 0 {
		vpnPullPolicy = corev1.PullPolicy(openVpnServerPullPolicy)
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        g.Name,
			Namespace:   g.Namespace,
			Annotations: map[string]string{"ns.networkservicemesh.io": "vl3-service-" + g.Spec.SliceName},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: "nsmgr-acc",
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{{
									MatchExpressions: []corev1.NodeSelectorRequirement{{
										Key:      "avesha/node-type",
										Operator: corev1.NodeSelectorOpIn,
										Values:   []string{"gateway"},
									}},
								}},
							},
						},
					},
					Containers: []corev1.Container{{
						Name:            "avesha-sidecar",
						Image:           sidecarImg,
						ImagePullPolicy: sidecarPullPolicy,
						Env: []corev1.EnvVar{
							{
								Name:  "SLICE_NAME",
								Value: g.Spec.SliceName,
							},
							{
								Name:  "CLUSTER_ID",
								Value: clusterName,
							},
							{
								Name:  "REMOTE_CLUSTER_ID",
								Value: g.Status.Config.SliceGatewayRemoteClusterID,
							},
							{
								Name:  "GATEWAY_ID",
								Value: g.Status.Config.SliceGatewayID,
							},
							{
								Name:  "REMOTE_GATEWAY_ID",
								Value: g.Status.Config.SliceGatewayRemoteGatewayID,
							},
							{
								Name:  "POD_TYPE",
								Value: "GATEWAY_POD",
							},
							{
								Name:  "NODE_IP",
								Value: nodeIP,
							},
							{
								Name:  "OPEN_VPN_MODE",
								Value: "SERVER",
							},
						},
						SecurityContext: &corev1.SecurityContext{
							Privileged:               &privileged,
							AllowPrivilegeEscalation: &privileged,
							Capabilities: &corev1.Capabilities{
								Add: []corev1.Capability{
									"NET_ADMIN",
								},
							},
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "shared-volume",
								MountPath: "/config",
							},
							{
								Name:      "vpn-certs",
								MountPath: "/var/run/vpn",
								ReadOnly:  true,
							},
						},
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								"memory": resource.MustParse("200Mi"),
								"cpu":    resource.MustParse("500m"),
							},
							Requests: corev1.ResourceList{
								"memory": resource.MustParse("50Mi"),
								"cpu":    resource.MustParse("50m"),
							},
						},
					}, {
						Name:            "avesha-openvpn-server",
						Image:           vpnImg,
						ImagePullPolicy: vpnPullPolicy,
						Command: []string{
							"/usr/local/bin/waitForConfigToRunCmd.sh",
						},
						Args: []string{
							"/etc/openvpn/openvpn.conf",
							"90",
							"ovpn_run",
						},
						SecurityContext: &corev1.SecurityContext{
							Privileged:               &privileged,
							AllowPrivilegeEscalation: &privileged,
							Capabilities: &corev1.Capabilities{
								Add: []corev1.Capability{
									"NET_ADMIN",
								},
							},
						},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "shared-volume",
							MountPath: "/etc/openvpn",
						}},
					}},
					Volumes: []corev1.Volume{
						{
							Name: "shared-volume",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						{
							Name: "vpn-certs",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName:  g.Name,
									DefaultMode: &vpnSecretDefaultMode,
								},
							},
						},
					},
					Tolerations: []corev1.Toleration{{
						Key:      "avesha/node-type",
						Operator: "Equal",
						Effect:   "NoSchedule",
						Value:    "gateway",
					}, {
						Key:      "avesha/node-type",
						Operator: "Equal",
						Effect:   "NoExecute",
						Value:    "gateway",
					}},
				},
			},
		},
	}

	if len(imagePullSecretName) != 0 {
		dep.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{{
			Name: imagePullSecretName,
		}}
	}

	// Set SliceGateway instance as the owner and controller
	ctrl.SetControllerReference(g, dep, r.Scheme)
	return dep
}

// deploymentForGateway returns a gateway Deployment object
func (r *SliceGwReconciler) deploymentForGateway(g *meshv1beta1.SliceGateway) *appsv1.Deployment {
	if g.Status.Config.SliceGatewayHostType == "Server" {
		return r.deploymentForGatewayServer(g)
	}
	return nil
}