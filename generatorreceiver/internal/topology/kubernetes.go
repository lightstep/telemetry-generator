package topology

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

const (
	defaultTarget  = 0.5
	defaultJitter  = 0.4
	defaultDisk    = 100
	defaultNetwork = 800
	megabyte       = 1024 * 1024

	// Templated variables, these will get replaced with real values at metric generation with ReplaceTags.

	Pod        = "$pod"
	Service    = "$service"
	Namespace  = "$namespace"
	Container  = "$container"
	Cluster    = "$cluster"
	ReplicaSet = "$replicaset"
)

type Kubernetes struct {
	ClusterName string   `json:"cluster_name" yaml:"cluster_name"`
	Request     Resource `json:"request" yaml:"request"`
	Limit       Resource `json:"limit" yaml:"limit"`
	Usage       Usage    `json:"usage" yaml:"usage"`
	Restart     Restart  `json:"restart" yaml:"restart"`

	mutex          sync.Mutex
	StartTime      time.Time `json:"-" yaml:"-"`
	Service        string    `json:"-" yaml:"-"`
	ReplicaSetName string    `json:"-" yaml:"-"`
	Namespace      string    `json:"-" yaml:"-"`
	PodName        string    `json:"-" yaml:"-"`
	Container      string    `json:"-" yaml:"-"`
}

type Resource struct {
	CPU    float64 `json:"cpu" yaml:"cpu"`
	Memory float64 `json:"memory" yaml:"memory"`
}

type Restart struct {
	Every time.Duration `json:"every" yaml:"every"`
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

func (k *Kubernetes) CreatePod(serviceName string) {
	k.mutex.Lock()
	defer k.mutex.Unlock()
	k.StartTime = time.Now()
	k.ReplicaSetName = serviceName + "-" + generateK8sName(10)
	k.PodName = k.ReplicaSetName + "-" + generateK8sName(5)
	k.Namespace = serviceName
	k.Container = serviceName
	k.Service = serviceName
}

func (k *Kubernetes) RestartPod() {
	k.mutex.Lock()
	defer k.mutex.Unlock()
	k.StartTime = time.Now()
	k.PodName = k.ReplicaSetName + "-" + generateK8sName(5)
}

func (k *Kubernetes) GetK8sTags() map[string]string {
	k.mutex.Lock()
	defer k.mutex.Unlock()
	// ref: https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/resource/semantic_conventions/k8s.md
	return map[string]string{
		"k8s.cluster.name":   k.ClusterName,
		"k8s.pod.name":       k.PodName,
		"k8s.namespace.name": k.Namespace,
		"k8s.container.name": k.Container,
	}
}

func (k *Kubernetes) ReplaceTags(tags map[string]string) map[string]string {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	replaced := make(map[string]string, len(tags))
	for key, value := range tags {
		switch value {
		case Pod:
			replaced[key] = k.PodName
		case Service:
			replaced[key] = k.Service
		case Namespace:
			replaced[key] = k.Namespace
		case Container:
			replaced[key] = k.Container
		case Cluster:
			replaced[key] = k.ClusterName
		case ReplicaSet:
			replaced[key] = k.ReplicaSetName
		default:
			replaced[key] = value
		}
	}

	return replaced
}

func (k *Kubernetes) GenerateMetrics() []Metric {
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

	restart := 1.
	memoryShape := Average
	if k.Restart.Every != 0 {
		restart = 0
		memoryShape = Leaking
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
				"pod":   Pod,
			},
		},
		{
			Name: "kube_pod_owner",
			Type: "Gauge",
			Min:  1,
			Max:  1,
			Tags: map[string]string{
				"pod":        Pod,
				"namespace":  Namespace,
				"owner_name": ReplicaSet,
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
				"pod":      Pod, // used to created multiple time series that will be summed up.
			},
		},
		{
			Name: "kube_node_status_allocatable",
			Type: "Gauge",
			Min:  memTotal,
			Max:  memTotal,
			Tags: map[string]string{
				"resource": "memory",
				"pod":      Pod, // used to created multiple time series that will be summed up.
			},
		},
		{
			Name: "kube_pod_container_resource_requests",
			Type: "Gauge",
			Min:  k.Request.CPU,
			Max:  k.Request.CPU,
			Tags: map[string]string{
				"resource":  "cpu",
				"namespace": Namespace,
				"container": Container,
				"pod":       Pod,
			},
		},
		{
			Name: "kube_pod_container_resource_requests",
			Type: "Gauge",
			Min:  k.Request.Memory * megabyte,
			Max:  k.Request.Memory * megabyte,
			Tags: map[string]string{
				"resource":  "memory",
				"namespace": Namespace,
				"container": Container,
				"pod":       Pod,
			},
		},
		{
			Name: "kube_pod_container_resource_limits",
			Type: "Gauge",
			Min:  k.Limit.CPU,
			Max:  k.Limit.CPU,
			Tags: map[string]string{
				"resource":  "cpu",
				"namespace": Namespace,
				"container": Container,
				"pod":       Pod,
			},
		},
		{
			Name: "kube_pod_container_resource_limits",
			Type: "Gauge",
			Min:  k.Limit.Memory * megabyte,
			Max:  k.Limit.Memory * megabyte,
			Tags: map[string]string{
				"resource":  "memory",
				"namespace": Namespace,
				"container": Container,
				"pod":       Pod,
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
				"net.host.name": Pod, // for this we assume each pod run on its own node.
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
				"net.host.name": Pod, // for this we assume each pod run on its own node.
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
				"net.host.name": Pod, // for this we assume each pod run on its own node.
			},
		},

		{
			Name:   "node_memory_MemTotal_bytes",
			Type:   "Gauge",
			Min:    memTotal,
			Max:    memTotal,
			Jitter: k.Usage.Memory.Jitter,
			Tags: map[string]string{
				"net.host.name": Pod, // for this we assume each pod run on its own node.
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
				"pod":       Pod,
				"container": Container,
				"image":     Service,
				"namespace": Namespace,
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
				"container":    Container,
				"device":       "/dev/sda",
				"namespace":    Namespace,
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
				"container":    Container,
				"device":       "/dev/sda",
				"namespace":    Namespace,
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
				"container":    Container,
				"device":       "/dev/sda",
				"namespace":    Namespace,
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
				"container":    Container,
				"device":       "/dev/sda",
				"namespace":    Namespace,
			},
		},
		{
			Name:   "container_memory_working_set_bytes",
			Type:   "Gauge",
			Period: &minute,
			// If k.restart.every is set, min should be 0 and max should be k.Limit.memory
			Min:    math.Max(memTarget*(1-memJitter)*restart, 0),
			Max:    math.Min(memTarget*(1+memJitter)+k.Limit.Memory*megabyte*(1-restart), k.Limit.Memory*megabyte),
			Shape:  memoryShape,
			Jitter: k.Usage.Memory.Jitter,
			Tags: map[string]string{
				"pod":       Pod,
				"container": Container,
				"image":     Service,
				"namespace": Namespace,
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
				"image": Service,
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
				"image": Service,
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
				"image": Service,
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
				"image": Service,
			},
		},
	}

	for i := range metrics {
		metrics[i].Kubernetes = k
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
