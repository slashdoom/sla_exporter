package ping

import (
    "time"

	"example.org/config"
	"example.org/logger"
	
	"go.uber.org/zap"
	"github.com/prometheus-community/pro-bing"
)

type PingResult struct {
	Completed bool
	Duration time.Duration
	PacketsRecv int
	PacketsSent int
	PacketsRecvDup int
	PacketLoss float64
	MinRtt time.Duration
	MaxRtt time.Duration
	AvgRtt time.Duration
	StdDevRtt time.Duration
}

func Test(test config.PingTestConfig) (PingResult, error) {
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

	logger.Log.Debug("ping count requested",
			zap.Int("count", test.Count),
		)
	c := 1 // Default value

	if test.Count > 0 {
		c = test.Count
	}
	logger.Log.Debug("ping count determined",
			zap.Int("count", c),
		)

	results := PingResult{
		Completed: false,
	}
	startTime := time.Now()

	pinger, err := probing.NewPinger(test.Host)
	if err != nil {
		logger.Log.Warn("failed to create pinger",
			zap.Error(err),
		)
		results.Duration = time.Since(startTime)
		return results, err
	}
	pinger.Timeout = t
	pinger.Count = c
	err = pinger.Run() // Blocks until finished.
	if err != nil {
		logger.Log.Warn("failed to run pinger",
			zap.Error(err),
		)
		results.Duration = time.Since(startTime)
		return results, err
	}
	stats := pinger.Statistics() // get send/receive/duplicate/rtt stats
	duration:= time.Since(startTime)
	logger.Log.Info("ping successful.",
		zap.String("host", test.Host),
		zap.Int("count", c),
		zap.Duration("duration", duration),
	)

	results = PingResult{
		Completed: true,
		Duration: duration,
		PacketsRecv: stats.PacketsRecv,
		PacketsSent: stats.PacketsSent,
		PacketsRecvDup: stats.PacketsRecvDuplicates,
		PacketLoss: stats.PacketLoss,
		MinRtt: stats.MinRtt,
		MaxRtt: stats.MaxRtt,
		AvgRtt: stats.AvgRtt,
		StdDevRtt: stats.StdDevRtt,
	}

	return results, nil
}