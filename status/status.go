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

package status

import (
	"fmt"
	"net/http"
)

// Response used to define the status response for Lostromos
type Response struct {
	Success bool
	Info    string
}

// Handler is used for managing calls to /status to inform of the current status of Lostromos.
func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, err := fmt.Fprint(w, "{\"success\": true}")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
