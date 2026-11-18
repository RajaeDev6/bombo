# Bombo

A small interpreted programming language with Jamaican Patois keywords, written in Go.

```bombo
wah gwaan

seh "Hello, world!"
```

## Installation

### Requirements

- [Go](https://go.dev/dl/) 1.26 or newer
- Git (only needed to build from source)
- On Linux, the X11/OpenGL and ALSA development packages, which the `game` library needs to build. See [Ebitengine's install guide](https://ebitengine.org/en/documents/install.html) for your distro. Windows and macOS need nothing extra.

### Quick install

```sh
go install github.com/RajaeDev6/bombo@latest
```

This downloads, builds and installs `bombo` into `$(go env GOPATH)/bin` (usually `~/go/bin`). Make sure that folder is on your `PATH`, then run `bombo` from anywhere.

### Or build from source

#### Get the source

```sh
git clone https://github.com/RajaeDev6/bombo.git
cd bombo
```

#### Build

```sh
go build -o bombo .
```

This creates a `bombo` executable (`bombo.exe` on Windows) in the current directory.

#### Install system-wide (optional)

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
| `tek "name"` | import a module |
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
| `luck(lo, hi)` | random integer (`luck()` gives a decimal between 0 and 1) |
| `min` / `max` | smallest / biggest of some numbers or a list |
| `sin` `cos` `tan` `atan2` `pi` | trigonometry |

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

## Maps and dot access

Map fields can be read and set with a dot, as well as with `[ ]`:

```bombo
mek player = {name: "Ann", score: 0}
player.score += 10
seh player.name, player["score"]
```

## Modules

`tek` loads a module and stores it in a variable named after it.

Split a program into files. `tek "shapes"` runs `shapes.bombo` (next to the current file) once and gives you its variables and functions:

```bombo
# shapes.bombo
wuk area(w, h) {
    sen back w * h
}
```

```bombo
# main.bombo
tek "shapes"
seh shapes.area(3, 4)

tek "lib/maths-helpers.bombo" as helpers   # paths an 'as' work too
```

Bombo also comes with a library module, `game`.

## Making games

The `game` library gives Bombo a window, images, shapes, text, keyboard input, sounds and saving. It's similar to pygame or LÖVE. Write the game in a `.bombo` file and run it with `bombo`, like any other program:

```bombo
tek "game"

game.window(480, 360, "Catch")

mek x = 240

wuk update(dt) {
    bombo game.held("left") { x -= 300 * dt }
    bombo game.held("right") { x += 300 * dt }
}

wuk draw() {
    game.clear("#101025")
    game.rect(x - 30, 320, 60, 10, "cyan")
    game.text("Hello from Bombo", 10, 10, 18, "white")
}

game.run(update, draw)
```

`game.run(update, draw)` opens the window. It calls `update(dt)` 60 times a second (`dt` is the time step in seconds) and `draw()` every frame. It returns when the window is closed or the game calls `game.quit()`. Errors inside your game come back as normal Bombo errors with the line number, so `tek a chance` works around `game.run` too.

The full version of this example is [`examples/catch.bombo`](examples/catch.bombo).

Games made with Bombo:

- [Starfall](https://github.com/RajaeDev6/starfall): a retro space shooter

### Window

| Function | Description |
|---|---|
| `game.window(w, h, title)` | Set the window size and title |
| `game.run(update, draw)` | Start the game loop |
| `game.quit()` | Close the window; `game.run` then returns |
| `game.width()` / `game.height()` | Screen size |

### Images

| Function | Description |
|---|---|
| `game.image(path)` | Load a `.png` or `.jpg`. Paths are relative to your `.bombo` file. The image has `.width` and `.height` |
| `game.draw(img, x, y)` | Draw an image with its top-left corner at `x, y` |
| `game.draw(img, x, y, options)` | Same, with a map of options: `center` (put the middle at `x, y`), `scale`, `angle` (radians), `alpha` (0 to 1), `tint` (a colour), `flip` (mirror left to right), `smooth` (smooth scaling instead of crisp pixels) |

```bombo
mek ship = game.image("ship.png")
game.draw(ship, 100, 200, {center: yeh mon, scale: 2, angle: pi / 4})
```

### Shapes and text

| Function | Description |
|---|---|
| `game.clear(colour)` | Fill the screen |
| `game.rect(x, y, w, h, colour)` | Filled rectangle |
| `game.box(x, y, w, h, thickness, colour)` | Rectangle outline |
| `game.circle(x, y, r, colour)` | Filled circle |
| `game.ring(x, y, r, thickness, colour)` | Circle outline |
| `game.line(x1, y1, x2, y2, thickness, colour)` | Line |
| `game.triangle(x1, y1, x2, y2, x3, y3, colour)` | Filled triangle |
| `game.text(words, x, y, size, colour)` | Text. Add `"center"` or `"right"` as a 6th argument to align it |
| `game.camera(x, y)` | Shift everything drawn after it, for scrolling or screen shake. Resets each frame |
| `game.screenshot(path)` | Save the screen so far to a PNG |

Drawing only works inside the `draw` function.

A colour can be a name (`"white"`, `"black"`, `"red"`, `"green"`, `"blue"`, `"cyan"`, `"yellow"`, `"gold"`, `"orange"`, `"purple"`, `"pink"`, `"grey"`), a hex string (`"#ff8800"`, `"#f80"`, or `"#ff880080"` with alpha), or `game.rgb(r, g, b)` / `game.rgb(r, g, b, alpha)` with values from 0 to 255.

### Keyboard

| Function | Description |
|---|---|
| `game.held(key)` | `yeh mon` while the key is down |
| `game.tapped(key)` | `yeh mon` only on the frame the key was pressed |

Key names: `"left"`, `"right"`, `"up"`, `"down"`, `"space"`, `"enter"`, `"escape"`, `"shift"`, `"ctrl"`, `"tab"`, `"backspace"`, `"a"` to `"z"` and `"0"` to `"9"`.

### Sound

| Function | Description |
|---|---|
| `game.sound(path)` | Load a `.wav`, `.ogg` or `.mp3` |
| `game.play(sound)` | Play it. `game.play(sound, 0.5)` plays it at half volume |
| `game.volume(v)` | Overall volume, from 0 (silent) to 1 |

### Saving

| Function | Description |
|---|---|
| `game.save(key, value)` | Keep a number, string or truth value between runs |
| `game.load(key)` | Get it back, or `kakafaat` if nothing was saved |

Each game gets its own save file, named after its window title, in your config folder (for example `~/.config/bombo/starfall.json` on Linux).

## Errors

Errors report the line number and a short message:

```
BOMBOCLAAT! Line 2: Mi cyaan find 'x'. Yuh fi 'mek' it first
```
