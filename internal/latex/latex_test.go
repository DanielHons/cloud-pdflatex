package latex

import (
	"os"
	"testing"
)

func TestInvalidLatex(t *testing.T) {
	renderer, err := NewRenderer()
	if err != nil {
		t.Fail()
	}

	_, err = renderer.Write([]byte("4711"))
	if err != nil {
		t.Fail()
	}

	err = renderer.MakeFile("test.pdf")
	if err == nil {
		// must fail, content is no valid tex
		t.Fail()
	}
}

func TestValidLatex(t *testing.T) {
	renderer, err := NewRenderer()
	if err != nil {
		t.Fail()
	}

	_, err = renderer.Write([]byte(minimalValidTexCode))
	if err != nil {
		t.Fail()
	}

	const outFile = "test-pdf.pdf"
	err = renderer.MakeFile(outFile)
	if err != nil {
		// must fail, content is no valid tex
		t.Fatal(err)
	}
	err = os.Remove(outFile)
	if err != nil {
		// must fail, content is no valid tex
		t.Fatal(err)
	}
}

const minimalValidTexCode = `
\documentclass{article}
\usepackage[utf8]{inputenc}
\begin{document}
(Test was created by test generator.)
\end{document}
`
