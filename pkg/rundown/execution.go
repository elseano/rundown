package rundown

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/charmbracelet/glamour/ansi"
	rdexec "github.com/elseano/rundown/pkg/exec"
	"github.com/elseano/rundown/pkg/markdown"
	"github.com/elseano/rundown/pkg/util"
	"github.com/logrusorgru/aurora"
)

func rpcLoop(messages chan string, spinner util.Spinner, context *Context, logger *log.Logger, rpcDoneChan chan struct{}, rpcDoneCommand string) {
	for {
		var incoming = <-messages
		var splitLine = strings.SplitN(incoming, " ", 2)
		var cmd = splitLine[0]
		var args = ""

		if len(splitLine) > 1 {
			args = splitLine[1]
		}

		switch strings.ToLower(cmd) {
		case "name:":
			logger.Printf("[RPC] Set Spinner Message: %s\n", args)
			spinner.SetMessage(args)
			break
		case "env:":
			logger.Printf("[RPC] Set Environment: %s\n", args)
			context.SetEnvString(args)
			break
		case "envdiff:":
			logger.Printf("[RPC] Set all new environment variables\n")
			var existing = map[string]bool{"RUNDOWN": true}

			for _, env := range os.Environ() {
				splitEnv := strings.SplitN(env, "=", 2)
				existing[splitEnv[0]] = true
			}

			logger.Printf("[RPC] Got existing envs\n")

			for line := range messages {
				logger.Printf("[RPC] Got env %s\n", line)
				if line == ":done" {
					break
				}
				args := strings.SplitN(line, "=", 2)

				if existing[args[0]] {
					continue
				} // Ignore existing environment variables
				if len(args) < 2 {
					logger.Printf("[RPC] Garbage: %s\n", args[0])
					continue
				}

				logger.Printf("[RPC] Environments to set: %v\n", args)
				context.SetEnv(args[0], args[1])
			}

		case strings.ToLower(rpcDoneCommand):
			logger.Printf("[RPC] Got RPC DONE\n")
			rpcDoneChan <- struct{}{}
			return

		default:
			logger.Printf("[RPC] Ignoring unknown %s", incoming)

		}

	}
}

func saveContentsToTemp(context *Context, contents string, filenamePreference string) string {
	if tmpFile, err := ioutil.TempFile(context.TempDir, "saved-*-"+filenamePreference); err == nil {
		tmpFile.WriteString(contents)
		tmpFile.Close()

		return tmpFile.Name()
	} else {
		panic(err)
	}
}

/**
* Modifiers:
* 	nospin - No Spinner
*		interactive - Support STDIN
*		skip_on_success - Skip to next heading when code executes successfully
*		skip_on_failure - Skip to next...
*   stdout - Display stdout
*   stderr - Display stderr
*   reveal - Render the code block before running it
*   norun - Show the code, don't run it
*   named - The first line is a comment marker, space, then text for the spinner
*   abort - Implies named and stdout. Displays a named execution failure, and exits without showing the script.
 */

func Execute(context *Context, executionBlock *markdown.ExecutionBlock, source []byte, logger *log.Logger, out io.Writer) ExecutionResult {
	var spinner util.Spinner = util.NewDummySpinner()
	var spinnerName = "Running"
	var rpcDoneCommand = "RD-DDDX"
	var doneChannel = make(chan struct{})
	var modifiers = executionBlock.Modifiers
	var contentSwitchback = ""

	indent := 2

	// if section, ok := executionBlock.Parent().(*markdown.Section); ok {
	// 	indent = section.Level
	// }

	content := util.NodeLines(executionBlock, source)

	logger.Printf("Block mods: %s\n", modifiers)

	if save, ok := modifiers.Values[SaveParameter]; ok {
		content := util.NodeLines(executionBlock, source)

		if modifiers.Flags[EnvAwareFlag] == true {
			content, _ = SubEnv(content, context)
		}
		path := saveContentsToTemp(context, content, save)
		varName := strings.Split(save, ".")[0]
		context.SetEnv(strings.ToUpper(varName), path)
		modifiers.Flags[NoRunFlag] = true
	}

	if modifiers.HasAny("named", "spinner") && !modifiers.Flags[NoSpinFlag] {
		logger.Println("Block has a custom spinner title")

		if modifiers.Flags[NamedFlag] {
			firstLine := strings.Split(strings.TrimSpace(content), "\n")[0]
			matcher := regexp.MustCompile(`\s*.{1,2}\s+(.*)`)
			matches := matcher.FindStringSubmatch(firstLine)

			if len(matches) > 1 {
				logger.Printf("Spinner title extracted from block")
				spinnerName = matches[1]
			} else {
				logger.Println("No title detected")
			}
		} else if name, ok := modifiers.Values[markdown.Parameter("spinner")]; ok {
			logger.Printf("Spinner title defined")
			spinnerName = name
		}
	}

	if modifiers.Flags[NoSpinFlag] != true {
		logger.Printf("Block requires spinner at indent: %d, title: %s", int(math.Max(0, float64(indent-1))), spinnerName)
		spinner = util.NewSpinner(int(math.Max(0, float64(indent-1))), spinnerName, out)
	}

	spinner.Start()

	go rpcLoop(context.Messages, spinner, context, logger, doneChannel, rpcDoneCommand)

	var tmpFile *os.File
	var err interface{}

	if tmpFile, err = ioutil.TempFile(context.TempDir, "rundown-exec-*"); err != nil {
		panic(err)
	}
	defer tmpFile.Close()
	filename := tmpFile.Name()

	logger.Printf("Writing script into %s for execution\n", filename)

	executable := executionBlock.Syntax

	if prog, ok := modifiers.Values[WithParameter]; ok {
		executable = prog
	}

	if strings.Contains(executable, "$FILE") {
		context.SetEnv("FILE", saveContentsToTemp(context, content, "source"))
		util.Debugf("Moving script into secondary file\n")

		contentSwitchback = content
		content = executable
		executable = "bash"
	}

	contents := "#!/usr/bin/env " + executable + "\n\n"

	if executable == "bash" {
		contents = contents + "set -Eeuo pipefail\n\n"
	}

	// Convert every comment into a RPC call to set heading.
	if modifiers.Flags[NamedAllFlag] {
		matcher := regexp.MustCompile(`^[\/\#]{1,2}\s+(.*)$`)
		for _, line := range strings.Split(content, "\n") {
			commentLine := matcher.ReplaceAllString(line, "echo \"Name: $1\" >> $$RUNDOWN")
			contents = contents + commentLine + "\n"
		}
	} else {
		contents = contents + content
	}

	if modifiers.Flags[CaptureEnvFlag] && executable == "bash" {
		contents = contents + "\necho \"envdiff:\" >> $RUNDOWN\nenv >> $RUNDOWN\necho \":done\" >> $RUNDOWN"
	}

	if _, err := tmpFile.Write([]byte(contents)); err != nil {
		panic(err)
	}

	util.Debugf("Script is:\n\n%s\n\n", contents)

	os.Chmod(filename, 0700)
	tmpFile.Close()

	if modifiers.Flags[BorgFlag] {
		var tmpFileRepeat *os.File

		if context.Repeat {
			// Wrap the script in a script to relaunch

			if tmpFileRepeat, err = ioutil.TempFile(context.TempDir, "rundown-repeat-*"); err != nil {
				panic(err)
			}
			defer tmpFileRepeat.Close()

			repeatContents := "#!/bin/sh\n" + filename + "\n" + context.Invocation + "\n"

			logger.Printf("Repeat file: \n %s \n", repeatContents)

			tmpFileRepeat.Write([]byte(repeatContents))

			filename = tmpFileRepeat.Name()
			os.Chmod(filename, 0700)
		}

		execErr := syscall.Exec(filename, []string{}, context.EnvStringList())
		if execErr != nil {
			return ExecutionResult{
				Message: execErr.Error(),
				Kind:    "Error",
				Source:  contents,
				Output:  "",
				IsError: true,
			}
		}
	}

	cmd := exec.Command(filename)

	// logger.Printf("Execution Command: %v\n", cmd.

	cmd.Env = os.Environ()

	for key, value := range context.Env {
		cmd.Env = append(cmd.Env, key+"="+value)
	}

	util.Debugf("Script:\r\n%s\r\nEnv:%v\r\n", contents, cmd.Env)

	if contentSwitchback != "" {
		contents = contentSwitchback
	}

	if modifiers.Flags[NoRunFlag] != true {
		process := rdexec.NewProcess(cmd)

		stdout, err := process.Start()
		if err != nil {
			spinner.Error("Error")
			return ExecutionResult{Message: fmt.Sprintf("%v", err), Kind: "Error", Source: contents, IsError: true}
		}

		var captureBuffer bytes.Buffer
		endedWithoutNewline := false
		var output io.Writer = nil
		var waiter sync.WaitGroup

		if modifiers.Flags[StdoutFlag] {
			if _, ok := executionBlock.PreviousSibling().(*markdown.ExecutionBlock); ok {
				out.Write([]byte("\r\n"))
			}

			stdoutReader, stdoutWriter, err := os.Pipe()
			if err != nil {
				spinner.Error("Error")
				return ExecutionResult{Message: fmt.Sprintf("%v", err), Kind: "Error", Source: contents, IsError: true}
			}

			stdoutDist := io.MultiWriter(stdoutWriter, &captureBuffer)

			go func() {
				defer stdoutWriter.Close()
				_, _ = io.Copy(stdoutDist, stdout)
			}()

			logger.Println("Displaying process output\n\rLevel is ", indent, "\n\r")
			output = out

			logger.Println("Setting up output formatter...")

			outputHeading := spinnerName
			if outputHeading == "Running" {
				outputHeading = "Output"
			}

			go func() {
				waiter.Add(1)

				prefix := ansi.Ssprintf(context.Profile, context.Style.Heading.StylePrimitive, "‣ ")
				endedWithoutNewline = util.ReadAndFormatOutput(stdoutReader, indent, prefix, spinner, bufio.NewWriter(output), logger, aurora.Faint(outputHeading).String())
				logger.Printf("endedWithoutNewline? %v\r\n", endedWithoutNewline)
				waiter.Done()
			}()
		} else {
			output = ioutil.Discard
			go func() {
				waiter.Add(1)
				_, _ = io.Copy(&captureBuffer, stdout)
				waiter.Done()
			}()
		}

		logger.Println("Waiting for command completion")

		waitErr := process.Wait()

		waiter.Wait()

		time.Sleep(100 * time.Millisecond) // 2x receiveLoop delay

		os.Chmod(filename, 0644)

		// Wait for pending RPC commands.
		context.Messages <- rpcDoneCommand
		<-doneChannel

		if modifiers.Flags[StopOkFlag] {
			spinner.Success("Complete")
		} else if modifiers.Flags[StopFailFlag] {
			spinner.Error("Failed")
		} else {
			if ex, ok := waitErr.(*exec.ExitError); ok {
				logger.Printf("Error condition detected. Err: %v, SOF: %v, SOS: %v\n", ex, modifiers.Flags[SkipOnFailureFlag], modifiers.Flags[SkipOnSuccessFlag])
				if modifiers.Flags[SkipOnFailureFlag] {
					spinner.Skip("Passed")
				} else if modifiers.Flags[SkipOnSuccessFlag] {
					spinner.Success("Required")
				} else if modifiers.Flags[IgnoreFailureFlag] {
					spinner.Error("Ignoring Failure")
				} else {
					spinner.Error("Failed")
				}
			} else {
				if modifiers.Flags[SkipOnSuccessFlag] {
					spinner.Skip("Passed")
				} else if modifiers.Flags[SkipOnFailureFlag] {
					spinner.Success("Required")
				} else {
					spinner.Success("Complete")
				}
			}
		}

		if we, ok := waitErr.(*exec.ExitError); ok {
			if modifiers.Flags[SkipOnFailureFlag] {
				return SkipToNextHeading
			} else if modifiers.Flags[SkipOnSuccessFlag] || modifiers.Flags[IgnoreFailureFlag] {
				return SuccessfulExecution
			} else {
				fl := -1

				if f, ok := DetectErrorLine(filename, captureBuffer.String()); ok {
					fl = f
				}

				return ExecutionResult{
					Message:   we.String(),
					Kind:      "Error",
					Source:    contents,
					Output:    strings.ReplaceAll(captureBuffer.String(), filename, "SCRIPT"),
					FocusLine: fl,
					IsError:   true,
				}
			}
		} else {
			if modifiers.Flags[SkipOnSuccessFlag] {
				return SkipToNextHeading
			}

			if modifiers.Flags[StopOkFlag] {
				return StopOkResult
			}

			if modifiers.Flags[StopFailFlag] {
				return StopFailResult
			}

			return SuccessfulExecution
		}
	}

	return SuccessfulExecution // NORUN = true
}
