package main

import (
	"flag"
	"fmt"
	"os"
	"net/http"
    "sync"
	"time"

	"example.org/config"
	"example.org/curl"
	"example.org/dns"
	"example.org/logger"
	"example.org/ping"
	"example.org/tcping"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type testResult struct {
	test     string
	target   string
	success  bool
	duration time.Duration
}

const version string = "0.0.1"

var (
	showVersion        = flag.Bool("version", false, "Print version information.")
	listenAddress      = flag.String("web.listen-address", "0.0.0.0:9909", "Address on which to expose metrics and web interface.")
	metricsPath        = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	level              = flag.String("level", "info", "Set logging verbose level")
	timeout            = flag.Int("Timeout", 5, "Set default test timeout value")
	configFile         = flag.String("config.file", "config.yaml", "Path to config file (required)")
)

var (
	registry = prometheus.NewRegistry()
	testResults = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sla_result",
			Help: "Total number of test results by type, and target.",
		},
		[]string{"test", "target"},
	)
	testDurations = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "sla_duration",
            Help: "Duration of tests by test and target.",
        },
        []string{"test", "target"},
    )
)

func init() {
	flag.Usage = func() {
		fmt.Println("Usage: sla_exporter [ ... ]\n\nParameters:")
		fmt.Println()
		flag.PrintDefaults()
	}
}


func main() {
	flag.Parse()

	if *showVersion {
		printVersion()
		os.Exit(0)
	}

	err := initialize()
	if err != nil {
		logger.Log.Fatal("could not initialize exporter.",
			zap.Error(err),
		)
	}

	// Register metrics with the global Prometheus registry
    registry.MustRegister(testResults)
    registry.MustRegister(testDurations)
	dns.Register(registry)
	curl.Register(registry)
	ping.Register(registry)
	tcping.Register(registry)

	startServer()
}


func initialize() error {
	var err error
	c := config.New()
	c.ConfigFile = *configFile

	c, err = loadConfig(c)
	if err != nil {
		return err
	}

	config.AppConfig = c
	logger.Config()

	return nil
}


func printVersion() {
	fmt.Println("sla_exporter")
	fmt.Printf("Version: %s\n", version)
	fmt.Println("Author(s): slashdoom (Patrick Ryon)")
	fmt.Println("Metric exporter for simple network SLA tests")
}


func startServer() {
	logger.Log.Info("starting sla_exporter",
		zap.String("version", version),
	)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
<html>
  <head>
    <title>SLA Exporter (Version ` + version + `)</title>
  </head>
  <body>
    <h1>SLA Exporter</h1>
    <p><a href="` + *metricsPath + `">Metrics</a></p>
    <h2>More information:</h2>
    <p><a href="https://github.com/slashdoom/sla_exporter">github.com/slashdoom/sla_exporter</a></p>
  </body>
</html>
`))
	})
	http.HandleFunc(*metricsPath, handleMetricsRequest)

	logger.Log.Info("listening for connections...",
		zap.String("metricsPath", *metricsPath),
		zap.String("listenAddress", *listenAddress),
	)
	http.ListenAndServe(*listenAddress, nil)
}


func handleMetricsRequest(w http.ResponseWriter, r *http.Request) {
	reg := registry

	startTimeTotal := time.Now()

	// Create a WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Channels for passing results back
	resultChan := make(chan testResult, 100)

	// Variables to track overall success and duration for each test category
	var dnsSuccess bool
	var dnsDuration time.Duration
	var pingSuccess bool
	var pingDuration time.Duration
	var curlSuccess bool
	var curlDuration time.Duration
	var tcpingSuccess bool
	var tcpingDuration time.Duration

	// Run cURL tests asynchronously
	wg.Add(1)
	go func() {
		defer wg.Done()
		startTimeCurl := time.Now()
		var curlAllSuccess bool
		for _, curlTest := range config.AppConfig.CurlTests {
			result, _ := curl.Test(curlTest)
			curl.Record(fmt.Sprintf("%s", curlTest.URL), "GET", result)

			// Collect the result
			resultChan <- testResult{
				test:     "curl",
				target:   fmt.Sprintf("%s", curlTest.URL),
				success:  result.Completed,
				duration: result.Duration,
			}

			if !result.Completed {
				curlAllSuccess = false
			}
		}
		curlDuration = time.Since(startTimeCurl)
		curlSuccess = curlAllSuccess
	}()

	// Run DNS tests asynchronously
	wg.Add(1)
	go func() {
		defer wg.Done()
		startTimeDns := time.Now()
		var dnsAllSuccess bool
		for _, dnsTest := range config.AppConfig.DnsTests {
			result, _ := dns.Test(dnsTest)
			serverName := ""
			if dnsTest.Server != "" {
				serverName = fmt.Sprintf("%s", dnsTest.Server)
			}
			dns.Record(fmt.Sprintf("%s", dnsTest.Host), serverName, result)

			// Collect the result
			resultChan <- testResult{
				test:     "dns",
				target:   fmt.Sprintf("%s", dnsTest.Host),
				success:  result.Completed,
				duration: result.Duration,
			}

			if !result.Completed {
				dnsAllSuccess = false
			}
		}
		dnsDuration = time.Since(startTimeDns)
		dnsSuccess = dnsAllSuccess
	}()

	// Run Ping tests asynchronously
	wg.Add(1)
	go func() {
		defer wg.Done()
		startTimePing := time.Now()
		var pingAllSuccess bool
		for _, pingTest := range config.AppConfig.PingTests {
			result, _ := ping.Test(pingTest)
			ping.Record(fmt.Sprintf("%s", pingTest.Host), result)

			// Collect the result
			resultChan <- testResult{
				test:     "ping",
				target:   fmt.Sprintf("%s", pingTest.Host),
				success:  result.Completed,
				duration: result.Duration,
			}

			if !result.Completed {
				pingAllSuccess = false
			}
		}
		pingDuration = time.Since(startTimePing)
		pingSuccess = pingAllSuccess
	}()

	// Run TCPing tests asynchronously
	wg.Add(1)
	go func() {
		defer wg.Done()
		startTimeTcping := time.Now()
		var tcpingAllSuccess bool
		for _, tcpingTest := range config.AppConfig.TcpingTests {
			result, _ := tcping.Test(tcpingTest)
			tcping.Record(fmt.Sprintf("%s", tcpingTest.Host), fmt.Sprintf("%d", tcpingTest.Port), result)

			// Collect the result
			resultChan <- testResult{
				test:     "tcping",
				target:   fmt.Sprintf("%s:%d", tcpingTest.Host, tcpingTest.Port),
				success:  result.Completed,
				duration: result.Duration,
			}

			if !result.Completed {
				tcpingAllSuccess = false
			}
		}
		tcpingDuration = time.Since(startTimeTcping)
		tcpingSuccess = tcpingAllSuccess
	}()

	// Wait for all goroutines to finish
	wg.Wait()

	// Close the result channel now that all results have been sent
	close(resultChan)

	// Record the overall results for each test type
	recordTestResult("curl", "all_curl_tests", curlSuccess, curlDuration)
	recordTestResult("dns", "all_dns_tests", dnsSuccess, dnsDuration)
	recordTestResult("ping", "all_ping_tests", pingSuccess, pingDuration)
	recordTestResult("tcping", "all_tcping_tests", tcpingSuccess, tcpingDuration)

	// Record the total result
	recordTestResult("all_tests", "total_run_time", true, time.Since(startTimeTotal))
    recordTestResult("all_tests", "total_test_duration", true, (curlDuration+dnsDuration+pingDuration+tcpingDuration))


	// Expose metrics via Prometheus
	promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		ErrorHandling: promhttp.ContinueOnError,
	}).ServeHTTP(w, r)
}


func recordTestResult(test string, target string, result bool, duration time.Duration) {
	succeeded := 0.00
	if result {
		succeeded = 1.00
	}
	testResults.WithLabelValues(test, target).Set(succeeded)
	testDurations.WithLabelValues(test, target).Set(duration.Seconds())
}
