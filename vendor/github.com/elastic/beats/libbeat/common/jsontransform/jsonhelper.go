// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
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

package jsontransform

import (
	"fmt"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
)

// WriteJSONKeys writes the json keys to the given event based on the overwriteKeys option and the addErrKey
func WriteJSONKeys(event *beat.Event, keys map[string]interface{}, overwriteKeys bool, addErrKey bool) {
	if !overwriteKeys {
		for k, v := range keys {
			if _, exists := event.Fields[k]; !exists && k != "@timestamp" && k != "@metadata" {
				event.Fields[k] = v
			}
		}
		return
	}

	for k, v := range keys {
		switch k {
		case "@timestamp":
			vstr, ok := v.(string)
			if !ok {
				logp.Err("JSON: Won't overwrite @timestamp because value is not string")
				event.SetErrorWithOption(createJSONError("@timestamp not overwritten (not string)"), addErrKey)
				continue
			}

			// @timestamp must be of format RFC3339
			ts, err := time.Parse(time.RFC3339, vstr)
			if err != nil {
				logp.Err("JSON: Won't overwrite @timestamp because of parsing error: %v", err)
				event.SetErrorWithOption(createJSONError(fmt.Sprintf("@timestamp not overwritten (parse error on %s)", vstr)), addErrKey)
				continue
			}
			event.Timestamp = ts

		case "@metadata":
			switch m := v.(type) {
			case map[string]string:
				for meta, value := range m {
					event.Meta[meta] = value
				}

			case map[string]interface{}:
				event.Meta.Update(common.MapStr(m))

			default:
				event.SetErrorWithOption(createJSONError("failed to update @metadata"), addErrKey)
			}

		case "type":
			vstr, ok := v.(string)
			if !ok {
				logp.Err("JSON: Won't overwrite type because value is not string")
				event.SetErrorWithOption(createJSONError("type not overwritten (not string)"), addErrKey)
				continue
			}
			if len(vstr) == 0 || vstr[0] == '_' {
				logp.Err("JSON: Won't overwrite type because value is empty or starts with an underscore")
				event.SetErrorWithOption(createJSONError(fmt.Sprintf("type not overwritten (invalid value [%s])", vstr)), addErrKey)
				continue
			}
			event.Fields[k] = vstr

		default:
			event.Fields[k] = v
		}
	}
}

func createJSONError(message string) common.MapStr {
	return common.MapStr{"message": message, "type": "json"}
}
