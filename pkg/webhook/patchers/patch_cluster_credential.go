package patchers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func (v *MutatingWebhookHandler) Patch(req admission.Request) ([]byte, error) {
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

		curDs.Spec.Template.Spec = v.patchClusterCredential(curDs.Spec.Template.Spec)
		curDs.Spec.Template.Spec.Affinity = v.patchAffinity(curDs.Name)
		curDs.Spec.Template.Spec.Containers[0].Env = v.patchEnv(curDs.Spec.Template.Spec.Containers[0].Env)
		curDs.Spec.Template.Spec.ServiceAccountName = ErdaOnErdaServiceAccount

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

		curDeploy.Spec.Template.Spec = v.patchClusterCredential(curDeploy.Spec.Template.Spec)
		curDeploy.Spec.Template.Spec.Containers[0].Env = v.patchEnv(curDeploy.Spec.Template.Spec.Containers[0].Env)
		curDeploy.Spec.Template.Spec.ServiceAccountName = ErdaOnErdaServiceAccount

		newPodBytes, err := json.Marshal(curDeploy)
		if err != nil {
			return nil, fmt.Errorf("marshal new pod error: %v", err)
		}
		return newPodBytes, nil
	default:
		return nil, fmt.Errorf("resources doesn't match: %v", req.Resource)
	}
}

func (v *MutatingWebhookHandler) patchClusterCredential(curSpec corev1.PodSpec) corev1.PodSpec {
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
