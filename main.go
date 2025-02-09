//go:generate go run -modfile=tools/go.mod github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=generator/oapi/cfg.yaml cloud-pdflatex.openapi.yaml
package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/DanielHons/cloud-pdflatex/internal/latex"
	"github.com/DanielHons/cloud-pdflatex/internal/xgzip"
	"github.com/DanielHons/cloud-pdflatex/pkg/cloudpdf"
	"github.com/ory/graceful"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
)

type serverImpl struct {
}

func (s serverImpl) GetHeartbeat(
	ctx context.Context, request cloudpdf.GetHeartbeatRequestObject,
) (cloudpdf.GetHeartbeatResponseObject, error) {

	return cloudpdf.GetHeartbeat200Response{}, nil
}

func (s serverImpl) PostConvert(
	ctx context.Context, request cloudpdf.PostConvertRequestObject,
) (cloudpdf.PostConvertResponseObject, error) {

	temp, err := os.MkdirTemp("/tmp", "pdfzip")
	if err != nil {
		return cloudpdf.PostConvert500JSONResponse{"Fehler beim Anlegen des temporären Ordners"}, nil
	}

	err = xgzip.UnGzipIntoFolder(request.Body, temp)

	renderer, err := latex.NewFolderRenderer(temp, *request.Params.TexEntrypoint)
	if err != nil {
		return cloudpdf.PostConvert500JSONResponse{"Fehler beim Erstellen des LaTeX Renderers"}, nil
	}
	pdf, err := renderer.Render()
	if err != nil {
		return cloudpdf.PostConvert500JSONResponse{"Fehler beim Erstellen PDF"}, nil
	}

	return cloudpdf.PostConvert200ApplicationpdfResponse{
		Body:          bytes.NewReader(pdf),
		ContentLength: int64(len(pdf)),
	}, nil

}

var server cloudpdf.StrictServerInterface = &serverImpl{}

func main() {
	viper.SetDefault("port", "8080")
	viper.AutomaticEnv()
	server := graceful.WithDefaults(
		&http.Server{
			Addr:    fmt.Sprintf(":%s", viper.GetString("port")),
			Handler: cloudpdf.HandlerFromMux(cloudpdf.NewStrictHandler(server, nil), http.NewServeMux()),
		},
	)
	log.Println("main: Starting the server at port", viper.GetString("port"))
	if err := graceful.Graceful(server.ListenAndServe, server.Shutdown); err != nil {
		log.Fatalln("main: Failed to gracefully shutdown: ", err)
	}
	log.Println("main: Server was shutdown gracefully")
}
