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

const (
	// UnixConfigPath is the default configuration path in *ix platforms.
	UnixConfigPath = "/etc/slalite"
)

// IRepository expose the interface to be fulfilled by implementations of repositories.
type IRepository interface {
	/*
	 * GetAllProviders returns the list of providers.
	 *
	 * The list is empty when there are no providers;
	 * error != nil on error
	 */
	GetAllProviders() (Providers, error)

	/*
	 * GetProvider returns the Provider identified by id.
	 *
	 * error != nil on error;
	 * error is sql.ErrNoRows if the provider is not found
	 */
	GetProvider(id string) (*Provider, error)

	/*
	 * CreateProvider stores a new provider.
	 *
	 * error != nil on error;
	 * error is sql.ErrNoRows if the provider already exists
	 */
	CreateProvider(provider *Provider) (*Provider, error)

	/*
	 * DeleteProvider deletes from the repository the provider whose id is provider.Id.
	 *
	 * error != nil on error;
	 * error is sql.ErrNoRows if the provider does not exist.
	 */
	DeleteProvider(provider *Provider) error

	/*
	 * GetAllAgreements returns the list of agreements.
	 *
	 * The list is empty when there are no agreements;
	 * error != nil on error
	 */
	GetAllAgreements() (Agreements, error)

	/*
	 * GetAgreement returns the Agreement identified by id.
	 * error != nil on error;
	 * error is sql.ErrNoRows if the Agreement is not found
	 */
	GetAgreement(id string) (*Agreement, error)

	/*
	 * GetAgreementsByState returns the agreements that have one of the items in states.
	 *
	 * error != nil on error;
	 */
	GetAgreementsByState(states ...State) (Agreements, error)

	/*
	 * CreateAgreement stores a new Agreement.
	 *
	 * error != nil on error;
	 * error is sql.ErrNoRows if the Agreement already exists
	 */
	CreateAgreement(agreement *Agreement) (*Agreement, error)

	/*
	 *UpdateAgreement updates the information of an already saved instance of an agreement
	 */
	UpdateAgreement(agreement *Agreement) (*Agreement, error)

	/*
	 * DeleteAgreement deletes from the repository the Agreement whose id is provider.Id.
	 *
	 * error != nil on error;
	 * error is sql.ErrNoRows if the Agreement does not exist.
	 */
	DeleteAgreement(agreement *Agreement) error

	/*
	 * GetAllTemplates returns the list of templates.
	 *
	 * The list is empty when there are no templates;
	 * error != nil on error
	 */
	GetAllTemplates() (Templates, error)

	/*
	 * GetTemplate returns the Template identified by id.
	 * error != nil on error;
	 * error is sql.ErrNoRows if the Template is not found
	 */
	GetTemplate(id string) (*Template, error)

	/*
	 * CreateTemplate stores a new Template.
	 *
	 * error != nil on error;
	 * error is sql.ErrNoRows if the Template already exists
	 */
	CreateTemplate(template *Template) (*Template, error)

	/*
	 * CreateViolation stores a new Violation.
	 *
	 * error != nil on error;
	 * error is sql.ErrNoRows if the Violation already exists
	 */
	CreateViolation(v *Violation) (*Violation, error)

	/*
	 * GetViolation returns the Violation identified by id.
	 *
	 * error != nil on error;
	 * error is sql.ErrNoRows if the Violation is not found
	 */
	GetViolation(id string) (*Violation, error)

	/*
	 * UpdateAgreementState changes the state of an Agreement.
	 *
	 * Returns the updated agreement; error != nil on error
	 *
	 * error is sql.ErrNoRows if the Agreement does not exist
	 * Non-sentinel error is returned if not a valid transition
	 * (it is recommended to check a.IsValidTransition before UpdateAgreementState)
	 */
	UpdateAgreementState(id string, newState State) (*Agreement, error)
}
