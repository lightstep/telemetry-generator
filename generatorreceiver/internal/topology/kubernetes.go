package topology

import (
	"math"
	"math/rand"
	"time"
)

type Kubernetes struct {
	ClusterName string   `json:"cluster_name" yaml:"cluster_name"`
	Request     Resource `json:"request" yaml:"request"`
	Limit       Resource `json:"limit" yaml:"limit"`

	ReplicaSetName string
	Namespace      string
	PodName        string
	Container      string
}

type Resource struct {
	CPU    float64 `json:"cpu" yaml:"cpu"`
	Memory float64 `json:"memory" yaml:"memory"`
}

func (k *Kubernetes) CreatePod(service ServiceTier) {
	k.ReplicaSetName = service.ServiceName + "-" + generateK8sName(10)
	k.PodName = k.ReplicaSetName + "-" + generateK8sName(5)
	k.Namespace = service.ServiceName
	k.Container = service.ServiceName
}

func (k *Kubernetes) GetK8sTags() map[string]string {
	// ref: https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/resource/semantic_conventions/k8s.md
	return map[string]string{
		"k8s.cluster.name":   k.ClusterName,
		"k8s.pod.name":       k.PodName,
		"k8s.namespace.name": k.Namespace,
		"k8s.container.name": k.Container,
	}
}

func (k *Kubernetes) GenerateMetrics(service ServiceTier) []Metric {
	if k.ClusterName == "" {
		return nil
	}

	var metrics []Metric

	metrics = append(metrics, Metric{
		Name: "kube_pod_status_phase",
		Type: "Gauge",
		Min:  1,
		Max:  1,
		Tags: map[string]string{
			"phase": "Running",
			"pod":   k.PodName,
		},
	})

	metrics = append(metrics, Metric{
		Name: "kube_pod_owner",
		Type: "Gauge",
		Min:  1,
		Max:  1,
		Tags: map[string]string{
			"pod":        k.PodName,
			"namespace":  service.ServiceName,
			"owner_name": k.ReplicaSetName,
			"onwer_kind": "ReplicaSet",
		},
	})

	metrics = append(metrics, Metric{
		Name: "kube_node_status_allocatable",
		Type: "Gauge",
		Min:  k.Limit.CPU * 1.2, // make the node a little bigger than the limit
		Max:  k.Limit.CPU * 1.2, // make the node a little bigger than the limit
		Tags: map[string]string{
			"resource": "cpu",
			"pod":      k.PodName, // used to created multiple time series that will be summed up.
		},
	})

	metrics = append(metrics, Metric{
		Name: "kube_pod_container_resource_requests",
		Type: "Gauge",
		Min:  k.Request.CPU,
		Max:  k.Request.CPU,
		Tags: map[string]string{
			"resource":  "cpu",
			"namespace": service.ServiceName,
			"container": service.ServiceName,
			"pod":       k.PodName,
		},
	})

	metrics = append(metrics, Metric{
		Name: "kube_pod_container_resource_limits",
		Type: "Gauge",
		Min:  k.Limit.CPU,
		Max:  k.Limit.CPU,
		Tags: map[string]string{
			"resource":  "cpu",
			"namespace": service.ServiceName,
			"container": service.ServiceName,
			"pod":       k.PodName,
		},
	})

	metrics = append(metrics, Metric{
		Name: "node_cpu_seconds_total",
		Type: "Sum",
		Min:  k.Limit.CPU * 1.2,
		Max:  k.Limit.CPU * 1.2,
		Tags: map[string]string{
			"resource":      "cpu",
			"net.host.name": k.PodName, // for this we assume each pod run on its own node.
			"cpu":           "0",
		},
	})

	metrics = append(metrics, Metric{
		Name: "node_cpu_seconds_total",
		Type: "Sum",
		Min:  k.Request.CPU * 0.7,
		Max:  math.Min(k.Request.CPU*1.2, k.Limit.CPU),
		Tags: map[string]string{
			"resource":      "cpu",
			"net.host.name": k.PodName, // for this we assume each pod run on its own node.
			"cpu":           "0",
		},
	})

	t := time.Minute
	metrics = append(metrics, Metric{
		Name:   "container_cpu_usage_seconds_total",
		Type:   "Sum",
		Period: &t,
		Min:    k.Request.CPU * 0.7,
		Max:    math.Min(k.Request.CPU*1.2, k.Limit.CPU),
		Tags: map[string]string{
			"pod":       k.PodName,
			"container": service.ServiceName,
			"image":     service.ServiceName,
			"namespace": k.Namespace,
		},
	})

	return metrics
}

func generateK8sName(n int) string {
	var letters = []rune("bcdfghjklmnpqrstvwxz2456789")

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
