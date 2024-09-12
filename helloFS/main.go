package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

func FetchImg() string {
	imageRef := "docker.io/library/hello-world:latest"

	// Parse the image reference
	ref, err := name.ParseReference(imageRef)
	// ref: is an object name.Reference, value -> docker.io/library/hello-world:latest

	if err != nil {
		log.Fatalf("Failed to parse image ref: %v", err)
	}

	// Fetch OCI image from Docker HUB
	img, err := remote.Image(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	// img: object of v1.Image
	if err != nil {
		log.Fatalf("Failed to fetch image: %v", err)
	}

	// Create a file "hello-world.tar" and write the image in to it

	outFile, err := os.Create("hello-world.tar")
	if err != nil {
		log.Fatalf("Failed to save image as tar: %v", err)
	}
	defer outFile.Close()

	err = tarball.Write(ref, img, outFile)

	if err != nil {
		log.Fatalf("Failed to write image as tar: %v", err)
	}

	fmt.Println("Fetch Image Successfully !")
	return "hello-world.tar"
}

func Untar(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		log.Fatalf("Failed to open source(tar) file: %v", err)
	}
	defer file.Close()

	tarReader := tar.NewReader(file)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break //END OF FILE
		}
		if err != nil {
			log.Fatalf("Failed to traverse the tar file: %v", err)
		}

		//
		target := filepath.Join(dest, header.Name)

		//
		switch header.Typeflag {
		case tar.TypeDir:
			// create Directory
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				log.Fatalf("Failed to create Directory: %v", err)
			}
		case tar.TypeReg:
			// create file
			outFile, err := os.Create(target)
			if err != nil {
				log.Fatalf("Failed to create file: %v", err)
			}
			defer outFile.Close()

			// copy file content
			if _, err := io.Copy(outFile, tarReader); err != nil {
				log.Fatalf("Failed to copy file contents: %v", err)
			}

			// Check if the file is a .tar.gz
			if strings.HasSuffix(header.Name, ".tar.gz") {
				fmt.Println("Extracting nested .tar.gz file:", target)
				if err := UntarGz(target, filepath.Dir(target)); err != nil {
					return fmt.Errorf("failed to extract nested .tar.gz: %v", err)
				}
			}
		}
	}

	return err
}

// UntarGz function extracts a .tar.gz file to the specified destination
func UntarGz(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open .tar.gz file: %v", err)
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %v", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return fmt.Errorf("error reading .tar.gz file: %v", err)
		}

		target := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory: %v", err)
			}
		case tar.TypeReg:
			outFile, err := os.Create(target)
			if err != nil {
				return fmt.Errorf("failed to create file: %v", err)
			}
			defer outFile.Close()

			if _, err := io.Copy(outFile, tarReader); err != nil {
				return fmt.Errorf("failed to write file: %v", err)
			}

			// Check if the file is a .tar.gz (nested case)
			if strings.HasSuffix(header.Name, ".tar.gz") {
				fmt.Println("Extracting nested .tar.gz file:", target)
				if err := UntarGz(target, filepath.Dir(target)); err != nil {
					return fmt.Errorf("failed to extract nested .tar.gz: %v", err)
				}
			}
		default:
			// Handle other file types if necessary
		}
	}
	return nil
}

// check if dir is empty
func IsDirectoryEmpty(dirPath string) (bool, error) {
	// Open the directory
	dir, err := os.Open(dirPath)
	if err != nil {
		return false, err
	}
	defer dir.Close()

	// Read directory contents
	files, err := dir.Readdirnames(-1)
	if err != nil {
		return false, err
	}
	// Return true if there are no files
	return len(files) == 0, nil
}

func main() {
	srcFile := FetchImg()
	outputFile := "./contUntared"
	
	if result, err := IsDirectoryEmpty(outputFile); err != nil {
        // Handle error from IsDirectoryEmpty
        log.Fatalf("Error checking directory: %v", err)
    } else if result {
        // If the directory is not empty, proceed with untarring
        if err := Untar(srcFile, outputFile); err != nil {
            log.Fatalf("Failed to untar OCI: %v", err)
        }
    }

	cmd := exec.Command(outputFile + "/hello")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

}
