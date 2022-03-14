package cluster_credential

import (
	"context"
	"strings"

	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var (
	TelegrafDS              = "telegraf"
	TelegrafPlatform        = "telegraf-platform"
	FluentBitDS             = "fluent-bit"
	ErdaClusterCredential   = "erda-cluster-credential"
	defaultClusterAccessKey = map[string][]byte{
		"CLUSTER_ACCESS_KEY": []byte("init"),
	}

	dsResource     = metav1.GroupVersionResource{Group: "apps", Version: "v1", Resource: "daemonsets"}
	deployResource = metav1.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
)

type ValidatingHandler struct {
	CRClient client.Client
	Decoder  *admission.Decoder
}

func (v *ValidatingHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	if req.Resource == dsResource && (strings.Contains(req.Name, TelegrafDS) || strings.Contains(req.Name, FluentBitDS)) ||
		(req.Resource == deployResource && strings.Contains(req.Name, TelegrafPlatform)) {
		logrus.Infof("wating to patch, name: %s, resources: %s, namespace: %s", req.Name, req.Resource, req.Namespace)
		patchInfo, err := v.PatchClusterCredential(req)
		if err != nil {
			logrus.Error(err)
		} else {
			if len(patchInfo) != 0 {
				logrus.Infof("patch %s in namespace: %s", req.Name, req.Namespace)
				return admission.PatchResponseFromRaw(req.AdmissionRequest.Object.Raw, patchInfo)
			}
		}
	}

	return admission.Allowed("")
}
