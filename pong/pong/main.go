package main

import (
	"bufio"
	"errors"
	"log"
	"net"
	"os"

	"github.com/gdamore/tcell"
)

func main() {
	logFile := initLogger()
	conn := connect()

	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("Error creating screen: %s", err)
	}
	if err := screen.Init(); err != nil {
		log.Fatalf("Error initializing screen: %s", err)
	}

	defer func() {
		maybePanic := recover()
		screen.Fini()
		logFile.Close()
		conn.Close()
		if maybePanic != nil {
			panic(maybePanic)
		}
		os.Exit(1)
	}()

	kill := make(chan bool)
	keyEvents := make(chan tcell.Key)
	packets := make(chan Packet)
	go eventListener(keyEvents, kill, screen)
	socket(conn, keyEvents, packets)

	for {
		screen.Clear()
		select {
		case <-kill:
			return
		default:
			drawBorder(screen)
			screen.Show()
		}
	}
}

type Packet struct{}

func socket(conn net.Conn, out chan tcell.Key, p chan Packet) {
	// listen
	go listen(conn, p)

	// send
	go send(conn, out)
}

func listen(conn net.Conn, _ chan Packet) {

	r := bufio.NewReader(conn)
	for {
		msg, err := r.ReadString('\n')
		if err != nil {
			log.Printf("Connection closed by server: %s", err)
		}

		log.Printf("Recieved: %s", msg)
	}
}

func send(conn net.Conn, out chan tcell.Key) {
	w := bufio.NewWriter(conn)
	for {

		select {
		case msg := <-out:
			_, err := w.WriteString("up\n")
			if err != nil {
				log.Printf("Error writing to server on key event: %v\n", err)
			}
			log.Print("writing...", msg)
		}
	}
}

func initLogger() *os.File {
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	log.SetOutput(logFile)

	return logFile
}

func connect() net.Conn {
	conn, err := net.Dial("tcp", ":8000")
	if err != nil {
		log.Fatalf("Error connecting to server: %s", err)
	}

	return conn
}

func eventListener(ch chan tcell.Key, kill chan bool, screen tcell.Screen) {
	// main event loop listening to keyboard events
	for {
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			screen.Sync()
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
				log.Print("Got kill cmd")
				kill <- true
			}

			if ev.Key() == tcell.KeyUp || ev.Key() == tcell.KeyDown {
				log.Print("Got arrow key")
				ch <- ev.Key()
			}
		}
	}
}

func drawBorder(s tcell.Screen) error {
	MAX_WIDTH := 110
	MAX_HEIGHT := 40

	GAME_TOP := 0
	GAME_BOTTOM := 40
	GAME_LEFT := 0
	GAME_RIGHT := 110

	w, h := s.Size()
	if w < 150 || h < 60 {
		return errors.New("Window too small")
	}

	style := tcell.StyleDefault.Foreground(tcell.ColorWhite)

	/*
		* To calculate the placement of the borders we get the middle x and y
		 for x axis, we use the middle, minus half of our board size, and draw over until mid x + half of board size

		* we then bump the borders out by 1 to accomodate for the corners

		* repeat on y axis
	*/

	// Draw top and bottom horizontal borders

	for x := 0; x <= MAX_WIDTH; x++ {
		s.SetContent(x, GAME_TOP, tcell.RuneHLine, nil, style)    // Top border
		s.SetContent(x, GAME_BOTTOM, tcell.RuneHLine, nil, style) // Bottom border
	}

	// Draw left and right vertical borders
	for y := 0; y <= MAX_HEIGHT; y++ {
		s.SetContent(GAME_LEFT, y, tcell.RuneVLine, nil, style)  // Left border
		s.SetContent(GAME_RIGHT, y, tcell.RuneVLine, nil, style) // Right border
	}

	// Draw corners
	s.SetContent(GAME_LEFT, GAME_TOP, tcell.RuneULCorner, nil, style)     // Upper left corner
	s.SetContent(GAME_RIGHT, GAME_TOP, tcell.RuneURCorner, nil, style)    // Upper right corner
	s.SetContent(GAME_LEFT, GAME_BOTTOM, tcell.RuneLLCorner, nil, style)  // Lower left corner
	s.SetContent(GAME_RIGHT, GAME_BOTTOM, tcell.RuneLRCorner, nil, style) // Lower right corner

	return nil
}
