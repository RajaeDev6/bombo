# Bombo

A small interpreted programming language with Jamaican Patois keywords, written in Go.

```bombo
wah gwaan

seh "Hello, world!"
```

## Getting started

```sh
go build -o bombo .
./bombo yourfile.bombo
```

Run `./bombo` with no arguments to start the REPL. Type `likkle more` to exit.

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
