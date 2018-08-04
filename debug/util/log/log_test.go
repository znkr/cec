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

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

var logTests = []struct {
	name   string
	add    []string
	expect []string
}{
	{
		name:   "empty",
		add:    []string{},
		expect: []string{},
	},
	{
		name:   "add_1",
		add:    []string{"a"},
		expect: []string{"a"},
	},
	{
		name:   "add_2",
		add:    []string{"a", "b"},
		expect: []string{"a", "b"},
	},
	{
		name:   "add_5",
		add:    []string{"a", "b", "c", "d", "e"},
		expect: []string{"a", "b", "c", "d", "e"},
	},
	{
		name:   "add_6",
		add:    []string{"a", "b", "c", "d", "e", "f"},
		expect: []string{"b", "c", "d", "e", "f"},
	},
	{
		name:   "add_7",
		add:    []string{"a", "b", "c", "d", "e", "f", "g"},
		expect: []string{"c", "d", "e", "f", "g"},
	},
	{
		name:   "add_10",
		add:    []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"},
		expect: []string{"f", "g", "h", "i", "j"},
	},
	{
		name:   "add_11",
		add:    []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"},
		expect: []string{"g", "h", "i", "j", "k"},
	},
}

func TestLog(t *testing.T) {
	for _, test := range logTests {
		t.Run(test.name, func(t *testing.T) {
			l := New(5)
			for _, e := range test.add {
				l.Add(e)
			}

			actual := make([]string, 0)
			l.Do(func(e Entry) { actual = append(actual, e.(string)) })
			if diff := cmp.Diff(actual, test.expect); diff != "" {
				t.Errorf("Expected log to be %q, but got %q: %s", test.expect, actual, diff)
			}
		})
	}
}
