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
	OpenObserveCompatible bool
	LogDataStartString    string
	LogDataEndString      string
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
		metricLogger.LogDataStartString = "[\n"
		metricLogger.LogDataEndString = "\n]"
	} else {
		metricLogger.ElasticEndpoint = metricLogger.ElasticEndpoint + "/_bulk"
		metricLogger.LogDataStartString = `{ "index":{} }` + "\n"
		metricLogger.LogDataEndString = "\n"
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
	logData := strings.Builder{}
	logAvailable := false
outer:
	for { // Flush everything in the queue
		select {
		case LogEntry, ok := <-metricLogger.LogQueue:
			if !ok {
				break
			}
			LogBytes, err := json.Marshal(LogEntry)
			if err == nil {
				logData.WriteString(metricLogger.LogDataStartString)
				logData.Write(LogBytes)
				logData.WriteString(metricLogger.LogDataEndString)
			}
			logAvailable = true
		default:
			break outer
		}
	}
	if logAvailable {
		lerr := metricLogger.Insert(logData.String())
		if lerr != nil {
			log.Println(lerr)
		}
	}
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
