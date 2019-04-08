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

package utils

import "time"

const (
	// ConfigPrefix is the prefix of env vars that configure the SLALite
	ConfigPrefix string = "sla"

	// DefaultCheckPeriod is the default number of seconds of the periodic assessment execution
	DefaultCheckPeriod time.Duration = 60

	// DefaultRepositoryType is the name of the default repository
	DefaultRepositoryType string = "memory"

	// DefaultExternalIDs is the default value of externalIDs
	DefaultExternalIDs bool = false

	// CheckPeriodPropertyName is the name of the property CheckPeriod
	CheckPeriodPropertyName = "checkPeriod"

	// RepositoryTypePropertyName is the name of the property repository type
	RepositoryTypePropertyName = "repository"

	// ExternalIDsPropertyName is a boolean value that indicates if the used repository
	// auto assigns the ID of entities when they are stored on repository
	ExternalIDsPropertyName = "externalIDs"

	// SingleFilePropertyName is the name of the property single file
	// If singlefile is set, all configuration is retrieved from a single file.
	// If not, configuration may be obtained from several files: e.g. mongodb configuration
	// is obtained from mongodb.yml file.
	SingleFilePropertyName = "singlefile"

	// UnixConfigPath is the ":" separated list of directories where to search for config files
	UnixConfigPath = "/etc/slalite:."

	// ConfigName is the default filename of the configuration file
	ConfigName = "slalite"
)
