package curl

import (
	"crypto/tls"
	"errors"
	"net/http"
	"net/http/httptrace"
	"strings"
	"sync"
    "time"

	"example.org/config"
	"example.org/logger"
	
	"go.uber.org/zap"
	"github.com/prometheus-community/pro-bing"
)

type CurlResult struct {
	Completed         bool
	Duration          time.Duration
	dnsDuration       time.Duration
	connDuration      time.Duration
	tlsDuration       time.Duration
	ReqToRespDuration time.Duration
	statusCode        int
}

var (
	startTime  time.Time
	dnsStart   time.Duration
	dnsDone    time.Duration
	connStart  time.Duration
	connDone   time.Duration
	tlsStart   time.Duration
	tlsDone    time.Duration
	ReqAt      time.Duration
	RespAt     time.Duration
	statusCode int
)

var wg sync.WaitGroup
var stop bool

func Test(test config.CurlTestConfig) (CurlResult, error) {
    t, err := time.ParseDuration(test.Timeout)
    if err != nil {
		logger.Log.Info("cURL timeout not accepted, defaulting to global",
			zap.Error(err))
		t, err = time.ParseDuration(config.AppConfig.Timeout)
		if err != nil {
			logger.Log.Info("cURL timeout not accepted, defaulting to 5s",
				zap.Error(err))
        	t, err = time.ParseDuration("5s")
		}
    }
	// FIXME - Waiting on upstream library
	//method, err := parseHttpMethod(test.Method)
	//if err != nil {
	//	logger.Log.Warn("cURL method not accepted, defaulting to GET",
	//	 	zap.Error(err))
	//	method, err = parseHttpMethod("GET")
	//}
	//headers := test.Headers
	
	c := 1
	i := 0
	stop = false

	results := CurlResult{
		Completed: false,
	}
    startTime = time.Now()
	
	

    httpCaller := probing.NewHttpCaller(test.URL,
		probing.WithHTTPCallerTimeout(t),
        probing.WithHTTPCallerCallFrequency(time.Second),
        probing.WithHTTPCallerOnDNSStart(func(suite *probing.TraceSuite, info httptrace.DNSStartInfo) {
			dnsStart = time.Since(startTime)
        }),
        probing.WithHTTPCallerOnDNSDone(func(suite *probing.TraceSuite, info httptrace.DNSDoneInfo) {
			dnsDone = time.Since(startTime)
        }),
		probing.WithHTTPCallerOnConnStart(func(suite *probing.TraceSuite, network string, addr string) {
			connStart = time.Since(startTime)
        }),
        probing.WithHTTPCallerOnConnDone(func(suite *probing.TraceSuite, network string, addr string, err error) {
			connDone = time.Since(startTime)
        }),
		probing.WithHTTPCallerOnTLSStart(func(suite *probing.TraceSuite)  {
			tlsStart = time.Since(startTime)
        }),
        probing.WithHTTPCallerOnTLSDone(func(suite *probing.TraceSuite, state tls.ConnectionState, err error) {
			tlsDone = time.Since(startTime)
        }),
		probing.WithHTTPCallerOnReq(func(suite *probing.TraceSuite) {
			ReqAt = time.Since(startTime)
        }),
        probing.WithHTTPCallerOnResp(func(suite *probing.TraceSuite, info *probing.HTTPCallInfo) {
			i++
			RespAt = time.Since(startTime)
			statusCode = info.StatusCode
            elapsed := suite.GetGeneralEnd().Sub(suite.GetGeneralStart())
            logger.Log.Info("cURL response received.",
                zap.String("url", test.URL),
                zap.Int("status_code", info.StatusCode),
                zap.Duration("latency", elapsed),
            )
        }),
    )

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i < c && !stop {
			time.Sleep(time.Millisecond * 10)
		}
		if !stop {
			httpCaller.Stop()
		}
	}()
    httpCaller.Run()
	wg.Wait()
	stop = true

    duration := time.Since(startTime)
	results = CurlResult{
		Completed: true,
		Duration: duration,
	    dnsDuration:  dnsDone - dnsStart,
	    connDuration: connDone - connStart,
	    tlsDuration: tlsDone - tlsStart,
	    ReqToRespDuration: RespAt - ReqAt,
	    statusCode: statusCode,
	}

	return results, nil
}

func parseHttpMethod(methodStr string) (string, error) {
    methodStr = strings.ToUpper(methodStr)

    switch methodStr {
    case "GET":
        return http.MethodGet, nil
    case "POST":
        return http.MethodPost, nil
    case "PUT":
        return http.MethodPut, nil
    case "DELETE":
        return http.MethodDelete, nil
    case "PATCH":
        return http.MethodPatch, nil
    case "HEAD":
        return http.MethodHead, nil
    case "OPTIONS":
        return http.MethodOptions, nil
    case "CONNECT":
        return http.MethodConnect, nil
    case "TRACE":
        return http.MethodTrace, nil
    default:
        return http.MethodGet, errors.New("invalid HTTP method") //default to GET
    }
}

