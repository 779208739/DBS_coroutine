package main

import (
	"go_projects/controller"
	"go_projects/routers"
)

func main() {
	var info controller.AllInfo
	fin := make(chan int)
	go controller.Controller(fin, &info)
	routers.SetupRouter(fin, &info).Run(":9000")
}
