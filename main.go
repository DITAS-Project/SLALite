package main

import (
	"SLALite/model"
	"time"
	"log"
	"strconv"
)



func main() {

	log.Println("Initializing")
	//repo := model.MemRepository{}
	repo,err := model.CreateRepository("test.db")
	if err == nil {
		a := App{}
		a.Initialize(repo)
		go createValidationThread(repo)
		a.Run(":8090")
	}
}

func createValidationThread(repo model.IRepository) {
	ticker := time.NewTicker(5 * time.Second)

	for {
		<-ticker.C
		validateProviders(repo)
	}
	
}

func validateProviders(repo model.IRepository) {
	providers,err := repo.GetAllProviders()

	if err == nil {
		log.Println("There are " + strconv.Itoa(len(providers)) + " providers")
	} else {
		log.Println("Error: " + err.Error())
	}
}
