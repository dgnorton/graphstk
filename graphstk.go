package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var (
	title = flag.String("t", "", "title for the stack trace")
	filename = flag.String("f", "", "text file containing stack trace")
	graphType = flag.String("T", "png", "output type (e.g., png, svg)")
)

func init() {
	flag.Parse()
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func readLines(r io.Reader) ([]string, error) {
	stack, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(stack), "\n"), nil
}

type StackFile struct {
	Path string
	Line int
	Address int64
}

func ParseStackFile(s string) (*StackFile, error) {
	re := regexp.MustCompile(`((?:/.*)*):([0-9]*) \(0x([0-9A-Fa-f]+)\)`)
	groups := re.FindAllStringSubmatch(s, -1)
	if len(groups) != 1 || len(groups[0]) != 4 {
		return nil, errors.New("invalid stack file string")
	}

	line, err := strconv.ParseInt(groups[0][2], 10, 32)
	if err != nil {
		return nil, err
	}

	address, err := strconv.ParseInt(groups[0][3], 16, 64)
	if err != nil {
		return nil, err
	}

	sf := &StackFile{
		Path: groups[0][1],
		Line: int(line),
		Address: address,
	}

	return sf, nil
}

type StackCall struct {
	CallerType string
	CallerFunc string
	Callee string
}

func (sc *StackCall) Caller() string {
	if len(sc.CallerType) > 0 {
		return strings.Join([]string{sc.CallerType, sc.CallerFunc}, ".")
	}
	return sc.CallerFunc
}

func ParseStackCall(s string) (*StackCall, error) {
	re := regexp.MustCompile(`\((.*)\)\.(.*): (.*)`)
	groups := re.FindAllStringSubmatch(s, -1)
	if len(groups) == 1 && len(groups[0]) == 4 {
		sc := &StackCall{
			CallerType: groups[0][1],
			CallerFunc: groups[0][2],
			Callee: groups[0][3],
		}

		return sc, nil
	}

	re = regexp.MustCompile(` *(.*): (.*)`)
	groups = re.FindAllStringSubmatch(s, -1)
	if len(groups) == 1 && len(groups[0]) == 3 {
		sc := &StackCall{
			CallerFunc: groups[0][1],
			Callee: groups[0][2],
		}

		return sc, nil
	}

	return nil, errors.New("invalid stack file string")
}

func writeGraphNode(w io.Writer, nodeName string, sf *StackFile, sc *StackCall) error {
	if sf == nil {
		return errors.New("sf is nil")
	} else if sc == nil {
		return errors.New("sc is nil")
	}

	dirs, file := path.Split(sf.Path)
	_, dir := path.Split(strings.TrimRight(dirs, `/\`))

	fmt.Fprintf(w, "\t%s [label=\"%s/%s:%d\n%s\",shape=box];\n", nodeName, dir, file, sf.Line, sc.Caller())

	return nil
}

func main() {
	var f *os.File
	var err error

	if *filename != "" {
		f, err = os.Open(*filename)
		check(err)
		defer f.Close()
	} else {
		f = os.Stdin
	}

	stack, err := readLines(f)
	check(err)

	dotParms := fmt.Sprintf("-T%s", *graphType)
	cmd := exec.Command("dot", dotParms)
	in, err := cmd.StdinPipe()
	check(err)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	check(cmd.Start())

	fmt.Fprintf(in, "digraph {\n")
	if len(*title) > 0 {
		fmt.Fprintf(in, "\tlabelloc=\"t\"\n")
		fmt.Fprintf(in, "\nlabel=\"%s\"\n", *title)
	}

	var sf *StackFile
	var sc *StackCall
	var nodes []string

	for i, line := range stack {
		if line == "" {
			continue
		}
		if i % 2 == 0 {
			sf, err = ParseStackFile(line)
			if err != nil {
				log.Fatalf("%d: %s\n\t%s\n", i, line, err.Error())
			}
		} else {
			sc, err = ParseStackCall(line)
			if err != nil {
				log.Fatalf("%d: %s\n\t%s\n", i, line, err.Error())
			}

			name := fmt.Sprintf("N%d", i)
			err = writeGraphNode(in, name, sf, sc)
			if err != nil {
				log.Fatalln(err)
			}
			nodes = append(nodes, name);
		}
	}

	var lastNode string
	for i, node := range nodes {
		if i > 0 {
			fmt.Fprintf(in, "%s -> %s [weight=1,dir=back];\n", lastNode, node)
		}
		lastNode = node
	}

	fmt.Fprintf(in, "}\n")
	in.Close()

	check(cmd.Wait())
}
