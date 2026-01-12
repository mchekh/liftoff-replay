package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/mchekh/liftoff-replay/liftoffreplay"
	telemetry "github.com/mchekh/liftoff-replay/liftofftelemetry"
	telemetry_encoder "github.com/mchekh/liftoff-replay/liftofftelemetry/encoder"
)

type Temp struct {
	r io.Reader
}

func (t *Temp) Read(p []byte) (n int, err error) {
	return t.r.Read(p)
}

func (t *Temp) Close() (err error) {
	return nil
}

func main() {
	var (
		format = flag.String("format", "csv", "Output format: csv | json | ndjson")
		// outPath = flag.String("out", "", "Output file (default: stdout)")
		nested = flag.Bool("nested", false, "Use nested JSON structure (position/orientation/controls)")
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

	// _out, closer, err := getOutWriter(*outPath)
	// if err != nil {
	// 	fmt.Fprintln(os.Stderr, "error:", err)
	// 	os.Exit(1)
	// }
	// if closer != nil {
	// 	defer func() {
	// 		if err := closer.Close(); err != nil {
	// 			fmt.Fprintf(os.Stderr, "warning: close output %q: %v\n", *outPath, err)
	// 		}
	// 	}()
	// }

	schema, err := telemetry_encoder.SchemaFromStreamFormat([]string{
		"Position",
		"Attitude",
		"Input",
		"Timestamp",
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	decoded, err := liftoffreplay.ExtractReplayBinaryData(in)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	src := telemetry.NewFileTelemetrySource(&Temp{r: decoded}, schema.Size())

	reader, err := telemetry_encoder.NewReader[TestData](src, schema)

	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	for {
		rec, err := reader.Next()
		if err != nil {
			if err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}
		}
		fmt.Printf("\n%v\n", *rec.Timestapm)
	}

	// enc:= telemetry.NewCSVEncoder(schema, out)

	// err = enc.EncodeAll(src)
	//
	// if err != nil {
	// 	fmt.Fprintln(os.Stderr, "error:", err)
	// 	os.Exit(1)
	// }
}

type TestData struct {
	Timestapm *float32 `telemetry:"Timestamp"`
	PositionX float32  `telemetry:"PositionX"`
	Something float32  `telemetry:"asdfsad fsdf "`
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
