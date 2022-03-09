package cluster_credential

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
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

func (v *ValidatingHandler) PatchClusterCredential(req admission.Request) ([]byte, error) {
	clusterCredential := &corev1.Secret{}

	if err := v.CRClient.Get(context.Background(), client.ObjectKey{
		Name:      ErdaClusterCredential,
		Namespace: req.Namespace,
	}, clusterCredential); err != nil {
		if k8serrors.IsNotFound(err) {
			logrus.Infof("can't find secret %s in namespace: %s", ErdaClusterCredential, req.Namespace)

			if err = v.CRClient.Create(context.Background(), &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ErdaClusterCredential,
					Namespace: req.Namespace,
				},
				Data: defaultClusterAccessKey,
			}); err != nil {
				logrus.Errorf("create secret error: %v", err)
			}
		} else {
			return nil, fmt.Errorf("get secret error: %v", err)
		}
	}

	switch req.Resource {
	case dsResource:
		curDs := &appsv1.DaemonSet{}
		if err := v.Decoder.Decode(req, curDs); err != nil {
			return nil, fmt.Errorf("decode error: %v", err)
		}

		curDs.Spec.Template.Spec = v.checkClusterCredential(curDs.Spec.Template.Spec)

		newPodBytes, err := json.Marshal(curDs)
		if err != nil {
			return nil, fmt.Errorf("marshal new pod error: %v", err)
		}
		return newPodBytes, nil
	case deployResource:
		curDeploy := &appsv1.Deployment{}
		if err := v.Decoder.Decode(req, curDeploy); err != nil {
			return nil, fmt.Errorf("decode error: %v", err)
		}

		curDeploy.Spec.Template.Spec = v.checkClusterCredential(curDeploy.Spec.Template.Spec)

		newPodBytes, err := json.Marshal(curDeploy)
		if err != nil {
			return nil, fmt.Errorf("marshal new pod error: %v", err)
		}
		return newPodBytes, nil
	default:
		return nil, fmt.Errorf("resources doesn't match: %v", req.Resource)
	}
}

func (v *ValidatingHandler) checkClusterCredential(curSpec corev1.PodSpec) corev1.PodSpec {
	// Check volumes
	volumeExisted := false
	for _, volume := range curSpec.Volumes {
		if volume.Name == ErdaClusterCredential {
			volumeExisted = true
			break
		}
	}
	if !volumeExisted {
		curSpec.Volumes = append(curSpec.Volumes, corev1.Volume{
			Name: ErdaClusterCredential,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: ErdaClusterCredential,
				},
			},
		})
	}

	// Check mounted
	for i, container := range curSpec.Containers {
		isMounted := false
		for _, mount := range container.VolumeMounts {
			if mount.Name == ErdaClusterCredential {
				isMounted = true
				break
			}
		}
		if !isMounted {
			curSpec.Containers[i].VolumeMounts = append(curSpec.Containers[i].VolumeMounts, corev1.VolumeMount{
				Name:      ErdaClusterCredential,
				ReadOnly:  true,
				MountPath: fmt.Sprintf("/%s", ErdaClusterCredential),
			})
		}
	}

	return curSpec
}
