package dns

import (
	"context"
	"fmt"
	"net"
	"time"

	"example.org/config"
	"example.org/logger"

	"go.uber.org/zap"
)

type DnsTestResult struct {
	Completed bool
	Duration  time.Duration
}

func Test(test config.DnsTestConfig) (DnsTestResult, error) {
	t, err := time.ParseDuration(test.Timeout)
	if err != nil {
		logger.Log.Info("DNS resolution timeout not accepted, defaulting to global",
			zap.Error(err))
		t, err = time.ParseDuration(config.AppConfig.Timeout)
		if err != nil {
			logger.Log.Info("DNS resolution timeout not accepted, defaulting to 5s",
				zap.Error(err))
			t, err = time.ParseDuration("5s")
		}
	}

	var dial func(ctx context.Context, network, address string) (net.Conn, error)

	if test.Server == "" {
		// If no server, use the default dialer with timeout
		defaultDialer := &net.Dialer{Timeout: t}
    	dial = func(ctx context.Context, network, address string) (net.Conn, error) {
        	return defaultDialer.DialContext(ctx, network, address)
		}
	} else {
		// If server is specified, create a custom dialer for that server
		port := fmt.Sprintf(":%d", 53) // Default DNS port
		if test.Port > 0 {
			port = fmt.Sprintf(":%d", test.Port) // Use specified port if available
		}
		dial = func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: t,
			}
			return d.DialContext(ctx, "udp", test.Server+port)
		}
	}
	
	resolver := net.Resolver{
		PreferGo: true,
		Dial:     dial,
	}

	results := DnsTestResult{
		Completed: false,
	}

	startTime := time.Now()

	ips, err := resolver.LookupHost(context.Background(), test.Host)
	if err != nil {
		logger.Log.Warn("Failed to resolve DNS",
			zap.String("host", test.Host),
			zap.Error(err),
		)
		return results, err
	}

	duration := time.Since(startTime)

	logger.Log.Info("DNS resolution successful",
		zap.String("host", test.Host),
		zap.String("server", test.Server),
		zap.Duration("duration", duration),
		zap.Any("ips", ips),
	)

	results = DnsTestResult{
		Completed: true,
		Duration:  duration,
	}

	return results, nil
}
