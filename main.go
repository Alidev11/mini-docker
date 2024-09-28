package main

import (
	"mini-docker/pkg/image"
)

func main() {
	imageName := "redis"

	srcFile := image.FetchImg(imageName)
	outputDir := "./" + imageName

	image.SetupFileDir(srcFile, outputDir)
}
