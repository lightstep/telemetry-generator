package topology

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestReplaceTags(t *testing.T) {
	k := &Kubernetes{
		PodName:        "testapp-abbab-abb",
		Service:        "testapp-service",
		Namespace:      "testapp-namespace",
		Container:      "testapp-container",
		ClusterName:    "testapp-cluster",
		ReplicaSetName: "testapp-replica",
	}

	tags := map[string]string{
		"my_pod":       Pod,
		"my_service":   Service,
		"my_namespace": Namespace,
		"my_container": Container,
		"my_cluster":   Cluster,
		"my_replica":   ReplicaSet,
		"my_key":       "my_value",
	}

	tags = k.ReplaceTags(tags)

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
