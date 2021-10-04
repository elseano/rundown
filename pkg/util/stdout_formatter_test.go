package util

import (
	"bufio"
	"fmt"
	"testing"
	"time"

	// "bytes"
	"io"
	"log"

	// "io/ioutil"
	"os"

	"github.com/stretchr/testify/assert"
)

func TestFormatBasic(t *testing.T) {

	inputStream := []string{
		"One\n",
		"Two\n",
		"Three",
		"Four\n",
	}

	expectedStream := []string{
		"^E^* One\n^S^",
		"^E^* Two\n^S^",
		"^E^* Three",
		"Four\n^S^",
	}

	validateStreams(t, inputStream, expectedStream, true)
}

func TestFormatIgnoreMultipleBlanks(t *testing.T) {

	inputStream := []string{
		"One\n",
		"\n",
		"\n",
		"Four\n",
	}

	expectedStream := []string{
		"^E^* One\n^S^",
		"^E^\n^S^",
		"^E^\n^S^",
		"^E^* Four\n^S^",
	}

	validateStreams(t, inputStream, expectedStream, true)
}

func TestFormatBlankLineAfter(t *testing.T) {

	inputStream := []string{
		"One",
		"\n",
		"Two",
		"\n",
		"Three\n",
	}

	expectedStream := []string{
		"^E^* One",
		"\n^S^",
		"^E^* Two",
		"\n^S^",
		"^E^* Three\n^S^",
	}

	validateStreams(t, inputStream, expectedStream, true)
}

func TestFormatBlankLineWithFormattingAfter(t *testing.T) {

	inputStream := []string{
		"One",
		"\n",
		"Two",
		"\n",
		"Three\n",
		"\033[0m",
		"Four\n",
	}

	expectedStream := []string{
		"^E^* One",
		"\n^S^",
		"^E^* Two",
		"\n^S^",
		"^E^* Three\n^S^",
		"^E^* \033[0mFour\n^S^",
	}

	validateStreams(t, inputStream, expectedStream, true)
}

func TestFormatInputPrompt(t *testing.T) {

	inputStream := []string{
		"This is some text\n",
		"Here is a prompt ",
		"This is user typing\n",
	}

	expectedStream := []string{
		"^E^* This is some text\n^S^",
		"^E^* Here is a prompt ",
		"This is user typing\n^S^",
	}

	validateStreams(t, inputStream, expectedStream, true)
}

func TestFormatEdge1(t *testing.T) {

	inputStream := []string{
		"No USB token found yet. Waiting for USB token...\r\n1. Please insert the token if you haven't done so and touch its key\r\n2. If you have inserted your token already, please try shutting down browsers first or fol",
		"low the passgo troubleshooting steps at: https://internal.acme.corp/USBTokens/UserGuide.\r\n3. If you're on a remote computer, USB won't work and you need to run passgo --remote\r\n",
		"\r\nUSB Authentication failed, reason :",
		"USB_TIMEOUT",
		"(2)\r\n",
	}

	expectedStream := []string{
		"^E^* No USB token found yet. Waiting for USB token...\r\n* 1. Please insert the token if you haven't done so and touch its key\r\n* 2. If you have inserted your token already, please try shutting down browsers first or fol",
		"low the passgo troubleshooting steps at: https://internal.acme.corp/USBTokens/UserGuide.\r\n* 3. If you're on a remote computer, USB won't work and you need to run passgo --remote\r\n^S^",
		"^E^\r\n* USB Authentication failed, reason :",
		"USB_TIMEOUT",
		"(2)\r\n^S^",
	}

	validateStreams(t, inputStream, expectedStream, true)
}

func TestFormatEdge2(t *testing.T) {
	// Replicates STDIN typing. Adds \r\n to the end when completed with the stream as we're
	// not on a newline.

	inputStream := []string{
		"T", "h", "i", "s", " ", "i", "n",
	}

	expectedStream := []string{
		"^E^* T", "h", "i", "s", " ", "i", "n", "\r\n^S^",
	}

	validateStreams(t, inputStream, expectedStream, false)
}

func TestFormatAsProgress(t *testing.T) {

	inputStream := []string{
		"1%\r",
		"2%\r",
		"3%\r8%\r",
		"100%\r",
		"Done.\r\n",
	}

	expectedStream := []string{
		"^E^* 1%\r\x1b[1B^S^",
		"^E^\x1b[1A* 2%\r\x1b[1B^S^",
		"^E^\x1b[1A* 3%\r* 8%\r\x1b[1B^S^",
		"^E^\x1b[1A* 100%\r\x1b[1B^S^",
		"^E^\x1b[1A* Done.\r\n^S^",
	}

	validateStreams(t, inputStream, expectedStream, true)
}

func TestFormatAsLineJumps(t *testing.T) {

	inputStream := []string{
		"The push refers to repository [863423692916.dkr.ecr.ap-southeast-2.amazonaws.com/kamailio]\r\n",
		"\r\n\033[1A\033[2K\re186af104679: Preparing \r\033[1B\r\n\033[1A\033[2K\r97acaa5c7388: Preparing \r\033[1B",
		"\033[1A\033[2K\r97acaa5c7388: Layer already exists \r\033[1B",
	}

	expectedStream := []string{
		"^E^* The push refers to repository [863423692916.dkr.ecr.ap-southeast-2.amazonaws.com/kamailio]\r\n^S^",
		"^E^\r\n\x1b[1A\x1b[2K\r* e186af104679: Preparing \r\x1b[1B\r\n\x1b[1A\x1b[2K\r* 97acaa5c7388: Preparing \r\x1b[1B^S^",
		"^E^\033[1A\033[2K\r* 97acaa5c7388: Layer already exists \r\033[1B^S^",
	}

	validateStreams(t, inputStream, expectedStream, true)
}

func TestFormatDocker(t *testing.T) {
	inputStream := []string{
		string(
			[]byte{0x53, 0x74, 0x6f, 0x70, 0x70, 0x69, 0x6e, 0x67, 0x20, 0x62, 0x64, 0x64, 0x5f, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x5f, 0x31, 0x20, 0x20, 0x2e, 0x2e, 0x2e, 0x20, 0xd, 0xd, 0xa, 0x53, 0x74, 0x6f, 0x70, 0x70, 0x69, 0x6e, 0x67, 0x20, 0x62, 0x64, 0x64, 0x5f, 0x61, 0x70, 0x69, 0x5f, 0x31, 0x20, 0x20, 0x20, 0x20, 0x20, 0x2e, 0x2e, 0x2e, 0x20, 0xd, 0xd, 0xa, 0x53, 0x74, 0x6f, 0x70, 0x70, 0x69, 0x6e, 0x67, 0x20, 0x62, 0x64, 0x64, 0x5f, 0x72, 0x65, 0x64, 0x69, 0x73, 0x5f, 0x31, 0x20, 0x20, 0x20, 0x2e, 0x2e, 0x2e, 0x20, 0xd, 0xd, 0xa, 0x53, 0x74, 0x6f, 0x70, 0x70, 0x69, 0x6e, 0x67, 0x20, 0x62, 0x64, 0x64, 0x5f, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x5f, 0x31, 0x20, 0x2e, 0x2e, 0x2e, 0x20, 0xd, 0xd, 0xa, 0x53, 0x74, 0x6f, 0x70, 0x70, 0x69, 0x6e, 0x67, 0x20, 0x62, 0x64, 0x64, 0x5f, 0x64, 0x62, 0x5f, 0x31, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x2e, 0x2e, 0x2e, 0x20, 0xd, 0xd, 0xa},
		),
		string(
			[]byte{0x1b, 0x5b, 0x35, 0x41, 0x1b, 0x5b, 0x32, 0x4b, 0xd, 0x53, 0x74, 0x6f, 0x70, 0x70, 0x69, 0x6e, 0x67, 0x20, 0x62, 0x64, 0x64, 0x5f, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x65, 0x5f, 0x31, 0x20, 0x20, 0x2e, 0x2e, 0x2e, 0x20, 0x1b, 0x5b, 0x33, 0x32, 0x6d, 0x64, 0x6f, 0x6e, 0x65, 0x1b, 0x5b, 0x30, 0x6d, 0xd, 0x1b, 0x5b, 0x35, 0x42},
		),
	}

	expectedStream := []string{
		"^E^* Stopping bdd_chrome_1  ... \r\r\n" +
			"* Stopping bdd_api_1     ... \r\r\n" +
			"* Stopping bdd_redis_1   ... \r\r\n" +
			"* Stopping bdd_connect_1 ... \r\r\n" +
			"* Stopping bdd_db_1      ... \r\r\n" +
			"^S^",
		"^E^\x1b[5A\x1b[2K\r* Stopping bdd_chrome_1  ... \x1b[32mdone\x1b[0m\r\x1b[5B" +
			"^S^",
	}

	validateStreams(t, inputStream, expectedStream, true)
}

type MockSpinner struct {
	Out      io.Writer
	isActive bool
	Heading  string
}

func (s *MockSpinner) Start() {
	s.isActive = true
	s.Out.Write([]byte("^S^"))
}

func (s *MockSpinner) Stop() {
	s.isActive = false
	s.Out.Write([]byte("^E^"))
}

func (s *MockSpinner) Active() bool {
	return s.isActive
}

func (s *MockSpinner) CurrentHeading() string {
	return s.Heading
}

func validateStreams(t *testing.T, inputStream []string, expectedStream []string, endedOnBlank bool) {

	var outputReader, outputWriter, _ = os.Pipe()
	var inputReader, inputWriter, _ = os.Pipe()
	var result = []string{}
	var done = make(chan bool)
	var formatterResult = make(chan bool)

	bufWriter := bufio.NewWriter(outputWriter)

	logger := log.New(os.Stdout, "", 0)
	spinner := &MockSpinner{Out: bufWriter, isActive: true}

	go func() {
		result := ReadAndFormatOutput(inputReader, 0, "* ", spinner, bufWriter, logger, "")
		outputWriter.Close()
		formatterResult <- result
	}()

	defer outputWriter.Close()
	defer outputReader.Close()
	defer inputWriter.Close()
	defer inputReader.Close()

	// if len(expectedStream) != len(inputStream) {
	// 	t.Fatalf("Input and output stream sizes differ: %d vs %d", len(inputStream), len(expectedStream))
	// }

	go func() {
		for {
			var buf []byte = make([]byte, 256)
			n, err := outputReader.Read(buf)
			if err == io.EOF {
				done <- true
				return
			} else if err == nil {
				result = append(result, string(buf[0:n]))
			}
		}
	}()

	for i := 0; i < len(inputStream); i = i + 1 {
		inputWriter.WriteString(inputStream[i])
		time.Sleep(20 * time.Millisecond)
	}

	inputWriter.Close()

	<-done

	fmt.Println("In/Out done. Validating")

	fmt.Printf("Rendered:\n%#v\n", result)

	for i := 0; i < len(expectedStream); i = i + 1 {
		if i > len(result)-1 {
			t.Fatalf("Expected %s, got nil", InspectString(expectedStream[i]))
			return
		} else if !assert.Equal(t, expectedStream[i], result[i]) {
			return
		}
	}

	if len(result) != len(expectedStream) {
		t.Fatalf("Result lengths differ %d vs %d: %v", len(result), len(expectedStream), InspectList(result))
		return
	}

	assert.Equal(t, endedOnBlank, <-formatterResult, "Result newline setting wasn't expected")

}
