// Licensed to Apache Software Foundation (ASF) under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Apache Software Foundation (ASF) licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package core

import (
	"sync"

	"github.com/apache/skywalking-go/plugins/core/metrics"
)

var (
	instance *So11y
	once     sync.Once
)

type So11y struct {
	propagatedContextCounter    metrics.Counter
	samplerIgnoreContextCounter metrics.Counter
	finishIgnoreContextCounter  metrics.Counter
	finishContextCounter        metrics.Counter
	leakedIgnoreContextCounter  metrics.Counter
	leakedContextCounter        metrics.Counter

	errorCounterMap sync.Map
}

func GetSo11y() *So11y {
	once.Do(func() {
		instance = &So11y{}
	})
	return instance
}

func (s *So11y) MeasureTracingContextCreation(t *Tracer, isForceSample, isIgnored bool) {
	if isForceSample {
		if isIgnored {
			if s.propagatedContextCounter == nil {
				s.propagatedContextCounter = t.NewCounter("sw_go_created_ignored_context_counter",
					metrics.WithLabel("created_by", "propagated")).(metrics.Counter)
			}
			s.propagatedContextCounter.Inc(1)
		} else {
			if s.propagatedContextCounter == nil {
				s.propagatedContextCounter = t.NewCounter("sw_go_created_tracing_context_counter",
					metrics.WithLabel("created_by", "propagated")).(metrics.Counter)
			}
			s.propagatedContextCounter.Inc(1)
		}
	} else {
		if isIgnored {
			if s.samplerIgnoreContextCounter == nil {
				s.samplerIgnoreContextCounter = t.NewCounter("sw_go_created_ignored_context_counter",
					metrics.WithLabel("created_by", "sampler")).(metrics.Counter)
			}
			s.samplerIgnoreContextCounter.Inc(1)
		} else {
			if s.samplerIgnoreContextCounter == nil {
				s.samplerIgnoreContextCounter = t.NewCounter("sw_go_created_tracing_context_counter",
					metrics.WithLabel("created_by", "sampler")).(metrics.Counter)
			}
			s.samplerIgnoreContextCounter.Inc(1)
		}
	}
}

func (s *So11y) MeasureTracingContextCompletion(t *Tracer, isIgnored bool) {
	if isIgnored {
		if s.finishIgnoreContextCounter == nil {
			s.finishIgnoreContextCounter = t.NewCounter(
				"sw_go_finished_ignored_context_counter", nil).(metrics.Counter)
		}
		s.finishIgnoreContextCounter.Inc(1)
	} else {
		if s.finishContextCounter == nil {
			s.finishContextCounter = t.NewCounter(
				"sw_go_finished_tracing_context_counter", nil).(metrics.Counter)
		}
		s.finishContextCounter.Inc(1)
	}
}

func (s *So11y) MeasureLeakedTracingContext(t *Tracer, isIgnored bool) {
	if isIgnored {
		if s.leakedIgnoreContextCounter == nil {
			s.leakedIgnoreContextCounter = t.NewCounter("sw_go_possible_leaked_context_counter",
				metrics.WithLabel("source", "ignore")).(metrics.Counter)
		}
		s.leakedIgnoreContextCounter.Inc(1)
	} else {
		if s.leakedContextCounter == nil {
			s.leakedContextCounter = t.NewCounter("sw_go_possible_leaked_context_counter",
				metrics.WithLabel("source", "tracing")).(metrics.Counter)
		}
		s.leakedContextCounter.Inc(1)
	}
}

func (t *Tracer) So11y() interface{} {
	return t
}

func (t *Tracer) CollectErrorOfPlugin(pluginName string) {
	if counter, ok := GetSo11y().errorCounterMap.Load(pluginName); ok {
		if c, ok := counter.(metrics.Counter); ok {
			c.Inc(1)
		}
	} else {
		counter, _ := GetSo11y().errorCounterMap.LoadOrStore(pluginName, t.NewCounter(
			"sw_go_interceptor_error_counter", metrics.WithLabel("plugin_name", pluginName)).(metrics.Counter))
		if c, ok := counter.(metrics.Counter); ok {
			c.Inc(1)
		}
	}
}
