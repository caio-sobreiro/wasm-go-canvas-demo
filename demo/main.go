package main

import (
	"fmt"
	"math"
	"math/rand"
	"syscall/js"
	"time"
)

const (
	width        = 800
	height       = 600
	numParticles = 100
)

type Particle struct {
	x, y   float64
	vx, vy float64
	r      float64
	hue    float64
}

type Canvas struct {
	ctx       js.Value
	particles []Particle
	frame     int
}

func NewCanvas(ctx js.Value) *Canvas {
	rand.Seed(time.Now().UnixNano())

	particles := make([]Particle, numParticles)
	for i := range particles {
		particles[i] = Particle{
			x:   rand.Float64() * width,
			y:   rand.Float64() * height,
			vx:  (rand.Float64() - 0.5) * 3,
			vy:  (rand.Float64() - 0.5) * 3,
			r:   rand.Float64()*3 + 2,
			hue: rand.Float64() * 360,
		}
	}

	return &Canvas{
		ctx:       ctx,
		particles: particles,
	}
}

func (c *Canvas) update() {
	for i := range c.particles {
		p := &c.particles[i]

		// Update position
		p.x += p.vx
		p.y += p.vy

		// Bounce off walls
		if p.x < 0 || p.x > width {
			p.vx = -p.vx
			p.x = math.Max(0, math.Min(width, p.x))
		}
		if p.y < 0 || p.y > height {
			p.vy = -p.vy
			p.y = math.Max(0, math.Min(height, p.y))
		}

		// Apply slight gravity
		p.vy += 0.05

		// Cycle hue
		p.hue += 0.5
		if p.hue > 360 {
			p.hue -= 360
		}
	}
}

func (c *Canvas) draw() {
	// Clear canvas with semi-transparent black for trail effect
	c.ctx.Set("fillStyle", "rgba(0, 0, 0, 0.1)")
	c.ctx.Call("fillRect", 0, 0, width, height)

	// Draw particles
	for _, p := range c.particles {
		c.ctx.Set("fillStyle", hslToString(p.hue, 100, 60))
		c.ctx.Call("beginPath")
		c.ctx.Call("arc", p.x, p.y, p.r, 0, 2*math.Pi)
		c.ctx.Call("fill")
	}

	// Draw connections between nearby particles
	c.ctx.Set("strokeStyle", "rgba(255, 255, 255, 0.1)")
	c.ctx.Set("lineWidth", 0.5)

	for i := 0; i < len(c.particles); i++ {
		for j := i + 1; j < len(c.particles); j++ {
			dx := c.particles[i].x - c.particles[j].x
			dy := c.particles[i].y - c.particles[j].y
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist < 100 {
				c.ctx.Call("beginPath")
				c.ctx.Call("moveTo", c.particles[i].x, c.particles[i].y)
				c.ctx.Call("lineTo", c.particles[j].x, c.particles[j].y)
				c.ctx.Call("stroke")
			}
		}
	}

	c.frame++
}

func (c *Canvas) animate() {
	c.update()
	c.draw()
}

func hslToString(h, s, l float64) string {
	return fmt.Sprintf("hsl(%.0f, %.0f%%, %.0f%%)", h, s, l)
}

var canvasInstance *Canvas

func animateFrame(this js.Value, args []js.Value) interface{} {
	canvasInstance.animate()
	return nil
}

func main() {
	// Get canvas and context
	doc := js.Global().Get("document")
	canvas := doc.Call("getElementById", "canvas")
	ctx := canvas.Call("getContext", "2d")

	// Set canvas size
	canvas.Set("width", width)
	canvas.Set("height", height)

	// Create canvas manager
	canvasInstance = NewCanvas(ctx)

	// Export animate function to JavaScript
	js.Global().Set("goAnimate", js.FuncOf(animateFrame))

	// Keep program running
	select {}
}
