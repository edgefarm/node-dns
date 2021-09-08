package config

type FeedConfig struct {
	// K8sapi configures the k8s api feed
	K8sapi K8sApiConfig
}

type K8sApiConfig struct {
	// Enabled indicates if the k8s api feed is used
	// default: false
	Enabled bool `json:"enabled"`
	// URI is where the api server is reacheble. Format: 'host:port', optional with 'http://' or 'https://'
	// default: 127.0.0.1:10550
	URI string `json:"URI"`
	// InsecureTLS indicates if there is any TLS certificate used that is self signed (optional)
	// default: true
	InsecureTLS bool `json:"insecureTLS"`
	// Token is the token to communicate with the API server (optional)
	// default: ""
	Token string `json:"token"`
}

func NewFeedConfig() *FeedConfig {
	return &FeedConfig{
		K8sapi: K8sApiConfig{
			Enabled:     true,
			URI:         "127.0.0.1:10550",
			InsecureTLS: true,
			Token:       "",
		},
	}
}
