package main

import (
	"bytes"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
)

type SoundManager struct {
	audioContext  *audio.Context
	players       []*audio.Player
	maxPlayers    int
	currentPlayer int
	volume        float64 // 0.0 to 1.0
	muted         bool
}

// NewSoundManager Call this only once, else AudioContext complains!
// NewContext panics when an audio context is already created.
func NewSoundManager(poolSize int) *SoundManager {
	return &SoundManager{
		audioContext: audio.NewContext(44100),
		players:      make([]*audio.Player, 0, poolSize),
		maxPlayers:   poolSize,
		volume:       1.0,
		muted:        false,
	}
}

func (sm *SoundManager) SetVolume(v float64) {
	sm.volume = v
	for _, player := range sm.players {
		if sm.muted {
			player.SetVolume(0)
		} else {
			player.SetVolume(v)
		}
	}
}

func (sm *SoundManager) SetMute(muted bool) {
	sm.muted = muted
	for _, player := range sm.players {
		if muted {
			player.SetVolume(0)
		} else {
			player.SetVolume(sm.volume)
		}
	}
}

func (sm *SoundManager) PlaySound(soundData []byte) error {
	decoded, err := mp3.DecodeWithSampleRate(44100, bytes.NewReader(soundData))
	if err != nil {
		return err
	}

	// Create new player or reuse existing one
	var player *audio.Player
	if len(sm.players) < sm.maxPlayers {
		player, err = sm.audioContext.NewPlayer(decoded)
		if err != nil {
			return err
		}
		sm.players = append(sm.players, player)
	} else {
		// Reuse existing player
		player = sm.players[sm.currentPlayer]
		_ = player.Close()
		player, err = sm.audioContext.NewPlayer(decoded)
		if err != nil {
			return err
		}
		sm.players[sm.currentPlayer] = player
		sm.currentPlayer = (sm.currentPlayer + 1) % sm.maxPlayers
	}

	if sm.muted {
		player.SetVolume(0)
	} else {
		player.SetVolume(sm.volume)
	}

	player.Play()
	return nil
}

func (sm *SoundManager) Close() {
	for _, player := range sm.players {
		if player != nil {
			_ = player.Close()
		}
	}
}
