package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"image/color"
)

type Button struct {
	x, y, w, h float32 // Changed to float32 for vector package
	text       string
	isHovered  bool
	isPressed  bool
	face       *text.GoTextFace
	onClick    func(this *Button)
	hotkey     ebiten.Key
	right      float32
}

func (b *Button) Measure() (float64, float64) {
	return text.Measure(b.text, b.face, b.face.Size)
}

func NewButton(x, y float32, str string, face *text.GoTextFace, onClick func(*Button)) *Button {
	paddingX := float32(0) // e.g., 5px padding on each side
	paddingY := float32(0) // e.g., 4px padding top/bottom
	w, h := text.Measure(str, face, face.Size)
	return &Button{
		x:       x,
		y:       y,
		w:       float32(w) + paddingX, // Add padding
		h:       float32(h) + paddingY, // Add padding
		text:    str,
		face:    face,
		onClick: onClick,
		right:   -1.0, // not right aligned by default
	}

}

func (b *Button) Align(screenWidth int) {
	if b.right >= 0.0 {
		paddingX := float32(0) // Consistent padding
		paddingY := float32(0) // Consistent padding

		w, h := text.Measure(b.text, b.face, b.face.Size)
		b.w = float32(w) + paddingX
		b.h = float32(h) + paddingY
		b.x = float32(screenWidth) - b.w - b.right
	}
}

func (b *Button) SetText(newText string, screenWidth int) {
	b.text = newText
	b.Align(screenWidth)
}

func (b *Button) Update() {
	x, y := ebiten.CursorPosition()
	b.isHovered = float32(x) >= b.x && float32(x) < b.x+b.w &&
		float32(y) >= b.y && float32(y) < b.y+b.h

	if b.hotkey != 0 && inpututil.IsKeyJustPressed(b.hotkey) {
		b.onClick(b)
	}

	if b.isHovered && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		b.isPressed = true
		if b.onClick != nil {
			b.onClick(b)
		}
	} else {
		b.isPressed = false
	}
}

func (b *Button) Draw(screen *ebiten.Image) {

	// Choose color based on button state
	bgColor := color.RGBA{100, 100, 100, 255}
	if b.isPressed {
		bgColor = color.RGBA{80, 80, 80, 255}
	} else if b.isHovered {
		bgColor = color.RGBA{150, 150, 150, 255}
	}

	vector.DrawFilledRect(screen, b.x, b.y, b.w, b.h, bgColor, true)

	op := &text.DrawOptions{}

	op.GeoM.Translate(float64(b.x), float64(b.y))
	text.Draw(screen, b.text, b.face, op)
}
