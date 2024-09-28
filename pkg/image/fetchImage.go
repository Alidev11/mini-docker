package image

import (
	"archive/tar"
	// "compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
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

	imageFileName := fmt.Sprintf("./%v", imageName)

	err = os.Mkdir(imageFileName, 0600)
	if err != nil{
		log.Fatalf("Failed to create directory: %v", err)
	}

	if err := writeImageLayersToTarFiles(img, imageFileName); err != nil {
		log.Fatalf("Failed to write image layers to tar files: %v", err)
	}

	fmt.Println("Fetch Image Successfully!")
	return "./" + imageName
}

func writeImageLayersToTarFiles(img v1.Image, imageFileName string) error {
	// Get all layers
	layers, err := img.Layers()
	if err != nil {
		return fmt.Errorf("failed to get image layers: %v", err)
	}

	// Use waitgroup to wait for all threads to finish
	var wg sync.WaitGroup

	// Channel to capture errors from goroutines
	layerChan := make(chan error, len(layers))

	for i, layer := range layers {
		wg.Add(1)
		go func(i int, layer v1.Layer) {
			defer wg.Done()

			// Create a tar file for this layer
			layerFilename := fmt.Sprintf("%v/layer-%d.tar", imageFileName, i)

			outFile, err := os.Create(layerFilename)
			if err != nil {
				layerChan <- fmt.Errorf("failed to create file %s: %v", layerFilename, err)
				return
			}
			defer outFile.Close()

			tw := tar.NewWriter(outFile)
			defer tw.Close()

			if err := writeLayerContentToTar(tw, layer); err != nil {
				layerChan <- err
				return
			}
			fmt.Printf("Layer %d written to %s successfully.\n", i, layerFilename)

			layerChan <- nil
		}(i, layer)
	}

	wg.Wait()
	close(layerChan)

	for err := range layerChan {
		if err != nil {
			return err
		}
	}

	return nil
}

func writeLayerContentToTar(tw *tar.Writer, layer v1.Layer) error {
	// Open the uncompressed stream
	reader, err := layer.Uncompressed()
	if err != nil {
		return fmt.Errorf("Failed to get layer uncompressed stream: %v", err)
	}
	defer reader.Close()

	// Create a new tar reader to read the contents of the layer
	tarReader := tar.NewReader(reader)

	// Iterate over each file in the layer tar
	for {
		// Read the next header from the tar file
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of tar archive
		}
		if err != nil {
			return fmt.Errorf("Failed to read layer content: %v", err)
		}

		// Write the header to the new tar file
		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("Failed to write header to layer tar: %v", err)
		}

		// Copy the file content to the tar writer
		if _, err := io.Copy(tw, tarReader); err != nil {
			return fmt.Errorf("Failed to copy file content to layer tar: %v", err)
		}
	}

	return nil
}
