package main

import (
	"SLALite/model"
	"SLALite/repositories"
	"github.com/spf13/viper"
	"log"
	"strconv"
	"time"
)

const (
	defaultPort           string        = "8090"
	defaultCheckPeriod    time.Duration = 60
	defaultRepositoryType string        = "memory"

	portPropertyName           = "port"
	checkPeriodPropertyName    = "checkPeriod"
	repositoryTypePropertyName = "repository"

	unixConfigPath = "/etc/slalite"
	configName     = "slalite"
)

func main() {

	log.Println("Initializing")
	viper.SetConfigName(configName)
	viper.AddConfigPath(unixConfigPath)
	// TODO: Add windows path

	viper.SetDefault(portPropertyName, defaultPort)
	viper.SetDefault(checkPeriodPropertyName, defaultCheckPeriod)
	viper.SetDefault(repositoryTypePropertyName, defaultRepositoryType)

	errConfig := viper.ReadInConfig()
	if errConfig != nil {
		log.Println("Can't find configuration file: " + errConfig.Error())
		log.Println("Using defaults")
	}

	port := viper.GetString(portPropertyName)
	checkPeriod := viper.GetDuration(checkPeriodPropertyName)
	repoType := viper.GetString(repositoryTypePropertyName)

	var repo model.IRepository = nil
	switch repoType {
	case defaultRepositoryType:
		repo = repositories.MemRepository{}
	case "bbolt":
		boltRepo, errRepo := repositories.CreateBBoltRepository()
		if errRepo != nil {
			log.Fatal("Error creating bbolt repository: ", errRepo.Error())
		}
		repo = boltRepo
	case "mongodb":
		mongoRepo, errMongo := repositories.CreateMongoDBRepository()
		if errMongo != nil {
			log.Fatal("Error creating mongo repository: ", errMongo.Error())
		}
		repo = mongoRepo
	}
	if repo != nil {
		a := App{}
		a.Initialize(repo)
		go createValidationThread(repo, checkPeriod)
		a.Run(":" + port)
	}
}

func createValidationThread(repo model.IRepository, checkPeriod time.Duration) {
	ticker := time.NewTicker(checkPeriod * time.Second)

	for {
		<-ticker.C
		validateProviders(repo)
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
