package main

import (
	"SLALite/model"
)

func main() {

	//repo := model.MemRepository{}
	repo,err := model.CreateRepository("test.db")
	if err != nil {
		a := App{}
		a.Initialize(repo)
		a.Run(":8090")
	}
}
