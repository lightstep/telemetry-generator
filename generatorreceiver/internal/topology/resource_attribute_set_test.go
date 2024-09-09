package topology

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResourceAttributeSet_GetAttributes(t *testing.T) {
	tests := []struct {
		name               string
		service            string
		resourceAttributes TagMap
		kubernetes         *Kubernetes
		expected           TagMap
	}{
		{
			name:    "resource attributes and kubernetes both defined",
			service: "fake-service",
			resourceAttributes: map[string]interface{}{
				"cloud":  "nimbus",
				"region": "us-central1",
			},
			kubernetes: &Kubernetes{
				ClusterName: "cluster-1",
				Cfg: &Config{
					Kubernetes: &KubernetesConfig{
						PodCount: 1,
					},
				},
			},
			expected: TagMap{
				"cloud":               "nimbus",
				"region":              "us-central1",
				"k8s.cluster.name":    "cluster-1",
				"k8s.container.name":  "fake-service",
				"k8s.namespace.name":  "fake-service",
				"k8s.deployment.name": "",
			},
		},
		{
			name:    "resource attributes defined but not kubernetes",
			service: "some-service",
			resourceAttributes: map[string]interface{}{
				"cloud":  "nimbus",
				"region": "europe-west1",
			},
			kubernetes: nil,
			expected: TagMap{
				"cloud":  "nimbus",
				"region": "europe-west1",
			},
		},
		{
			name:               "neither resource attributes nor kubernetes defined",
			service:            "another-service",
			resourceAttributes: map[string]interface{}{},
			kubernetes:         nil,
			expected:           TagMap{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resourceAttrSet := ResourceAttributeSet{
				Kubernetes:         tt.kubernetes,
				ResourceAttributes: tt.resourceAttributes,
			}

			if k := resourceAttrSet.Kubernetes; k != nil {
				random := rand.New(rand.NewSource(123))
				k.CreatePods(tt.service, random)

				// k8s.pod.name structure was copied from CreatePods()
				random = rand.New(rand.NewSource(123))
				tt.expected["k8s.pod.name"] = tt.service + "-" + generateK8sName(10, random) + "-" + generateK8sName(5, random)
			}

			random := rand.New(rand.NewSource(123))
			require.Equal(t, tt.expected, *resourceAttrSet.GetAttributes(random))
		})
	}
}
