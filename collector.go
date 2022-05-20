// Copyright 2021 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"os"
	"path"
	"strings"
	"time"

	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/prometheus-community/ipmi_exporter/freeipmi"
)

const (
	namespace   = "ipmi"
	targetLocal = ""
)

var sdrCacheDirectoy = "/root/.freeipmi/sdr-cache/"

type collector interface {
	Name() CollectorName
	Cmd() string
	Args() []string
	Collect(output freeipmi.Result, ch chan<- prometheus.Metric, target ipmiTarget) (int, error)
}

type metaCollector struct {
	job    string
	target string
	module string
	config *SafeConfig
}

type ipmiTarget struct {
	host   string
	config IPMIConfig
}

var (
	upDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "up"),
		"'1' if a scrape of the IPMI device was successful, '0' otherwise.",
		[]string{"collector"},
		nil,
	)

	durationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape_duration", "seconds"),
		"Returns how long the scrape took to complete in seconds.",
		nil,
		nil,
	)
)

// Describe implements Prometheus.Collector.
func (c metaCollector) Describe(ch chan<- *prometheus.Desc) {
	// all metrics are described ad-hoc
}

func markCollectorUp(ch chan<- prometheus.Metric, name string, up int) {
	ch <- prometheus.MustNewConstMetric(
		upDesc,
		prometheus.GaugeValue,
		float64(up),
		name,
	)
}

// Collect implements Prometheus.Collector.
func (c metaCollector) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		level.Debug(logger).Log("msg", "Scrape duration", "target", targetName(c.target), "duration", duration)
		ch <- prometheus.MustNewConstMetric(
			durationDesc,
			prometheus.GaugeValue,
			duration,
		)
	}()

	config := c.config.ConfigForTarget(c.target, c.module)
	target := ipmiTarget{
		host:   c.target,
		config: config,
	}
	flushSensorSDRCache(target, config.GetFreeipmiConfig())

	for _, collector := range config.GetCollectors() {
		var up int
		level.Debug(logger).Log("msg", "Running collector", "target", target.host, "collector", collector.Name())

		fqcmd := path.Join(*executablesPath, collector.Cmd())
		args := collector.Args()
		cfg := config.GetFreeipmiConfig()

		result := freeipmi.Execute(fqcmd, args, cfg, target.host, logger)

		up, _ = collector.Collect(result, ch, target)
		markCollectorUp(ch, string(collector.Name()), up)
	}
}

func targetName(target string) string {
	if target == targetLocal {
		return "[local]"
	}
	return target
}

func flushSensorSDRCache(target ipmiTarget, cfg string) error {
	dirRead, err := os.Open(sdrCacheDirectoy)
	if err != nil {
		return err
	}
	dirFiles, err := dirRead.Readdir(0)
	if err != nil {
		return err
	}
	now := time.Now().Local()
	nextTick := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	for index := range dirFiles {
		file := dirFiles[index]
		name := file.Name()
		if strings.Contains(name, target.host) {
			info, err := os.Stat(sdrCacheDirectoy + name)
			if err != nil {
				level.Error(logger).Log(err)
				continue
			}
			modTime := info.ModTime()
			level.Debug(logger).Log("SDR-Cache age: %s ", modTime)
			if nextTick.Day()-modTime.Day() > 0 || nextTick.Month() != modTime.Month() {
				level.Info(logger).Log("Starting to flush SDR Cache. Cache-Time %s, NextTick: %s", modTime, nextTick)
				freeipmi.Execute("ipmi-sensors", []string{"--flush-cache"}, cfg, target.host, logger)
			}
		}
	}
	return nil
}
