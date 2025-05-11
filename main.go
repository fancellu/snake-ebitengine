package main

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"image/color"
	_ "image/png"
	"log"
	"math"
	"math/rand/v2"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed food.mp3
var foodSound []byte

//go:embed gameover.mp3
var gameoverSound []byte

var gameSpeed = time.Second / 5

func init() {
	s, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.MPlus1pRegular_ttf))
	if err != nil {
		log.Fatal(err)
	}

	scoreface = &text.GoTextFace{
		Source: s,
		Size:   20,
	}

	face = &text.GoTextFace{
		Source: s,
		Size:   28,
	}
}

const (
	acceleration = 0.95
	screenWidth  = 640
	screenHeight = 480
	gridSize     = 20
	gridWidth    = screenWidth / gridSize
	gridHeight   = screenHeight / gridSize
	tongueLength = 8
)

type Point struct {
	x int
	y int
}

type Game struct {
	snake        []Point
	direction    Point
	lastUpdate   time.Time
	food         *Point
	foodWas      *Point
	gameOver     bool
	paused       bool
	particles    *ParticleSystem
	soundManager *SoundManager
	buttons      []*Button

	tongueCooldown int
	score          int
}

func NewGame() *Game {
	g := &Game{
		particles:    NewParticleSystem(),
		soundManager: NewSoundManager(5),
	}
	g.initGame()
	return g
}

func (g *Game) initGame() {
	gameSpeed = time.Second / 5
	startX := gridWidth / 2
	startY := gridHeight / 2
	snake := []Point{
		{startX, startY},
		{startX - 1, startY},
		{startX - 2, startY},
	}
	g.gameOver = false
	g.snake = snake
	g.direction = Point{1, 0}
	g.lastUpdate = time.Now()
	g.food = nil
	g.foodWas = nil
	g.paused = false
	g.score = 0
	g.tongueCooldown = 0

	newButton := NewButton(0, 0, " Restart ", face, func(me *Button) {
		g.initGame()
	})

	g.buttons = nil
	g.buttons = append(g.buttons, newButton)

	pauseButton := NewButton(0, 0, " Pause ", face, func(me *Button) {
		g.paused = !g.paused
		if g.paused {
			me.SetText(" Resume ", screenWidth)
		} else {
			me.SetText(" Pause ", screenWidth)
		}
	})

	pauseButton.right = float32(screenWidth / 4)
	pauseButton.Align(screenWidth)
	pauseButton.hotkey = ebiten.KeyP

	g.buttons = append(g.buttons, pauseButton)

	muteButton := NewButton(0, 0, " Mute ", face, func(me *Button) {
		g.soundManager.muted = !g.soundManager.muted
		if g.soundManager.muted {
			me.SetText(" Unmute ", screenWidth)

		} else {
			me.SetText(" Mute ", screenWidth)
		}
	})

	muteButton.hotkey = ebiten.KeyM
	muteButton.right = 0.0
	muteButton.Align(screenWidth)
	g.buttons = append(g.buttons, muteButton)

}

func (g *Game) updateSnake(snake *[]Point, direction Point) {
	head := (*snake)[0]
	newHead := Point{head.x + direction.x, head.y + direction.y}
	*snake = append([]Point{newHead}, (*snake)[:len(*snake)-1]...)
}

func (g *Game) removeFood() {
	g.foodWas = g.food
	g.food = nil
}

func (g *Game) shootTongue() {
	head := g.snake[0]

	// Calculate end point based on current direction
	endPoint := Point{head.x, head.y}
	for i := 0; i < tongueLength; i++ {
		endPoint.x += g.direction.x
		endPoint.y += g.direction.y

		// Check if tongue reaches food
		if g.food != nil && endPoint == *g.food {
			g.particles.Spawn(float64(g.food.x*gridSize+gridSize/2), float64(g.food.y*gridSize+gridSize/2),
				color.RGBA{255, 220, 100, 255})
			g.removeFood()
			g.score++
			g.growSnake()
			break
		}
	}

	// Visual and sound effects
	//g.particles.Emit(head.x, head.y) // Optional: emit particles when shooting
	//if g.soundManager != nil {
	//	g.soundManager.PlaySound("tongue") // Add appropriate sound effect
	//}

	g.tongueCooldown = 60 // Adjust cooldown frames as needed
}

func (g *Game) growSnake() {
	// Don't grow if already quite long, will still get faster though!!!
	if len(g.snake) < 40 {
		g.snake = append(g.snake, g.snake[len(g.snake)-1])
	}
	g.playSound(foodSound)
	// Gets faster as it eats!!
	gameSpeed = time.Duration(int64(float64(gameSpeed) * acceleration))
}

func (g *Game) Update() error {

	g.particles.Update()
	for _, b := range g.buttons {
		b.Update()
	}
	if g.gameOver {
		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			g.initGame()
		}
		return nil
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) && g.tongueCooldown == 0 {
		g.shootTongue()
	}
	if g.tongueCooldown > 0 {
		g.tongueCooldown--
	}

	// We want direction change to be instant, not rate limited, else direction changes ignored!!
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		g.direction = Point{1, 0}
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		g.direction = Point{-1, 0}
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		g.direction = Point{0, -1}
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		g.direction = Point{0, 1}
	}

	if g.paused {
		return nil
	}

	if time.Since(g.lastUpdate) < gameSpeed {
		return nil
	}
	g.lastUpdate = time.Now()
	g.updateSnake(&g.snake, g.direction)

	// handle collisions
	if g.checkCollision(g.snake[0]) {
		g.particles.Spawn(float64(g.snake[0].x*gridSize+gridSize/2), float64(g.snake[0].y*gridSize+gridSize/2),
			color.RGBA{255, 0, 0, 255})
		g.playSound(gameoverSound)
		g.gameOver = true
	} else if g.checkFoodCollision() {
		g.growSnake()
		// You get more scores for eating without tongue
		g.score += 2
	} else {
		err := g.generateFood()
		if err != nil {
			return err
		}
	}

	// randomly remove food, such is life!
	if rand.IntN(1000) < 10 && g.food != nil {
		// spawn particles when food disappears
		g.particles.Spawn(float64(g.food.x*gridSize+gridSize/2), float64(g.food.y*gridSize+gridSize/2), GREEN)
		g.removeFood()
	}

	return nil
}

func (g *Game) playSound(sound []byte) {
	if err := g.soundManager.PlaySound(sound); err != nil {
		log.Println("Failed to play sound:", err)
	}
}

var (
	RED       = color.RGBA{0xFF, 0, 0, 0xff}
	BLUE      = color.RGBA{0x0, 0, 0xFF, 0xff}
	GREEN     = color.RGBA{0x0, 0xFF, 0x00, 0xff}
	tailcolor = color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}
	face      *text.GoTextFace
	scoreface *text.GoTextFace
)

func (g *Game) drawTongue(screen *ebiten.Image) {
	if g.tongueCooldown > 40 { // Only show tongue for a few frames
		head := g.snake[0]
		startX := float32(head.x*gridSize) + gridSize/2
		startY := float32(head.y*gridSize) + gridSize/2

		// Debug prints for visual verification

		// Calculate end point based on direction
		endX := startX + float32(g.direction.x*gridSize*tongueLength)
		endY := startY + float32(g.direction.y*gridSize*tongueLength)

		// If foodWas exists, check if tongue intersects with it
		if g.foodWas != nil {

			foodX := float32(g.foodWas.x*gridSize) + gridSize/2
			foodY := float32(g.foodWas.y*gridSize) + gridSize/2

			// Check if food is within one cell vertically for horizontal movement
			if g.direction.x != 0 && math.Abs(float64(g.foodWas.y-head.y)) <= 1 {
				if (g.direction.x > 0 && g.foodWas.x > head.x && g.foodWas.x <= head.x+tongueLength) ||
					(g.direction.x < 0 && g.foodWas.x < head.x && g.foodWas.x >= head.x-tongueLength) {
					endX = foodX
				}
				// Check if food is within one cell horizontally for vertical movement
			} else if g.direction.y != 0 && math.Abs(float64(g.foodWas.x-head.x)) <= 1 {
				if (g.direction.y > 0 && g.foodWas.y > head.y && g.foodWas.y <= head.y+tongueLength) ||
					(g.direction.y < 0 && g.foodWas.y < head.y && g.foodWas.y >= head.y-tongueLength) {
					endY = foodY
				}
			}
		}

		// Draw thin red line for tongue
		vector.StrokeLine(screen,
			startX, startY,
			endX, endY, 1,
			color.RGBA{255, 0, 0, 255}, true)
	}
}

func (g *Game) Draw(screen *ebiten.Image) {

	// draw 1 pixel white frame
	vector.StrokeRect(screen, 0, 0, screenWidth, screenHeight, 1, color.White, true)

	str := fmt.Sprintf("Score: %d", g.score)
	w, _ := text.Measure(str, scoreface, scoreface.Size)
	op := &text.DrawOptions{}

	op.GeoM.Translate(float64(screenWidth-w)/3, 0)
	text.Draw(screen, str, scoreface, op)

	for _, b := range g.buttons {
		b.Draw(screen)
	}

	if g.food != nil {
		vector.DrawFilledRect(screen, float32(g.food.x*gridSize), float32(g.food.y*gridSize), gridSize, gridSize, BLUE, true)
	}

	g.drawSnake(screen)

	g.drawTongue(screen)

	g.particles.Draw(screen)

	if g.gameOver || g.paused {
		str := "Game Over!!! Press space to restart"
		if g.paused {
			str = "Game Paused. Press p to continue"
		}
		w, h := text.Measure(str, face, face.Size)
		op := &text.DrawOptions{}

		op.GeoM.Translate(float64(screenWidth-w)/2, float64(screenHeight-h)/2)
		text.Draw(screen, str, face, op)
	}

}

func (g *Game) drawSnake(screen *ebiten.Image) {
	for i, point := range g.snake {
		if i == 0 {
			vector.DrawFilledRect(screen, float32(point.x*gridSize), float32(point.y*gridSize), gridSize, gridSize, RED, true)
		} else {
			fadingFactor := math.Abs(1.0 - float64(i)/float64(len(g.snake)))
			// there is no inbuilt fade function on RGBA, surprisingly
			faded := color.RGBA{
				R: uint8(float64(tailcolor.R) * fadingFactor),
				G: uint8(float64(tailcolor.G) * fadingFactor),
				B: uint8(float64(tailcolor.B) * fadingFactor),
				A: tailcolor.A,
			}
			vector.DrawFilledRect(screen, float32(point.x*gridSize), float32(point.y*gridSize), gridSize, gridSize, faded, false)
		}

	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 640, 480
}

func (g *Game) checkCollision(point Point) bool {
	for _, body := range g.snake[1:] {
		if point == body {
			return true
		}
	}
	// Check if snake head is out of bounds
	if point.x < 0 || point.x >= gridWidth || point.y < 0 || point.y >= gridHeight {
		return true
	}
	return false
}

func (g *Game) checkFoodCollision() bool {
	head := g.snake[0]
	if g.food == nil {
		return false
	}
	if head == *g.food {
		g.particles.Spawn(float64(g.food.x*gridSize+gridSize/2), float64(g.food.y*gridSize+gridSize/2),
			color.RGBA{255, 220, 100, 255})
		g.removeFood()
		return true
	}
	return false
}

func (g *Game) generateFood() error {
	// If food already exists, don't generate new food
	if g.food != nil {
		return nil
	}

	// Check if the grid is full, obv v v unlikely
	if g.isGridFull() {
		return errors.New("no available space for food")
	}

	maxAttempts := gridWidth * gridHeight
	attempts := 0

	for attempts < maxAttempts {
		point := Point{
			x: rand.IntN(gridWidth),
			// don't want food near UI buttons
			y: rand.IntN(gridHeight-2) + 2,
		}

		if !g.checkCollision(point) {
			g.food = &point
			return nil
		}

		attempts++
		if attempts%10 == 0 {
			log.Printf("Failed to place food after %d attempts", attempts)
		}
	}

	return fmt.Errorf("could not place food after %d attempts", maxAttempts)
}

func (g *Game) isGridFull() bool {
	// Calculate total grid size
	totalSize := gridWidth * gridHeight

	// Count occupied spaces (snake body)
	occupiedSpaces := len(g.snake) // Assuming g.snake contains all body segments

	return occupiedSpaces >= totalSize
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Snake!")

	g := NewGame()
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
