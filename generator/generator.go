/*
Copyright 2018 Atos

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

package generator

import (
	"SLALite/model"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/google/uuid"
)

const (
	nonReplacedTag = "<no value>"
)

const (
	errValidation = "validation"
	errUnreplaced = "unreplaced"
	errOther      = ""
)

type Model struct {
	Template  model.Template         `json:"template"`
	Variables map[string]interface{} `json:"variables"`
}

type generatorError interface {
	IsErrValidation() bool
	IsErrUnreplaced() bool
}

func IsErrValidation(err error) bool {
	v, ok := err.(generatorError)
	return ok && v.IsErrValidation()
}

func IsErrUnreplaced(err error) bool {
	v, ok := err.(generatorError)
	return ok && v.IsErrUnreplaced()
}

type genError struct {
	msg  string
	kind string
}

func (e *genError) Error() string {
	return e.msg
}

func (e *genError) IsErrValidation() bool {
	return e.kind == errValidation
}

func (e *genError) IsErrUnreplaced() bool {
	return e.kind == errUnreplaced
}

// "meta": {
// 	"duration": "P1M"
// },
// "variables":{
// 	"providerid": "p01",
// 	"clientid": "c01",
// 	"M": 0.5,
// 	"N": 0.5
// }

// Do generates an agreement from a generator model
func Do(genmodel *Model) (*model.Agreement, error) {

	// marshal template
	marshalled, err := json.Marshal(genmodel.Template)
	if err != nil {
		/* should not happen */
		return nil, err
	}
	str := string(marshalled)

	// string replacement -> agreement generation
	tmpl, err := template.New(genmodel.Template.Name).Parse(str)
	if err != nil {
		/* should not happen */
		return nil, err
	}

	var b bytes.Buffer
	tmpl.Execute(&b, genmodel.Variables)
	if err != nil {
		return nil, err
	}

	// check all placeholders has been replaced
	s := b.String()
	if i := strings.Index(s, nonReplacedTag); i != -1 {
		return nil, &genError{
			kind: errUnreplaced,
			msg:  fmt.Sprintf("Found non-replaced placeholder at index %d. Agreement is %s", i, s),
		}
	}

	// unmarshal agreement
	var agreement model.Agreement
	err = json.NewDecoder(&b).Decode(&agreement)

	// modify agreement where needed
	agreement.Details.Type = model.AGREEMENT
	agreement.Details.Id = uuid.New().String()
	agreement.Id = agreement.Details.Id
	agreement.Details.Creation = time.Now()
	agreement.Name = agreement.Details.Name

	// validate agreement
	errs := agreement.Validate()
	if len(errs) != 0 {
		return &agreement, newValidationError(errs)
	}

	return &agreement, nil
}

func newValidationError(errs []error) *genError {
	var buffer bytes.Buffer
	for _, err := range errs {
		buffer.WriteString(err.Error())
		buffer.WriteString(". ")
	}
	return &genError{kind: errValidation, msg: buffer.String()}
}
