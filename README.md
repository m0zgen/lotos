# LoToS

Log to WebSocket realtime log reader with WebSocket re-translation.

### Configuration

```yaml
port: 3001
logFilePath: /path/to/log/file.log
showLogs: true
```

* `port` - Port to run the server on.
* `logFilePath` - Path to the log file to read.
* `showLogs` - Show logs in the realtime console.

## Building

```bash
go build
```

## Running

```bash
./lotos config.yml
```