package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/bryanaustin/yaarp"
	"io"
	"os"
	"text/template"
)

type TemplateData struct {
	N     int
	Name  map[string]string
	Index []string
}

// Constant
var (
	nullbyte    = []byte{0x00}
	newlinebyte = []byte("\n")
)

// Config values
var (
	input       string
	output      string
	noheader    bool
	templatestr string
	zerodelim   bool
	newline     bool
)

// Runtime values
var (
	in  io.Reader
	out io.Writer
	ct  *template.Template
)

func main() {
	configure()
	run()
}

func configure() {
	flag.StringVar(&input, "input", input, "input csv (from stdin if not provided)")
	flag.StringVar(&output, "output", output, "output file (to stdout if not provided)")
	flag.BoolVar(&noheader, "no-header", noheader, "first line is not a header, address by index only")
	flag.BoolVar(&newline, "new-line", newline, "add a new line at the end of each output")
	flag.BoolVar(&zerodelim, "zero", zerodelim, "seperate each line with a null character")

	yaarp.Parse()
	templatestr = yaarp.Arg(0)

	if len(templatestr) < 1 {
		fmt.Fprintln(os.Stderr, `Ouput template expected.
Example:
echo -e "num,value\n1,one\n2,two\n8,eight" | csveach --new-line '{{ index .Index 0 }}:{{ index .Name "value"}}'`)
		os.Exit(1)
	}
}

func run() {
	var err error
	ct, err = template.New("out").Parse(templatestr)
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("Error parsing output template: %s", err))
		os.Exit(4)
	}

	if len(input) > 0 {
		f, err := os.Open(input)
		if err != nil {
			fmt.Fprintln(os.Stderr, fmt.Sprintf("Error opening input file: %s", err))
			os.Exit(2)
		}
		defer f.Close()
		in = bufio.NewReader(f)
	} else {
		in = os.Stdin
	}

	if len(output) > 0 {
		f, err := os.Create(output)
		if err != nil {
			fmt.Fprintln(os.Stderr, fmt.Sprintf("Error opening output file: %s", err))
			os.Exit(3)
		}
		defer f.Close()
		out = f
	} else {
		out = os.Stdout
	}

	csvr := csv.NewReader(in)
	csvr.ReuseRecord = true

	var header []string
	if !noheader {
		header, err = csvr.Read()
		if err != nil {
			if err != io.EOF {
				fmt.Fprintln(os.Stderr, fmt.Sprintf("Error reading csv header: %s", err))
				os.Exit(6)
			}
			return
		}
		header = append([]string(nil), header...) //copy
	}

	var data TemplateData
	data.Name = make(map[string]string, len(header))
	for {
		data.Index, err = csvr.Read()
		if err != nil {
			if err != io.EOF {
				fmt.Fprintln(os.Stderr, fmt.Sprintf("Error reading csv: %s", err))
				os.Exit(7)
			}
			return
		}

		zeromap(data.Name)
		if !noheader {
			headertomap(header, data.Name, data.Index)
		}

		ct.Execute(out, data)
		data.N++

		if newline {
			out.Write(newlinebyte)
		}

		if zerodelim {
			out.Write(nullbyte)
		}
	}
}

func headertomap(columns []string, m map[string]string, line []string) {
	for i, v := range line {
		if i < len(columns) {
			m[columns[i]] = v
		}
	}
}

func zeromap(x map[string]string) {
	for k := range x {
		x[k] = ""
	}
}
