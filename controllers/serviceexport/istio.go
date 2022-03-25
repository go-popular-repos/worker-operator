package serviceexport

import (
	"context"

	meshv1beta1 "bitbucket.org/realtimeai/kubeslice-operator/api/v1beta1"
	"bitbucket.org/realtimeai/kubeslice-operator/controllers"
	"bitbucket.org/realtimeai/kubeslice-operator/internal/logger"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *Reconciler) ReconcileIstio(ctx context.Context, serviceexport *meshv1beta1.ServiceExport) (ctrl.Result, error, bool) {
	log := logger.FromContext(ctx).WithValues("type", "Istio")
	debugLog := log.V(1)

	slice, err := controllers.GetSlice(ctx, r.Client, serviceexport.Spec.Slice)
	if err != nil {
		log.Error(err, "Unable to fetch slice for serviceexport")
		return ctrl.Result{}, err, true
	}

	if slice.Status.SliceConfig.ExternalGatewayConfig == nil ||
		slice.Status.SliceConfig.ExternalGatewayConfig.Ingress == nil ||
		slice.Status.SliceConfig.ExternalGatewayConfig.Ingress.Enabled == false {
		debugLog.Info("istio ingress not enabled for slice, skipping reconcilation")
		return ctrl.Result{}, nil, false
	}

	debugLog.Info("reconciling istio")

	res, err, requeue := r.ReconcileServiceEntries(ctx, serviceexport)
	if requeue {
		return res, err, requeue
	}

	res, err, requeue = r.ReconcileVirtualService(ctx, serviceexport)
	if requeue {
		return res, err, requeue
	}

	return ctrl.Result{}, nil, false
}
