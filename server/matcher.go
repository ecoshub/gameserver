package server

import (
	"bufio"
	"gameserver/client"
	"gameserver/config"
	"gameserver/frame"
	"gameserver/utils"
	"log"
	"net"
)

func (s *Server) StartMatcher(ip, port string) {
	s.listen(ip, port)
}

func (s *Server) listen(ip, port string) {
	listener, err := net.Listen("tcp", config.ServerListenAddress+":"+port)
	if err != nil {
		// listen error must be handle.
		// it can control by an error channel
		log.Println(err)
		return
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			return
		}
		s.matchingRoutine(conn)
	}
}

func (s *Server) matchingRoutine(conn net.Conn) {
	reader := bufio.NewReader(conn)
	msg, err := utils.ReadNBytes(reader, utils.HashLength)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("[req] game request arrived from: " + conn.RemoteAddr().String())
	valid := utils.ValidateHash(msg)
	if !valid {
		log.Println("[auth] auth failed. remote: " + conn.RemoteAddr().String())
		conn.Close()
		return
	}
	log.Println("[auth] auth success!. remote: " + conn.RemoteAddr().String())
	c := client.NewClient(s.currentClientID, conn)
	s.gameQueue = append(s.gameQueue, c)
	s.currentClientID++
	s.checkQueue()
}

// Check queue if there are enough participant to fill a game
func (s *Server) checkQueue() {
	group := make([]*client.Client, 0, config.GameSize)
	for _, c := range s.gameQueue {
		if c.State == client.ClientState.InQueue {
			group = append(group, c)
		}
		if len(group) == config.GameSize {
			s.createGame(group)
			setStateAll(group, client.ClientState.InPool)
			log.Printf("[game on] there are enough participant to create a game. game size: %v\n", config.GameSize)
			return
		}
	}
	log.Println("[wait] Adding player to game queue.")
}

// create the game and attach it to gameList
func (s *Server) createGame(players []*client.Client) {
	// send all clients its own client and game ID
	for _, p := range players {
		pack := frame.PackGameIDAndClientID(s.currentGameID, p.ClientID)
		_, err := p.TCPconn.Write(pack)
		if err != nil {
			// if something went wrong change all states to 'InQueue' again
			abortGameCreation(players)
			// close connection with player who has issue
			p.TCPconn.Close()
			// and remove the player from gameQueue
			// to avoid any other problem.
			// some attempt base approach might be good for this kind situations
			s.removeFromGameQueue(p)
			return
		}
	}
	s.gameLobby[s.currentGameID] = players
	for _, p := range players {
		p.TCPconn.Close()
		p.ChangeState(client.ClientState.InGame)
	}
	s.currentGameID++
	s.clearGameQueue()
}

func (s *Server) clearGameQueue() {
	newGameQueue := make([]*client.Client, 0, len(s.gameQueue))
	for _, c := range s.gameQueue {
		if c.State != client.ClientState.InGame {
			newGameQueue = append(newGameQueue, c)
		}
	}
	s.gameQueue = newGameQueue
}

func (s *Server) removeFromGameQueue(p *client.Client) {
	newGameQueue := make([]*client.Client, 0, len(s.gameQueue)-1)
	for _, c := range s.gameQueue {
		if c.ClientID != p.ClientID {
			newGameQueue = append(newGameQueue, c)
		}
	}
	s.gameQueue = newGameQueue
}

func abortGameCreation(players []*client.Client) {
	setStateAll(players, client.ClientState.InQueue)
}

func setStateAll(clients []*client.Client, state string) {
	for _, c := range clients {
		c.ChangeState(state)
	}
}
