package topology

import (
	"math"
	"math/rand"
	"time"
)

const (
	defaultTarget  = 0.5
	defaultJitter  = 0.4
	defaultDisk    = 100
	defaultNetwork = 800
	megabyte       = 1024 * 1024
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
	CPU     ResourceUsage `json:"cpu" yaml:"cpu"`
	Memory  ResourceUsage `json:"memory" yaml:"memory"`
	Disk    ResourceUsage `json:"disk" yaml:"disk"`
	Network ResourceUsage `json:"network" yaml:"network"`
}

type ResourceUsage struct {
	Target float64 `json:"target" yaml:"target"`
	Jitter float64 `json:"jitter" yaml:"jitter"`
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

	if k.Usage.CPU.Target == 0 {
		k.Usage.CPU.Target = defaultTarget
	}

	if k.Usage.Memory.Target == 0 {
		k.Usage.Memory.Target = defaultTarget
	}

	if k.Usage.CPU.Jitter == 0 {
		k.Usage.CPU.Jitter = defaultJitter
	}

	if k.Usage.Disk.Target == 0 {
		k.Usage.Disk.Target = defaultDisk
	}

	if k.Usage.Disk.Jitter == 0 {
		k.Usage.Disk.Jitter = defaultJitter
	}

	if k.Usage.Memory.Jitter == 0 {
		k.Usage.Memory.Jitter = defaultJitter
	}

	if k.Usage.Network.Target == 0 {
		k.Usage.Network.Target = defaultNetwork
	}

	if k.Usage.Network.Jitter == 0 {
		k.Usage.Network.Jitter = defaultJitter
	}

	cpuTarget := k.Request.CPU * k.Usage.CPU.Target
	cpuJitter := k.Usage.CPU.Jitter / 2
	cpuTotal := k.Limit.CPU * 1.2 // make the node a little bigger than the limit

	diskTarget := k.Usage.Disk.Target
	diskJitter := k.Usage.Disk.Jitter / 2

	memTarget := k.Request.Memory * megabyte * k.Usage.Memory.Target
	memJitter := k.Usage.Memory.Jitter / 2
	memTotal := k.Limit.Memory * megabyte * 1.2 // make the node a little bigger than the limit

	networkTarget := k.Usage.Network.Target
	networkJitter := k.Usage.Network.Jitter / 2

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
			Min:  cpuTotal,
			Max:  cpuTotal,
			Tags: map[string]string{
				"resource": "cpu",
				"pod":      k.PodName, // used to created multiple time series that will be summed up.
			},
		},
		{
			Name: "kube_node_status_allocatable",
			Type: "Gauge",
			Min:  memTotal,
			Max:  memTotal,
			Tags: map[string]string{
				"resource": "memory",
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
			Name: "kube_pod_container_resource_requests",
			Type: "Gauge",
			Min:  k.Request.Memory * megabyte,
			Max:  k.Request.Memory * megabyte,
			Tags: map[string]string{
				"resource":  "memory",
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
		{
			Name: "kube_pod_container_resource_limits",
			Type: "Gauge",
			Min:  k.Limit.Memory * megabyte,
			Max:  k.Limit.Memory * megabyte,
			Tags: map[string]string{
				"resource":  "memory",
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
			Period: &minute,
			Min:    math.Max(cpuTarget*(1-cpuJitter), 0),
			Max:    math.Min(cpuTarget*(1+cpuJitter), k.Limit.CPU),
			Shape:  Average,
			Jitter: k.Usage.CPU.Jitter,
			Tags: map[string]string{
				"resource":      "cpu",
				"net.host.name": k.PodName, // for this we assume each pod run on its own node.
				"cpu":           "0",
			},
		},

		{
			Name:   "node_memory_MemAvailable_bytes",
			Type:   "Gauge",
			Min:    math.Max(memTotal-memTarget*(1+memJitter), 0),
			Max:    math.Min(memTotal-memTarget*(1-memJitter), k.Limit.Memory*megabyte),
			Shape:  Average,
			Jitter: k.Usage.Memory.Jitter,
			Tags: map[string]string{
				"net.host.name": k.PodName, // for this we assume each pod run on its own node.
			},
		},

		{
			Name:   "node_memory_MemTotal_bytes",
			Type:   "Gauge",
			Min:    memTotal,
			Max:    memTotal,
			Jitter: k.Usage.Memory.Jitter,
			Tags: map[string]string{
				"net.host.name": k.PodName, // for this we assume each pod run on its own node.
			},
		},

		// container metrics
		{
			Name:   "container_cpu_usage_seconds_total",
			Type:   "Sum",
			Min:    math.Max(cpuTarget*(1-cpuJitter), 0),
			Max:    math.Min(cpuTarget*(1+cpuJitter), k.Limit.CPU),
			Shape:  Average,
			Jitter: k.Usage.CPU.Jitter,
			Tags: map[string]string{
				"pod":       k.PodName,
				"container": service.ServiceName,
				"image":     service.ServiceName,
				"namespace": k.Namespace,
			},
		},
		{
			Name:   "container_fs_reads_total",
			Type:   "Sum",
			Min:    math.Max(diskTarget*(1-diskJitter), 0),
			Max:    diskTarget * (1 + diskJitter),
			Shape:  Average,
			Jitter: k.Usage.Disk.Jitter,
			Tags: map[string]string{
				"job":          "kubelet",
				"metrics_path": "/metrics/cadvisor",
				"container":    service.ServiceName,
				"device":       "/dev/sda",
				"namespace":    k.Namespace,
			},
		},
		{
			Name:   "container_fs_writes_total",
			Type:   "Sum",
			Min:    math.Max(diskTarget*(1-diskJitter), 0),
			Max:    diskTarget * (1 + diskJitter),
			Shape:  Average,
			Jitter: k.Usage.Disk.Jitter,
			Tags: map[string]string{
				"job":          "kubelet",
				"metrics_path": "/metrics/cadvisor",
				"container":    service.ServiceName,
				"device":       "/dev/sda",
				"namespace":    k.Namespace,
			},
		},
		{
			Name:   "container_fs_reads_bytes_total",
			Type:   "Sum",
			Min:    math.Max(diskTarget*(1-diskJitter), 0),
			Max:    diskTarget * (1 + diskJitter),
			Shape:  Average,
			Jitter: k.Usage.Disk.Jitter,
			Tags: map[string]string{
				"job":          "kubelet",
				"metrics_path": "/metrics/cadvisor",
				"container":    service.ServiceName,
				"device":       "/dev/sda",
				"namespace":    k.Namespace,
			},
		},
		{
			Name:   "container_fs_writes_bytes_total",
			Type:   "Sum",
			Min:    math.Max(diskTarget*(1-diskJitter), 0),
			Max:    diskTarget * (1 + diskJitter),
			Shape:  Average,
			Jitter: k.Usage.Disk.Jitter,
			Tags: map[string]string{
				"job":          "kubelet",
				"metrics_path": "/metrics/cadvisor",
				"container":    service.ServiceName,
				"device":       "/dev/sda",
				"namespace":    k.Namespace,
			},
		},
		{
			Name:   "container_memory_working_set_bytes",
			Type:   "Gauge",
			Period: &minute,
			Min:    math.Max(memTarget*(1-memJitter), 0),
			Max:    math.Min(memTarget*(1+memJitter), k.Limit.Memory*megabyte),
			Shape:  Average,
			Jitter: k.Usage.Memory.Jitter,
			Tags: map[string]string{
				"pod":       k.PodName,
				"container": service.ServiceName,
				"image":     service.ServiceName,
				"namespace": k.Namespace,
			},
		},
		{
			Name:   "container_network_receive_bytes_total",
			Type:   "Sum",
			Min:    networkTarget * (1 + networkJitter),
			Max:    networkTarget * (2000 + networkJitter),
			Shape:  Average,
			Jitter: k.Usage.Network.Jitter,
			Tags: map[string]string{
				"image": service.ServiceName,
			},
		},
		{
			Name:   "container_network_transmit_bytes_total",
			Type:   "Sum",
			Min:    networkTarget * (1 + networkJitter),
			Max:    networkTarget * (2000 + networkJitter),
			Shape:  Average,
			Jitter: k.Usage.Network.Jitter,
			Tags: map[string]string{
				"image": service.ServiceName,
			},
		},
		{
			Name:   "container_network_receive_packets_total",
			Type:   "Sum",
			Min:    math.Max(networkTarget*(1-networkJitter), 0),
			Max:    networkTarget * (1 + networkJitter),
			Shape:  Average,
			Jitter: k.Usage.Network.Jitter,
			Tags: map[string]string{
				"image": service.ServiceName,
			},
		},
		{
			Name:   "container_network_transmit_packets_total",
			Type:   "Sum",
			Min:    math.Max(networkTarget*(1-networkJitter), 0),
			Max:    networkTarget * (1 + networkJitter),
			Shape:  Average,
			Jitter: k.Usage.Network.Jitter,
			Tags: map[string]string{
				"image": service.ServiceName,
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
