package tcping

import (
	"fmt"
	"net"
    "time"

	"example.org/config"
	"example.org/logger"
	
	"go.uber.org/zap"
)

type TcpingResult struct {
	Completed bool
	Duration time.Duration
}

func Test(test config.TcpingTestConfig) (TcpingResult, error) {
	t, err := time.ParseDuration(test.Timeout)
    if err != nil {
		logger.Log.Info("ping timeout not accepted, defaulting to global",
			zap.Error(err))
		t, err = time.ParseDuration(config.AppConfig.Timeout)
		if err != nil {
			logger.Log.Info("ping timeout not accepted, defaulting to 5s",
				zap.Error(err))
        	t, err = time.ParseDuration("5s")
		}
    }

	results := TcpingResult{
		Completed: false,
	}
	startTime := time.Now()
	address := fmt.Sprintf("%s:%d", test.Host, test.Port)

    conn, err := net.DialTimeout("tcp", address, t)
    if err != nil {
		logger.Log.Warn("failed to tcping",
			zap.String("host", test.Host),
			zap.Int("port", test.Port),
			zap.Error(err),
		)
        return results, err
    }
    defer conn.Close()

	duration := time.Since(startTime)

	logger.Log.Info("Tcping successful",
		zap.String("host", test.Host),
		zap.Int("port", test.Port),
		zap.Duration("duration", duration),
	)

	results = TcpingResult{
		Completed: true,
		Duration: duration,
	}

	return results, nil
}