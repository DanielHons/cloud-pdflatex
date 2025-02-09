package xgzip

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func UnGzipIntoFolder(r io.Reader, folder string) error {
	uncompressedStream, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	tarReader := tar.NewReader(uncompressedStream)
	for true {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(header.Name, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			outFile, err := os.Create(filepath.Join(folder, header.Name))
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return err
			}
			_ = outFile.Close()

		default:
			return err
		}
	}
	return nil
}

func GzipFolder(folder string) (*bytes.Buffer, error) {
	dir, err := os.ReadDir(folder)
	if err != nil {
		return nil, err
	}
	out := new(bytes.Buffer)

	err = createArchive(
		mapSLice(
			dir, func(v os.DirEntry) string {
				return filepath.Join(folder, v.Name())
			},
		), out, folder,
	)
	if err != nil {
		return nil, err
	}
	fmt.Println("Archive created successfully")
	permissions := 0644 // or whatever you need

	err = os.WriteFile("file.tar.gz", out.Bytes(), os.FileMode(permissions))
	f, err := os.Open("file.tar.gz")
	if err != nil {
		return nil, err
	}
	err = UnGzipIntoFolder(f, ".")
	if err != nil {
		return nil, err
	}

	return out, nil
}

func createArchive(files []string, buf io.Writer, tmpDir string) error {
	// concat writers
	gw := gzip.NewWriter(buf)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Iterate over files and add them to the tar archive
	for _, file := range files {
		err := addToArchive(tw, file, tmpDir)
		if err != nil {
			return err
		}
	}

	return nil
}

func addToArchive(tw *tar.Writer, filename string, trimfolder string) error {
	// Open the file which will be written into the archive
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Get FileInfo about our file providing file size, mode, etc.
	info, err := file.Stat()
	if err != nil {
		return err
	}

	// Create a tar Header from the FileInfo data
	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		return err
	}

	// We strip away the root folder
	header.Name = strings.TrimPrefix(filename, trimfolder)

	// Write file header to the tar archive
	err = tw.WriteHeader(header)
	if err != nil {
		return err
	}

	// Copy file content to tar archive
	_, err = io.Copy(tw, file)
	if err != nil {
		return err
	}

	return nil
}

// batch conversion helper
func mapSLice[V any, W comparable](s []V, extractor func(V) W) []W {
	unique := make(map[W]struct{})
	result := make([]W, 0)

	for _, value := range s {
		prop := extractor(value)
		if _, exists := unique[prop]; !exists {
			unique[prop] = struct{}{}
			result = append(result, prop)
		}
	}
	return result
}
