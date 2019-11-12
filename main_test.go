package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

type TestLogger struct{}

func (l *TestLogger) Infof(m string, v ...interface{})  { log.Printf(m, v...) }
func (l *TestLogger) Info(v ...interface{})             { log.Println(v...) }
func (l *TestLogger) Debugf(m string, v ...interface{}) { log.Printf(m, v...) }
func (l *TestLogger) Debug(v ...interface{})            { log.Println(v...) }
func (l *TestLogger) Errorf(m string, v ...interface{}) { log.Printf(m, v...) }
func (l *TestLogger) Error(v ...interface{})            { log.Println(v...) }
func (l *TestLogger) Panicf(m string, v ...interface{}) { log.Panicf(m, v...) }
func (l *TestLogger) Panic(v ...interface{})            { log.Panic(v...) }

// MockWebClient : fake http client
type MockWebClient struct {
	logger Logger
}

var LogLevelTests = []struct {
	testID            int    // testcase id (used to easily find out which test case had issues)
	input             string // input
	expected          bool   // expected result
	expectedErrString string // expected error string
}{
	{1, "INFO", true, ""},
	{2, "warn", true, ""},
	{3, "Error", true, ""},
	{4, "DebuG", true, ""},
	{5, "FATAL", true, ""},
	{6, "PANIC", true, ""},
	{7, "TRACE", true, ""},
	{8, "BLAH", false, "loglevel : BLAH not valid, must be one of : DEBUG ERROR FATAL INFO PANIC TRACE WARN"},
	{9, "BLAHBLAHBLAH", false, "loglevel : BLAHBLAHBLAH not valid, must be one of : DEBUG ERROR FATAL INFO PANIC TRACE WARN"},
}

var CanConnectTests = []struct {
	testID            int    // testcase id
	input             string // host to try to connect to
	connType          string // connection type (valid values : telnet | http)
	expected          bool   // expected result (can connect or not)
	expectedErrString string // expected error string
}{
	{1, "localhost:33333", "telnet", true, ""},
	{2, "nonexistanthost.local:8080", "telnet", false, "error connecting to nonexistanthost.local:8080"},
	{3, "http://anothertest.com", "http", true, ""},
	{4, "http://youtube.com", "http", true, ""},
	{5, "http://giveme500.com", "http", false, "got back 502 http status code from host : http://giveme500.com"},
	{6, "http://throwerror.com", "http", false, "failed to get response from host : http://throwerror.com got error : unable to connect to : http://throwerror.com"},
	{7, "invalidurl.com", "http", false, "URL : invalidurl.com is not valid...must be in the format <protocol>://<hostname> like https://example.com"},
	{8, "https://validurl.com:443/path/to/my/resource", "http", true, ""},
}

func TestLogLevelInput(t *testing.T) {
	levels := map[string]interface{}{
		"INFO":  logrus.InfoLevel,
		"WARN":  logrus.WarnLevel,
		"ERROR": logrus.ErrorLevel,
		"DEBUG": logrus.DebugLevel,
		"FATAL": logrus.FatalLevel,
		"PANIC": logrus.PanicLevel,
		"TRACE": logrus.TraceLevel,
	}
	for _, tt := range LogLevelTests {
		ok, err := ValidateLogLevel(levels, strings.ToUpper(tt.input))
		if err != nil {
			if err.Error() != tt.expectedErrString {
				t.Errorf("got : %s but expected : %s...test case ID : %d", err.Error(), tt.expectedErrString, tt.testID)
			}
		}
		if ok != tt.expected {
			t.Errorf("got : %t but expected : %t...given a valid input, ValidateLogLevel function should return true...test case ID : %d", ok, tt.expected, tt.testID)
		}
	}
}

func TestCanConnect(t *testing.T) {
	logger := &TestLogger{}
	fakeHTTPClient := &HTTPClient{
		HTTP: &MockWebClient{
			logger: logger,
		},
	}
	for _, tt := range CanConnectTests {
		ok, err := CanConnect(tt.input, tt.connType, mockDialWithTimeout, fakeHTTPClient, logger)
		if err != nil {
			if err.Error() != tt.expectedErrString {
				t.Errorf("got : %s but expected : %s...test case ID : %d", err.Error(), tt.expectedErrString, tt.testID)
			}
		}
		if ok != tt.expected {
			t.Errorf("got : %t but expected : %t with input : %s for test case ID : %d", ok, tt.expected, tt.input, tt.testID)
		}
	}
}

// mock of net.DialTimeout function
func mockDialWithTimeout(protocol string, address string, timeout time.Duration) (net.Conn, error) {
	cli, svr := net.Pipe()
	defer svr.Close()
	defer cli.Close()
	if address == "localhost:33333" {
		return cli, nil
	}
	return cli, fmt.Errorf("error connecting to %s", address)
}

// Get : fake Get function for MockWebClient
func (mwc *MockWebClient) Get(url string) (*int, error) {
	switch url {
	case "http://giveme500.com":
		return createPointerToInt(http.StatusBadGateway), nil
	case "http://throwerror.com":
		return nil, fmt.Errorf("unable to connect to : %s", url)
	default:
		return createPointerToInt(http.StatusOK), nil
	}
}

// helper function that takes an int and returns a pointer to int
func createPointerToInt(x int) *int {
	return &x
}
