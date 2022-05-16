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
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/prometheus-community/ipmi_exporter/freeipmi"
)

const (
	ChassisCollectorName CollectorName = "chassis"
)

var (
	chassisPowerStateDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "chassis", "power_state"),
		"Current power state (1=on, 0=off).",
		[]string{},
		nil,
	)
)

type ChassisCollector struct{}

func (c ChassisCollector) Name() CollectorName {
	return ChassisCollectorName
}

func (c ChassisCollector) Cmd() string {
	return "ipmi-chassis"
}

func (c ChassisCollector) Args() []string {
	return []string{"--get-chassis-status"}
}

func (c ChassisCollector) Collect(result freeipmi.Result, ch chan<- prometheus.Metric, target ipmiTarget) (int, error) {
	currentChassisPowerState, err := freeipmi.GetChassisPowerState(result)
	if err != nil {
		level.Error(logger).Log("msg", "Failed to collect chassis data", "target", targetName(target.host), "error", err)
		return 0, err
	}
	ch <- prometheus.MustNewConstMetric(
		chassisPowerStateDesc,
		prometheus.GaugeValue,
		currentChassisPowerState,
	)
	return 1, nil
}
