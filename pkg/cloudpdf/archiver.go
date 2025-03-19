package cloudpdf

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
)

// Archiver speichert die Buffer und Writer für fortlaufendes Archivieren
type Archiver struct {
	buf    *bytes.Buffer
	gw     *gzip.Writer
	tw     *tar.Writer
	closed bool
}

// NewArchiver erstellt eine neue Instanz mit offenen gzip- und tar-Writer
func NewArchiver() *Archiver {
	buf := new(bytes.Buffer)
	gw := gzip.NewWriter(buf)
	tw := tar.NewWriter(gw)
	return &Archiver{
		buf:    buf,
		gw:     gw,
		tw:     tw,
		closed: false,
	}
}

func (a *Archiver) RenderPDF(ctx context.Context, client ClientInterface) ([]byte, error) {
	if !a.closed {
		err := a.Close()
		if err != nil {
			return nil, err
		}
	}
	body, err := client.PostConvertWithBody(ctx, nil, "application/octed-stream", a.GetArchive())
	if err != nil {
		return nil, err
	}
	if body.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("creator api returned %d", body.StatusCode))
	}

	defer body.Body.Close()

	return io.ReadAll(body.Body)
}

// AddFile fügt eine Datei zum Archiv hinzu (Schreibt in den tar.Writer)
func (a *Archiver) AddFile(filePath string) error {
	return addToArchive(a.tw, filePath, filePath)
}

// AddFile fügt eine Datei zum Archiv hinzu (Schreibt in den tar.Writer)
func (a *Archiver) AddFileWithInternalName(filePath string, internalName string) error {
	return addToArchive(a.tw, filePath, internalName)
}

// AddReader fügt einen Reader zum Archiv hinzu
func (a *Archiver) AddReader(filename string, r io.Reader) error {
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, r)
	if err != nil {
		return err
	}
	return addReaderToArchive(
		a.tw, &tar.Header{
			Name: filename,
			Size: int64(len(buf.Bytes())),
			Mode: 0644,
		}, buf,
	)
}

// Close schließt die Writer und finalisiert das Archiv
func (a *Archiver) Close() error {
	if err := a.tw.Close(); err != nil {
		return err
	}
	if err := a.gw.Close(); err != nil {
		return err
	}
	a.closed = true
	return nil
}

// GetArchive gibt das fertige Archiv als Byte-Buffer zurück
func (a *Archiver) GetArchive() *bytes.Buffer {
	return a.buf
}

func addToArchive(tw *tar.Writer, filename string, internalName string) error {
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
	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		return err
	}

	header.Name = internalName

	return addReaderToArchive(tw, header, file)
	// Create a tar Header from the FileInfo data

}

func addReaderToArchive(tw *tar.Writer, header *tar.Header, file io.Reader) error {
	// Write file header to the tar archive
	err := tw.WriteHeader(header)
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
