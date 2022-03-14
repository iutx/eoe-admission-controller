package cluster_credential

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
)

func (v *ValidatingHandler) PatchEdgeAffinity(name string) *corev1.Affinity {
	if !strings.Contains(name, "edge") {
		return nil
	}
	return &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      "node.kubernetes.io/edge",
								Operator: corev1.NodeSelectorOpExists,
							},
						},
					},
				},
			},
		},
	}
}
