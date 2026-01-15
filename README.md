# liftoff-telemetry

**liftoff-telemetry** is a command-line tool for extracting, decoding, and converting telemetry data from the **Liftoff FPV drone simulator**.

It supports:
- parsing telemetry from Liftoff replay files
- receiving live telemetry over UDP
- converting telemetry into machine-readable formats (CSV for now)

The tool is designed to be:
- streaming-oriented
- schema-driven
- suitable for data analysis and machine-learning workflows

---

## How it works

Liftoff provides telemetry data in two main ways:

1. **Replay files** (`.xml`) containing recorded telemetry frames
2. **Live UDP telemetry stream** configured via `TelemetryConfiguration.json`

This tool:
1. Reads raw telemetry frames (from replay file or UDP)
2. Decodes them using a predefined telemetry schema
3. Encodes the decoded records into an output format (CSV)

```

Replay XML / UDP
↓
Telemetry source
↓
Schema-based decoder
↓
Encoder (CSV)
↓
stdout / file

```

---

## Supported commands

### parse-replay

Parse a Liftoff replay XML file and convert telemetry to the chosen output format.

```

liftoff-telemetry parse-replay [flags] <replay.xml>

````

**Flags**
- `--format` — output format (`csv`, default: `csv`)
- `--out` — output file path (`-` or empty = stdout)

**Examples**
```bash
liftoff-telemetry parse-replay ./replay.xml
liftoff-telemetry parse-replay --format csv --out output.csv ./replay.xml
````

---

### listen

Listen for live telemetry over UDP using a Liftoff telemetry configuration file.

```
liftoff-telemetry listen --config <TelemetryConfiguration.json> [flags]
```

**Required flags**

* `--config` — path to `TelemetryConfiguration.json`

**Optional flags**

* `--format` — output format (`csv`)
* `--out` — output file path (`-` or empty = stdout)
* `--queue` — internal frame queue size (default: `256`)
* `--read-buf` — OS UDP receive buffer size in bytes

**Examples**

```bash
liftoff-telemetry listen \
  --config ./TelemetryConfiguration.json \
  --out stream.csv
```

The command runs until interrupted (`Ctrl+C`).

---

## Output formats

Currently supported:

* **CSV**

The design allows additional encoders (e.g. JSON, NDJSON, Parquet) to be added without changing the core pipeline.

---

## Telemetry schema

The telemetry schema defines:

* which telemetry fields are expected
* their order and binary layout
* how raw frames are decoded

Schemas are derived from Liftoff’s telemetry stream format configuration.

---

## Liftoff telemetry documentation

Official Liftoff documentation describing telemetry configuration and UDP streaming:

* [https://steamcommunity.com/sharedfiles/filedetails/?id=3160488434](https://steamcommunity.com/sharedfiles/filedetails/?id=3160488434)

This page explains:

* how to enable telemetry in Liftoff
* how to configure `TelemetryConfiguration.json`
* how the UDP telemetry stream is structured

---

## Installation

### Build from source

```bash
git clone https://github.com/mchekh/liftoff-telemetry.git
cd liftoff-telemetry
make build
```

Binary will be available at:

```
bin/liftoff-telemetry
```

### Run without building

```bash
go run ./cmd/liftoff-telemetry --help
```
---

## License

MIT
