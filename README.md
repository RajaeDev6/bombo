# Bombo

A small interpreted programming language with Jamaican Patois keywords, written in Go.

```bombo
wah gwaan

seh "Hello, world!"
```

## Installation

### Requirements

- [Go](https://go.dev/dl/) 1.22 or newer
- Git

### Get the source

```sh
git clone https://github.com/RajaeDev6/bombo.git
cd bombo
```

### Build

```sh
go build -o bombo .
```

This creates a `bombo` executable (`bombo.exe` on Windows) in the current directory.

### Install system-wide (optional)

To run `bombo` from anywhere, move the binary somewhere on your `PATH`:

```sh
# Linux / macOS
sudo mv bombo /usr/local/bin/
```

On Windows, put `bombo.exe` in a folder that is listed in your `PATH` environment variable.

### Run the tests

```sh
go test ./...
```

## Usage

```sh
bombo yourfile.bombo   # run a program
bombo                  # start the REPL (type 'likkle more' to exit)
bombo help             # language reference
bombo version          # show the version
```

Try one of the examples:

```sh
bombo examples/hello.bombo
```

## Keywords

| Bombo | Meaning |
|---|---|
| `mek` | declare a variable |
| `seh` | print |
| `bombo` | if |
| `bomboclaat` | else if |
| `bloodclaat` | else |
| `gwaan` | while loop |
| `fi i from 1 tuh 10` | counting loop (inclusive) |
| `fi x inna list` | for-each loop |
| `rassclaat` | break |
| `gallang` | continue |
| `wuk` | define a function |
| `sen back` | return |
| `tek a chance` | try |
| `if it bruk` | catch |
| `bruk` | throw an error |
| `big up` / `bun` | increment / decrement |
| `yeh mon` / `nah mon` | true / false |
| `kakafaat` | nil |
| `an` / `or` / `nuh` | and / or / not |
| `watch yah` or `#` | comment |
| `likkle more` | exit |
| `wah gwaan` | optional greeting (does nothing) |

## Built-in functions

| Function | Description |
|---|---|
| `howmuch(x)` | length of a string, list or map |
| `ax(prompt)` | read a line of input |
| `pon(list, x)` / `tekoff(list)` | push / pop |
| `wahkind(x)` | type of a value |
| `numba(x)` / `wud(x)` | convert to number / string |
| `luck(lo, hi)` | random integer |

Run `./bombo help` for the full list.

## Example

```bombo
wuk fizzbuzz(n) {
    fi i from 1 tuh n {
        bombo i % 15 == 0 {
            seh "BOMBOCLAAT!"
        } bomboclaat i % 3 == 0 {
            seh "Bombo"
        } bomboclaat i % 5 == 0 {
            seh "Claat"
        } bloodclaat {
            seh i
        }
    }
}

fizzbuzz(15)
```

More examples are in [`examples/`](examples).

## Errors

Errors report the line number and a short message:

```
BOMBOCLAAT! Line 2: Mi cyaan find 'x'. Yuh fi 'mek' it first
```
