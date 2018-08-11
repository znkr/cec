// Copyright 2017 Google LLC
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

package raspberrypi

//go:generate stringer -type=notify

type notify uint32

const (
	notifyTx              notify = 1 << 0
	notifyRx              notify = 1 << 1
	notifyButtonPressed   notify = 1 << 2
	notifyButtonRelease   notify = 1 << 3
	notifyRemotePressed   notify = 1 << 4
	notifyRemoteRelease   notify = 1 << 5
	notifyLogicalAddr     notify = 1 << 6
	notifyTopology        notify = 1 << 7
	notifyLogicalAddrLost notify = 1 << 15
)
