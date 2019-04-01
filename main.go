package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func printLineNumbered(s string) {
	lines := strings.Split(s, "\n")
	for i, v := range lines {
		fmt.Printf("%d\t%s\n", i+1, v)
	}
}

func formatGoCode(in io.Reader, out io.Writer) error {
	cmd := exec.Command("gofmt")
	cmd.Stdin = in
	cmd.Stdout = out
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func main() {
	fname := os.Args[1]
	fi, err := os.Open(fname)
	if err != nil {
		panic(err)
	}

	outputBuffer := &bytes.Buffer{}
	codebuffer := &bytes.Buffer{}
	codeLang := ""
	var inCodeBlock bool
	scan := bufio.NewScanner(fi)
	for scan.Scan() {
		if strings.HasPrefix(scan.Text(), "```") {
			if inCodeBlock {
				if codeLang == "go" {
					str := codebuffer.String()
					if err := formatGoCode(codebuffer, outputBuffer); err != nil {
						fmt.Println("The following code block has formatting errors:")
						fmt.Println("----------------------------------------")
						printLineNumbered(str)
						fmt.Println("----------------------------------------")
						fmt.Println("Please fix these issues and rerun mdfmt")
						return
					}
				} else {
					codebuffer.WriteTo(outputBuffer)
				}
				fmt.Fprintln(outputBuffer, "```")
				inCodeBlock = false
				codebuffer.Reset()
				continue
			}

			codeLang = strings.Trim(scan.Text(), " `\n")
			inCodeBlock = true
			fmt.Fprintln(outputBuffer, "```"+codeLang)
			continue
		}

		if inCodeBlock {
			codebuffer.Write(scan.Bytes())
			codebuffer.WriteByte('\n')
			continue
		}

		fmt.Fprintln(outputBuffer, scan.Text())
	}

	fi.Close()

	outfi, err := os.Create(fname)
	if err != nil {
		panic(err)
	}
	outputBuffer.WriteTo(outfi)
	outfi.Close()
}
