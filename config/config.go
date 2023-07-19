package config

import (
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/prebid/prebid-cache/utils"
)

func NewConfig(filename string) Configuration {
	v := viper.New()

	setConfigDefaults(v)

	setEnvVarsLookup(v)

	setConfigFilePath(v, filename)

	// Read configuration file
	err := v.ReadInConfig()
	if err != nil {
		// Make sure the configuration file was not defective
		if _, fileNotFound := err.(viper.ConfigFileNotFoundError); fileNotFound {
			// Config file not found. Just log at info level and start Prebid Cache with default values
			log.Info("Configuration file not detected. Initializing with default values and environment variable overrides.")
		} else {
			// Config file was found but was defective, Either `UnsupportedConfigError` or `ConfigParseError` was thrown
			log.Fatalf("Configuration file could not be read: %v", err)
		}
	}

	cfg := Configuration{}
	if err := v.Unmarshal(&cfg); err != nil {
		log.Fatalf("Failed to unmarshal config: %v", err)
	}

	return cfg
}

func setConfigDefaults(v *viper.Viper) {
	v.SetDefault("port", 2424)
	v.SetDefault("admin_port", 2525)
	v.SetDefault("index_response", "This application stores short-term data for use in Prebid.")
	v.SetDefault("status_response", "")
	v.SetDefault("log.level", "info")
	v.SetDefault("backend.type", "memory")
	v.SetDefault("backend.aerospike.host", "")
	v.SetDefault("backend.aerospike.hosts", []string{})
	v.SetDefault("backend.aerospike.port", 0)
	v.SetDefault("backend.aerospike.namespace", "")
	v.SetDefault("backend.aerospike.user", "")
	v.SetDefault("backend.aerospike.password", "")
	v.SetDefault("backend.aerospike.default_ttl_seconds", 0)
	v.SetDefault("backend.aerospike.max_read_retries", 2)
	v.SetDefault("backend.aerospike.max_write_retries", 0)
	v.SetDefault("backend.aerospike.connection_idle_timeout_seconds", 0)
	v.SetDefault("backend.aerospike.connection_queue_size", 0)
	v.SetDefault("backend.cassandra.hosts", "")
	v.SetDefault("backend.cassandra.keyspace", "")
	v.SetDefault("backend.cassandra.default_ttl_seconds", utils.CASSANDRA_DEFAULT_TTL_SECONDS)
	v.SetDefault("backend.memcache.hosts", []string{})
	v.SetDefault("backend.redis.host", "")
	v.SetDefault("backend.redis.port", 0)
	v.SetDefault("backend.redis.password", "")
	v.SetDefault("backend.redis.db", 0)
	v.SetDefault("backend.redis.expiration", utils.REDIS_DEFAULT_EXPIRATION_MINUTES)
	v.SetDefault("backend.redis.tls.enabled", false)
	v.SetDefault("backend.redis.tls.insecure_skip_verify", false)
	v.SetDefault("compression.type", "snappy")
	v.SetDefault("metrics.influx.enabled", false)
	v.SetDefault("metrics.influx.host", "")
	v.SetDefault("metrics.influx.database", "")
	v.SetDefault("metrics.influx.measurement", "")
	v.SetDefault("metrics.influx.username", "")
	v.SetDefault("metrics.influx.password", "")
	v.SetDefault("metrics.influx.align_timestamps", false)
	v.SetDefault("metrics.prometheus.port", 0)
	v.SetDefault("metrics.prometheus.namespace", "")
	v.SetDefault("metrics.prometheus.subsystem", "")
	v.SetDefault("metrics.prometheus.timeout_ms", 0)
	v.SetDefault("metrics.prometheus.enabled", false)
	v.SetDefault("rate_limiter.enabled", true)
	v.SetDefault("rate_limiter.num_requests", utils.RATE_LIMITER_NUM_REQUESTS)
	v.SetDefault("request_limits.allow_setting_keys", false)
	v.SetDefault("request_limits.max_size_bytes", utils.REQUEST_MAX_SIZE_BYTES)
	v.SetDefault("request_limits.max_num_values", utils.REQUEST_MAX_NUM_VALUES)
	v.SetDefault("request_limits.max_ttl_seconds", utils.REQUEST_MAX_TTL_SECONDS)
	v.SetDefault("routes.allow_public_write", true)
}

func setConfigFilePath(v *viper.Viper, filename string) {
	v.SetConfigName(filename)              // name of config file (without extension)
	v.AddConfigPath("/etc/prebid-cache/")  // path to look for the config file in
	v.AddConfigPath("$HOME/.prebid-cache") // call multiple times to add many search paths
	v.AddConfigPath(".")
}

func setEnvVarsLookup(v *viper.Viper) {
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.SetEnvPrefix("PBC")
	v.AutomaticEnv()
}

type Configuration struct {
	Port           int           `mapstructure:"port"`
	AdminPort      int           `mapstructure:"admin_port"`
	IndexResponse  string        `mapstructure:"index_response"`
	Log            Log           `mapstructure:"log"`
	RateLimiting   RateLimiting  `mapstructure:"rate_limiter"`
	RequestLimits  RequestLimits `mapstructure:"request_limits"`
	StatusResponse string        `mapstructure:"status_response"`
	Backend        Backend       `mapstructure:"backend"`
	Compression    Compression   `mapstructure:"compression"`
	Metrics        Metrics       `mapstructure:"metrics"`
	Routes         Routes        `mapstructure:"routes"`
}

// ValidateAndLog validates the config, terminating the program on any errors.
// It also logs the config values that it used.
func (cfg *Configuration) ValidateAndLog() {

	log.Infof("config.port: %d", cfg.Port)
	log.Infof("config.admin_port: %d", cfg.AdminPort)
	cfg.Log.validateAndLog()
	cfg.RateLimiting.validateAndLog()
	cfg.RequestLimits.validateAndLog()

	if err := cfg.Backend.validateAndLog(); err != nil {
		log.Fatalf("%s", err.Error())
	}

	cfg.Compression.validateAndLog()
	cfg.Metrics.validateAndLog()
	cfg.Routes.validateAndLog()
}

type Log struct {
	Level LogLevel `mapstructure:"level"`
}

func (cfg *Log) validateAndLog() {
	log.Infof("config.log.level: %s", cfg.Level)
}

type LogLevel string

const (
	Trace   LogLevel = "trace"
	Debug   LogLevel = "debug"
	Info    LogLevel = "info"
	Warning LogLevel = "warning"
	Error   LogLevel = "error"
	Fatal   LogLevel = "fatal"
	Panic   LogLevel = "panic"
)

type RateLimiting struct {
	Enabled              bool  `mapstructure:"enabled"`
	MaxRequestsPerSecond int64 `mapstructure:"num_requests"`
}

func (cfg *RateLimiting) validateAndLog() {
	log.Infof("config.rate_limiter.enabled: %t", cfg.Enabled)
	log.Infof("config.rate_limiter.num_requests: %d", cfg.MaxRequestsPerSecond)
}

type RequestLimits struct {
	MaxSize          int  `mapstructure:"max_size_bytes"`
	MaxNumValues     int  `mapstructure:"max_num_values"`
	MaxTTLSeconds    int  `mapstructure:"max_ttl_seconds"`
	AllowSettingKeys bool `mapstructure:"allow_setting_keys"`
}

func (cfg *RequestLimits) validateAndLog() {
	log.Infof("config.request_limits.allow_setting_keys: %v", cfg.AllowSettingKeys)

	if cfg.MaxTTLSeconds >= 0 {
		log.Infof("config.request_limits.max_ttl_seconds: %d", cfg.MaxTTLSeconds)
	} else {
		log.Fatalf("invalid config.request_limits.max_ttl_seconds: %d. Value cannot be negative.", cfg.MaxTTLSeconds)
	}

	if cfg.MaxSize >= 0 {
		log.Infof("config.request_limits.max_size_bytes: %d", cfg.MaxSize)
	} else {
		log.Fatalf("invalid config.request_limits.max_size_bytes: %d. Value cannot be negative.", cfg.MaxSize)
	}

	if cfg.MaxNumValues >= 0 {
		log.Infof("config.request_limits.max_num_values: %d", cfg.MaxNumValues)
	} else {
		log.Fatalf("invalid config.request_limits.max_num_values: %d. Value cannot be negative.", cfg.MaxNumValues)
	}
}

type Compression struct {
	Type CompressionType `mapstructure:"type"`
}

func (cfg *Compression) validateAndLog() {
	switch cfg.Type {
	case CompressionNone:
		fallthrough
	case CompressionSnappy:
		log.Infof("config.compression.type: %s", cfg.Type)
	default:
		log.Fatalf(`invalid config.compression.type: %s. It must be "none" or "snappy"`, cfg.Type)
	}
}

type CompressionType string

const (
	CompressionNone   CompressionType = "none"
	CompressionSnappy CompressionType = "snappy"
)

type Metrics struct {
	Type       MetricsType       `mapstructure:"type"`
	Influx     InfluxMetrics     `mapstructure:"influx"`
	Prometheus PrometheusMetrics `mapstructure:"prometheus"`
}

func (cfg *Metrics) validateAndLog() {

	if cfg.Type == MetricsInflux || cfg.Influx.Enabled {
		cfg.Influx.validateAndLog()
		cfg.Influx.Enabled = true
	}

	if cfg.Prometheus.Enabled {
		cfg.Prometheus.validateAndLog()
		cfg.Prometheus.Enabled = true
	}

	metricsEnabled := cfg.Influx.Enabled || cfg.Prometheus.Enabled
	if cfg.Type == MetricsNone || cfg.Type == "" {
		if !metricsEnabled {
			log.Infof("Prebid Cache will run without metrics")
		}
	} else if cfg.Type != MetricsInflux {
		// Was any other metrics system besides "InfluxDB" or "Prometheus" specified in `cfg.Type`?
		if metricsEnabled {
			// Prometheus, Influx or both, are enabled. Log a message explaining that `prebid-cache` will
			// continue with supported metrics and non-supported metrics will be disabled
			log.Infof("Prebid Cache will run without unsupported metrics \"%s\".", cfg.Type)
		} else {
			// The only metrics engine specified in the configuration file is a non-supported
			// metrics engine. We should log error and exit program
			log.Fatalf("Metrics \"%s\" are not supported, exiting program.", cfg.Type)
		}
	}
}

type MetricsType string

const (
	MetricsNone   MetricsType = "none"
	MetricsInflux MetricsType = "influx"
)

type InfluxMetrics struct {
	Enabled         bool   `mapstructure:"enabled"`
	Host            string `mapstructure:"host"`
	Database        string `mapstructure:"database"`
	Measurement     string `mapstructure:"measurement"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	AlignTimestamps bool   `mapstructure:"align_timestamps"`
}

func (influxMetricsConfig *InfluxMetrics) validateAndLog() {
	// validate
	if influxMetricsConfig.Host == "" {
		log.Fatalf(`Despite being enabled, influx metrics came with no host info: config.metrics.influx.host = "".`)
	}
	if influxMetricsConfig.Database == "" {
		log.Fatalf(`Despite being enabled, influx metrics came with no database info: config.metrics.influx.database = "".`)
	}
	if influxMetricsConfig.Measurement == "" {
		log.Fatalf(`Despite being enabled, influx metrics came with no measurement info: config.metrics.influx.measurement = "".`)
	}

	// log
	log.Infof("config.metrics.influx.host: %s", influxMetricsConfig.Host)
	log.Infof("config.metrics.influx.database: %s", influxMetricsConfig.Database)
	log.Infof("config.metrics.influx.measurement: %s", influxMetricsConfig.Measurement)
	log.Infof("config.metrics.influx.align_timestamps: %v", influxMetricsConfig.AlignTimestamps)
}

type PrometheusMetrics struct {
	Port             int    `mapstructure:"port"`
	Namespace        string `mapstructure:"namespace"`
	Subsystem        string `mapstructure:"subsystem"`
	TimeoutMillisRaw int    `mapstructure:"timeout_ms"`
	Enabled          bool   `mapstructure:"enabled"`
}

// validateAndLog will error out when the value of port is 0
func (promMetricsConfig *PrometheusMetrics) validateAndLog() {
	if promMetricsConfig.Port == 0 {
		log.Fatalf(`Despite being enabled, prometheus metrics came with an empty port number: config.metrics.prometheus.port = 0`)
	}

	log.Infof("config.metrics.prometheus.namespace: %s", promMetricsConfig.Namespace)
	log.Infof("config.metrics.prometheus.subsystem: %s", promMetricsConfig.Subsystem)
	log.Infof("config.metrics.prometheus.port: %d", promMetricsConfig.Port)
}

func (m *PrometheusMetrics) Timeout() time.Duration {
	return time.Duration(m.TimeoutMillisRaw) * time.Millisecond
}

type Routes struct {
	AllowPublicWrite bool `mapstructure:"allow_public_write"`
}

func (cfg *Routes) validateAndLog() {
	if !cfg.AllowPublicWrite {
		log.Infof("Main server will only accept GET requests")
	}
}
