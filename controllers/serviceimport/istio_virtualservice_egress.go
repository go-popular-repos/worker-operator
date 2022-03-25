package serviceimport

import (
	"context"

	meshv1beta1 "bitbucket.org/realtimeai/kubeslice-operator/api/v1beta1"
	"bitbucket.org/realtimeai/kubeslice-operator/controllers"
	"bitbucket.org/realtimeai/kubeslice-operator/internal/logger"
	networkingv1beta1 "istio.io/api/networking/v1beta1"
	istiov1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *Reconciler) ReconcileVirtualServiceEgress(ctx context.Context, serviceimport *meshv1beta1.ServiceImport) (ctrl.Result, error, bool) {
	log := logger.FromContext(ctx).WithValues("type", "Istio VS with egress")
	debugLog := log.V(1)

	debugLog.Info("reconciling istio vs with egress")

	vs, err := r.getVirtualServiceFromAppPod(ctx, serviceimport)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("vs to egress not found; creating")
			vs = r.virtualServiceToEgress(serviceimport)
			err := r.Create(ctx, vs)
			if err != nil {
				log.Error(err, "Failed to create virtualService for", "serviceimport", serviceimport)
				return ctrl.Result{}, err, true
			}
			log.Info("virtualService resource created")
			return ctrl.Result{Requeue: true}, nil, true
		}
		log.Error(err, "unable to get virtualService")
		return ctrl.Result{}, err, true
	}

	vs, err = r.getVirtualServiceFromEgress(ctx, serviceimport)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("vs from egress not found; creating")

			if len(serviceimport.Status.Endpoints) == 0 {
				log.Info("Endpoints are 0, skipping virtualService creation from egress")
				return ctrl.Result{}, nil, false
			}

			vs = r.virtualServiceFromEgress(serviceimport)
			err := r.Create(ctx, vs)
			if err != nil {
				log.Error(err, "Failed to create virtualService egress for", "serviceimport", serviceimport)
				return ctrl.Result{}, err, true
			}
			log.Info("virtualService resource created")
			return ctrl.Result{Requeue: true}, nil, true
		}
		log.Error(err, "unable to get virtualService")
		return ctrl.Result{}, err, true
	}

	if len(serviceimport.Status.Endpoints) == 0 {
		log.Info("Endpoints are 0, deleting existing virtualService from egress")
		err = r.Delete(ctx, vs)
		if err != nil {
			log.Error(err, "Unable to delete virtualService from egress")
			return ctrl.Result{}, err, true
		}
		log.Info("deleted virtualService from egress")
		return ctrl.Result{}, nil, false
	}

	if hasVirtualServiceRoutesChanged(vs, serviceimport) {
		log.Info("virtualService routes changed for egress, updating")
		if getServiceProtocol(serviceimport) == meshv1beta1.ServiceProtocolHTTP {
			httpRoutes := getVirtualServiceHTTPRoutes(serviceimport)
			debugLog.Info("new routes", "http", httpRoutes)
			vs.Spec.Http = []*networkingv1beta1.HTTPRoute{{
				Route: httpRoutes,
			}}
		} else {
			tcpRoutes := getVirtualServiceTCPRoutes(serviceimport)
			debugLog.Info("new routes", "tcp", tcpRoutes)
			vs.Spec.Tcp = []*networkingv1beta1.TCPRoute{{
				Route: tcpRoutes,
			}}
		}
		err = r.Update(ctx, vs)
		if err != nil {
			log.Error(err, "Unable to update virtualService routes")
			return ctrl.Result{}, err, true
		}
		log.Info("virtualService routes updated")
		return ctrl.Result{Requeue: true}, nil, true
	}

	return ctrl.Result{}, nil, false
}

// Create a VirtualService which routes all traffic to the service to go to egress
func (r *Reconciler) virtualServiceToEgress(serviceImport *meshv1beta1.ServiceImport) *istiov1beta1.VirtualService {

	egressHost := serviceImport.Spec.Slice + "-istio-egressgateway." + controllers.ControlPlaneNamespace + ".svc.cluster.local"

	vs := &istiov1beta1.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      virtualServiceFromAppPodName(serviceImport),
			Namespace: serviceImport.Namespace,
		},
		Spec: networkingv1beta1.VirtualService{
			Hosts: []string{
				serviceImport.Spec.DNSName,
				serviceImport.Name + "." + serviceImport.Namespace + ".svc.cluster.local",
			},
			ExportTo: []string{"."},
		},
	}

	if getServiceProtocol(serviceImport) == meshv1beta1.ServiceProtocolHTTP {
		vs.Spec.Http = []*networkingv1beta1.HTTPRoute{{
			Route: []*networkingv1beta1.HTTPRouteDestination{{
				Destination: &networkingv1beta1.Destination{
					Host: egressHost,
					Port: &networkingv1beta1.PortSelector{
						Number: 80,
					},
				},
			}},
		}}
	} else {
		vs.Spec.Tcp = []*networkingv1beta1.TCPRoute{{
			Route: []*networkingv1beta1.RouteDestination{{
				Destination: &networkingv1beta1.Destination{
					Host: egressHost,
					Port: &networkingv1beta1.PortSelector{
						Number: uint32(serviceImport.Spec.Ports[0].ContainerPort),
					},
				},
			}},
		}}
	}

	ctrl.SetControllerReference(serviceImport, vs, r.Scheme)

	return vs
}

func (r *Reconciler) virtualServiceFromEgress(serviceImport *meshv1beta1.ServiceImport) *istiov1beta1.VirtualService {

	gw := controllers.ControlPlaneNamespace + "/" + serviceImport.Spec.Slice + "-istio-egressgateway"

	vs := &istiov1beta1.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      virtualServiceFromEgressName(serviceImport),
			Namespace: controllers.ControlPlaneNamespace,
		},
		Spec: networkingv1beta1.VirtualService{
			Hosts: []string{
				serviceImport.Spec.DNSName,
				serviceImport.Name + "." + serviceImport.Namespace + ".svc.cluster.local",
			},
			Gateways: []string{
				gw,
			},
			ExportTo: []string{"."},
		},
	}

	if getServiceProtocol(serviceImport) == meshv1beta1.ServiceProtocolHTTP {
		vs.Spec.Http = []*networkingv1beta1.HTTPRoute{{
			Route: getVirtualServiceHTTPRoutes(serviceImport),
		}}
	} else {
		vs.Spec.Tcp = []*networkingv1beta1.TCPRoute{{
			Route: getVirtualServiceTCPRoutes(serviceImport),
		}}
	}

	ctrl.SetControllerReference(serviceImport, vs, r.Scheme)

	return vs
}

func (r *Reconciler) getVirtualServiceFromEgress(ctx context.Context, serviceimport *meshv1beta1.ServiceImport) (*istiov1beta1.VirtualService, error) {
	vs := &istiov1beta1.VirtualService{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      virtualServiceFromEgressName(serviceimport),
		Namespace: controllers.ControlPlaneNamespace,
	}, vs)

	if err != nil {
		return nil, err
	}

	return vs, nil
}
