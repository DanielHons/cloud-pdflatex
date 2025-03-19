package cloudpdf

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"testing"
)

// MockClient simuliert den Client für Tests
// Er sollte Methoden wie PostConvertWithBody enthalten
// Die Implementierung hängt von deinem eigentlichen Client ab

type MockClient struct{}

func (c *MockClient) GetHeartbeat(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	//TODO implement me
	panic("implement me")
}

func (c *MockClient) PostConvertWithBody(
	ctx context.Context, params *PostConvertParams, contentType string, body io.Reader, reqEditors ...RequestEditorFn,
) (*http.Response, error) {
	return &http.Response{StatusCode: 200, Body: io.NopCloser(body)}, nil
}

type Response struct {
	StatusCode int
	Body       io.ReadCloser
}

func TestArchiver_AddFile(t *testing.T) {
	archiver := NewArchiver()
	tmpFile, err := os.CreateTemp("", "testfile-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte("Hello, Test!")); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	if err := archiver.AddFile(tmpFile.Name()); err != nil {
		t.Errorf("AddFile failed: %v", err)
	}
}

func TestArchiver_AddReader(t *testing.T) {
	archiver := NewArchiver()
	reader := bytes.NewReader([]byte("Hello, Archive!"))

	if err := archiver.AddReader("testfile.txt", reader); err != nil {
		t.Errorf("AddReader failed: %v", err)
	}
}

func TestArchiver_RenderPDF(t *testing.T) {
	archiver := NewArchiver()
	client := &MockClient{}
	ctx := context.Background()

	data, err := archiver.RenderPDF(ctx, client)
	if err != nil {
		t.Errorf("RenderPDF failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("RenderPDF returned empty data")
	}
}

func TestArchiver_Close(t *testing.T) {
	archiver := NewArchiver()
	if err := archiver.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}
