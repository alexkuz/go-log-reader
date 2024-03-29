# go-log-reader

### Install
```sh
curl -fsSL https://raw.githubusercontent.com/alexkuz/go-log-reader/main/install.sh | sh
```

### Usage

```sh
go-log-reader [-c <config_file>] [--param_name <param_value>]
```

### Config example

**.go-log-reader.json**
```json
{
  "logs": [
    {
      "title": "My remote service",
      "command": "ssh -tt ${url} tail -100f /path/to/logfile.log",
      "entry_pattern": "(INFO|DEBUG|WARN)"
    },
    {
      "title": "My local docker container",
      "command": "docker logs -f my_container",
      "entry_pattern": ""
    }
  ]
}
```

**.go-log-reader.yaml**
```yaml
logs:
  - title: "My remote service"
    command: "ssh -tt ${url} tail -100f /path/to/logfile.log"
    entry_pattern: "(INFO|DEBUG|WARN)"

  - title: "My local docker container"
    command: "docker logs -f my_container",
    entry_pattern: ""
```

```sh
go-log-reader --url my-server.com
```
