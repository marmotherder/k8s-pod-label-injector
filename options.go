package main

// ServerOptions is used for the webhook server specific configuration
type ServerOptions struct {
	ServerPort     int    `short:"d" long:"port" description:"Port to run the server against" default:"443"`
	TLSCertPath    string `short:"c" long:"cert" description:"Path to the TLS certificate" required:"true"`
	TLSKeyPath     string `short:"l" long:"key" description:"Path to the TLS key" required:"true"`
	Hook           string `short:"h" long:"hook" description:"The identifier for this webhook" required:"true"`
	KubeConfigPath string `short:"k" long:"kcp" description:"Path to the kubeconfig file, will try to load as in cluster if not set" required:"false"`
}
