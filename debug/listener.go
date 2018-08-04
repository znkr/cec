// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package debug provides utils to help debugging HDMI
package debug

import (
	"sync"
	"time"

	"github.com/krynr/cec"
	"github.com/krynr/cec/debug/util/log"
)

type LogEntry struct {
	time time.Time
	msg  cec.Message
}

func (e *LogEntry) Time() time.Time      { return e.time }
func (e *LogEntry) Message() cec.Message { return e.msg }

type LoggingListener struct {
	log  *log.Log
	size int
	mtx  sync.Mutex
}

func NewLoggingListener(size int) *LoggingListener {
	return &LoggingListener{
		log:  log.New(size),
		size: size,
	}
}

func (l *LoggingListener) Message(msg cec.Message) {
	t := time.Now()

	l.mtx.Lock()
	defer l.mtx.Unlock()
	l.log.Add(&LogEntry{
		time: t,
		msg:  msg,
	})
}

func (l *LoggingListener) GetLogged() []*LogEntry {
	r := make([]*LogEntry, 0, l.size)

	l.mtx.Lock()
	defer l.mtx.Unlock()
	l.log.Do(func(e log.Entry) {
		r = append(r, e.(*LogEntry))
	})
	return r
}
