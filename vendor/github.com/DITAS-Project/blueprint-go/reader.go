/*
Copyright 2017 Atos

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package blueprint

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
)

func ReadBlueprint(path string) (*BlueprintType, error) {
	var blueprint BlueprintType

	raw, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(err.Error())
		log.Errorf("Error reading blueprint from %s: %s", path, err.Error())
		return nil, err
	} else {
		err = json.Unmarshal(raw, &blueprint)
		if err != nil {
			log.Errorf("Error reading blueprint: %s", err.Error())
			return nil, err
		}
	}

	return &blueprint, nil
}
