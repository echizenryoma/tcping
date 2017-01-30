package main

import (
	"fmt"
	"math"
	"net"
	"reflect"
	"strconv"
	"time"
)

type Result struct {
	Addr     string
	Timout   int
	Total    int
	MinDelay float64
	MaxDelay float64
	AvgDelay float64
}

func dial(ipPool chan string, resultChan chan Result, cfg *Config) {
	for ip := range ipPool {
		result := Result{
			Addr:     fmt.Sprintf("%s:%d", ip, cfg.Port),
			Timout:   0,
			Total:    cfg.Repeat,
			MinDelay: math.MaxFloat64,
			MaxDelay: 0,
			AvgDelay: 0,
		}
		for i := 0; i < cfg.Repeat; i++ {
			start := time.Now()
			conn, err := net.DialTimeout("tcp", result.Addr, time.Duration(cfg.Timeout)*time.Millisecond)
			end := time.Now()
			if err != nil {
				result.Timout = result.Timout + 1
			} else {
				delay := end.Sub(start).Seconds()
				if delay < result.MinDelay {
					result.MinDelay = delay
				}
				if delay > result.MaxDelay {
					result.MaxDelay = delay
				}
				result.AvgDelay = result.AvgDelay + delay
				conn.Close()
			}
		}
		if result.Total == result.Timout {
			result = result.NaN()
		} else {
			result.AvgDelay = result.AvgDelay / float64(result.Total-result.Timout)
		}
		resultChan <- result
	}
}

func (result Result) NaN() Result {
	result.AvgDelay = math.NaN()
	result.MinDelay = math.NaN()
	result.MaxDelay = math.NaN()
	return result
}

func convertStrings(results []Result) [][]string {
	records := [][]string{}
	for _, result := range results {
		records = append(records, result.String())
	}
	return records
}

func (result Result) String() []string {
	ResultType := reflect.TypeOf(result)
	ResultValue := reflect.ValueOf(result)
	line := []string{}
	for i := 0; i < ResultType.NumField(); i++ {
		val := ResultValue.FieldByName(ResultType.Field(i).Name)
		var value string
		switch val.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			value = strconv.FormatInt(val.Int(), 10)
		case reflect.Float32, reflect.Float64:
			value = strconv.FormatFloat(val.Float(), 'f', 4, 64)
		case reflect.String:
			value = val.String()
		default:
			value = val.String()
		}
		line = append(line, value)
	}
	return line
}

func (result Result) getHearders() []string {
	ResultType := reflect.TypeOf(result)
	line := []string{}
	for i := 0; i < ResultType.NumField(); i++ {
		line = append(line, ResultType.Field(i).Name)
	}
	return line
}
