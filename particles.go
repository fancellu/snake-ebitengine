package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"image/color"
	"math"
	"math/rand/v2"
)

var (
	maxLife = 25
)

type Particle struct {
	x, y   float64
	dx, dy float64
	life   int
	color  color.Color
}

type ParticleSystem struct {
	particles []Particle
}

func NewParticleSystem() *ParticleSystem {
	return &ParticleSystem{
		particles: make([]Particle, 0, 100),
	}
}

func (ps *ParticleSystem) Spawn(x, y float64, pcolor color.Color) {
	// Create 10 particles in a burst
	for i := 0; i < 10; i++ {
		angle := rand.Float64() * 2 * math.Pi
		speed := 2 + rand.Float64()*2
		particle := Particle{
			x:     x,
			y:     y,
			dx:    math.Cos(angle) * speed,
			dy:    math.Sin(angle) * speed,
			life:  maxLife,
			color: pcolor,
		}
		ps.particles = append(ps.particles, particle)
	}
}

func (ps *ParticleSystem) Update() {
	var alive []Particle
	for _, p := range ps.particles {
		p.x += p.dx
		p.y += p.dy
		p.life--

		if p.life > 0 {
			alive = append(alive, p)
		}
	}
	ps.particles = alive
}

func (ps *ParticleSystem) Draw(screen *ebiten.Image) {
	for _, p := range ps.particles {
		// Get the original color components
		r, g, b, _ := p.color.RGBA()

		// Convert from color.Color's 16-bit per channel to 8-bit per channel
		// and calculate alpha based on remaining life
		alpha := uint8(float64(p.life) / float64(maxLife) * 255)

		col := color.RGBA{
			R: uint8(r >> 8),
			G: uint8(g >> 8),
			B: uint8(b >> 8),
			A: alpha,
		}

		vector.DrawFilledCircle(screen, float32(p.x), float32(p.y), 2, col, true)
	}
}
