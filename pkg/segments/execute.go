package segments

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	// "bufio"
	// "time"
	"bytes"
	// "context"
	"bufio"
	"syscall"

	// "strings"

	"github.com/logrusorgru/aurora"
	"github.com/yuin/goldmark/renderer"
)

func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}

func ScanFullLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, dropCR(data[0:i]), nil
	}

	// Request more data.
	return 0, nil, nil
}

func receiveLoop(filename string, messages chan<- string, logger *log.Logger) {
	logger.Printf("Setting up receive loop\r\n")

	os.Remove(filename)
	err := syscall.Mkfifo(filename, 0666)
	if err != nil {
		logger.Printf("Error openeing receive loop %v\n", err)
		return
	}

	// RDWR so it doesn't block on opening.
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, os.ModeNamedPipe)
	if err != nil {
		logger.Printf("Error openeing receive loop %v\n", err)
		return
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			logger.Printf("[RPC] Error reading bytes %#v", err)
			return
		}

		logger.Printf("[RPC] Got line")
		messages <- string(bytes.TrimRight(line, "\r\n"))
	}
}

func ExecuteRundown(context *Context, rundown []Segment, renderer renderer.Renderer, logger *log.Logger, out io.Writer) ExecutionResult {
	// Using a buffered channel allows us to capture all messages correctly. Unsure how to test that reliably though.
	messages := make(chan string, 200)

	tmpDir, err := ioutil.TempDir("", "rundown")
	if err != nil {
		panic(err)
	}
	os.MkdirAll(tmpDir, 0644)

	tmpFile, err := ioutil.TempFile(tmpDir, "rpc-*")
	if err != nil {
		panic(err)
	}
	tmpFile.Close()

	logger.Printf("Created rundown RPC file at %s", tmpFile.Name())

	context.Messages = messages
	context.TempDir = tmpDir
	context.ForcedLevelZero = false
	context.Env["RUNDOWN"] = tmpFile.Name()

	go receiveLoop(tmpFile.Name(), messages, logger)

	// defer os.Remove(tmpFile.Name())

	var skipToHeading = false
	var lastSegment Segment
	var currentHeading *HeadingMarker = nil

	for _, segment := range rundown {
		if skipToHeading {
			if _, ok := segment.(*HeadingMarker); !ok {
				continue
			}

			out.Write([]byte("\r\n")) // Add spacing between skipping code block and next heading.
			skipToHeading = false
		}

		var result ExecutionResult

		// Ensure all rerequisites have been run, when running via shortcodes.
		if segment.Kind() == "HeadingMarker" {
			headingMarker := segment.(*HeadingMarker)
			currentHeading = headingMarker

			if headingMarker.ParentHeading != nil {
				// Only run the parent pre-reqs, as the ones at the current level will be run
				// as part of the current loop.
				var count = 0
				context.ForcedLevelZero = true
				result, count = headingMarker.ParentHeading.RunSetups(context, renderer, lastSegment, logger, out)
				context.ForcedLevelZero = false

				if count > 0 {
					fmt.Fprintf(out, "\r\n") // Blank line between setups and the we're running heading.
				}
			}
		}

		if !result.IsError {
			result = segment.Execute(context, renderer, lastSegment, logger, out)
		}

		lastSegment = segment

		if result == SkipToNextHeading {
			logger.Printf("Block returned SkipToNextHeading")
			skipToHeading = true
		} else if result == SuccessfulExecution {
			logger.Printf("Block returned SuccessfulExecution")
		} else if result == StopOkResult || result == StopFailResult {
			os.RemoveAll(tmpDir)
			return result
		} else { // Error
			if currentHeading != nil && len(currentHeading.Handlers) > 0 {
				context.SetEnv("RUNDOWN_ERROR", result.Output)
				fmt.Fprintf(out, "\n")
				currentHeading.RunHandlers(result.Output, context, renderer, lastSegment, logger, out)
				fmt.Fprintf(out, "\n%s\n\n", result.Message)
				fmt.Fprintf(out, "%s %s\n\n", aurora.Red("✖"), "Aborted due to failure.")
			} else {
				logger.Printf("Block returned FailedExecution")
				fmt.Fprintf(out, "\n\n%s\n\n%s\n\nError: %s\n", aurora.Bold("Error executing script:"), aurora.Faint(result.Source), result.Message)
				fmt.Fprintf(out, "\n%s\n", result.Output)
				fmt.Fprintf(out, "%s %s\n\n", aurora.Red("✖"), "Aborted due to failure.")
			}

			return result
		}
	}

	os.RemoveAll(tmpDir)
	return SuccessfulExecution
}
