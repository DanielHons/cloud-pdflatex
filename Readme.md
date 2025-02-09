# cloud-pdflatex



A WebService tranforming a `.targ.gz` with latex sources to a pdf.

API described at [cloud-pdflatex.openapi.yaml](cloud-pdflatex.openapi.yaml)

### Example usage in Go
```Go
package main

import (
	"context"
	"github.com/DanielHons/cloud-pdflatex/pkg/cloudpdf"
	"io"
	"os"
)

func main() {
	client, err := cloudpdf.NewClient("http://localhost:8081/")
	if err != nil {
		panic(err)
	}

	// should contain a main.tex and additional required resources, like images, lco-files,...
	file, err := os.Open("sample.tar.gz") 
	if err != nil {
		panic(err)
	}
	body, err := client.PostConvertWithBody(context.TODO(), nil, "application/octed-stream", file)
	switch body.StatusCode {

	}
	defer body.Body.Close()

	all, err := io.ReadAll(body.Body)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile("result.pdf", all, 0x660)
}

```


## Create Docker Image
```shell
docker build -t danielhons.de/office/cloud-pdflatex:latest .

```