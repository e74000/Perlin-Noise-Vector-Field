package main

import (
	"flag"
	"fmt"
	"github.com/aquilax/go-perlin"
	"github.com/hajimehoshi/ebiten"
	"github.com/lucasb-eyer/go-colorful"
	"log"
	"math"
	"math/rand"
	"os"
	"time"
)

type Game struct {
	pixels []byte
	grid   Grid

	sx int
	sy int
	ws int

	mx int
	my int

	ta float64

	enableAgents bool
}

type Grid struct {
	dir []float64
	mag []float64

	xpe perlin.Perlin
	ype perlin.Perlin

	sx int
	sy int

	agents       []Agent
	enableAgents bool

	nScale float64
}

type Vector2 struct {
	x float64
	y float64
}

type Agent struct {
	loc Vector2
	vel Vector2
}

func (v Vector2) mag() float64 {
	return math.Sqrt(math.Pow(v.x, 2) + math.Pow(v.y, 2))
}

func Clamp1(x float64) float64 {
	if x > 1 {
		return 1
	}

	if x < 0 {
		return 0
	}

	return x
}

func (g *Grid) Draw(pixels []byte) {
	var (
		red   float64
		green float64
		blue  float64

		c colorful.Color

		x int
		y int
		l int

		v float64
	)

	for i := range g.dir {
		c = colorful.Hsv((g.dir[i]/math.Pi)*180, 1, g.mag[i])

		red, green, blue = c.R, c.G, c.B

		pixels[4*i+0] = uint8(red * 255)
		pixels[4*i+1] = uint8(green * 255)
		pixels[4*i+2] = uint8(blue * 255)
		pixels[4*i+3] = 0xff
	}

	if g.enableAgents {
		for _, agent := range g.agents {
			x = int(agent.loc.x)
			y = int(agent.loc.y)

			v = agent.vel.mag()
			l = y*g.sx + x

			c = colorful.Hsv((g.dir[l]/math.Pi)*180, Clamp1(1-v), Clamp1(v+g.mag[l]))

			red, green, blue = c.R, c.G, c.B

			pixels[4*l+0] = uint8(red * 255)
			pixels[4*l+1] = uint8(green * 255)
			pixels[4*l+2] = uint8(blue * 255)
			pixels[4*l+3] = 0xff
		}
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.pixels == nil {
		g.pixels = make([]byte, g.sx*g.sy*4)
	}

	g.grid.Draw(g.pixels)
	err := screen.ReplacePixels(g.pixels)

	if err != nil {
		return
	}
}

func (g *Game) Layout(_, _ int) (screenWidth, screenHeight int) {
	return g.sx, g.sy
}

func (g *Game) Update(_ *ebiten.Image) error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) || ebiten.IsKeyPressed(ebiten.KeyQ) {
		os.Exit(0)
	}

	var (
		v Vector2
		a float64
		m float64
	)

	for y := 0; y < g.sy; y++ {
		for x := 0; x < g.sx; x++ {
			v.x = g.grid.xpe.Noise3D(float64(x)/g.grid.nScale, float64(y)/g.grid.nScale, g.ta)
			v.y = g.grid.ype.Noise3D(float64(x)/g.grid.nScale, float64(y)/g.grid.nScale, g.ta)

			a = math.Atan2(v.y, v.x) + math.Pi
			m = v.mag()

			g.grid.dir[y*g.sx+x] = a
			g.grid.mag[y*g.sx+x] = m
		}
	}

	if g.enableAgents {
		for i, agent := range g.grid.agents {
			a = g.grid.dir[int(agent.loc.y)*g.sx+int(agent.loc.x)]
			m = g.grid.dir[int(agent.loc.y)*g.sx+int(agent.loc.x)]

			g.grid.agents[i].vel.x += math.Cos(a) * m * 0.001
			g.grid.agents[i].vel.y += math.Sin(a) * m * 0.001

			g.grid.agents[i].loc.x += g.grid.agents[i].vel.x
			g.grid.agents[i].loc.y += g.grid.agents[i].vel.y

			if g.grid.agents[i].loc.x < 0 || g.grid.agents[i].loc.x > float64(g.sx) || g.grid.agents[i].loc.y < 1 || g.grid.agents[i].loc.y > float64(g.sy) {
				g.grid.agents[i].loc.x = float64(rand.Int() % g.sx)
				g.grid.agents[i].loc.y = float64(rand.Int() % g.sy)
				g.grid.agents[i].vel.x = 0
				g.grid.agents[i].vel.y = 0
			}

			if g.grid.agents[i].vel.mag() > 1 {
				g.grid.agents[i].vel.x = math.Cos(a)
				g.grid.agents[i].vel.y = math.Sin(a)
			}
		}
	}

	g.ta += 0.005

	return nil
}

func main() {
	var enableAgents bool

	flag.BoolVar(&enableAgents, "a", true, "Enable agents")

	flag.Parse()

	fmt.Println(enableAgents)

	g := &Game{
		sx:           320,
		sy:           180,
		ws:           6,
		enableAgents: enableAgents,
	}

	g.grid = Grid{
		mag: make([]float64, g.sx*g.sy),
		dir: make([]float64, g.sx*g.sy),

		xpe: *perlin.NewPerlin(2, 1, 1, time.Now().Unix()),
		ype: *perlin.NewPerlin(2, 1, 1, time.Now().Unix()+1),

		agents:       make([]Agent, 2000),
		enableAgents: enableAgents,

		sx: g.sx,
		sy: g.sy,

		nScale: 125,
	}

	rand.Seed(time.Now().Unix() + 2)

	if enableAgents {
		for i := range g.grid.agents {
			g.grid.agents[i].loc.x = float64(rand.Int() % g.sx)
			g.grid.agents[i].loc.y = float64(rand.Int() % g.sy)
		}
	}

	ebiten.SetWindowSize(g.sx*g.ws, g.sy*g.ws)
	ebiten.SetFullscreen(true)
	ebiten.SetWindowTitle("Pixel Screensaver")
	ebiten.SetCursorMode(ebiten.CursorModeHidden)

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
