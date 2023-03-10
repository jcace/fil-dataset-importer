# Installation

Building from Source
1. Clone `git clone https://github.com/jcace/fil-dataset-importer.git` 
2. `make all`
3. `make install`


# Usage

```
NAME:
   Filecoin Offline Dataset Importer 

USAGE:
   Filecoin Offline Dataset Importer [global options] command [command options] [arguments...]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --boost value           192.168.1.1
   --debug                 set to enable debug logging output (default: false)
   --dir value             /home/filecoin/path/to/mount
   --gql value             8080 (default: 8080)
   --help, -h              show help (default: false)
   --interval value        interval, in seconds, to re-run the importer (default: 0)
   --key value             eyJ....XXX
   --max_concurrent value  stop importing if # of deals in AP or PC1 are above this threshold. 0 = unlimited. (default: 0)
   --port value            1288 (default: 1288)
```

## Example run
The followng command will import a deal every 240 seconds, until there are 80 deals currently in the AP/PC1 state. Then, it will stop untill some deals clear out. 

`fil-dataset-importer --boost 192.168.1.1 --port 1288 --gql 8080 --key eyJ...XXX --dir /home/filecoin/datasets --interval 240 --max_concurrent 80 --debug`


## Datasets Config
You must provide the tool with a file named `datasets.json` , in the same directory that the command is being executed from. This file maintains a mapping between client `wallets` (i.e, who is making deals) with a `dataset slug`. This dataset slug is appended to the `--dir` flag when importing data. 

For example, given a `datasets.json` that looks like this:
```json
{
   "f1234": "test-dataset"
}
```

And a `--dir` flag or `/home/filecoin/datasets`, 

Then, when an offline deal comes in from address `f1234`, the importer will search in the directory `/home/filecoin/datasets/test-dataset` for a CAR file `<pieceCID>.car` to import. 

You can find an example of the datasets file in `sample_datasets.json`