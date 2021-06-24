package server

import (
	"gameserver/client"
	"gameserver/config"
	"gameserver/frame"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Matcher struct {
}

type Server struct {
	gameQueue       []*client.Client
	gameLobby       map[uint16][]*client.Client
	gameState       map[uint16]bool
	currentGameID   uint16
	currentClientID uint16
}

func NewServer() *Server {
	return &Server{
		gameLobby:       make(map[uint16][]*client.Client),
		gameQueue:       make([]*client.Client, 0, 8),
		gameState:       make(map[uint16]bool),
		currentGameID:   1,
		currentClientID: 1,
	}
}

func (s *Server) InterruptHandle() {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// signal catch routine
	go func(s *Server) {
		<-signalChannel
		for gameID := range s.gameLobby {
			s.broadCastWithGameID(frame.CreateEventPacket(gameID, frame.Events.GameOver, config.NullData))
		}
		time.Sleep(time.Millisecond * 500)
		os.Exit(0)
	}(s)
}
