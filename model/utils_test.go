/*
Copyright 2019 Atos

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
	"testing"
)

func TestReadTemplate(t *testing.T) {
	var err error

	_, err = ReadTemplate("testdata/template.json")
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	_, err = ReadTemplate("testdata/notfound.json")
	if err == nil {
		t.Error("FileNotFound error expected")
	}
	_, err = ReadTemplate("testdata/agreement.json")
	if err == nil {
		t.Error("ValidationError expected")
	}
}

func TestReadAgreement(t *testing.T) {
	var err error

	_, err = ReadAgreement("testdata/agreement.json")
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	_, err = ReadAgreement("testdata/notfound.json")
	if err == nil {
		t.Error("FileNotFound error expected")
	}
	_, err = ReadAgreement("testdata/template.json")
	if err == nil {
		t.Error("ValidationError expected")
	}
}
