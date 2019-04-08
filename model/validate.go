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

import "fmt"

/*
Validator is the interface that contains validate functions for the model entities.

Each function takes an entities and calls Validate(val) on it
*/
type Validator interface {
	ValidateProvider(p *Provider, mode ValidationMode) []error
	ValidateClient(c *Client, mode ValidationMode) []error
	ValidateAgreement(a *Agreement, mode ValidationMode) []error
	ValidateTemplate(t *Template, mode ValidationMode) []error
	ValidateAssessment(as *Assessment, mode ValidationMode) []error
	ValidateDetails(t *Details, mode ValidationMode) []error
	ValidateGuarantee(g *Guarantee, mode ValidationMode) []error
	ValidateViolation(v *Violation, mode ValidationMode) []error
}

// ValidationMode is the type of possible validations
type ValidationMode string

const (
	// CREATE is the mode to be used to validate an entity on creation in repository
	CREATE ValidationMode = "create"

	// UPDATE is the mod to be used to validate an entity after creation
	UPDATE ValidationMode = "update"
)

/*
DefaultValidator is an implementation of Validator.

It validates inputs to the system and should cover most of the cases.
*/
type DefaultValidator struct {
	externalIDs bool
	equalIDs    bool
}

// NewDefaultValidator returns a default Validator.
//
// The externalIDs parameter is true when the Id of the entity is set by the repository,
// and therefore, out of the control of the SLALite; in this case, we cannot enforce that
// the Id is set when creating an entity.
//
// If externalIDs is false, Id and Details.Id can be forced to be equal passing
// equalIDs=true. externalIDs is consider false regardless of its value if externalIDs=true
func NewDefaultValidator(externalIDs bool, equalIDs bool) Validator {
	return DefaultValidator{
		externalIDs: externalIDs,
		equalIDs:    !externalIDs && equalIDs,
	}
}

// ValidateProvider implements model.Validator.ValidateProvider
func (val DefaultValidator) ValidateProvider(p *Provider, mode ValidationMode) []error {
	result := make([]error, 0, 2)

	result = checkEmpty(mode == CREATE && val.externalIDs, p.Id, "Provider.Id", result)
	result = checkNotEmpty(p.Name, "Provider.Name", result)

	return result
}

// ValidateClient implements model.Validator.ValidateClient
func (val DefaultValidator) ValidateClient(c *Client, mode ValidationMode) []error {
	result := make([]error, 0, 2)

	result = checkEmpty(mode == CREATE && val.externalIDs, c.Id, "Client.Id", result)
	result = checkNotEmpty(c.Name, "Client.Name", result)

	return result
}

// ValidateTemplate implements model.Validator.ValidateTemplate
func (val DefaultValidator) ValidateTemplate(t *Template, mode ValidationMode) []error {
	result := make([]error, 0)

	result = checkEmpty(mode == CREATE && val.externalIDs, t.Id, "Template.Id", result)
	result = checkNotEmpty(t.Name, "Template.Name", result)

	for _, e := range t.Details.Validate(val, mode) {
		result = append(result, e)
	}

	if val.equalIDs {
		result = checkEquals(t.Id, "Template.Id", t.Details.Id, "Template.Details.Id", result)
	}
	if t.Details.Type != TEMPLATE {
		result = append(result, fmt.Errorf("Template.Details.Type must be equal to '%s'", TEMPLATE))
	}
	return result
}

// ValidateAgreement implements model.Validator.ValidateAgreement
func (val DefaultValidator) ValidateAgreement(a *Agreement, mode ValidationMode) []error {
	result := make([]error, 0)

	a.State = normalizeState(a.State)
	result = checkEmpty(mode == CREATE && val.externalIDs, a.Id, "Agreement.Id", result)
	result = checkNotEmpty(a.Name, "Agreement.Name", result)
	for _, e := range a.Assessment.Validate(val, mode) {
		result = append(result, e)
	}
	for _, e := range a.Details.Validate(val, mode) {
		result = append(result, e)
	}

	if val.equalIDs {
		result = checkEquals(a.Id, "Agreement.Id", a.Details.Id, "Agreement.Details.Id", result)
	}
	result = checkEquals(a.Name, "Agreement.Name", a.Details.Name, "Agreement.Details.Name", result)

	return result
}

// ValidateAssessment implements model.Validator.ValidateAssessment
func (val DefaultValidator) ValidateAssessment(as *Assessment, mode ValidationMode) []error {
	return []error{}
}

// ValidateDetails implements model.Validator.ValidateDetails
func (val DefaultValidator) ValidateDetails(t *Details, mode ValidationMode) []error {
	result := make([]error, 0)
	result = checkNotEmpty(t.Id, "Text.Id", result)
	result = checkNotEmpty(t.Name, "Text.Name", result)
	/*
	 * On creation, we do not want to pass Mode=CREATE to validate parties
	 */
	for _, e := range t.Provider.Validate(val, UPDATE) {
		result = append(result, e)
	}
	for _, e := range t.Client.Validate(val, UPDATE) {
		result = append(result, e)
	}
	for _, g := range t.Guarantees {
		for _, e := range g.Validate(val, mode) {
			result = append(result, e)
		}
	}
	return result
}

// ValidateViolation implements model.Validator.ValidateViolation
func (val DefaultValidator) ValidateViolation(v *Violation, mode ValidationMode) []error {
	result := make([]error, 0)

	result = checkEmpty(mode == CREATE && val.externalIDs, v.Id, "Violation.Id", result)
	result = checkNotEmpty(v.AgreementId, "Violation.AgreementId", result)
	result = checkNotEmpty(v.Guarantee, "Violation.Guarantee", result)
	if v.Datetime.IsZero() {
		result = append(result, fmt.Errorf("%v is not a valid date", v.Datetime))
	}
	if v.Values == nil || len(v.Values) == 0 {
		result = append(result, fmt.Errorf("Violation.Values cannot be empty"))
	}
	result = checkNotEmpty(v.Constraint, "Violation.Constraint", result)

	return result
}

// ValidateGuarantee implements model.Validator.ValidateGuarantee
func (val DefaultValidator) ValidateGuarantee(g *Guarantee, mode ValidationMode) []error {
	result := make([]error, 0)
	result = checkNotEmpty(g.Name, "Guarantee.Name", result)
	result = checkNotEmpty(g.Constraint, fmt.Sprintf("Guarantee['%s'].Constraint", g.Name), result)

	return result
}

func checkNotEmpty(field string, description string, current []error) []error {
	if field == "" {
		current = append(current, fmt.Errorf("%s is empty", description))
	}
	return current
}

/*
Generic checkEmpty function to test if a string is empty or is not empty
*/
func checkEmpty(empty bool, field string, description string, current []error) []error {
	if empty && field != "" || !empty && field == "" {
		var qualifier string
		if empty {
			qualifier = "not"
		}
		current = append(current, fmt.Errorf("%s is %s empty", description, qualifier))
	}
	return current
}

func checkEquals(f1 string, f1desc, f2 string, f2desc string, current []error) []error {
	if f1 != f2 {
		current = append(current, fmt.Errorf("%s and %s do not match", f1desc, f2desc))
	}
	return current
}

func normalizeState(s State) State {
	for _, v := range States {
		if s == v {
			return s
		}
	}
	return STOPPED
}
