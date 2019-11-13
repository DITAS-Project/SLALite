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

package model

import (
	"encoding/json"
	"fmt"
	"os"
)

// ReadAgreement returns the agreement read from the file pointed by path.
// The CWD is the location of the test.
//
// Ex:
//    a, err := readAgreement("testdata/a.json")
//    if err != nil {
//      t.Errorf("Error reading agreement: %v", err)
//    }
func ReadAgreement(path string) (Agreement, error) {
	res, err := readEntity(path, new(Agreement))
	a := res.(*Agreement)

	if a.Details.Type != AGREEMENT {
		return *a, fmt.Errorf("%v is not an agreement", a)
	}
	return *a, err
}

// ReadTemplate returns the template read from the file pointed by path.
// The CWD is the location of the test.
//
// Ex:
//    a, err := readAgreement("testdata/a.json")
//    if err != nil {
//      t.Errorf("Error reading agreement: %v", err)
//    }
func ReadTemplate(path string) (Template, error) {
	res, err := readEntity(path, new(Template))
	t := res.(*Template)
	if t.Details.Type != TEMPLATE {
		return *t, fmt.Errorf("%v is not a template", t)
	}
	return *t, err
}

func readEntity(path string, result interface{}) (interface{}, error) {

	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		return result, err
	}
	json.NewDecoder(f).Decode(&result)
	return result, err
}
