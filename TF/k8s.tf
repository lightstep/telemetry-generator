terraform {
  required_providers {
    lightstep = {
      source  = "lightstep/lightstep"
      version = "~> 1.60.2"
    }
  }
  required_version = ">= v1.0.11"
}
provider "lightstep" {
  api_key         = "eyJhbGciOiJIUzI1NiIsImtpZCI6IjIwMTktMDMtMDEiLCJ0eXAiOiJKV1QifQ.eyJzY3AiOnsicm9sZSI6ImU1MTI5NmJkLTFjYjktMTFlOC05M2Y1LTQyMDEwYWYwMGFkNiJ9LCJ2ZXIiOjEsImRlYnVnIjp7Im9yZyI6IkxpZ2h0U3RlcCIsInJvbGUiOiJPcmdhbml6YXRpb24gQWRtaW4ifSwiYXVkIjoiYXBwLmxpZ2h0c3RlcC5jb20iLCJleHAiOjE2OTI5MTM4MzksImp0aSI6InByemlta2ttcXFiYzV3NjJ1NW1rYnA2MmM0dXUzdmx1M3BiNmozYmdmY2pjZ3pzdCIsImlhdCI6MTY2MTM3NzgzOSwiaXNzIjoibGlnaHRzdGVwLmNvbSIsInN1YiI6IjEyOWY2ZWIzNGE3ZGJjZDcxZTI2NDM1Yzc4OSJ9.er8uyGmBvC98Fw0LiRl3CBKomEuGpHFDrhRdGn94b6Y"
  organization    = "LightStep"
  environment     = "public"
}
module "kube-dashboards" {
  source            = "./collector-dashboards/otel-collector-kubernetes-dashboard"
  lightstep_project = "dev-parker"
  workloads = [
    {
      namespace = "cert-manager"
      workload  = "cert-manager"
    },
    {
      namespace = "kube-system"
      workload  = "fluentbit-gke"
    },
    {
      namespace = "kube-system"
      workload  = "gke-metrics-agent"
    },
    {
      namespace = "kube-system"
      workload  = "konnectivity-agent"
    },
    {
      namespace = "testapp"
      workload  = "testapp"
    }
  ]
}