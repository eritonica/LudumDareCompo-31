package main

// Small stand-ins for the SFML types the game was originally written
// against (RectangleShape, Text, Music), implemented on top of Ebitengine.

import (
	"bytes"
	"embed"
	"image"
	_ "image/png"
	"io"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

//go:embed *.png *.wav deja.ttf
var assets embed.FS

type vec2 struct {
	X, Y float64
}

type rect struct {
	X, Y, W, H float64
}

func (r rect) Intersects(o rect) bool {
	return r.X < o.X+o.W && o.X < r.X+r.W && r.Y < o.Y+o.H && o.Y < r.Y+r.H
}

// shape mirrors sf.RectangleShape with a texture stretched over its size.
type shape struct {
	Image    *ebiten.Image
	Size     vec2
	Position vec2
	Origin   vec2
	Rotation float64 // degrees, clockwise
}

func newShape(img *ebiten.Image, size, pos vec2) *shape {
	return &shape{Image: img, Size: size, Position: pos}
}

func (s *shape) Move(d vec2) {
	s.Position.X += d.X
	s.Position.Y += d.Y
}

func (s *shape) Rotate(deg float64) {
	s.Rotation = math.Mod(s.Rotation+deg, 360)
}

func (s *shape) transform() ebiten.GeoM {
	var g ebiten.GeoM
	g.Translate(-s.Origin.X, -s.Origin.Y)
	g.Rotate(s.Rotation * math.Pi / 180)
	g.Translate(s.Position.X, s.Position.Y)
	return g
}

// GlobalBounds is the axis-aligned bounding box of the transformed shape.
func (s *shape) GlobalBounds() rect {
	g := s.transform()
	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)
	for _, c := range [4]vec2{{0, 0}, {s.Size.X, 0}, {0, s.Size.Y}, {s.Size.X, s.Size.Y}} {
		x, y := g.Apply(c.X, c.Y)
		minX, maxX = math.Min(minX, x), math.Max(maxX, x)
		minY, maxY = math.Min(minY, y), math.Max(maxY, y)
	}
	return rect{minX, minY, maxX - minX, maxY - minY}
}

func (s *shape) Draw(dst *ebiten.Image) {
	b := s.Image.Bounds()
	var g ebiten.GeoM
	g.Scale(s.Size.X/float64(b.Dx()), s.Size.Y/float64(b.Dy()))
	g.Concat(s.transform())
	op := &ebiten.DrawImageOptions{GeoM: g, Filter: ebiten.FilterLinear}
	dst.DrawImage(s.Image, op)
}

// label mirrors sf.Text: black text drawn with its top-left at Position.
type label struct {
	Face     *text.GoTextFace
	String   string
	Position vec2
}

func (l *label) Draw(dst *ebiten.Image) {
	op := &text.DrawOptions{}
	m := l.Face.Metrics()
	op.LineSpacing = m.HAscent + m.HDescent + m.HLineGap
	op.GeoM.Translate(l.Position.X, l.Position.Y)
	op.ColorScale.Scale(0, 0, 0, 1)
	text.Draw(dst, l.String, l.Face, op)
}

func loadImage(name string) (*ebiten.Image, error) {
	data, err := assets.ReadFile(name)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return ebiten.NewImageFromImage(img), nil
}

func loadFace(name string, size float64) (*text.GoTextFace, error) {
	data, err := assets.ReadFile(name)
	if err != nil {
		return nil, err
	}
	src, err := text.NewGoTextFaceSource(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return &text.GoTextFace{Source: src, Size: size}, nil
}

const sampleRate = 44100

var audioContext = audio.NewContext(sampleRate)

// sound mirrors sf.Music: Play restarts it from the beginning.
type sound struct {
	player *audio.Player
}

func loadSound(name string) (*sound, error) {
	data, err := assets.ReadFile(name)
	if err != nil {
		return nil, err
	}
	stream, err := wav.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	pcm, err := io.ReadAll(stream)
	if err != nil {
		return nil, err
	}
	return &sound{player: audioContext.NewPlayerFromBytes(pcm)}, nil
}

func (s *sound) Play() {
	_ = s.player.SetPosition(0)
	s.player.Play()
}
