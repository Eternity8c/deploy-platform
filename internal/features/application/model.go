package features_application

type Application struct {
	APIVersion string   `json:"apiVersion"`
	Kind       string   `json:"kind"`
	Metadata   Metadata `json:"metadata"`
	Spec       Spec     `json:"spec"`
}

type Metadata struct {
	Name string `json:"name"`
}

type Spec struct {
	Tenant   string            `json:"tenant"`
	Image    string            `json:"image"`
	Tag      string            `json:"tag"`
	Replicas int               `json:"replicas"`
	Port     int               `json:"port"`
	Env      map[string]string `json:"env"`
	Ingress  Ingress           `json:"ingress"`
}

type Ingress struct {
	Host string `json:"host"`
	Path string `json:"path"`
}
