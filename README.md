# def

Definitions fetched from [vocabulary.com](https://vocabulary.com) are printed and stored offline.

## Installation

```sh
go get github.com/nilsocket/def
```

## Usage

All valid words, searched are stored offline.

```sh
def         # list of words available offline

def -lp     # Long format and play audio
def bharat  # Short format by default

def -h      # help

def dslkdfj # invalid word, would give word suggestions
```

## External Dependencies

`mpg123` for playing audio

## Internal Dependencies

[Goquery - PuerkitoBio](github.com/PuerkitoBio/goquery)  
[Badger - Dgraph](github.com/dgraph-io/badger/v2)  
[Cli - Urfave](github.com/urfave/cli/v2)
