# Go Snake Game

A classic Snake game implementation written in Go and [Ebitengine](https://ebitengine.org/), featuring sound effects, particle systems, and progressive difficulty.

Mostly created as a good way to play with Golang and to add another Ebitengine example

Majority of the work in a weekend after discovering ebitengine

I am strong believer in monkey see, monkey do

## Description

This is a modern take on the classic Snake game where players control a snake that grows longer as it consumes food. 

## How to Play

### Controls
- Use arrow keys to control the snake's direction
- Guide the snake to eat the food items that appear on the screen
- Avoid hitting the walls or the snake's own body
- Press p to pause the game and resume it at any time
- Press m to toggle sound effects on and off
- Press Spacebar to fire snake tongue
- When gameover, press Spacebar to restart
- UI buttons for new game, mute and pause

### Game Mechanics
- The snake grows longer each time it eats food (caps at 40 segments)
- Eat the food to gain 2 points, or use tongue to do it more quickly for 1 points
- The game speed increases with each food item eaten, making it progressively more challenging
- Food items may randomly disappear, adding an extra challenge
- Game ends if the snake collides with walls or itself

### Features
- Smooth snake movement
- Visual particle effects
- Sound effects for:
- Food collection
- Game over events
- Progressive difficulty (snake speeds up as it eats)
- Random food disappearance mechanics

## To run

1. Ensure you have Go installed on your system
2. Clone the repository
3. Run the following commands:
```bash
go mod tidy
go run .
```

## To run on web via wasm

[Click here to Preview](https://acid.seedhost.eu/seedbod/snake)

## To build for web

Set env vars
```
GOARCH=wasm GOOS=js go build -o main.wasm
```

Creates a `main.wasm` file, copy to web server along with `index.html` and `wasm_exec.js`

Note, your web server should support correct mime type for wasm i.e.  'application/wasm', or it won't load
