package config

import (
	"net/http"
)

// Config represents the configuration for the exporter
type Config struct {
	ConfigFile    string            `yaml:"config_file,omitempty"`
	Level         string            `yaml:"level,omitempty"`
	Timeout	      string            `yaml:"timeout,omitempty"`
	Web           WebConfig         `yaml:"web,omitempty"`
	CurlTests   []CurlTestConfig    `yaml:"curl_tests"`
	DnsTests    []DnsTestConfig     `yaml:"dns_tests"`
	PingTests   []PingTestConfig    `yaml:"ping_tests"`
	TcpingTests []TcpingTestConfig  `yaml:"tcping_tests"`
}

type WebConfig struct {
	ListenAddress int   `yaml:"listen_address,omitempty"`
	MetricsPath  string `yaml:"metrics_path,omitempty"`
}

type DnsTestConfig struct {
	Host    string `yaml:"host"`
	Server  string `yaml:"server,omitempty"`
	Port    int    `yaml:"port,omitempty"`
	Timeout string `yaml:"timeout,omitempty"`
}


type PingTestConfig struct {
	Host    string `yaml:"host"`
	Count   int    `yaml:"count,omitempty"`
	Timeout string `yaml:"timeout,omitempty"`
}

type CurlTestConfig struct {
	URL     string      `yaml:"url"`
	Timeout string      `yaml:"timeout,omitempty"`
	Method  string      `yaml:"method,omitempty"`
	Headers http.Header `yaml:"headers,omitempty"`
}

type TcpingTestConfig struct {
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	Timeout string `yaml:"timeout,omitempty"`
}

var AppConfig Config


// New creates a new config
func New() Config {
	c := setDefaultValues()

	return c
}


func setDefaultValues() Config {
	var c Config
	c.Web.ListenAddress = 9909
	c.Web.MetricsPath = "/metrics"
	c.Level = "info"
	c.Timeout = "5s"
	c.ConfigFile = "config.yaml"

	return c
}
