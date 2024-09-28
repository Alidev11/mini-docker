package image

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

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

func DeleteTars(dirPath string) {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		log.Fatalf("Failed to open source(tar) file: %v", err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.Contains(file.Name(), ".tar") {
			fmt.Println("Match: " + file.Name())
			err := os.Remove(dirPath + "/" + file.Name())
			if err != nil {
				log.Printf("Failed to delete %s: %v\n", file.Name(), err)
			} else {
				fmt.Printf("Deleted: %s\n", file.Name())
			}
		} else if !file.IsDir() && strings.Contains(file.Name(), "sha") {
			os.Rename(dirPath+"/"+file.Name(), dirPath+"/"+"redis")
		}
	}
}

func SetupFileDir(srcFile string, outputDir string) {
	slices := []string{"/bin", "/lib", "/lib64"}
	_, err := os.Stat(outputDir)

	// check if directory already exists
	if os.IsNotExist(err) {
		err := os.Mkdir(outputDir, 0755)
		if err != nil {
			log.Fatalf("Failed to create directory: %v", err)
		}
		fmt.Println("Directory created:", outputDir)
	} else if err != nil {
		log.Fatalf("Error with checking if folder exists: %v", err)
	} else {
		fmt.Println("Directory already exists: ", outputDir)
	}

	// check if directory empty before extracting FS
	if result, err := IsDirectoryEmpty(outputDir); err != nil {
		// Handle error from IsDirectoryEmpty
		log.Fatalf("Error checking directory: %v", err)
	} else if result {
		// If the directory is not empty, proceed with untarring
		if err := Untar(srcFile, outputDir); err != nil {
			log.Fatalf("Failed to untar OCI: %v", err)
		}
	}

	// delete .tar files
	DeleteTars(outputDir)

	fmt.Println(slices)
}
