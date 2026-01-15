package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/mchekh/liftoff-telemetry/encode/csv"
	"github.com/mchekh/liftoff-telemetry/source"
	"github.com/mchekh/liftoff-telemetry/source/replay"
	"github.com/mchekh/liftoff-telemetry/source/udp"
	"github.com/mchekh/liftoff-telemetry/telemetry"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		printUsage(os.Stderr)
		return 2
	}

	switch args[0] {
	case "parse-replay":
		return cmdParseReplay(args[1:])
	case "listen":
		return cmdListen(args[1:])
	case "-h", "--help", "help":
		printUsage(os.Stdout)
		return 0
	default:
		fmt.Fprintln(os.Stderr, "unknown command:", args[0])
		printUsage(os.Stderr)
		return 2
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintf(w, `liftoff-telemetry — Liftoff telemetry processing tool

Usage:
  liftoff-telemetry <command> [flags] [args]

Commands:
  parse-replay   Convert a Liftoff replay XML file to telemetry records
  listen         Listen for a UDP telemetry stream using a configuration file

parse-replay:
  liftoff-telemetry parse-replay [--format csv] [--out FILE|-] <replay.xml>

listen:
  liftoff-telemetry listen --config <TelemetryConfiguration.json> [--format csv] [--out FILE|-]
                             [--queue N] [--read-buf BYTES]

Flags:
  --format   Output format (csv for now)
  --out      Output file path; "-" or empty means stdout
  --config   Path to TelemetryConfiguration.json (listen only)
  --queue    Internal frame queue size (listen only)
  --read-buf OS UDP receive buffer size in bytes (listen only)

Notes:
  • Flags must appear before positional arguments.
  • parse-replay reads a replay XML file and writes telemetry records.
  • listen runs until interrupted (Ctrl+C).

Examples:
  liftoff-telemetry parse-replay --format csv --out output.csv ./replay.xml
  liftoff-telemetry parse-replay ./replay.xml
  liftoff-telemetry listen --config ./TelemetryConfiguration.json --out stream.csv
`)
}

func cmdParseReplay(args []string) int {
	fs := flag.NewFlagSet("parse-replay", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	format := fs.String("format", "csv", "output format (csv)")
	outPath := fs.String("out", "", "output file path (default: stdout)")

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "parse-replay requires <replay.xml>")
		return 2
	}
	replayPath := fs.Arg(0)

	schema, err := telemetry.SchemaFromStreamFormat([]string{
		"Position",
		"Attitude",
		"Input",
		"Timestamp",
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	src, err := replay.ReplayTelemetrySourceFromFile(replayPath, schema.Size())
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	defer func() { _ = src.Close() }()

	out, closeOut, err := openOut(*outPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	defer closeOut()

	enc, err := newEncoder(*format, schema, out)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	if err := enc.EncodeAll(src); err != nil && !errors.Is(err, io.EOF) {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	return 0
}

func cmdListen(args []string) int {
	fs := flag.NewFlagSet("listen", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	cfgPath := fs.String("config", "", "path to TelemetryConfiguration.json")
	format := fs.String("format", "csv", "output format (csv)")
	outPath := fs.String("out", "", "output file path (default: stdout)")

	queue := fs.Int("queue", 256, "internal frame queue size (drop when full)")
	readBuf := fs.Int("read-buf", 0, "OS UDP receive buffer (bytes), 0 = don't set")

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *cfgPath == "" {
		fmt.Fprintln(os.Stderr, "listen requires --config <TelemetryConfiguration.json>")
		return 2
	}

	cfg, err := readTelemetryConfig(*cfgPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	schema, err := telemetry.SchemaFromStreamFormat(cfg.StreamFormat)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	src, err := udp.NewUdpTelemetrySource(udp.UDPConfig{
		Addr:            cfg.EndPoint,
		MaxDatagramSize: schema.Size(),
		QueueSize:       *queue,
		ReadBufferBytes: *readBuf,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	defer func() { _ = src.Close() }()

	go func() {
		<-ctx.Done()
		_ = src.Close()
	}()

	out, closeOut, err := openOut(*outPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	defer closeOut()

	enc, err := newEncoder(*format, schema, out)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	if err := enc.EncodeAll(src); err != nil && !errors.Is(err, io.EOF) {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	return 0
}

type telemetryConfig struct {
	EndPoint     string   `json:"EndPoint"`
	StreamFormat []string `json:"StreamFormat"`
}

func readTelemetryConfig(path string) (telemetryConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return telemetryConfig{}, err
	}

	var cfg telemetryConfig
	if err := json.Unmarshal(b, &cfg); err != nil {
		return telemetryConfig{}, err
	}

	cfg.EndPoint = strings.TrimSpace(cfg.EndPoint)
	if cfg.EndPoint == "" {
		return telemetryConfig{}, errors.New("config: EndPoint is required")
	}
	if len(cfg.StreamFormat) == 0 {
		return telemetryConfig{}, errors.New("config: StreamFormat must be non-empty")
	}

	if _, err := net.ResolveUDPAddr("udp", cfg.EndPoint); err != nil {
		return telemetryConfig{}, fmt.Errorf("config: invalid EndPoint %q: %w", cfg.EndPoint, err)
	}

	return cfg, nil
}

func openOut(path string) (io.Writer, func() error, error) {
	if path == "" || path == "-" {
		return os.Stdout, func() error { return nil }, nil
	}

	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, nil, err
		}
	}

	f, err := os.Create(path)
	if err != nil {
		return nil, nil, err
	}
	return f, f.Close, nil
}

type encoder interface {
	EncodeAll(src source.TelemetrySource) error
}

func newEncoder(format string, schema *telemetry.TelemetrySchema, out io.Writer) (encoder, error) {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "csv":
		return csv.NewCSVEncoder(schema, out)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}
