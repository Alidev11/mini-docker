package main

import (
	"fmt"
	"mini-docker/pkg/image"
)

func main() {
	imageName := "redis"

	srcFile := image.FetchImg(imageName)
	fmt.Println(srcFile)
	// titleCaser := cases.Title(language.English)
	// outputDir := "../contUntared" + titleCaser.String(imageName)

	// image.SetupFileDir(srcFile, outputDir)
}
