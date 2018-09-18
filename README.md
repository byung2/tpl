# tpl

`tpl` is a command line tool for Golang templates.

Inspired by Jinja2 [kolypto/j2cli](https://github.com/kolypto/j2cli).

Features:

* Execute Golang templates using JSON, YAML, INI file data sources
* Show all missing keys and processed key:value pairs in JSON, YAML, INI format
* Search for missing keys and input values from stdin (Do not support if template files including 'Actions' or 'Fuctions')
* Allows to use environment variables
* Ensure no missing keys

## Install

### Download the binary

1. Download the binary for your platform:
```
$ curl -o tpl -L0 https://github.com/byung2/tpl/releases/download/v0.1.0/tpl_v0.1.0_linux_amd64
```

2. Install the binary:
```
$ chmod +x ./tpl
$ sudo mv ./tpl /usr/local/bin/tpl
```

3. Verify that the CLI is working properly:
```
$ tpl -v
```

### Install latest version using Golang
```
$ go get github.com/byung2/tpl/cmd/tpl
```


## Usage

Execute template(s) using JSON-file data source:

    $ tpl exec config.tmpl -d data.json

Execute template(s) using YAML-file data source:

    $ tpl exec config.yml.tmpl config2.yml.tmpl -d data.yml

Execute template(s) using INI-file data source:

    $ tpl exec config.yml config2.yml -d data.ini

Execute template(s) using environment variables:

    $ tpl exec config -e

Execute template(s) using interactively input values from stdin:

    $ tpl exec config -i

Show all missing keys:

    $ tpl keys config

Show all missing keys and processed key:value pairs:

    $ tpl keys config -d data.yaml


## Simple examples

1. Create a template and data file:
```
$ echo 'version: '3'
services:
  web:
    build: .
    ports:
     - "5000:5000"
  redis:
    image: "{{.redis.image}}:{{.redis.tag}}"' > docker-compose.yml.tmpl

$ echo '---

redis:
  image: "redis"
  tag: "4.0.11-alpine"' > data.yml
```

2-A. Execute a template and store the processed template:
```
$ tpl exec docker-compose.yml.tmpl -d data.yml --outdir .

$ cat docker-compose.yml
version: 3
services:
  web:
    build: .
    ports:
     - "5000:5000"
  redis:
    image: "redis:4.0.11-alpine"
```

2-B. Execute a template using interactively input values from stdin:
```
$ tpl exec docker-compose.yml.tmpl -i
[docker-compose.yml.tmpl]
missing key found
services:
  redis:
    image: "{{.redis.image}}:{{.redis.tag}}"
value for '.redis.image': redis
value for '.redis.tag': 4.0.11-alpine

version: 3
services:
  web:
    build: .
    ports:
     - "5000:5000"
  redis:
    image: "redis:4.0.11-alpine"
```

2-C. Execute a template using environment variables:
```
$ image=redis tag=4.0.11-alpine tpl exec docker-compose.yml.tmpl --env-prefix redis
version: 3
services:
  web:
    build: .
    ports:
     - "5000:5000"
  redis:
    image: "redis:4.0.11-alpine"
```

3. Show all missing keys:
```
$ tpl keys docker-compose.yml.tmpl
---

redis:
  image: ""
  tag: ""
```

4. Check for missing keys:
```
$ tpl ensure docker-compose.yml.tmpl
missing key found: template: docker-compose.yml.tmpl:8:20: ...

$ echo $?
1
```


## Commands

tpl:

```
Usage:
  tpl [flags]
  tpl [command]

Available Commands:
  completion  Emit bash completion
  ensure      Check for missing keys
  exec        Execute Go templates
  help        Help about any command
  keys        Show all missing keys and processed key:value pairs

Flags:
  -h, --help      help for tpl
  -v, --version   version
```

tpl exec:

```
Execute go templates

Usage:  tpl exec [OPTIONS] TMPL_FILE [TMPL_FILE...] [flags]

Flags:
  -d, --datafile string      Colon separated files containing data objects
  -e, --env                  Load the environment variables into the data objects
  -p, --env-prefix string    Key prefix to load environment variables.
                             If a template key has a dot chain of the given value as a prefix,
                             load the corresponding environment variable into the data objects
  -x, --export-data string   Output file to store the data. Omit to do not store data.
                             The data also contains the values obtained in interactive mode
  -c, --fold-context         Folds the parent context of missing keys when searching.
                             Only meaningful if the template file is yaml|json format
  -f, --format string        Default format for input data file without extention (default "yaml")
  -h, --help                 help for exec
  -i, --interactive          Search for missing keys and input values from the stdin.
                             (Do not support template files including 'Actions' or 'Fuctions')
  -m, --missingkey string    The missingkey gotemplate option (default "error")
  -o, --out string           Output file to store processed templates. Omit to use stdout,
                             but if 'outdir' flag is specified, output will not be stdout
      --outdir string        Directory to store the processed templates.
                             If multiple template files are given, name of each file will be used
                             instead of the 'out' flag ($outdir/$TMPL_FILE_WITHOUT_TMPL_EXT)"
      --overwrite            Overwrite file if it exists
  -s, --show-file            Show processed file info
```

tpl keys:

```
Show all missing keys and processed key:value pairs

Usage:  tpl keys [OPTIONS] TMPL_FILE [TMPL_FILE...] [flags]

Flags:
  -d, --datafile string        Colon separated files containing data objects
                               to execute templates to retrieve processed key:value pairs.
                               Omit to get only the keys of unprocessed TMPL FILES
  -e, --env                    Load the environment variables into the data objects
  -p, --env-prefix string      Key prefix to load environment variables.
                               If a template key has a dot chain of the given value as a prefix,
                               load the corresponding environment variable into the data objects
  -f, --format string          Default format for input data file without extention (default "yaml")
  -h, --help                   help for keys
  -m, --missing                Show only missing keys of processed template.
                               Only used for --datafile is specified
  -o, --out string             Output file to store the generated data. Omit to use stdout
  -t, --output-format string   Output format for data object (default "yaml")
```

