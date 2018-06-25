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
package ditas

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func checkMethodList(t *testing.T, methodList *MethodListType, name string) {
	if methodList == nil {
		t.Fatalf("Can't find method list %s", name)
	}

	if len(methodList.Methods) == 0 {
		t.Fatalf("No methods found in %s section", name)
	}

	if *methodList.Methods[0].Name != "patient-details" {
		t.Fatalf("Unexpected name for method in %s. Found %s", name, *methodList.Methods[0].Name)
	}

}

func TestReader(t *testing.T) {
	blueprint := ReadBlueprint("resources/vdc_blueprint_example_1.json")

	checkMethodList(t, blueprint.DataManagement, "Data Management")
	checkMethodList(t, blueprint.AbstractProperties, "Abstract Properties")

	slas := CreateAgreements(blueprint)

	if slas == nil || len(slas) == 0 {
		t.Fatalf("Did not get any SLA from the blueprint")
	}

	sla := slas[0]
	if sla.Id != "patient-details" {
		t.Fatalf("Did not find patient-details SLA")
	}

	guarantees := sla.Details.Guarantees
	if len(guarantees) != 3 {
		t.Fatalf("Unexpected number of guarantees. Expected 3 but found %d", len(guarantees))
	}

}
