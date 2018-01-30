// Copyright 2017 the lostromos Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package crwatcher

import (
	"github.com/golang/mock/gomock"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type StatusMatcher struct {
	status CustomResourceStatus
}

func (m StatusMatcher) Matches(x interface{}) bool {
	s := StatusFor(x.(*unstructured.Unstructured))
	return m.status.Phase == s.Phase && m.status.Reason == s.Reason && m.status.Message == s.Message
}

func (m StatusMatcher) String() string {
	return ""
}

func MatchStatus(expected CustomResourceStatus) gomock.Matcher {
	return StatusMatcher{status: expected,}
}
