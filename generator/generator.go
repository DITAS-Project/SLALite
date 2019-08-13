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

/*
Package generator builds an Agreement from a Template.

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

// Model contains the parameters needed to do an agreement generation
type Model struct {
	Template  model.Template         `json:"template"`
	Variables map[string]interface{} `json:"variables"`
}

type generatorError interface {
	IsErrValidation() bool
	IsErrUnreplaced() bool
}

// IsErrValidation checks that the err is an ErrValidation error
func IsErrValidation(err error) bool {
	v, ok := err.(generatorError)
	return ok && v.IsErrValidation()
}

// IsErrUnreplaced checks that the err is an ErrUnreplaced error
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

/*
Do generates an agreement from a generator model.

This generator receives a model, which contains a Template and a list of variables
with their values, and substitutes the placeholders in the template the appropriate
variable values. The output is an agreement.

The generator uses text/template syntax for the substitution of the placeholds,
i.e., placeholders are of the type {{.var}} to substitute the value of {{.var}} with
the value of the key 'var' in the model.Variables map.

A right template should define placeholders in the following paths:

- Details.Client

- Details.Provider if the template is used by more than one provider
(if not, the provider can be hardcoded in the template)

- Details.Name

Placeholders may be defined in other paths, e.g.
Guarantee[i].Constraint, Details.Expiration

After the substitution, the following fields are modified,
regardless of the template content:

- Details.Type: set to agreement type

- Details.Id: UUID value

- Id: equals to agreement.Details.Id

- Details.Creation: current time

- Name: equal to agreement.Details.Name

An error of type validation is returned if the validation on the generated agreement
fails. An error of type unreplaced is returned if there is a placeholder that is
not substituted. Use IsErrValidation and IsErrUnreplaced to check type of an error.
*/
func Do(genmodel *Model, val model.Validator, externalIDs bool) (*model.Agreement, error) {

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
	if !externalIDs {
		agreement.Id = agreement.Details.Id
	}
	agreement.Details.Creation = time.Now()
	agreement.Name = agreement.Details.Name

	// validate agreement
	errs := agreement.Validate(val, model.CREATE)
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
