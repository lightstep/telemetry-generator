package topology

import (
	"github.com/lightstep/telemetry-generator/generatorreceiver/internal/flags"
	"go.uber.org/zap"
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

	PodName    = "$pod"
	Service    = "$service"
	Namespace  = "$namespace"
	Container  = "$container"
	Cluster    = "$cluster"
	ReplicaSet = "$replicaset"
	Deployment = "$deployment"
)

type Pod struct {
	StartTime       time.Time
	RestartDuration time.Duration
	PodName         string
	Container       string
	Kubernetes      *Kubernetes
}

type Kubernetes struct {
	ClusterName string   `json:"cluster_name" yaml:"cluster_name"`
	Request     Resource `json:"request" yaml:"request"`
	Limit       Resource `json:"limit" yaml:"limit"`
	Usage       Usage    `json:"usage" yaml:"usage"`
	Restart     Restart  `json:"restart" yaml:"restart"`
	PodCount    int      `json:"pod_count" yaml:"pod_count"`
	Deployment  string   `json:"deployment" yaml:"deployment"`

	ReplicaSetName string
	Service        string
	Namespace      string

	mutex sync.Mutex
	pods  []*Pod
	Cfg   *Config
}

type Resource struct {
	CPU    float64 `json:"cpu" yaml:"cpu"`
	Memory float64 `json:"memory" yaml:"memory"`
}

type Restart struct {
	Every  time.Duration `json:"every" yaml:"every"`
	Jitter time.Duration `json:"jitter" yaml:"jitter"`
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

func (k *Kubernetes) CreatePods(serviceName string, random *rand.Rand) {
	k.mutex.Lock()
	defer k.mutex.Unlock()
	k.ReplicaSetName = serviceName + "-" + generateK8sName(10, random)
	k.Namespace = serviceName
	k.Service = serviceName
	k.pods = make([]*Pod, k.GetPodCount())
	for i := 0; i < len(k.pods); i++ {
		k.pods[i] = &Pod{
			StartTime:       time.Now(),
			PodName:         k.ReplicaSetName + "-" + generateK8sName(5, random),
			Container:       serviceName,
			Kubernetes:      k,
			RestartDuration: k.RestartDurationWithJitter(random),
		}
	}
}

func (k *Kubernetes) GetPodCount() int {
	if k.PodCount > 0 {
		return k.PodCount
	} else if k.Cfg != nil && k.Cfg.Kubernetes.PodCount > 0 {
		return k.Cfg.Kubernetes.PodCount
	} else {
		return 1
	}
}

func (p *Pod) RestartIfNeeded(flags flags.EmbeddedFlags, logger *zap.Logger, random *rand.Rand) bool {
	if p == nil || p.Kubernetes.Restart.Every == 0 {
		return false
	}

	p.Kubernetes.mutex.Lock()
	defer p.Kubernetes.mutex.Unlock()

	flagTime := flags.GenerateStartTime()
	if flagTime.After(p.StartTime) {
		// consider that the pod started at the time that a flag was enabled/disabled.
		// TODO: restart with some jitter
		p.restart(logger, random)
		return true
	} else if time.Since(p.StartTime) >= p.RestartDuration {
		// TODO: restart with some jitter
		p.restart(logger, random)
		return true
	}
	return false

}

func (p *Pod) restart(logger *zap.Logger, random *rand.Rand) {
	// this is locked by RestartIfNeeded
	p.StartTime = time.Now()
	p.RestartDuration = p.Kubernetes.RestartDurationWithJitter(random)
	p.PodName = p.Kubernetes.ReplicaSetName + "-" + generateK8sName(5, random)
	logger.Info("pod restarted", zap.String("service", p.Kubernetes.Service), zap.String("pod", p.PodName))
}

func (k *Kubernetes) randomPod(random *rand.Rand) *Pod {
	return k.pods[random.Intn(len(k.pods))]
}

// only called from tag generator!
func (k *Kubernetes) GetRandomK8sTags(random *rand.Rand) map[string]string {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	pod := k.randomPod(random)
	// ref: https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/resource/semantic_conventions/k8s.md
	return k.GetK8sTags(pod)
}

func (k *Kubernetes) GetK8sTags(pod *Pod) map[string]string {
	return map[string]string{
		"k8s.cluster.name":    k.ClusterName,
		"k8s.pod.name":        pod.PodName,
		"k8s.namespace.name":  k.Namespace,
		"k8s.container.name":  pod.Container,
		"k8s.deployment.name": k.Deployment,
	}
}

func (p *Pod) ReplaceTags(tags map[string]string) map[string]string {
	p.Kubernetes.mutex.Lock()
	defer p.Kubernetes.mutex.Unlock()

	replaced := make(map[string]string, len(tags))
	for key, value := range tags {
		switch value {
		case PodName:
			replaced[key] = p.PodName
		case Service:
			replaced[key] = p.Kubernetes.Service
		case Namespace:
			replaced[key] = p.Kubernetes.Namespace
		case Container:
			replaced[key] = p.Container
		case Cluster:
			replaced[key] = p.Kubernetes.ClusterName
		case Deployment:
			replaced[key] = p.Kubernetes.Deployment
		case ReplicaSet:
			replaced[key] = p.Kubernetes.ReplicaSetName
		default:
			replaced[key] = value
		}
	}

	return replaced
}

func (k *Kubernetes) RestartDurationWithJitter(random *rand.Rand) time.Duration {
	return k.Restart.Every + time.Duration(float64(k.Restart.Jitter)*(random.Float64()-0.5))
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

	var metrics []Metric

	for _, pod := range k.pods {
		k8sTags := k.GetK8sTags(pod)
		podMetrics := []Metric{
			// kube_pod metrics
			{
				Name: "kube_pod_status_phase",
				Type: "Gauge",
				Min:  1,
				Max:  1,
				Tags: mergeTags(k8sTags, map[string]string{
					"phase": "Running",
					"pod":   PodName,
				}),
			},
			{
				Name: "kube_pod_owner",
				Type: "Gauge",
				Min:  1,
				Max:  1,
				Tags: mergeTags(k8sTags, map[string]string{
					"pod":        PodName,
					"namespace":  Namespace,
					"owner_name": ReplicaSet,
					"owner_kind": "ReplicaSet",
				}),
			},
			{
				Name: "kube_node_status_allocatable",
				Type: "Gauge",
				Min:  cpuTotal,
				Max:  cpuTotal,
				Tags: mergeTags(k8sTags, map[string]string{
					"resource": "cpu",
					"pod":      PodName, // used to created multiple time series that will be summed up.
				}),
			},
			{
				Name: "kube_node_status_allocatable",
				Type: "Gauge",
				Min:  memTotal,
				Max:  memTotal,
				Tags: mergeTags(k8sTags, map[string]string{
					"resource": "memory",
					"pod":      PodName, // used to created multiple time series that will be summed up.
				}),
			},
			{
				Name: "kube_pod_container_resource_requests",
				Type: "Gauge",
				Min:  k.Request.CPU,
				Max:  k.Request.CPU,
				Tags: mergeTags(k8sTags, map[string]string{
					"resource":  "cpu",
					"namespace": Namespace,
					"container": Container,
					"pod":       PodName,
				}),
			},
			{
				Name: "kube_pod_container_resource_requests",
				Type: "Gauge",
				Min:  k.Request.Memory * megabyte,
				Max:  k.Request.Memory * megabyte,
				Tags: mergeTags(k8sTags, map[string]string{
					"resource":  "memory",
					"namespace": Namespace,
					"container": Container,
					"pod":       PodName,
				}),
			},
			{
				Name: "kube_pod_container_resource_limits",
				Type: "Gauge",
				Min:  k.Limit.CPU,
				Max:  k.Limit.CPU,
				Tags: mergeTags(k8sTags, map[string]string{
					"resource":  "cpu",
					"namespace": Namespace,
					"container": Container,
					"pod":       PodName,
				}),
			},
			{
				Name: "kube_pod_container_resource_limits",
				Type: "Gauge",
				Min:  k.Limit.Memory * megabyte,
				Max:  k.Limit.Memory * megabyte,
				Tags: mergeTags(k8sTags, map[string]string{
					"resource":  "memory",
					"namespace": Namespace,
					"container": Container,
					"pod":       PodName,
				}),
			},
			// node metrics
			{
				Name: "node_cpu_seconds_total",
				Type: "Sum",
				Min:  k.Limit.CPU * 1.2,
				Max:  k.Limit.CPU * 1.2,
				Tags: mergeTags(k8sTags, map[string]string{
					"resource":      "cpu",
					"net.host.name": PodName, // for this we assume each pod run on its own node.
					"cpu":           "0",
				}),
			},
			{
				Name:   "node_cpu_seconds_total",
				Type:   "Sum",
				Period: &minute,
				Min:    math.Max(cpuTarget*(1-cpuJitter), 0),
				Max:    math.Min(cpuTarget*(1+cpuJitter), k.Limit.CPU),
				Shape:  Average,
				Jitter: k.Usage.CPU.Jitter,
				Tags: mergeTags(k8sTags, map[string]string{
					"resource":      "cpu",
					"net.host.name": PodName, // for this we assume each pod run on its own node.
					"cpu":           "0",
				}),
			},

			{
				Name:   "node_memory_MemAvailable_bytes",
				Type:   "Gauge",
				Min:    math.Max(memTotal-memTarget*(1+memJitter), 0),
				Max:    math.Min(memTotal-memTarget*(1-memJitter), k.Limit.Memory*megabyte),
				Shape:  Average,
				Jitter: k.Usage.Memory.Jitter,
				Tags: mergeTags(k8sTags, map[string]string{
					"net.host.name": PodName, // for this we assume each pod run on its own node.
				}),
			},

			{
				Name:   "node_memory_MemTotal_bytes",
				Type:   "Gauge",
				Min:    memTotal,
				Max:    memTotal,
				Jitter: k.Usage.Memory.Jitter,
				Tags: mergeTags(k8sTags, map[string]string{
					"net.host.name": PodName, // for this we assume each pod run on its own node.
				}),
			},
			{
				Name:   "container_fs_reads_total",
				Type:   "Sum",
				Min:    math.Max(diskTarget*(1-diskJitter), 0),
				Max:    diskTarget * (1 + diskJitter),
				Shape:  Average,
				Jitter: k.Usage.Disk.Jitter,
				Tags: mergeTags(k8sTags, map[string]string{
					"job":          "kubelet",
					"metrics_path": "/metrics/cadvisor",
					"container":    Container,
					"device":       "/dev/sda",
					"namespace":    Namespace,
				}),
			},
			{
				Name:   "container_fs_writes_total",
				Type:   "Sum",
				Min:    math.Max(diskTarget*(1-diskJitter), 0),
				Max:    diskTarget * (1 + diskJitter),
				Shape:  Average,
				Jitter: k.Usage.Disk.Jitter,
				Tags: mergeTags(k8sTags, map[string]string{
					"job":          "kubelet",
					"metrics_path": "/metrics/cadvisor",
					"container":    Container,
					"device":       "/dev/sda",
					"namespace":    Namespace,
				}),
			},
			{
				Name:   "container_fs_reads_bytes_total",
				Type:   "Sum",
				Min:    math.Max(diskTarget*(1-diskJitter), 0),
				Max:    diskTarget * (1 + diskJitter),
				Shape:  Average,
				Jitter: k.Usage.Disk.Jitter,
				Tags: mergeTags(k8sTags, map[string]string{
					"job":          "kubelet",
					"metrics_path": "/metrics/cadvisor",
					"container":    Container,
					"device":       "/dev/sda",
					"namespace":    Namespace,
				}),
			},
			{
				Name:   "container_fs_writes_bytes_total",
				Type:   "Sum",
				Min:    math.Max(diskTarget*(1-diskJitter), 0),
				Max:    diskTarget * (1 + diskJitter),
				Shape:  Average,
				Jitter: k.Usage.Disk.Jitter,
				Tags: mergeTags(k8sTags, map[string]string{
					"job":          "kubelet",
					"metrics_path": "/metrics/cadvisor",
					"container":    Container,
					"device":       "/dev/sda",
					"namespace":    Namespace,
				}),
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
				Tags: mergeTags(k8sTags, map[string]string{
					"pod":        PodName,
					"container":  Container,
					"image":      Service,
					"namespace":  Namespace,
					"deployment": Deployment,
				}),
			},
			{
				Name:   "container_network_receive_bytes_total",
				Type:   "Sum",
				Min:    networkTarget * (1 + networkJitter),
				Max:    networkTarget * (2000 + networkJitter),
				Shape:  Average,
				Jitter: k.Usage.Network.Jitter,
				Tags: mergeTags(k8sTags, map[string]string{
					"image": Service,
				}),
			},
			{
				Name:   "container_network_transmit_bytes_total",
				Type:   "Sum",
				Min:    networkTarget * (1 + networkJitter),
				Max:    networkTarget * (2000 + networkJitter),
				Shape:  Average,
				Jitter: k.Usage.Network.Jitter,
				Tags: mergeTags(k8sTags, map[string]string{
					"image": Service,
				}),
			},
			{
				Name:   "container_network_receive_packets_total",
				Type:   "Sum",
				Min:    math.Max(networkTarget*(1-networkJitter), 0),
				Max:    networkTarget * (1 + networkJitter),
				Shape:  Average,
				Jitter: k.Usage.Network.Jitter,
				Tags: mergeTags(k8sTags, map[string]string{
					"image": Service,
				}),
			},
			{
				Name:   "container_network_transmit_packets_total",
				Type:   "Sum",
				Min:    math.Max(networkTarget*(1-networkJitter), 0),
				Max:    networkTarget * (1 + networkJitter),
				Shape:  Average,
				Jitter: k.Usage.Network.Jitter,
				Tags: mergeTags(k8sTags, map[string]string{
					"image": Service,
				}),
			},

			// container metrics
			{
				Name:   "container_cpu_usage_seconds_total",
				Type:   "Sum",
				Min:    math.Max(cpuTarget*(1-cpuJitter), 0),
				Max:    math.Min(cpuTarget*(1+cpuJitter), k.Limit.CPU),
				Shape:  Average,
				Jitter: k.Usage.CPU.Jitter,
				Tags: mergeTags(k8sTags, map[string]string{
					"pod":       PodName,
					"container": Container,
					"image":     Service,
					"namespace": Namespace,
				}),
			},
			{
				Name:   "container_fs_reads_total",
				Type:   "Sum",
				Min:    math.Max(diskTarget*(1-diskJitter), 0),
				Max:    diskTarget * (1 + diskJitter),
				Shape:  Average,
				Jitter: k.Usage.Disk.Jitter,
				Tags: mergeTags(k8sTags, map[string]string{
					"job":          "kubelet",
					"metrics_path": "/metrics/cadvisor",
					"container":    Container,
					"device":       "/dev/sda",
					"namespace":    Namespace,
				}),
			},
			{
				Name:   "container_fs_writes_total",
				Type:   "Sum",
				Min:    math.Max(diskTarget*(1-diskJitter), 0),
				Max:    diskTarget * (1 + diskJitter),
				Shape:  Average,
				Jitter: k.Usage.Disk.Jitter,
				Tags: mergeTags(k8sTags, map[string]string{
					"job":          "kubelet",
					"metrics_path": "/metrics/cadvisor",
					"container":    Container,
					"device":       "/dev/sda",
					"namespace":    Namespace,
				}),
			},
			{
				Name:   "container_fs_reads_bytes_total",
				Type:   "Sum",
				Min:    math.Max(diskTarget*(1-diskJitter), 0),
				Max:    diskTarget * (1 + diskJitter),
				Shape:  Average,
				Jitter: k.Usage.Disk.Jitter,
				Tags: mergeTags(k8sTags, map[string]string{
					"job":          "kubelet",
					"metrics_path": "/metrics/cadvisor",
					"container":    Container,
					"device":       "/dev/sda",
					"namespace":    Namespace,
				}),
			},
			{
				Name:   "container_fs_writes_bytes_total",
				Type:   "Sum",
				Min:    math.Max(diskTarget*(1-diskJitter), 0),
				Max:    diskTarget * (1 + diskJitter),
				Shape:  Average,
				Jitter: k.Usage.Disk.Jitter,
				Tags: mergeTags(k8sTags, map[string]string{
					"job":          "kubelet",
					"metrics_path": "/metrics/cadvisor",
					"container":    Container,
					"device":       "/dev/sda",
					"namespace":    Namespace,
				}),
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
				Tags: mergeTags(k8sTags, map[string]string{
					"pod":       PodName,
					"container": Container,
					"image":     Service,
					"namespace": Namespace,
				}),
			},
			{
				Name:   "container_network_receive_bytes_total",
				Type:   "Sum",
				Min:    networkTarget * (1 + networkJitter),
				Max:    networkTarget * (2000 + networkJitter),
				Shape:  Average,
				Jitter: k.Usage.Network.Jitter,
				Tags: mergeTags(k8sTags, map[string]string{
					"image": Service,
				}),
			},
			{
				Name:   "container_network_transmit_bytes_total",
				Type:   "Sum",
				Min:    networkTarget * (1 + networkJitter),
				Max:    networkTarget * (2000 + networkJitter),
				Shape:  Average,
				Jitter: k.Usage.Network.Jitter,
				Tags: mergeTags(k8sTags, map[string]string{
					"image": Service,
				}),
			},
			{
				Name:   "container_network_receive_packets_total",
				Type:   "Sum",
				Min:    math.Max(networkTarget*(1-networkJitter), 0),
				Max:    networkTarget * (1 + networkJitter),
				Shape:  Average,
				Jitter: k.Usage.Network.Jitter,
				Tags: mergeTags(k8sTags, map[string]string{
					"image": Service,
				}),
			},
			{
				Name:   "container_network_transmit_packets_total",
				Type:   "Sum",
				Min:    math.Max(networkTarget*(1-networkJitter), 0),
				Max:    networkTarget * (1 + networkJitter),
				Shape:  Average,
				Jitter: k.Usage.Network.Jitter,
				Tags: mergeTags(k8sTags, map[string]string{
					"image": Service,
				}),
			},
		}

		for i := range podMetrics {
			podMetrics[i].Pod = pod
			metrics = append(metrics, podMetrics[i])
		}

	}

	return metrics
}

func generateK8sName(nameLength int, random *rand.Rand) string {
	var letters = []rune("bcdfghjklmnpqrstvwxz2456789")

	b := make([]rune, nameLength)
	for i := range b {
		b[i] = letters[random.Intn(len(letters))]
	}
	return string(b)
}
