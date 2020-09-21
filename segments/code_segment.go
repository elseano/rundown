package segments

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"
	"math"
	"bufio"

	"github.com/elseano/rundown/util"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/creack/pty"
	"github.com/logrusorgru/aurora"
	"github.com/yuin/goldmark/renderer"
)

type CodeSegment struct {
	BaseSegment
	code      string
	Modifiers *Modifiers
	language  string
}

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
			var existing = map[string]bool{ "RUNDOWN": true }
			
			for _, env := range os.Environ() {
				splitEnv := strings.SplitN(env, "=", 2)
				existing[splitEnv[0]] = true
			}

			logger.Printf("[RPC] Got existing envs\n")

			for line := range messages {
				logger.Printf("[RPC] Got env %s\n", line)
				if line == ":done" { break }
				args := strings.SplitN(line, "=", 2)

				if existing[args[0]] { continue } // Ignore existing environment variables
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

// Takes carriage returns from the input, and passes them through to the output.
// Needed for formatting issues with accepting input from STDIN while showing STDOUT.
func inputHandler(reader *os.File, writer *os.File, logger *log.Logger) {
	var buf = make([]byte, 216)

	for {
		if n, err := reader.Read(buf); err == nil {
			logger.Printf("Got input: %d %q\n", n, buf[0:n])				
			newlines := bytes.Count(buf[0:n], []byte{ 13 })
			if newlines > 0 {
				logger.Printf("Found newlines: %d", newlines)
				writer.Write([]byte("\r\n"))
			}
		}
	}
}


func (c *CodeSegment) Kind() string { return "CodeSegment" }

func (s *CodeSegment) displayBlock(source []byte, context *Context, renderer renderer.Renderer, lastSegment *Segment, logger *log.Logger, out io.Writer) {
	// if lastSegment != nil {

	// 	if lastCode, ok := CodeSegment(*lastSegment); ok {
	// 		if lastCode.mods[Reveal] != true && lastCode.mods[NoRun] != true {
	// 			out.write([]byte("\r\n"))
	// 		}
	// 	}
	// }

	seg := &DisplaySegment{
		BaseSegment{
			Indent: s.Indent,
			Nodes:  s.Nodes,
			Source: &source,
		},
	}
	
	seg.Execute(context, renderer, lastSegment, logger, out)
}

func (s *CodeSegment) String() string {
	var buf bytes.Buffer

	buf.WriteString("CodeSegment {\n")
	buf.WriteString(fmt.Sprintf("  Mods: %s", s.Modifiers))
	out := util.CaptureStdout(func() { 
		for _, n := range s.Nodes {
			n.Dump(*s.Source, s.Indent)
		}
	})
	buf.WriteString(out)
	buf.WriteString("}\n")

	return buf.String()
}

type SafeWriter struct {
	io.Writer
	spinner util.Spinner
}

func (w SafeWriter) Write(p []byte) (n int, err error) {
	if w.spinner != nil {
		w.spinner.HideAndExecute(func() {
			n, err = w.Writer.Write(p)
		})
	} else {
		n, err = w.Writer.Write(p)
	}

	return n, err
}

func saveContentsToTemp(context *Context, contents string, filenamePreference string) string {
	if tmpFile, err := ioutil.TempFile(context.TempDir, "saved-*-" + filenamePreference); err == nil {
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

func (s *CodeSegment) Execute(context *Context, renderer renderer.Renderer, lastSegment *Segment, loggerInner *log.Logger, out io.Writer) ExecutionResult {
	var spinner util.Spinner = util.NewDummySpinner()
	var spinnerName = "Running"
	var rpcDoneCommand = "RD-DDDX"
	var doneChannel = make(chan struct{})

	writer := SafeWriter{Writer: loggerInner.Writer(), spinner: spinner}

	logger := log.New(writer, loggerInner.Prefix(), loggerInner.Flags())

	indent := s.Indent

	if context.ForcedIndentZero {
		indent = 0
	}

	logger.Printf("Block mods: %s\n", s.Modifiers)

	if strings.TrimSpace(s.language) == "" {
		s.Modifiers.Flags[NoRunFlag] = true
		s.Modifiers.Flags[RevealFlag] = true
	}

	if save, ok := s.Modifiers.Values[SaveParameter]; ok {
		content := s.code

		if s.Modifiers.Flags[EnvAwareFlag] == true {
			for k,v := range context.Env {
				content = strings.ReplaceAll(content, "$" + k, v)
			}
		}
		path := saveContentsToTemp(context, content, save)
		varName := strings.Split(save, ".")[0]
		context.SetEnv(strings.ToUpper(varName), path)
		s.Modifiers.Flags[NoRunFlag] = true
	}

	if s.Modifiers.Flags[AbortFlag] {
		s.Modifiers.Flags[NamedFlag] = true
		s.Modifiers.Flags[StdoutFlag] = true
	}

	if s.Modifiers.Flags[NoRunFlag] {
		logger.Println("Block is NORUN")
		if !s.Modifiers.Flags[RevealFlag] {
			logger.Println("Block is noop")
			return SuccessfulExecution
		} else {
			logger.Println("Block is REVEAL. Rendering.")
			var outCap bytes.Buffer
			s.displayBlock(*s.Source, context, renderer, lastSegment, logger, &outCap)

			content := outCap.String()

			if s.Modifiers.Flags[EnvAwareFlag] == true {
				logger.Println("Block is ENV_AWARE. Substituting environment")

				envMatch := regexp.MustCompile("(\\$[A-Z0-9_]+)")

				for k,v := range context.Env {
					content = strings.ReplaceAll(content, "$" + k, v)
				}

				if match := envMatch.FindString(content); match != "" {
					return ExecutionResult{ 
						Message: match + " is not set",
						Kind: "Error", 
						Source: content, 
						Output: "",
						IsError: true,
					}
				}
			}
			
			out.Write([]byte(content))
			// out.Write([]byte("\r\n")) // Add some space between code output and next block.

			return SuccessfulExecution
		}
	}

	if s.Modifiers.Flags[NamedFlag] && !s.Modifiers.Flags[NoSpinFlag] {
		logger.Println("Block is NAMED")
		firstLine := strings.Split(strings.TrimSpace(s.code), "\n")[0]
		matcher := regexp.MustCompile(`\s*.{1,2}\s+(.*)`)
		matches := matcher.FindStringSubmatch(firstLine)

		if len(matches) > 1 {
			logger.Printf("Name is %s", matches[1])
			spinnerName = matches[1]
		} else {
			logger.Println("No name detected")
		}
	}

	if s.Modifiers.Flags[NoSpinFlag] != true {
		logger.Printf("Block requires spinner at indent: %d, title: %s", int(math.Max(0, float64(indent-1))), spinnerName)
		spinner = util.NewSpinner(int(math.Max(0, float64(indent-1))), spinnerName, out)
		writer.spinner = spinner
	}

	spinner.Start()

	go rpcLoop(context.Messages, spinner, context, logger, doneChannel, rpcDoneCommand)

	// fmt.Print("\033[s")

	var tmpFile *os.File
	var err interface{}

	if tmpFile, err = ioutil.TempFile(context.TempDir, "rundown-exec-*"); err != nil {
		panic(err)
	}
	defer tmpFile.Close()
	filename := tmpFile.Name()

	logger.Printf("Writing script into %s for execution\n", filename)

	executable := s.language

	if prog, ok := s.Modifiers.Values[WithParameter]; ok {
		executable = prog
	}

	contents := "#!/usr/bin/env " + executable + "\n\n"

	if executable == "bash" {
		contents = contents + "set -Eeuo pipefail\n\n"
	}

	// Convert every comment into a RPC call to set heading.
	if s.Modifiers.Flags[NamedAllFlag] {
		matcher := regexp.MustCompile(`^[\/\#]{1,2}\s+(.*)$`)
		for _, line := range strings.Split(s.code, "\n") {
			commentLine := matcher.ReplaceAllString(line, "echo \"Name: $1\" >> $$RUNDOWN")
			contents = contents + commentLine + "\n"
		}
	} else {
		contents = contents + s.code
	}

	if s.Modifiers.Flags[CaptureEnvFlag] {
		contents = contents + "\necho \"envdiff:\" >> $RUNDOWN\nenv >> $RUNDOWN\necho \":done\" >> $RUNDOWN"
	}


	if _, err := tmpFile.Write([]byte(contents)); err != nil {
		panic(err)
	}

	cmd := exec.Command(filename)
	cmd.Env = os.Environ()

	for key, value := range context.Env {
		cmd.Env = append(cmd.Env, key+"="+value)
	}

	logger.Printf("Script:\r\n%s\r\nEnv:%v\r\n", contents, cmd.Env)

	if s.Modifiers.Flags[RevealFlag] {
		spinner.HideAndExecute(func() {
			s.displayBlock(*s.Source, context, renderer, lastSegment, logger, out)
			// out.Write([]byte("\r\n")) // Add some space between code output and spinner.
		})
	}

	os.Chmod(filename, 0700)

	if s.Modifiers.Flags[NoRunFlag] != true {

		ptmx, err := pty.Start(cmd)
		if err != nil {
			spinner.Error("Error")
			return ExecutionResult{ Message: fmt.Sprintf("%v", err), Kind: "Error", Source: contents, IsError: true }
		}
		defer func() { _ = ptmx.Close() }() // Best effort.

		var captureBuffer bytes.Buffer
		endedWithoutNewline := false
		var output io.Writer = nil


		if s.Modifiers.Flags[StdoutFlag] {
			output = os.Stdout

			ch := make(chan os.Signal, 1)
			signal.Notify(ch, syscall.SIGWINCH)
			go func() {
				for range ch {
					if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
						// log.Printf("error resizing pty: %s", err) // Don't care.
					}
				}
			}()
			ch <- syscall.SIGWINCH // Initial resize.

			oldState, err := terminal.MakeRaw(int(os.Stdin.Fd()))
			if err != nil {
				// Don't care.
			}
			defer func() { if(oldState != nil) { terminal.Restore(int(os.Stdin.Fd()), oldState) } }() // Best effort.

			logger.Println("Setting up STDIN...")

			// Copy stdin to our newline capturer for inserting blank lines when password input is requested.
			piper, pipew, err := os.Pipe()
			if err != nil { panic(err) }
			mw := io.MultiWriter(pipew, ptmx)
			go func() { _, _ = io.Copy(mw, os.Stdin) }()
			// go func() { _, _ = io.Copy(ptmx, mw) }()
			go func() { inputHandler(piper, ptmx, logger) }()

			outr, outw, err := os.Pipe()
			if err != nil { panic(err) }

			logger.Println("Setting up STDOUT...")

			// Setup a distribution of STDOUT/STDERR into outw (which will be used to display output) and captureBuffer
			// which is used for error reporting.
			ptmxMulti := io.MultiWriter(outw, &captureBuffer)
			go func() {
				defer outw.Close()
				_, _ = io.Copy(ptmxMulti, ptmx) 
			}()

			logger.Println("Displaying process output\n\rIndent is ", indent, "\n\r")
			output = out

			logger.Println("Setting up output formatter...")

			outputHeading := spinnerName
			if outputHeading == "Running" {
				outputHeading = "Output"
			}


			if s.Modifiers.Flags[StdoutFlag] {
				endedWithoutNewline = util.ReadAndFormatOutput(outr, indent, aurora.Blue("‣ ").Bold().Faint().String(), spinner, bufio.NewWriter(output), logger, aurora.Faint(outputHeading).String())
				logger.Printf("endedWithoutNewline? %v\r\n", endedWithoutNewline)
			}
		} else {
			output = ioutil.Discard
			go func() {
				_, _ = io.Copy(&captureBuffer, ptmx) 
			}()
		}

		logger.Println("Waiting for command completion")

		waitErr := cmd.Wait()

		time.Sleep(100 * time.Millisecond)	// 2x receiveLoop delay

		os.Chmod(filename, 0644)

		// Wait for pending RPC commands.
		context.Messages <- rpcDoneCommand
		<- doneChannel

		if s.Modifiers.Flags[AbortFlag] {
			spinner.Error("Aborted")
		} else {
			if ex, ok := waitErr.(*exec.ExitError); ok {
				logger.Printf("Error condition detected. Err: %v, SOF: %v, SOS: %v\n", ex, s.Modifiers.Flags[SkipOnFailureFlag], s.Modifiers.Flags[SkipOnSuccessFlag])
				if s.Modifiers.Flags[SkipOnFailureFlag] {
					spinner.Skip("Passed")
				} else if s.Modifiers.Flags[SkipOnSuccessFlag] {
					spinner.Success("Required")
				} else if s.Modifiers.Flags[IgnoreFailureFlag] {
					spinner.Error("Ignoring Failure")
				} else {
					spinner.Error("Failed")
				}
			} else {
				if s.Modifiers.Flags[SkipOnSuccessFlag] {
					spinner.Skip("Passed")
				} else if s.Modifiers.Flags[SkipOnFailureFlag] {
					spinner.Success("Required")
				} else {
					spinner.Success("Complete")
				}
			}
		}
		
		if !endedWithoutNewline {
			logger.Println("Injecting newline")
			fmt.Fprint(output, "\r\n")
		}

		if we, ok := waitErr.(*exec.ExitError); ok {
			if s.Modifiers.Flags[SkipOnFailureFlag] {
				return SkipToNextHeading
			} else if s.Modifiers.Flags[SkipOnSuccessFlag] || s.Modifiers.Flags[IgnoreFailureFlag] {
				return SuccessfulExecution
			} else {
				return ExecutionResult{ 
					Message: we.String(), 
					Kind: "Error", 
					Source: contents, 
					Output: strings.ReplaceAll(captureBuffer.String(), filename, "SCRIPT"),
					IsError: true,
				}
			}
		} else {
			if s.Modifiers.Flags[SkipOnSuccessFlag] {
				return SkipToNextHeading
			}
			
			if s.Modifiers.Flags[AbortFlag] {
				return AbortedExecution
			}
			
			return SuccessfulExecution
		}
	}

	return SuccessfulExecution  // NORUN = true
}
