package dto

import "time"

type APIConfig struct {
	Server   serverConfig   `mapstructure:"server"`
	Database databaseConfig `mapstructure:"database"`
	Nats     natsConfig     `mapstructure:"nats"`
}

type serverConfig struct {
	Port            string        `mapstructure:"port"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"db_name"`
}

// databaseConfig is an alias for backward compatibility
type databaseConfig = DatabaseConfig

type NatsConfig struct {
	URL      string         `mapstructure:"url"`
	Producer ProducerConfig `mapstructure:"producer"`
}

// ProducerConfig holds NATS producer stream and channel names
type ProducerConfig struct {
	StreamName               string `mapstructure:"stream_name"`
	DeploymentRequestChannel string `mapstructure:"deployment_request_channel"`
	DeploymentUpdateChannel  string `mapstructure:"deployment_update_channel"`
}

// WorkerConfig holds configuration for the worker consumer
type WorkerConfig struct {
	Database DatabaseConfig `mapstructure:"database"`
	K8s      K8sConfig      `mapstructure:"k8s"`
	Nats     NatsConfig     `mapstructure:"nats"`
	Consumer ConsumerConfig `mapstructure:"consumer"`
	Watcher  WatcherConfig  `mapstructure:"watcher"`
}

// WatcherConfig holds configuration for the deployment informer (resync, task timeout)
type WatcherConfig struct {
	ResyncPeriod  time.Duration `mapstructure:"resync_period"`
	TaskTimeout   time.Duration `mapstructure:"task_timeout"`
}

// K8sConfig holds Kubernetes client configuration
type K8sConfig struct {
	// InCluster when true uses in-cluster config (service account). When false uses kubeconfig.
	InCluster bool `mapstructure:"in_cluster"`
	// Kubeconfig path when InCluster is false. Empty uses default (KUBECONFIG env or ~/.kube/config).
	Kubeconfig string `mapstructure:"kubeconfig"`
	// ManagerTag is the value for the managed-by label on created resources (from config key manager-tag).
	ManagerTag string `mapstructure:"manager_tag"`
}

// ConsumerConfig holds worker consumer settings
type ConsumerConfig struct {
	ShutdownTimeout       time.Duration      `mapstructure:"shutdown_timeout"`
	DeploymentRequestTask ConsumerTypeConfig `mapstructure:"deployment_request_task"`
	DeploymentUpdateTask  ConsumerTypeConfig `mapstructure:"deployment_update_task"`
}

// ConsumerTypeConfig holds per-task-type configuration for a consumer route
type ConsumerTypeConfig struct {
	Channel     string         `mapstructure:"channel"`
	QueueGroup  string         `mapstructure:"queue_group"`
	TaskTimeout *time.Duration `mapstructure:"task_timeout"` // optional, uses consumer default if nil
	RetryCount  *int           `mapstructure:"retry_count"`  // optional, uses consumer default if nil
}

// natsConfig is an alias for backward compatibility
type natsConfig = NatsConfig
