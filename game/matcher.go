package game

import (
	"bufio"
	"gameserver/client"
	"gameserver/config"
	"gameserver/event"
	"gameserver/utils"
	"log"
	"net"
)

var (
	MainMatcher *Matcher
)

type Matcher struct {
	gameQueue       []*client.Client
	gameList        map[uint16][]*client.Client
	currentGameID   uint16
	currentClientID uint16
}

func StartMatcher(ip, port string) {
	MainMatcher = &Matcher{
		gameQueue:       make([]*client.Client, 0, 8),
		gameList:        make(map[uint16][]*client.Client),
		currentGameID:   1,
		currentClientID: 1,
	}
	// go MainMatcher.connectionControlRoutine()
	MainMatcher.listen(ip, port)
}

func (m *Matcher) listen(ip, port string) {
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
		m.matchingRoutine(conn)
	}
}

func (m *Matcher) matchingRoutine(conn net.Conn) {
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
	c := client.NewClient(m.currentClientID, conn)
	m.gameQueue = append(m.gameQueue, c)
	m.currentClientID++
	m.checkQueue()
}

// Check queue if there are enough participant to fill a game
func (m *Matcher) checkQueue() {
	group := make([]*client.Client, 0, config.GameSize)
	for _, c := range m.gameQueue {
		if c.State == client.ClientState.InQueue {
			group = append(group, c)
		}
		if len(group) == config.GameSize {
			m.createGame(group)
			setStateAll(group, client.ClientState.InPool)
			log.Printf("[game on] there are enough participant to create a game. game size: %v\n", config.GameSize)
			return
		}
	}
	log.Println("[wait] Adding player to game queue.")
}

// create the game and attach it to gameList
func (m *Matcher) createGame(players []*client.Client) {
	// send all clients its own client and game ID
	for _, p := range players {
		pack := event.PackGameIDAndClientID(m.currentGameID, p.ClientID)
		_, err := p.TCPconn.Write(pack)
		if err != nil {
			// if something went wrong change all states to 'InQueue' again
			abortGameCreation(players)
			// close connection with player who has issue
			p.TCPconn.Close()
			// and remove the player from gameQueue
			// to avoid any other problem.
			// some attempt base approach might be good for this kind situations
			m.removeFromGameQueue(p)
			return
		}
	}
	m.gameList[m.currentGameID] = players
	for _, p := range players {
		p.TCPconn.Close()
		p.ChangeState(client.ClientState.InGame)
	}
	m.currentGameID++
	m.clearGameQueue()
}

func (m *Matcher) clearGameQueue() {
	newGameQueue := make([]*client.Client, 0, len(m.gameQueue))
	for _, c := range m.gameQueue {
		if c.State != client.ClientState.InGame {
			newGameQueue = append(newGameQueue, c)
		}
	}
	m.gameQueue = newGameQueue
}

func (m *Matcher) removeFromGameQueue(p *client.Client) {
	newGameQueue := make([]*client.Client, 0, len(m.gameQueue)-1)
	for _, c := range m.gameQueue {
		if c.ClientID != p.ClientID {
			newGameQueue = append(newGameQueue, c)
		}
	}
	m.gameQueue = newGameQueue
}

func abortGameCreation(players []*client.Client) {
	setStateAll(players, client.ClientState.InQueue)
}

func setStateAll(clients []*client.Client, state string) {
	for _, c := range clients {
		c.ChangeState(state)
	}
}
