package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

type MetricLogger struct {
	LogQueue              chan Metric // Queue of Metrics that is to be flushed
	LoggingActive         *atomic.Bool
	FlushInterval         time.Duration
	ElasticServerUrl      string
	ElasticIndexName      string
	ElasticEndpoint       string
	LogMarshaller         func(logs []Metric) string
	OpenObserveCompatible bool
}

type Metric struct {
	Type            string
	Time            string
	Country         string
	Referrer        string
	UserUid         string
	SessionDuration string
}

var MLogger = MetricLogger{}

func (metricLogger *MetricLogger) StartLogger(queueSize int, flushInterval int,
	elasticServerHost string, elasticIndexName string,
	elasticUsername string, elasticPassword string, openObserveCompatible bool) {
	metricLogger.LogQueue = make(chan Metric, queueSize)
	metricLogger.LoggingActive = &atomic.Bool{}
	metricLogger.LoggingActive.Store(true)
	metricLogger.FlushInterval = time.Minute * time.Duration(flushInterval)
	metricLogger.ElasticIndexName = elasticIndexName
	metricLogger.ElasticServerUrl = fmt.Sprintf("https://%s:%s@%s", elasticUsername, elasticPassword, elasticServerHost)
	metricLogger.ElasticEndpoint = metricLogger.ElasticServerUrl + "/" +
		metricLogger.ElasticIndexName
	metricLogger.OpenObserveCompatible = openObserveCompatible

	if openObserveCompatible {
		metricLogger.ElasticEndpoint = metricLogger.ElasticEndpoint + "/_json"
		metricLogger.LogMarshaller = openObserveGetLogString
	} else {
		metricLogger.ElasticEndpoint = metricLogger.ElasticEndpoint + "/_bulk"
		metricLogger.LogMarshaller = elasticGetLogString
	}
	go metricLogger.periodicFlush()
}

func (metricLogger *MetricLogger) AddLogEntry(log *Metric) {
	if metricLogger.LoggingActive.Load() {
		log.Time = time.Now().Format(time.RFC3339)
		select {
		case metricLogger.LogQueue <- *log:
		default: // Channel full
		}
	}
}

func (metricLogger *MetricLogger) periodicFlush() {
	for {
		time.Sleep(metricLogger.FlushInterval)
		metricLogger.FlushQueue()
	}
}

func (metricLogger *MetricLogger) FlushQueue() {
	logAvailable := false
	logsToFlush := []Metric{}
outer:
	for { // Extract everything from the queue
		select {
		case LogEntry, ok := <-metricLogger.LogQueue:
			if !ok {
				break
			}
			logsToFlush = append(logsToFlush, LogEntry)
			logAvailable = true
		default:
			break outer
		}
	}

	if logAvailable {
		logString := metricLogger.LogMarshaller(logsToFlush)
		lerr := metricLogger.Insert(logString)
		if lerr != nil {
			log.Println(lerr)
		}
	}
}

func elasticGetLogString(logs []Metric) string {
	logData := strings.Builder{}
	for _, log := range logs {
		LogBytes, err := json.Marshal(log)
		if err == nil {
			logData.WriteString(`{ "index":{} }` + "\n")
			logData.Write(LogBytes)
			logData.WriteString("\n")
		}
	}

	return logData.String()
}

func openObserveGetLogString(logs []Metric) string {
	logData := strings.Builder{}
	logData.WriteString("[\n")

	logsLength := len(logs) - 1

	for i, log := range logs {
		LogBytes, err := json.Marshal(log)
		if err == nil {
			logData.Write(LogBytes)
			if i != logsLength {
				logData.WriteString(",")
			}
		}
	}

	logData.WriteString("\n]")

	return logData.String()
}

func (metricLogger *MetricLogger) Insert(Data string) error {
	client := &http.Client{}
	req, err := http.NewRequest("POST", metricLogger.ElasticEndpoint, strings.NewReader(Data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		return nil
	}

	return fmt.Errorf("code:%d Insert Failed", resp.StatusCode)
}
