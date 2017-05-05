//-----------------------------------------------------------------------------

//-----------------------------------------------------------------------------

package main

import (
	"fmt"
	"os"

	cli "github.com/deadsy/go-cli"
	//"github.com/deadsy/slamx/lidar"
	//"github.com/deadsy/slamx/view"
)

//-----------------------------------------------------------------------------
// cli related leaf functions

var cmd_help = cli.Leaf{
	Descr: "general help",
	F: func(c *cli.CLI, args []string) {
		c.GeneralHelp()
	},
}

var cmd_history = cli.Leaf{
	Descr: "command history",
	F: func(c *cli.CLI, args []string) {
		c.SetLine(c.DisplayHistory(args))
	},
}

var cmd_exit = cli.Leaf{
	Descr: "exit application",
	F: func(c *cli.CLI, args []string) {
		c.Exit()
	},
}

//-----------------------------------------------------------------------------
// LIDAR menu

var lidar_start = cli.Leaf{
	Descr: "start lidar scanning",
	F: func(c *cli.CLI, args []string) {
		c.Put("TODO")
	},
}

var lidar_status = cli.Leaf{
	Descr: "show lidar status",
	F: func(c *cli.CLI, args []string) {
		c.Put("TODO")
	},
}

var lidar_stop = cli.Leaf{
	Descr: "stop lidar scanning",
	F: func(c *cli.CLI, args []string) {
		c.Put("TODO")
	},
}

// lidar submenu items
var lidar_menu = cli.Menu{
	{"start", lidar_start},
	{"status", lidar_status},
	{"stop", lidar_stop},
}

//-----------------------------------------------------------------------------

// root menu
var menu_root = cli.Menu{
	{"exit", cmd_exit},
	{"help", cmd_help},
	{"history", cmd_history, cli.HistoryHelp},
	{"lidar", lidar_menu, "lidar functions"},
}

//-----------------------------------------------------------------------------

type user_app struct {
}

func NewUserApp() *user_app {
	app := user_app{}
	return &app
}

func (user *user_app) Put(s string) {
	fmt.Printf("%s", s)
}

//-----------------------------------------------------------------------------

func main() {

	hpath := "history.txt"
	c := cli.NewCLI(NewUserApp())
	c.HistoryLoad(hpath)
	c.SetRoot(menu_root)
	c.SetPrompt("slamx> ")
	for c.Running() {
		c.Run()
	}
	c.HistorySave(hpath)
	os.Exit(0)
}

//-----------------------------------------------------------------------------

/*

func main() {

	view0, err := view.Open("view0")
	if err != nil {
		log.Fatal("unable to open view window")
	}

	lidar0, err := lidar.Open("lidar0", lidar_serial, lidar_pwm)
	if err != nil {
		log.Fatal("unable to open lidar device")
	}

	quit := make(chan bool)
	wg := &sync.WaitGroup{}

	// start the LIDAR process
	wg.Add(1)
	scan_ch := make(chan lidar.Scan_2D)
	go lidar0.Process(quit, wg, scan_ch)

	angle := float32(0)

	// run the event loop
	running := true
	for running {
		select {
		case scan := <-scan_ch:
			log.Printf("rxed %d", len(scan.Samples))
			//view0.Render(&scan)
			view0.Render2(angle)
			angle += 1
		default:
			running = view0.Events()
			view0.Delay(30)
		}
	}

	// stop all go routines
	close(quit)
	wg.Wait()

	lidar0.Close()
	view0.Close()

	os.Exit(0)
}

*/

//-----------------------------------------------------------------------------
