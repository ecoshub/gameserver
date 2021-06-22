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
	// A lists that holds all client that requesting a game
	gameQueue []*client.Client = make([]*client.Client, 0, 8)

	// A map of clients in that same game attached with game ID
	gameList map[uint16][]*client.Client = make(map[uint16][]*client.Client)

	// gameID and client ID is designed as incrementables
	// some uuid decleration can be more suitable
	// '0' is reserved
	currentGameID   uint16 = 1
	currentClientID uint16 = 1
)

func StartMatcher(ip, port string) {
	listen(ip, port)
}

func listen(ip, port string) {
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
		matchingRoutine(conn)
	}
}

func matchingRoutine(conn net.Conn) {
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
	c := client.NewClient(currentClientID, conn)
	gameQueue = append(gameQueue, c)
	currentClientID++
	checkQueue()
}

// Check queue if there are enough participant to fill a game
func checkQueue() {
	group := make([]*client.Client, 0, config.GameSize)
	for _, c := range gameQueue {
		if c.State == client.ClientState.InQueue {
			group = append(group, c)
		}
		if len(group) == config.GameSize {
			createGame(group)
			setStateAll(group, client.ClientState.InPool)
			log.Printf("[game on] there are enough participant to create a game. game size: %v\n", config.GameSize)
			return
		}
	}
	log.Println("[wait] Adding player to game queue.")
}

// create the game and attach it to gameList
func createGame(players []*client.Client) {
	// send all clients its own client and game ID
	for _, p := range players {
		pack := event.PackGameIDAndClientID(currentGameID, p.ClientID)
		_, err := p.TCPconn.Write(pack)
		if err != nil {
			// if something went wrong change all states to 'InQueue' again
			abortGameCreation(players)
			// close connection with player who has issue
			p.TCPconn.Close()
			// and remove the player from gameQueue
			// to avoid any other problem.
			// some attempt base approach might be good for this kind situations
			removeFromGameQueue(p)
			return
		}
	}
	gameList[currentGameID] = players
	for _, p := range players {
		p.TCPconn.Close()
		p.ChangeState(client.ClientState.InGame)
	}
	currentGameID++
	clearGameQueue()
}

func abortGameCreation(players []*client.Client) {
	setStateAll(players, client.ClientState.InQueue)
}

func setStateAll(clients []*client.Client, state string) {
	for _, c := range clients {
		c.ChangeState(state)
	}
}

func clearGameQueue() {
	newGameQueue := make([]*client.Client, 0, len(gameQueue))
	for _, c := range gameQueue {
		if c.State != client.ClientState.InGame {
			newGameQueue = append(newGameQueue, c)
		}
	}
	gameQueue = newGameQueue
}

func removeFromGameQueue(p *client.Client) {
	newGameQueue := make([]*client.Client, 0, len(gameQueue))
	for _, c := range gameQueue {
		if c.ClientID != p.ClientID {
			newGameQueue = append(newGameQueue, c)
		}
	}
	gameQueue = newGameQueue
}
