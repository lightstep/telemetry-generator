package topology

type File struct {
	Topology *Topology `json:"topology"`
	RootRoutes []struct {
		Service       string `json:"service"`
		Route         string `json:"route"`
		TracesPerHour int    `json:"tracesPerHour"`
	} `json:"rootRoutes"`
}