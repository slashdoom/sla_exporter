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
	alias    string
	success  bool
	duration time.Duration
}

const version string = "0.0.3"

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
			Help: "Overall result of tests.",
		},
		[]string{"test", "description"},
	)
	testRunDurations = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "sla_run_duration",
            Help: "Run Duration of tests.",
        },
        []string{"test", "description"},
    )
	testCumulativeDurations = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "sla_cumulative_duration",
            Help: "Cumulative Duration of tests.",
        },
        []string{"test", "description"},
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
    registry.MustRegister(testRunDurations)
    registry.MustRegister(testCumulativeDurations)
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

	// Channels for passing results
	resultChan := make(chan testResult,
		(len(config.AppConfig.CurlTests)+
			len(config.AppConfig.DnsTests)+
			len(config.AppConfig.PingTests)+
			len(config.AppConfig.TcpingTests)),
	)


    // Variables to track overall success and duration for each test category
    var curlSuccess, dnsSuccess, pingSuccess, tcpingSuccess bool
    var curlRunDuration, dnsRunDuration, pingRunDuration, tcpingRunDuration time.Duration
	var curlCumulativeDuration, dnsCumulativeDuration, pingCumulativeDuration, tcpingCumulativeDuration time.Duration

	// Create a WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	wg.Add(4) // One for each test type

    go func() {
        defer wg.Done()
        curlSuccess, curlRunDuration, curlCumulativeDuration = runCurlTests(resultChan)
    }()
    go func() {
        defer wg.Done()
        dnsSuccess, dnsRunDuration, dnsCumulativeDuration = runDnsTests(resultChan)
    }()
    go func() {
        defer wg.Done()
        pingSuccess, pingRunDuration, pingCumulativeDuration = runPingTests(resultChan)
    }()
    go func() {
        defer wg.Done()
        tcpingSuccess, tcpingRunDuration, tcpingCumulativeDuration = runTcpingTests(resultChan)
    }()

	// Wait for all test types to complete
	wg.Wait()
	close(resultChan)

	// Record the overall results for each test type
	recordTestResult("curl", "all_curl_tests", curlSuccess, curlRunDuration, curlCumulativeDuration)
	recordTestResult("dns", "all_dns_tests", dnsSuccess, dnsRunDuration, dnsCumulativeDuration)
	recordTestResult("ping", "all_ping_tests", pingSuccess, pingRunDuration, pingCumulativeDuration)
	recordTestResult("tcping", "all_tcping_tests", tcpingSuccess, tcpingRunDuration, tcpingCumulativeDuration)

	// Record the total result
	recordTestResult(
		"all",
		"all_test",
		true,
		time.Since(startTimeTotal),
		(curlCumulativeDuration+
			dnsCumulativeDuration+
			pingCumulativeDuration+
			tcpingCumulativeDuration),
	)

	// Expose metrics via Prometheus
	promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		ErrorHandling: promhttp.ContinueOnError,
	}).ServeHTTP(w, r)
}


func recordTestResult(test string, description string, result bool, runDuration time.Duration, cumulativeDuration time.Duration) {
	succeeded := 0.00
	if result {
		succeeded = 1.00
	}
	testResults.WithLabelValues(test, description).Set(succeeded)
	testRunDurations.WithLabelValues(test, description).Set(runDuration.Seconds())
	testCumulativeDurations.WithLabelValues(test, description).Set(cumulativeDuration.Seconds())
}


func runCurlTests(resultChan chan<- testResult) (bool, time.Duration, time.Duration) {
    var wg sync.WaitGroup
    startTime := time.Now()

    allSuccess := true
    var cumulativeDuration time.Duration
    var mu sync.Mutex // To safely update cumulativeDuration across goroutines

    for _, curlTest := range config.AppConfig.CurlTests {
        wg.Add(1)
        go func(test config.CurlTestConfig) {
            defer wg.Done()
            result, err := curl.Test(test)
            if err != nil {
                logger.Log.Error("Curl test failed",
                    zap.String("url", test.URL),
                    zap.Error(err),
                )
                allSuccess = false
            } else if !result.Completed {
                allSuccess = false
            }

			// Record the test result
            curl.Record(
                fmt.Sprintf("%s", test.URL),
                fmt.Sprintf("%s", test.Alias),
                "GET",
                result,
            )

			// Safely update the cumulative duration
			mu.Lock()
			cumulativeDuration += result.Duration
			mu.Unlock()

            // Record the result
            resultChan <- testResult{
                test:     "curl",
                target:   fmt.Sprintf("%s", test.URL),
                alias:    fmt.Sprintf("%s", test.Alias),
                success:  result.Completed,
                duration: result.Duration,
            }
        }(curlTest)
    }

    wg.Wait()
    return allSuccess, time.Since(startTime), cumulativeDuration
}


func runDnsTests(resultChan chan<- testResult) (bool, time.Duration, time.Duration) {
    var wg sync.WaitGroup
    startTime := time.Now()

    allSuccess := true
    var cumulativeDuration time.Duration
    var mu sync.Mutex // To safely update cumulativeDuration across goroutines

    for _, dnsTest := range config.AppConfig.DnsTests {
        wg.Add(1)
        go func(test config.DnsTestConfig) {
            defer wg.Done()
            result, err := dns.Test(test)
			serverName := ""
			if dnsTest.Server != "" {
				serverName = fmt.Sprintf("%s", dnsTest.Server)
			}
            if err != nil {
                logger.Log.Error("DNS test failed",
                    zap.String("host", test.Host),
                    zap.Error(err),
                )
                allSuccess = false
            } else if !result.Completed {
                allSuccess = false
            }

			// Record the test result
            dns.Record(
                fmt.Sprintf("%s", test.Host),
                fmt.Sprintf("%s", test.Alias),
                serverName,
				result,
			)

			// Safely update the cumulative duration
			mu.Lock()
			cumulativeDuration += result.Duration
			mu.Unlock()

            // Record the result
            resultChan <- testResult{
                test:     "dns",
                target:   fmt.Sprintf("%s", test.Host),
                alias:    fmt.Sprintf("%s", test.Alias),
                success:  result.Completed,
                duration: result.Duration,
            }
        }(dnsTest)
    }

    wg.Wait()
    return allSuccess, time.Since(startTime), cumulativeDuration
}


func runPingTests(resultChan chan<- testResult) (bool, time.Duration, time.Duration) {
    var wg sync.WaitGroup
    startTime := time.Now()

    allSuccess := true
    var cumulativeDuration time.Duration
    var mu sync.Mutex // To safely update cumulativeDuration across goroutines

    for _, pingTest := range config.AppConfig.PingTests {
        wg.Add(1)
        go func(test config.PingTestConfig) {
            defer wg.Done()
            result, err := ping.Test(test)
            if err != nil {
                logger.Log.Error("Ping test failed",
                    zap.String("host", test.Host),
                    zap.Error(err),
                )
                allSuccess = false
            } else if !result.Completed {
                allSuccess = false
            }

			// Record the test result
            ping.Record(
                fmt.Sprintf("%s", test.Host),
                fmt.Sprintf("%s", test.Alias),
				result,
			)

			// Safely update the cumulative duration
			mu.Lock()
			cumulativeDuration += result.Duration
			mu.Unlock()

            // Record the result
            resultChan <- testResult{
                test:     "ping",
                target:   fmt.Sprintf("%s", test.Host),
                alias:    fmt.Sprintf("%s", test.Alias),
                success:  result.Completed,
                duration: result.Duration,
            }
        }(pingTest)
    }

    wg.Wait()
    return allSuccess, time.Since(startTime), cumulativeDuration
}


func runTcpingTests(resultChan chan<- testResult) (bool, time.Duration, time.Duration) {
    var wg sync.WaitGroup
    startTime := time.Now()

    allSuccess := true
    var cumulativeDuration time.Duration
    var mu sync.Mutex // To safely update cumulativeDuration across goroutines

    for _, tcpingTest := range config.AppConfig.TcpingTests {
        wg.Add(1)
        go func(test config.TcpingTestConfig) {
            defer wg.Done()
            result, err := tcping.Test(test)
            if err != nil {
                logger.Log.Error("TCPing test failed",
                    zap.String("host", test.Host),
					zap.String("port", fmt.Sprintf("%d", tcpingTest.Port)),
                    zap.Error(err),
                )
                allSuccess = false
            } else if !result.Completed {
                allSuccess = false
            }

			// Record the test result
            tcping.Record(
                fmt.Sprintf("%s", test.Host),
                fmt.Sprintf("%s", test.Alias),
				fmt.Sprintf("%d", tcpingTest.Port),
				result,
			)

			// Safely update the cumulative duration
			mu.Lock()
			cumulativeDuration += result.Duration
			mu.Unlock()

            // Record the result
            resultChan <- testResult{
                test:     "tcping",
                target:   fmt.Sprintf("%s:%d", tcpingTest.Host, tcpingTest.Port),
                alias:    fmt.Sprintf("%s", test.Alias),
                success:  result.Completed,
                duration: result.Duration,
            }
        }(tcpingTest)
    }

    wg.Wait()
    return allSuccess, time.Since(startTime), cumulativeDuration
}
