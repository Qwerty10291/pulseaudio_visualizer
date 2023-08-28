package main

import (
	_ "embed"
	"image/color"
	"math"
	"pulseaudio_visualizer/audio"

	"github.com/hajimehoshi/ebiten/v2"
)

var (
	//go:embed bar.kage
	barShader []byte
)

type Game struct {
	AudioProcessor *audio.AudioProcessor
	Samples        []float64
	Shader         *ebiten.Shader
}



func NewGame() *Game {
	shader, err := ebiten.NewShader(barShader)
	check(err)
	game := Game{
		Shader: shader,
	}
	processor, err := audio.NewAudioProcessor(4096, func(f []float64) {
		game.Samples = f
		ebiten.ScheduleFrame()
	})
	check(err)
	processor.Start()
	game.AudioProcessor = processor
	return &game
}

func (g *Game) Update() error {
	_, dy := ebiten.Wheel()
	if ebiten.IsKeyPressed(ebiten.KeyV){
		g.AudioProcessor.VolumeMult += float32(0.5 * dy)
		if g.AudioProcessor.VolumeMult < 0{
			g.AudioProcessor.VolumeMult = 0
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		g.AudioProcessor.SmoothStep += 0.5 * dy
		if g.AudioProcessor.SmoothStep < 1{
			g.AudioProcessor.SmoothStep = 1
		}
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	width, height := float64(screen.Bounds().Dx()), float64(screen.Bounds().Dy())
	screen.Fill(color.Black)

	cellWidth := math.Floor(width / float64(len(g.Samples)))
	for i := 0; i < len(g.Samples); i++ {
		x := float64(g.Samples[i])
		h := height * (1.0 - math.Pow(1-x/2, 4))
		// h := height * x
		var options ebiten.DrawRectShaderOptions
		options.Uniforms = map[string]interface{}{"Size": []float64{width, height}}
		options.GeoM.Translate(float64(i)*cellWidth, height-h)
		screen.DrawRectShader(int(cellWidth), int(height), g.Shader, &options)
	}
}

func (*Game) Layout(outsideWidth int, outsideHeight int) (screenWidth int, screenHeight int) {
	return outsideWidth, outsideHeight
}

func main() {
	ebiten.SetScreenTransparent(true)
	ebiten.SetWindowResizable(true)
	ebiten.SetWindowDecorated(false)
	game := NewGame()
	check(ebiten.RunGame(game))
}

// Helper function to check on any error.
func check(err error) {
	if err != nil {
		panic(err)
	}
}
