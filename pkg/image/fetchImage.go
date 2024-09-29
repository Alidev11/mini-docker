package image

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

func FetchImg(imageName string) string {
	//------------------ Fetch image
	imageRef := "docker.io/library/" + imageName + ":latest"

	// Parse the image reference
	ref, err := name.ParseReference(imageRef)
	if err != nil {
		log.Fatalf("Failed to parse image ref: %v", err)
	}

	var img v1.Image
	// Retry logic for fetching the image
	for i := 0; i < 3; i++ {
		img, err = remote.Image(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
		if err == nil {
			break
		}
		log.Printf("Attempt %d: Failed to fetch image: %v", i+1, err)
		time.Sleep(2 * time.Second) // Wait before retrying
	}

	if err != nil {
		log.Fatalf("Failed to fetch image after retries: %v", err)
	}

	imageFileName := fmt.Sprintf("./%v.tar", imageName)

	outFileImg, err := os.Create(imageFileName)
	if err != nil{
		log.Fatalf("Failed to create directory: %v", err)
	}

	err = tarball.Write(ref, img, outFileImg)

	fmt.Println("Fetch Image Successfully!")
	return "./" + imageName + ".tar"
}