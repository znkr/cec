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

// Package log provides utils to implement logging.
package log

// Single log entry.
type Entry interface{}

// Bounded log that evicts old entries.
type Log struct {
	es  []Entry
	pos int
}

func New(size int) *Log {
	return &Log{
		es:  make([]Entry, size),
		pos: 0,
	}
}

// Adds a new entry to the log, potentially evicting existing entries.
func (l *Log) Add(e Entry) {
	l.es[l.pos] = e
	l.pos = (l.pos + 1) % len(l.es)
}

func (l *Log) Do(f func(Entry)) {
	for i := 0; i < len(l.es); i++ {
		j := (i + l.pos) % len(l.es)
		if l.es[j] == nil {
			continue
		}
		f(l.es[j])
	}
}
