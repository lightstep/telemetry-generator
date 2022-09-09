package topology

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
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

func TestMultiPod(t *testing.T) {
	nPods := 7
	minute, _ := time.ParseDuration("1m")
	tenMinute, _ := time.ParseDuration("10m")
	k := &Kubernetes{
		ClusterName: "some-cluster",
		Restart: Restart{
			Every:  tenMinute,
			Jitter: minute,
		},
		PodCount: nPods,
	}
	k.CreatePods("some")

	// we should see more than one pod name in the tags
	// (odds of this test failing randomly are 1 in 7**100 =~ 3 in 10^85
	names := make(map[string]bool)
	for i := 0; i < 100; i++ {
		tags := k.GetK8sTags()
		names[tags["k8s.pod.name"]] = true
	}
	require.Greater(t, len(names), 1, "multiple pod names should be generated")

	// we should see different restart durations
	// (odds of this test failing randomly are 1 in (60e9)**6 =~ 5 in 10^64)
	unequal := false
	for i := 1; i < nPods; i++ {
		unequal = unequal || (k.pods[i].RestartDuration != k.pods[0].RestartDuration)
	}
	require.True(t, unequal, "different restart durations should be set")

	k = &Kubernetes{
		ClusterName: "some-cluster",
		Restart: Restart{
			Every:  tenMinute,
			Jitter: minute,
		},
		Cfg: &Config{
			Kubernetes: &KubernetesConfig{
				PodCount: 2,
			},
		},
	}
	k.CreatePods("some")
	require.Equal(t, 2, k.GetPodCount(), "pod count defaults to config value")
	k = &Kubernetes{
		ClusterName: "some-cluster",
		Restart: Restart{
			Every:  tenMinute,
			Jitter: minute,
		},
	}
	k.CreatePods("some")
	require.Equal(t, 1, k.GetPodCount(), "pod count defaults to 1 if no config value")
}
