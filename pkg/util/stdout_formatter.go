package util

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/elseano/rundown/pkg/spinner"
	"github.com/logrusorgru/aurora"
	// "github.com/logrusorgru/aurora"
)

// TODO - Maybe use with https://github.com/Azure/go-ansiterm

type Line struct {
	LastColumn        int
	FormattingStash   string
	CurrentlyIndented bool
	Dirty             bool
}

func (l *Line) MoveCursor(col int) {
	if col == StartOfLine {
		l.LastColumn = 0
		l.CurrentlyIndented = false
	} else {
		l.LastColumn = l.LastColumn + col
		if l.LastColumn < 0 {
			l.LastColumn = 0
		}
	}
}

type TokenResult int

const (
	TokenCursor TokenResult = iota + 1
	TokenColour
	TokenWhitespace
	TokenDisplayStream
	TokenEmpty
)

var formattingMatch = regexp.MustCompile("(\x1b\\[[0-9\\;]*m)|(\x1b\\[[0-9]*[A-Za-ln-z]|[\r\n])|([ \t]+)")

func NextToken(input string) (token TokenResult, data string, rest string) {
	if input == "" {
		return TokenEmpty, "", ""
	}

	i := 0

	match := formattingMatch.FindStringSubmatchIndex(input)

	if len(match) > 0 {
		if match[0] > 0 { // Text at the start
			nextMatch := formattingMatch.FindStringSubmatchIndex(input[i:])

			if len(nextMatch) < 2 { // No formatting
				return TokenDisplayStream, input, ""
			}

			// fmt.Printf("DisplayStream %#v in %q\n", nextMatch, input)

			return TokenDisplayStream, input[0:nextMatch[0]], input[nextMatch[0]:]
		} else if match[4] == 0 { // Cursor
			return TokenCursor, input[match[0]:match[1]], input[match[1]:]
		} else if match[2] == 0 { // Colour
			return TokenColour, input[match[0]:match[1]], input[match[1]:]
		} else if match[6] == 0 { // Whitespace
			return TokenWhitespace, input[match[0]:match[1]], input[match[1]:]
		}
	}

	return TokenDisplayStream, input, ""
}

const StartOfLine = -999

var movementCode = regexp.MustCompile("\x1b\\[([0-9]*)([A-Za-z])")

func DecodeCursor(cursor string) (lineDelta, columnDelta int) {
	switch cursor {
	case "\r":
		return 0, StartOfLine
	case "\n":
		return 1, 0
	case "\r\n":
		return 1, StartOfLine
	}

	if match := movementCode.FindStringSubmatch(cursor); match != nil {
		amount := 0
		if match[1] != "" {
			amount, _ = strconv.Atoi(match[1])
		}
		switch match[2] {
		case "A": // Up
			return amount * -1, 0
		case "B": //Down
			return amount, 0
		case "C": // Forward
			return 0, amount
		case "D": // Backward
			return 0, amount * -1
		}
	}

	return 0, 0
}

type ProgressIndicator interface {
	Start()
	Stop()
	Active() bool
	CurrentHeading() string
}

func moveCursor(out io.Writer, offset int) {
	if offset > 0 {
		Debugf("Moving cursor down: %d\n", offset)
		fmt.Fprintf(out, "\x1b[%dB", offset)
	} else if offset < 0 {
		Debugf("Moving cursor up: %d\n", offset*-1)
		fmt.Fprintf(out, "\x1b[%dA", offset*-1)
	}
}

func ReadAndFormatOutput(reader io.Reader, indent int, prefix string, progressIndicator ProgressIndicator, out *bufio.Writer, logger *log.Logger, initialHeading string) bool {
	if progressIndicator == nil {
		progressIndicator = spinner.NewDummySpinner()
	}

	var indentSpaces = strings.Repeat("  ", int(math.Max(float64(indent-1), 0)))
	var indentStr = indentSpaces + prefix
	var buf = make([]byte, 216)
	var hasShownSomething = false
	var lineTracking = map[int]*Line{0: &Line{}}
	var currentLine = 0
	var maxLine = 0
	var lastToken TokenResult
	var spinnerLineOffset = 0
	var cursorLine = 0
	var oldCurrentLine = 0
	var spinnerWasActive = progressIndicator.Active()
	var writtenHeading = ""

	// r := regexp.MustCompile("([^\r\n]*)([\r\n]{0,2})")

	Logger.Trace().Msgf("Reading and formatting output with spinner %#v\n", progressIndicator)

	for {
		if n, err := reader.Read(buf); err == nil {

			var toWrite strings.Builder

			Logger.Trace().Msgf("Got input %#v", buf[0:n])

			stringBuf := string(buf[0:n])
			segmentContainedText := false
			oldCurrentLine = currentLine

			for token, data, rest := NextToken(stringBuf); token != TokenEmpty; token, data, rest = NextToken(rest) {

				Logger.Trace().Msgf("Token: %#v, %#v, %#v\n", token, data, rest)

				// if lineIsEmpty && currentlyBlankLine {
				// 	logger.Println("This line is empty and we're already on a blank line. Skipping line.")
				// 	fmt.Fprintf(&toWrite, end)
				// 	continue
				// }

				var lineData = lineTracking[currentLine]

				switch token {

				case TokenCursor:
					lineDelta, columnDelta := DecodeCursor(data)
					Logger.Trace().Msgf("Cursor movement: %d, %d", lineDelta, columnDelta)

					// Moving multiple new lines downwards, indent the blank line.
					if lineDelta > 0 && lastToken == TokenCursor && !lineData.CurrentlyIndented {
						// fmt.Fprint(&toWrite, indentStr + lineData.FormattingStash)
						// lineData.FormattingStash = ""
						// lineData.CurrentlyIndented = true
					}

					currentLine = currentLine + lineDelta
					if lineTracking[currentLine] == nil {
						Logger.Trace().Msgf("We haven't written to this line (#%d) before.", currentLine)
						lineTracking[currentLine] = &Line{}
					}

					lineTracking[currentLine].MoveCursor(columnDelta)

					fmt.Fprint(&toWrite, data)

				case TokenColour:
					Logger.Trace().Msgf("Token is Colour code")
					fallthrough
				case TokenWhitespace:
					Logger.Trace().Msgf("Token is whitespace\n")
					if lineData.CurrentlyIndented {
						Logger.Trace().Msgf("We've written on this line. Writing %#v", lineData.FormattingStash+data)
						fmt.Fprint(&toWrite, lineData.FormattingStash+data)
						lineData.FormattingStash = ""
					} else {
						Logger.Trace().Msgf("It's a clean line. Stashing %#v", lineData.FormattingStash+data)
						lineData.FormattingStash = lineData.FormattingStash + data
					}

				case TokenDisplayStream:
					Logger.Trace().Msgf("Token is display stream")

					if !lineData.CurrentlyIndented {
						Logger.Trace().Msgf("Line %d not indented, indenting line with %q and formatting %q\n", currentLine, indentStr, lineData.FormattingStash)
						fmt.Fprintf(&toWrite, "%s%s", indentStr, lineData.FormattingStash)

						lineData.LastColumn = lineData.LastColumn + len(indentStr)
						lineData.FormattingStash = ""
						lineData.CurrentlyIndented = true
					}

					Logger.Trace().Msgf("Writing: %#v", data)
					fmt.Fprint(&toWrite, data)
					lineData.LastColumn = lineData.LastColumn + len(data)
					lineData.Dirty = true
					segmentContainedText = true
				}

				lastToken = token

				if currentLine > maxLine {
					maxLine = currentLine
				}

			}

			toWriteString := toWrite.String()
			anythingToWrite := len(toWriteString) > 0

			Logger.Trace().Msgf("Cursor currently at %d, output will finish at %d:%d", cursorLine, currentLine, lineTracking[currentLine].LastColumn)

			if progressIndicator.Active() && anythingToWrite {
				Logger.Trace().Msgf("Stopping spinner on line %d", cursorLine)
				spinnerWasActive = true
				progressIndicator.Stop()
				if hasShownSomething {
					movement := oldCurrentLine - cursorLine
					// If CL = 4 & ML = 5 and SO = 1, then we want to move CL - ML - SO, or 2 spots up
					moveCursor(out, movement)
				}
			}

			Logger.Trace().Msgf("Writing rendered content: %q", toWriteString)

			if segmentContainedText && !hasShownSomething {
				if progressIndicator.CurrentHeading() != writtenHeading {
					fmt.Fprintf(out, "%s%s\r\n", indentSpaces, aurora.Faint(initialHeading))
				}
				hasShownSomething = true
				spinnerLineOffset = 1
			}

			fmt.Fprint(out, toWriteString)
			cursorLine = currentLine

			Logger.Trace().Msgf("After render, current line %d, maxLines %d, current column %d", currentLine, maxLine, lineTracking[currentLine].LastColumn)

			// If we're on the last line of output, and we're clean, we're probably not waiting
			// for user input, so reveal the spinner.
			if lineTracking[currentLine].LastColumn == 0 && anythingToWrite {
				if lineTracking[currentLine].Dirty {
					movement := maxLine - currentLine + spinnerLineOffset
					// If ML = 5 and CL = 4 and SO = 1, then we need to move from 6 to 4, so ML - CL + SO = 2
					moveCursor(out, movement)
					cursorLine = cursorLine + movement
				}

				if spinnerWasActive {
					Logger.Trace().Msgf("Starting spinner on line %d", cursorLine)
					progressIndicator.Start()
				}
			}

			out.Flush()

			// fmt.Printf("End of segment: %+v", lineTracking)
		} else {
			if err == io.EOF {
				Logger.Trace().Msgf("GOT EOF")

				if lineTracking[currentLine].CurrentlyIndented {
					fmt.Fprint(out, "\r\n")
				}

				if !progressIndicator.Active() && spinnerWasActive {
					progressIndicator.Start()
				}

				out.Flush()

				return !lineTracking[currentLine].CurrentlyIndented

			}

			Logger.Trace().Msgf("ERR %s", err)
		}
	}
}
