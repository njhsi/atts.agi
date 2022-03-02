package sox

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

// Sox parent group for all functions in sox-go
type Sox struct {
}

// NewSox Create a new instance for sox
func NewSox() Sox {
	return Sox{}
}

//File - return output file define
type File struct {
	FilePath string
	Duration float32
}

//Trim - is func for crop the file sound -
func (s Sox) Trim(file string, outputFile string, start float32, duration float32) (*File, error) {
	cmd := exec.Command("sox", file, outputFile, "trim", fmt.Sprintf("%.2f", start), fmt.Sprintf("%.2f", duration))
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	return &File{
		FilePath: outputFile,
	}, nil
}

//Repeat - is func for crop the file sound -
func (s Sox) Repeat(file string, outputFile string, repeat int) (*File, error) {
	cmd := exec.Command("sox", file, outputFile, "repeat", fmt.Sprintf("%d", repeat))
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	log.Printf("in all caps: %q\n", out.String())
	return &File{
		FilePath: outputFile,
	}, nil
}

//Volume - is func for change volume
func (s Sox) Volume(file string, outputFile string, volume float32) (*File, error) {
	cmd := exec.Command("sox", "-v", fmt.Sprintf("%f", volume), file, outputFile)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	log.Printf("in all caps: %q\n", out.String())
	return &File{
		FilePath: outputFile,
	}, nil
}

func (s Sox) Info(file string) (*File, error) {
	cmdDuration := exec.Command("sox", "--info", "-D", file)
	var out bytes.Buffer
	cmdDuration.Stdout = &out
	err := cmdDuration.Run()
	if err != nil {
		return nil, err
	}

	duration, err := strconv.ParseFloat(strings.ReplaceAll(out.String(), "\n", ""), 32)
	if err != nil {
		return nil, err
	}
	return &File{
		FilePath: file,
		Duration: float32(duration),
	}, nil
}

func (s Sox) Join(files []string, outputFile string, mix bool) (*File, error) {
	commandCobine := "-M"
	if mix {
		commandCobine = "-m"
	}
	params := []string{commandCobine}
	for _, file := range files {
		params = append(params, file)
	}
	params = append(params, outputFile)
	cmd := exec.Command("sox", params...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	return &File{
		FilePath: outputFile,
	}, nil
}

func (s Sox) Combine(files []string, outputFile string) (*File, error) {

	params := []string{}
	for _, file := range files {
		params = append(params, file)
	}
	params = append(params, outputFile)
	cmd := exec.Command("sox", params...)
	log.Printf("ref %s", cmd.String())
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	return &File{
		FilePath: outputFile,
	}, nil
}

func (s Sox) Fade(file string, outputFile string, start float32, stop float32, typeFade string) (*File, error) {
	if typeFade != "in" && typeFade != "out" {
		return nil, errors.New("invalid type")
	}

	params := []string{}
	params = append(params, file)
	params = append(params, outputFile)
	params = append(params, "fade")
	switch typeFade {
	case "in":
		params = append(params, fmt.Sprintf("%f", start))
		params = append(params, fmt.Sprintf("%f", stop))
		params = append(params, "0")
	case "out":
		params = append(params, "0")
		params = append(params, fmt.Sprintf("%f", stop))
		params = append(params, fmt.Sprintf("%f", start))
	}
	cmd := exec.Command("sox", params...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	log.Printf("ref %s", cmd.String())
	if err != nil {
		return nil, err
	}
	return &File{
		FilePath: outputFile,
	}, nil
}
