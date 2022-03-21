package serviceexport

import (
	"context"
	"os"
	"strconv"
	"strings"

	meshv1beta1 "bitbucket.org/realtimeai/kubeslice-operator/api/v1beta1"
	"bitbucket.org/realtimeai/kubeslice-operator/controllers"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getClusterName() string {
	return os.Getenv("CLUSTER_NAME")
}

// portListToDisplayString converts list of ports to a single string
func portListToDisplayString(servicePorts []meshv1beta1.ServicePort) string {
	ports := []string{}
	for _, port := range servicePorts {
		protocol := "TCP"
		if port.Protocol != "" {
			protocol = string(port.Protocol)
		}
		ports = append(ports, strconv.Itoa(int(port.ContainerPort))+"/"+protocol)
	}
	return strings.Join(ports, ",")
}

// Get NSM Ip of an app pod
func getNsmIP(pod *corev1.Pod, appPods []meshv1beta1.AppPod) string {
	for _, appPod := range appPods {
		if pod.Name == appPod.PodName && pod.Namespace == appPod.PodNamespace {
			return appPod.NsmIP
		}
	}
	return ""
}

// Determine if there is a change in existing service pods list
func isServiceAppPodChanged(current []meshv1beta1.ServicePod, old []meshv1beta1.ServicePod) bool {
	if len(current) != len(old) {
		return true
	}

	s := make(map[string]meshv1beta1.ServicePod)

	for _, c := range old {
		s[c.Name] = c
	}

	for _, c := range current {
		if s[c.Name].NsmIP != c.NsmIP {
			return true
		}
		if s[c.Name].PodIp != c.PodIp {
			return true
		}
	}

	return false
}

// Get Apppods connected to a slice
func getAppPodsInSlice(ctx context.Context, c client.Client, sliceName string) ([]meshv1beta1.AppPod, error) {
	log := ctrl.Log.WithName("util")

	slice, err := controllers.GetSlice(ctx, c, sliceName)

	if err != nil {
		log.Error(err, "Failed to get Slice")
		return nil, err
	}

	return slice.Status.AppPods, nil
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}