package topology

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestReplaceTags(t *testing.T) {
	k := &Kubernetes{
		Service:        "testapp-service",
		Namespace:      "testapp-namespace",
		ClusterName:    "testapp-cluster",
		ReplicaSetName: "testapp-replica",
		pods:           make([]*Pod, 0, 1),
	}
	pod := Pod{
		PodName:    "testapp-abbab-abb",
		Container:  "testapp-container",
		Kubernetes: k,
	}
	k.pods = append(k.pods, &pod)

	tags := map[string]string{
		"my_pod":       PodName,
		"my_service":   Service,
		"my_namespace": Namespace,
		"my_container": Container,
		"my_cluster":   Cluster,
		"my_replica":   ReplicaSet,
		"my_key":       "my_value",
	}

	tags = k.pods[0].ReplaceTags(tags)

	require.Equal(t,
		map[string]string{
			"my_pod":       "testapp-abbab-abb",
			"my_service":   "testapp-service",
			"my_namespace": "testapp-namespace",
			"my_container": "testapp-container",
			"my_cluster":   "testapp-cluster",
			"my_replica":   "testapp-replica",
			"my_key":       "my_value",
		},
		tags,
	)
}
