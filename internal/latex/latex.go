package latex

import (
	"context"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type TimeSheet struct {
	CreationDate string
	Data         []map[string]interface{}
}

func Sanitize(string2 string) string {
	replacer := strings.NewReplacer("_", "\\_", "\"", "", "#", "\\#")
	return replacer.Replace(string2)
}

func NewFolderRenderer(renderDir string, main string) (*Renderer, error) {
	texFile, err := os.OpenFile(filepath.Join(renderDir, main), os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return nil, err
	}
	return &Renderer{renderDir: renderDir, texFile: texFile, TexConsoleWriter: os.Stdout}, nil
}

type Renderer struct {
	renderDir        string
	texFile          *os.File
	TexConsoleWriter io.Writer
}

func (r *Renderer) GetRenderDirectory() string {
	return r.renderDir
}

func (r *Renderer) Write(p []byte) (n int, err error) {
	return r.texFile.Write(p)
}

func (r *Renderer) MainTexFile() string {
	return r.texFile.Name()
}

func (r *Renderer) RemoveSourceFiles() error {
	errs := make([]error, 0)
	err := r.removeGeneratedFiles()
	if err != nil {
		errs = append(errs, err)
	}

	err = os.Remove(r.texFile.Name())
	if err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)

	}
	return nil
}

func (r *Renderer) RenamePdfFile(outFile string) error {
	pdfPath := r.getTempPdfPath()
	return os.Rename(pdfPath, outFile)
}

func (r *Renderer) PdfDoesExist() bool {
	if _, err := os.Stat(r.getTempPdfPath()); errors.Is(err, os.ErrNotExist) {
		return false
	} else {
		return true
	}
}

func (r *Renderer) getTempPdfPath() string {
	pdfPath := filepath.Join(
		r.renderDir, strings.TrimSuffix(filepath.Base(r.MainTexFile()), filepath.Ext(r.MainTexFile()))+".pdf",
	)
	return pdfPath
}

func (r *Renderer) removeGeneratedFiles() error {
	errs := make([]error, 0)
	tmpFileBase := r.texFileBase()
	err := os.Remove(tmpFileBase + ".aux")
	if err != nil {
		errs = append(errs, err)
	}
	err = os.Remove(tmpFileBase + ".log")
	if err != nil {
		errs = append(errs, err)
	}
	err = os.Remove(tmpFileBase + ".synctex.gz")
	if err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return errors.Join(errs...)

	}
	return nil

}

// putFile schreibt die Datei sicher in renderDir, ohne dass ein "Ausbruch" möglich ist
func (r *Renderer) PutFile(subPath string, content []byte) error {
	// Bereinige und baue den vollständigen Pfad
	safePath := filepath.Join(r.renderDir, filepath.Clean(subPath))

	// Stelle sicher, dass der Pfad innerhalb von renderDir bleibt
	if !filepath.HasPrefix(safePath, r.renderDir) {
		return errors.New("Pfad-Ausbruch erkannt: Zugriff auf renderDir verweigert")
	}

	// Erstelle das Verzeichnis, falls nicht vorhanden
	if err := os.MkdirAll(filepath.Dir(safePath), os.ModePerm); err != nil {
		return fmt.Errorf("Fehler beim Erstellen des Verzeichnisses: %w", err)
	}

	// Schreibe die Datei
	if err := os.WriteFile(safePath, content, 0644); err != nil {
		return fmt.Errorf("Fehler beim Schreiben der Datei: %w", err)
	}

	return nil
}

func (r *Renderer) deleteTempRenderDirectory() error {
	return os.RemoveAll(r.renderDir)
}

func (r *Renderer) texFileBase() string {
	tmpFileBase := filepath.Base(r.texFile.Name())
	return tmpFileBase
}

func (r *Renderer) MakeFile(outFile string) error {
	_, err := r.Render()
	if r.PdfDoesExist() {
		err = r.RenamePdfFile(outFile)

	}
	if err == nil {
		return r.deleteTempRenderDirectory()

	} else {
		return errors.Join(errors.New(fmt.Sprintf("failed to move pdf, temp files kept at %s", r.renderDir)), err)
	}

}

func (r *Renderer) Render() ([]byte, error) {
	command := r.pdfLaTeX()
	err := command.Run()
	if err != nil {
		return nil, err
	}

	if viper.GetString("latex.command") != "pdflatexmk" {
		// twice
		command = r.pdfLaTeX()
		err = command.Run()
		if err != nil {
			return nil, err
		}
	}
	file, err := os.ReadFile(r.getTempPdfPath())
	if err != nil {
		return nil, err
	}
	return file, err
}

func (r *Renderer) pdfLaTeX() *exec.Cmd {
	viper.SetDefault("latex.command", "pdflatex")
	timeout, _ := context.WithTimeout(context.Background(), 30*time.Second)
	command := exec.CommandContext(
		timeout,
		viper.GetString("latex.command"), "-synctex=1", "-interaction=batchmode", filepath.Base(r.MainTexFile()),
	)
	command.Stdout = r.TexConsoleWriter
	command.Dir = r.renderDir
	return command
}
