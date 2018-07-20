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
package main

import (
	"SLALite/assessment"
	"SLALite/assessment/monitor"
	"SLALite/assessment/notifier"
	"SLALite/ditas"
	"SLALite/model"
	"SLALite/repositories/memrepository"
	"SLALite/repositories/mongodb"
	"SLALite/repositories/validation"
	"SLALite/utils"
	"flag"
	"strconv"
	"strings"
	"time"

	"github.com/DITAS-Project/blueprint-go"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

// version and date are defined on compilation (see makefile)
var version string
var date string

func main() {

	// TODO: Add windows path
	configPath := flag.String("d", utils.UnixConfigPath, "Directories where to search config files")
	configBasename := flag.String("b", utils.ConfigName, "Filename (w/o extension) of config file")
	configFile := flag.String("f", "", "Path of configuration file. Overrides -b and -d")
	flag.Parse()

	log.Infof("Running SLALite %s compiled on %s", version, date)
	config := createMainConfig(configFile, configPath, configBasename)
	logMainConfig(config)

	singlefile := config.GetBool(utils.SingleFilePropertyName)
	checkPeriod := config.GetDuration(utils.CheckPeriodPropertyName)
	repoType := config.GetString(utils.RepositoryTypePropertyName)

	utils.AddTrustedCAs(config)

	var repoconfig *viper.Viper
	if singlefile {
		repoconfig = config
	}

	var repo model.IRepository
	var errRepo error
	switch repoType {
	case utils.DefaultRepositoryType:
		repo, errRepo = memrepository.New(repoconfig)
	case "mongodb":
		repo, errRepo = mongodb.New(repoconfig)
	}
	if errRepo != nil {
		log.Fatal("Error creating repository: ", errRepo.Error())
	}

	repo, _ = validation.New(repo, false)
	if repo != nil {
		a, _ := NewApp(config, repo)
		vdcId := config.GetString(utils.VDCId)
		_, ops := readBlueprint(repo)
		go createValidationThread(repo, ditas.NewAdapter(ops), ditas.NewNotifier(vdcId), checkPeriod)

		a.Run()
	}
}

//
// Creates the main Viper configuration.
// file: if set, is the path to a configuration file. If not set, paths and basename will be used
// paths: colon separated paths where to search a config file
// basename: basename of a configuration file accepted by Viper (extension is automatic)
//
func createMainConfig(file *string, paths *string, basename *string) *viper.Viper {
	config := viper.New()

	config.SetEnvPrefix(utils.ConfigPrefix) // Env vars start with 'SLA_'
	config.AutomaticEnv()
	config.SetDefault(utils.CheckPeriodPropertyName, utils.DefaultCheckPeriod)
	config.SetDefault(utils.RepositoryTypePropertyName, utils.DefaultRepositoryType)

	if *file != "" {
		config.SetConfigFile(*file)
	} else {
		config.SetConfigName(*basename)
		for _, path := range strings.Split(*paths, ":") {
			config.AddConfigPath(path)
		}
	}

	errConfig := config.ReadInConfig()
	if errConfig != nil {
		log.Println("Can't find configuration file: " + errConfig.Error())
		log.Println("Using defaults")
	}
	return config
}

func logMainConfig(config *viper.Viper) {

	checkPeriod := config.GetDuration(utils.CheckPeriodPropertyName)
	repoType := config.GetString(utils.RepositoryTypePropertyName)

	log.Infof("SLALite initialization\n"+
		"\tConfigfile: %s\n"+
		"\tRepository type: %s\n"+
		"\tCheck period:%d\n",
		config.ConfigFileUsed(), repoType, checkPeriod)

	caPath := config.GetString(utils.CAPathPropertyName)
	if caPath != "" {
		log.Infof("SLALite intialization. Trusted CAs file: %s", caPath)
	}
}

func createValidationThread(repo model.IRepository, ma monitor.MonitoringAdapter,
	not notifier.ViolationNotifier, checkPeriod time.Duration) {

	ticker := time.NewTicker(checkPeriod * time.Second)

	for {
		<-ticker.C
		assessment.AssessActiveAgreements(repo, ma, not)
	}

}

func validateProviders(repo model.IRepository) {
	providers, err := repo.GetAllProviders()

	if err == nil {
		log.Println("There are " + strconv.Itoa(len(providers)) + " providers")
	} else {
		log.Println("Error: " + err.Error())
	}
}

func readBlueprint(repo model.IRepository) (model.Agreements, map[string]blueprint.ExtendedOps) {
	bp, err := blueprint.ReadBlueprint(utils.BlueprintPath)

	if err == nil {
		agreements, ops := ditas.CreateAgreements(bp)

		if agreements != nil {
			for _, agreement := range agreements {
				_, err := repo.CreateAgreement(&agreement)
				if err != nil {
					log.Errorf("Error creating agreement %s: %s", agreement.Id, err.Error())
				}
			}
		}
		return agreements, ops
	}
	log.Errorf("Error reading blueprint: %s", err.Error())
	return make(model.Agreements, 0), make(map[string]blueprint.ExtendedOps)
}
