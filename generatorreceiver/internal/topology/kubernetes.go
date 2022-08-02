package topology

import (
	"math"
	"math/rand"
	"time"
)

const (
	defaultTargetCPU = 0.5
	defaultJitter    = 0.4
)

type Kubernetes struct {
	ClusterName string   `json:"cluster_name" yaml:"cluster_name"`
	Request     Resource `json:"request" yaml:"request"`
	Limit       Resource `json:"limit" yaml:"limit"`
	Usage       Usage    `json:"usage" yaml:"usage"`

	ReplicaSetName string
	Namespace      string
	PodName        string
	Container      string
}

type Resource struct {
	CPU    float64 `json:"cpu" yaml:"cpu"`
	Memory float64 `json:"memory" yaml:"memory"`
}

type Usage struct {
	CPU    ResourceUsage `json:"cpu" yaml:"cpu"`
	Memory ResourceUsage `json:"memory" yaml:"memory"`
}

type ResourceUsage struct {
	TargetPercentage float64 `json:"target" yaml:"target"`
	Jitter           float64 `json:"jitter" yaml:"jitter"`
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

	minute := time.Minute

	if k.Usage.CPU.TargetPercentage == 0 {
		k.Usage.CPU.TargetPercentage = defaultTargetCPU
	}

	if k.Usage.CPU.Jitter == 0 {
		k.Usage.CPU.Jitter = defaultJitter
	}

	metrics := []Metric{
		// kube_pod metrics
		{
			Name: "kube_pod_status_phase",
			Type: "Gauge",
			Min:  1,
			Max:  1,
			Tags: map[string]string{
				"phase": "Running",
				"pod":   k.PodName,
			},
		},
		{
			Name: "kube_pod_owner",
			Type: "Gauge",
			Min:  1,
			Max:  1,
			Tags: map[string]string{
				"pod":        k.PodName,
				"namespace":  service.ServiceName,
				"owner_name": k.ReplicaSetName,
				"owner_kind": "ReplicaSet",
			},
		},
		{
			Name: "kube_node_status_allocatable",
			Type: "Gauge",
			Min:  k.Limit.CPU * 1.2, // make the node a little bigger than the limit
			Max:  k.Limit.CPU * 1.2, // make the node a little bigger than the limit
			Tags: map[string]string{
				"resource": "cpu",
				"pod":      k.PodName, // used to created multiple time series that will be summed up.
			},
		},
		{
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
		},
		{
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
		},
		// node metrics
		{
			Name: "node_cpu_seconds_total",
			Type: "Sum",
			Min:  k.Limit.CPU * 1.2,
			Max:  k.Limit.CPU * 1.2,
			Tags: map[string]string{
				"resource":      "cpu",
				"net.host.name": k.PodName, // for this we assume each pod run on its own node.
				"cpu":           "0",
			},
		},
		{
			Name:   "node_cpu_seconds_total",
			Type:   "Sum",
			Min:    math.Max(k.Request.CPU*k.Usage.CPU.TargetPercentage*(1-k.Usage.CPU.Jitter/2), 0),
			Max:    math.Min(k.Request.CPU*k.Usage.CPU.TargetPercentage*(1+k.Usage.CPU.Jitter/2), k.Limit.CPU),
			Shape:  Average,
			Jitter: k.Usage.CPU.Jitter,
			Tags: map[string]string{
				"resource":      "cpu",
				"net.host.name": k.PodName, // for this we assume each pod run on its own node.
				"cpu":           "0",
			},
		},
		// container metrics
		{
			Name:   "container_cpu_usage_seconds_total",
			Type:   "Sum",
			Period: &minute,
			Min:    math.Max(k.Request.CPU*k.Usage.CPU.TargetPercentage*(1-k.Usage.CPU.Jitter), 0),
			Max:    math.Min(k.Request.CPU*k.Usage.CPU.TargetPercentage*(1+k.Usage.CPU.Jitter), k.Limit.CPU),
			Shape:  Average,
			Jitter: k.Usage.CPU.Jitter,
			Tags: map[string]string{
				"pod":       k.PodName,
				"container": service.ServiceName,
				"image":     service.ServiceName,
				"namespace": k.Namespace,
			},
		},
	}

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
