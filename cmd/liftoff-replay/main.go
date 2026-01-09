package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/mchekh/liftoff-replay/liftoffreplay"
)

var flatEncoders = []liftoffreplay.RecordEncoder[liftoffreplay.ReplayRecordFlat]{
	&liftoffreplay.CsvRecordEncoder{},
	&liftoffreplay.JsonRecordEncoder[liftoffreplay.ReplayRecordFlat]{},
	&liftoffreplay.NdJsonRecordEncoder[liftoffreplay.ReplayRecordFlat]{},
}

var nestedEncoders = []liftoffreplay.RecordEncoder[liftoffreplay.ReplayRecordNested]{
	&liftoffreplay.JsonRecordEncoder[liftoffreplay.ReplayRecordNested]{},
	&liftoffreplay.NdJsonRecordEncoder[liftoffreplay.ReplayRecordNested]{},
}

func main() {
	var (
		format  = flag.String("format", "csv", "Output format: csv | json | ndjson")
		outPath = flag.String("out", "", "Output file (default: stdout)")
		nested  = flag.Bool("nested", false, "Use nested JSON structure (position/orientation/controls)")
	)

	flag.Usage = func() {
		fmt.Fprintf(
			flag.CommandLine.Output(),
			"Usage:\n  %s [--format csv|json|ndjson] [--out <file>] [--nested] <input.xml>\n\nOptions:\n",
			os.Args[0],
		)
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}

	if *nested && *format == "csv" {
		fmt.Fprintln(os.Stderr, "error: --nested is not supported for csv format")
		os.Exit(2)
	}

	inPath := flag.Arg(0)
	in, err := os.Open(inPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: open input %q: %v\n", inPath, err)
		os.Exit(1)
	}
	defer in.Close()

	out, closer, err := getOutWriter(*outPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	if closer != nil {
		defer func() {
			if err := closer.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "warning: close output %q: %v\n", *outPath, err)
			}
		}()
	}

	if *nested {
		enc, err := getEncodingStrategy(*format, nestedEncoders)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(2)
		}
		if err := liftoffreplay.ConvertLiftoffReplayNested(in, out, enc); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		return
	}

	enc, err := getEncodingStrategy(*format, flatEncoders)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(2)
	}
	if err := liftoffreplay.ConvertLiftoffReplayFlat(in, out, enc); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func getOutWriter(outPath string) (io.Writer, io.Closer, error) {
	if outPath == "" {
		return os.Stdout, nil, nil
	}
	f, err := os.Create(outPath)
	if err != nil {
		return nil, nil, fmt.Errorf("create output %q: %w", outPath, err)
	}
	return f, f, nil
}

func getEncodingStrategy[T liftoffreplay.ReplayRecordConstraint](
	format string,
	encoders []liftoffreplay.RecordEncoder[T],
) (liftoffreplay.RecordEncoder[T], error) {
	for _, enc := range encoders {
		if enc.Format() == format {
			return enc, nil
		}
	}
	return nil, fmt.Errorf("unknown format %q (expected csv|json|ndjson)", format)
}
