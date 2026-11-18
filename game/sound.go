package game

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/RajaeDev6/bombo/interpreter"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

const sampleRate = 44100

var audioContext *audio.Context

// Sound is a sound file loaded with game.sound(path).
type Sound struct {
	path string
	pcm  []byte // 16-bit stereo at sampleRate
}

func (s *Sound) TypeName() string { return "sound" }
func (s *Sound) String() string   { return "<sound " + s.path + ">" }
func (s *Sound) Field(name string) V {
	if name == "path" {
		return s.path
	}
	return nil
}

// sound(path) loads a .wav, .ogg or .mp3 file.
func (l *lib) loadSound(args []V, line int) (V, error) {
	if len(args) != 1 {
		return nil, errorf(line, "'game.sound' want a file name: game.sound(\"pew.wav\")")
	}
	path := interpreter.ToString(args[0])
	full := path
	if !filepath.IsAbs(full) {
		full = filepath.Join(l.it.Dir, full)
	}
	data, err := os.ReadFile(full)
	if err != nil {
		return nil, errorf(line, "Mi cyaan find di sound '%s'", path)
	}

	var stream io.Reader
	r := bytes.NewReader(data)
	switch strings.ToLower(filepath.Ext(path)) {
	case ".wav":
		stream, err = wav.DecodeWithSampleRate(sampleRate, r)
	case ".ogg":
		stream, err = vorbis.DecodeWithSampleRate(sampleRate, r)
	case ".mp3":
		stream, err = mp3.DecodeWithSampleRate(sampleRate, r)
	default:
		return nil, errorf(line, "Mi can only play .wav, .ogg an .mp3 sounds, not '%s'", path)
	}
	if err == nil {
		var pcm []byte
		if pcm, err = io.ReadAll(stream); err == nil {
			return &Sound{path: path, pcm: pcm}, nil
		}
	}
	return nil, errorf(line, "Di sound '%s' look broken: %s", path, err)
}

// play(sound) or play(sound, volume) plays a loaded sound.
func (l *lib) play(args []V, line int) (V, error) {
	if len(args) != 1 && len(args) != 2 {
		return nil, errorf(line, "'game.play' want a sound an maybe a volume: game.play(pew) or game.play(pew, 0.5)")
	}
	s, ok := args[0].(*Sound)
	if !ok {
		return nil, errorf(line, "'game.play' want a sound from game.sound(...), not a %s", interpreter.TypeName(args[0]))
	}
	vol := l.volume
	if len(args) == 2 {
		f, err := nums("play", args[1:], 1, line)
		if err != nil {
			return nil, err
		}
		vol *= f[0]
	}
	if vol <= 0 {
		return nil, nil
	}
	if audioContext == nil {
		audioContext = audio.NewContext(sampleRate)
	}
	p := audioContext.NewPlayerFromBytes(s.pcm)
	p.SetVolume(vol)
	p.Play()
	return nil, nil
}
