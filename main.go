package main

import (
	"sla/model"
)

func main() {

	repo := model.MemRepository{}
	a := App{}
	a.Initialize(repo)
	a.Run(":8090")
}
