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
	"SLALite/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
)

var tpl, tplIncomplete model.Template

var val = model.NewDefaultValidator(false, true)

func TestMain(m *testing.M) {
	var err error

	log.SetLevel(log.DebugLevel)

	tpl, err = utils.ReadTemplate("testdata/template.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	tplIncomplete, err = utils.ReadTemplate("testdata/incomplete-template.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	result := m.Run()
	os.Exit(result)
}

func TestGenerateAgreement(t *testing.T) {
	genmodel := Model{
		Template: tpl,
		Variables: map[string]interface{}{
			"provider":      model.Provider{Id: "<provider-id>", Name: "<provider-name>"},
			"client":        model.Client{Id: "<client-id>", Name: "<client-name>"},
			"M":             "500",
			"N":             "0.9",
			"agreementname": "<a-name>",
		},
	}
	a, err := Do(&genmodel, val, false)
	if err == nil {
		var b bytes.Buffer
		enc := json.NewEncoder(&b)
		enc.SetEscapeHTML(false)
		enc.SetIndent(" ", " ")
		enc.Encode(a)
		log.Debug(b.String())
	} else {
		enc := json.NewEncoder(os.Stdout)
		enc.SetEscapeHTML(false)
		enc.SetIndent(" ", " ")
		enc.Encode(a)
		t.Errorf("%v", err)
	}
}

func TestGenerateAgreementMissingFields(t *testing.T) {
	genmodel := Model{
		Template: tpl,
		Variables: map[string]interface{}{
			"agreementname": "<a-name>",
			"provider":      model.Provider{Id: "<provider-id>", Name: "<provider-name>"},
			"client":        model.Client{Id: "<client-id>", Name: "<client-name>"},
			"N":             "0.9",
		},
	}
	a, err := Do(&genmodel, val, false)
	if err == nil || !IsErrUnreplaced(err) {
		enc := json.NewEncoder(os.Stdout)
		enc.SetEscapeHTML(false)
		enc.SetIndent(" ", " ")
		enc.Encode(a)
		t.Errorf("Unexpected err. Expected: ErrUnreplaced; actual: %v", err)
	}
}

func TestGenerateAgreementNonValid(t *testing.T) {
	genmodel := Model{
		Template: tplIncomplete,
		Variables: map[string]interface{}{
			"agreementname": "<a-name>",
			"provider":      model.Provider{Id: "<provider-id>", Name: "<provider-name>"},
			"client":        model.Client{Id: "<client-id>", Name: "<client-name>"},
			"M":             "500",
			"N":             "0.9",
		},
	}
	a, err := Do(&genmodel, val, false)
	if err == nil || !IsErrValidation(err) {
		enc := json.NewEncoder(os.Stdout)
		enc.SetEscapeHTML(false)
		enc.SetIndent(" ", " ")
		enc.Encode(a)
		t.Errorf("Unexpected err. Expected: ErrValidation; actual: %v", err)
	}
}
