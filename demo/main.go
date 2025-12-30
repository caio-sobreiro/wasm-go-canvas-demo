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
	ctx          js.Value
	particles    []Particle
	frame        int
	mouseX       float64
	mouseY       float64
	mouseDown    bool
	lastTime     float64
	fps          float64
	displayedFPS int
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
		ctx:          ctx,
		particles:    particles,
		lastTime:     js.Global().Get("performance").Call("now").Float(),
		fps:          60.0,
		displayedFPS: 60,
	}
}

func (c *Canvas) update() {
	for i := range c.particles {
		p := &c.particles[i]

		// Mouse interaction - attract/repel particles
		if c.mouseDown {
			dx := c.mouseX - p.x
			dy := c.mouseY - p.y
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist > 0 && dist < 200 {
				// Attract to mouse
				force := (200 - dist) / 200 * 0.5
				p.vx += (dx / dist) * force
				p.vy += (dy / dist) * force
			}
		}

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

		// Add damping
		p.vx *= 0.99
		p.vy *= 0.99

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

	// Draw connections between nearby particles (optimized)
	// Limit to first 150 particles to maintain performance
	maxConnectionParticles := len(c.particles)
	if maxConnectionParticles > 150 {
		maxConnectionParticles = 150
	}

	c.ctx.Set("strokeStyle", "rgba(255, 255, 255, 0.1)")
	c.ctx.Set("lineWidth", 0.5)
	c.ctx.Call("beginPath")

	connectionCount := 0
	maxConnections := 200 // Limit total connections drawn

	for i := 0; i < maxConnectionParticles && connectionCount < maxConnections; i++ {
		for j := i + 1; j < maxConnectionParticles && connectionCount < maxConnections; j++ {
			dx := c.particles[i].x - c.particles[j].x
			dy := c.particles[i].y - c.particles[j].y
			distSq := dx*dx + dy*dy

			if distSq < 10000 { // 100*100, avoid sqrt
				c.ctx.Call("moveTo", c.particles[i].x, c.particles[i].y)
				c.ctx.Call("lineTo", c.particles[j].x, c.particles[j].y)
				connectionCount++
			}
		}
	}
	c.ctx.Call("stroke")

	// Update displayed FPS every 10 frames to reduce blur
	if c.frame%10 == 0 {
		c.displayedFPS = int(c.fps + 0.5)
	}

	// Draw solid background behind text
	c.ctx.Set("fillStyle", "rgba(0, 0, 0, 0.7)")
	c.ctx.Call("fillRect", 5, 5, 120, 50)

	// Draw FPS counter
	c.ctx.Set("fillStyle", "rgba(255, 255, 255, 0.95)")
	c.ctx.Set("font", "14px monospace")
	c.ctx.Call("fillText", fmt.Sprintf("FPS: %d", c.displayedFPS), 10, 20)

	// Draw particle count
	c.ctx.Call("fillText", fmt.Sprintf("Particles: %d", len(c.particles)), 10, 40)

	c.frame++
}

func (c *Canvas) animate() {
	// Calculate FPS
	currentTime := js.Global().Get("performance").Call("now").Float()
	delta := currentTime - c.lastTime
	c.lastTime = currentTime

	if delta > 0 {
		// Smooth FPS using exponential moving average
		instantFPS := 1000.0 / delta
		c.fps = c.fps*0.9 + instantFPS*0.1
	}

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

func handleMouseMove(this js.Value, args []js.Value) interface{} {
	event := args[0]
	canvasInstance.mouseX = event.Get("offsetX").Float()
	canvasInstance.mouseY = event.Get("offsetY").Float()
	return nil
}

func handleMouseDown(this js.Value, args []js.Value) interface{} {
	canvasInstance.mouseDown = true
	return nil
}

func handleMouseUp(this js.Value, args []js.Value) interface{} {
	canvasInstance.mouseDown = false
	return nil
}

func handleClick(this js.Value, args []js.Value) interface{} {
	event := args[0]
	x := event.Get("offsetX").Float()
	y := event.Get("offsetY").Float()

	// Add new particles at click location
	for i := 0; i < 5; i++ {
		canvasInstance.particles = append(canvasInstance.particles, Particle{
			x:   x + (rand.Float64()-0.5)*20,
			y:   y + (rand.Float64()-0.5)*20,
			vx:  (rand.Float64() - 0.5) * 6,
			vy:  (rand.Float64() - 0.5) * 6,
			r:   rand.Float64()*3 + 2,
			hue: rand.Float64() * 360,
		})
	}

	// Cap at 300 particles, remove oldest
	if len(canvasInstance.particles) > 300 {
		canvasInstance.particles = canvasInstance.particles[len(canvasInstance.particles)-300:]
	}

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

	// Add event listeners - all handled in Go!
	canvas.Call("addEventListener", "mousemove", js.FuncOf(handleMouseMove))
	canvas.Call("addEventListener", "mousedown", js.FuncOf(handleMouseDown))
	canvas.Call("addEventListener", "mouseup", js.FuncOf(handleMouseUp))
	canvas.Call("addEventListener", "click", js.FuncOf(handleClick))

	// Export animate function to JavaScript
	js.Global().Set("goAnimate", js.FuncOf(animateFrame))

	// Keep program running
	select {}
}
