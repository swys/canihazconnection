package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// Logger interface
type Logger interface {
	Infof(string, ...interface{})
	Info(args ...interface{})
	Debugf(string, ...interface{})
	Debug(args ...interface{})
	Errorf(string, ...interface{})
	Error(args ...interface{})
	Panicf(string, ...interface{})
	Panic(args ...interface{})
}

// RealWebClient : to allow for easy mocking of http requests
type RealWebClient struct {
	logger Logger
}

// WebClient : interface for Web Clients (real and fake)
type WebClient interface {
	Get(string) (*int, error)
}

// HTTPClient : struct that implements WebClient interface
type HTTPClient struct {
	HTTP WebClient
}

// DialTimeout : function def to abstract the net.dialTimeout function
type DialTimeout func(string, string, time.Duration) (net.Conn, error)

var (
	telnetHosts      = Getenv("TELNET_HOSTS")
	httpRequestHosts = Getenv("HTTP_REQUESTS")
	loglevel         = strings.ToUpper(Getenv("LOG_LEVEL"))
)

func main() {
	levels := map[string]interface{}{
		"INFO":  logrus.InfoLevel,
		"WARN":  logrus.WarnLevel,
		"ERROR": logrus.ErrorLevel,
		"DEBUG": logrus.DebugLevel,
		"FATAL": logrus.FatalLevel,
		"PANIC": logrus.PanicLevel,
		"TRACE": logrus.TraceLevel,
	}
	_, err := ValidateLogLevel(levels, loglevel)
	if err != nil {
		log.Panicf("%s", err)
	}

	log := logrus.New()
	log.SetOutput(os.Stdout)
	log.SetLevel(levels[loglevel].(logrus.Level))
	log.SetReportCaller(true)
	hostsSuccess := make([]string, 0)
	hostsFailed := make([]string, 0)
	telnetRequests := StringToArray(telnetHosts)
	httpRequests := StringToArray(httpRequestHosts)
	allRequests := map[string][]string{
		"telnet": telnetRequests,
		"http":   httpRequests,
	}
	realHTTPClient := &HTTPClient{
		HTTP: &RealWebClient{
			logger: log,
		},
	}

	for connType, hosts := range allRequests {
		for i := range hosts {
			if hosts[i] == "" {
				log.Warnf("Host is empty string, won't try to connect : %v", hosts[i])
				continue
			}
			_, err := CanConnect(hosts[i], connType, net.DialTimeout, realHTTPClient, log)
			if err != nil {
				log.Errorf("Failed to connect to host : %s with Error : %v", hosts[i], err)
				hostsFailed = append(hostsFailed, hosts[i])
				continue
			}
			log.Debugf("Successfully connected to hosts : %s", hosts[i])
			hostsSuccess = append(hostsSuccess, hosts[i])

		}
	}
	log.Infof("Successfully connected to %d out of %d hosts", len(hostsSuccess), len(telnetRequests)+len(httpRequests))
	PrintHosts("failed to connect to host", hostsFailed, log)
	// exit with proper code
	if len(hostsFailed) > 0 {
		log.Debugf("We could not connect to [%d] host(s) so exiting process with exit code 1", len(hostsFailed))
		os.Exit(1)
	} else {
		log.Debugf("We were able to connect to all hosts, exiting 0")
		os.Exit(0)
	}
}

// Getenv : function to retrieve environment variable and panic if we can't find it
func Getenv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		panic("missing required environment variable " + name)
	}
	return v
}

// PrintHosts : function to print hosts from array with message prefix (used for logging)
func PrintHosts(msg string, h []string, l Logger) {
	for i := range h {
		l.Infof("%s : %s", msg, h[i])
	}
}

// StringToArray : function that takes a delimited string and returns an array of strings
func StringToArray(s string) []string {
	a := make([]string, 0)
	vals := strings.Split(s, ",")
	if len(vals) > 0 {
		for i := range vals {
			a = append(a, strings.TrimSpace(vals[i]))
		}
	}
	return a
}

// Keys : fucntion that takes a map of string keys with interface values and returns an array of keys of type []string
func Keys(m map[string]interface{}) (keys []string) {
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// CanConnect : function that takes a host, dialer function, success and fail pointer to string arrays
// and a Logger and returns a bool
func CanConnect(h string, connectionType string, dialer DialTimeout, httpClient *HTTPClient, l Logger) (bool, error) {
	l.Debugf("Attempting to connect to host : %s via %s", h, connectionType)
	switch connectionType {
	case "telnet":
		conn, err := dialer("tcp", h, time.Second*5)
		if err != nil {
			return false, err
		}
		defer conn.Close()
		return true, nil
	case "http":
		if !IsValidURL(h) {
			return false, fmt.Errorf("URL : %s is not valid...must be in the format <protocol>://<hostname> like https://example.com", h)
		}
		statusCode, err := httpClient.HTTP.Get(h)
		if err != nil {
			return false, fmt.Errorf("failed to get response from host : %s got error : %s", h, err)
		}
		l.Debugf("got %d response from %s", *statusCode, h)
		// we are considering all 2xx and 3xx as successful calls, if we get non 2xx or 3xx we will return error
		// most of these cases fallthrough because otherwise I'd have to make the widest case statement in history
		switch *statusCode {
		case http.StatusOK, http.StatusCreated, http.StatusAccepted, http.StatusNonAuthoritativeInfo:
			fallthrough
		case http.StatusNoContent, http.StatusResetContent, http.StatusPartialContent, http.StatusMultiStatus:
			fallthrough
		case http.StatusAlreadyReported, http.StatusIMUsed, http.StatusMultipleChoices, http.StatusMovedPermanently:
			fallthrough
		case http.StatusFound, http.StatusSeeOther, http.StatusNotModified, http.StatusUseProxy:
			fallthrough
		case http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
			fallthrough
		case http.StatusUnauthorized, http.StatusForbidden:
			return true, nil
		default:
			return false, fmt.Errorf("got back %d http status code from host : %s", *statusCode, h)
		}
	default:
		return false, fmt.Errorf("encountered unknown connection type : %s...not implimented", connectionType)
	}
}

// ValidateLogLevel : function to validate if LOG_LEVEL variable value is a valid log level
func ValidateLogLevel(m map[string]interface{}, level string) (bool, error) {
	if _, ok := m[level]; !ok {
		keys := Keys(m)
		sort.Strings(keys)
		return false, fmt.Errorf("loglevel : %v not valid, must be one of : %s", level, strings.Join(keys[:], " "))
	}
	return true, nil
}

// Get : function to make a real http request
func (r *RealWebClient) Get(url string) (*int, error) {
	client := &http.Client{}
	r.logger.Debugf("Attempting to make GET request to %s", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting Requst object from NewRequst : %s for url : %s", err, url)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error connecting to %s : %s", url, err)
	}
	// don't forget to close body (also we are not using it, we only care about the response code)
	defer resp.Body.Close()
	return &resp.StatusCode, nil
}

// IsValidURL : function to validate if url is valid
func IsValidURL(testURL string) bool {
	if _, err := url.ParseRequestURI(testURL); err != nil {
		return false
	}
	return true
}
