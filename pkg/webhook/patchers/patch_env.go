package patchers

import corev1 "k8s.io/api/core/v1"

const (
	envNodeName = "NODE_NAME"
)

func (v *MutatingWebhookHandler) patchEnv(envs []corev1.EnvVar) []corev1.EnvVar {
	for _, env := range envs {
		if env.Name == envNodeName {
			return envs
		}
	}
	envs = append(envs, corev1.EnvVar{
		Name: envNodeName,
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: "spec.nodeName",
			},
		},
	})
	return envs
}
