//-----------------------------------------------------------------------------

//-----------------------------------------------------------------------------

package main

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/deadsy/go-cli"
	"github.com/deadsy/slamx/lidar"
	"github.com/deadsy/slamx/pwm"
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
		app := c.User.(*slam)
		app.lidar.Ctrl <- lidar.Start
	},
}

var lidar_status = cli.Leaf{
	Descr: "show lidar status",
	F: func(c *cli.CLI, args []string) {
		app := c.User.(*slam)
		l := app.lidar
		rows := make([][]string, 0, 10)
		rows = append(rows, []string{"name", l.Name})
		rows = append(rows, []string{"port name", l.PortName})
		rows = append(rows, []string{"pwm name", l.PWMName})
		rows = append(rows, []string{"running", fmt.Sprintf("%t", l.Running)})
		rows = append(rows, []string{"rpm", fmt.Sprintf("%f", l.RPM)})
		rows = append(rows, []string{"good frames", fmt.Sprintf("%d", l.GoodFrames)})
		rows = append(rows, []string{"bad frames", fmt.Sprintf("%d", l.BadFrames)})
		c.Put(cli.TableString(rows, []int{10, 10}, 1) + "\n")
	},
}

var lidar_stop = cli.Leaf{
	Descr: "stop lidar scanning",
	F: func(c *cli.CLI, args []string) {
		app := c.User.(*slam)
		app.lidar.Ctrl <- lidar.Stop
	},
}

// lidar submenu items
var lidar_menu = cli.Menu{
	{"start", lidar_start},
	{"status", lidar_status},
	{"stop", lidar_stop},
}

//-----------------------------------------------------------------------------
// PWM testing

var pwm_off = cli.Leaf{
	Descr: "turn pwm off",
	F: func(c *cli.CLI, args []string) {
		app := c.User.(*slam)
		app.pwm.Set(0.0)
	},
}

var pwm_on = cli.Leaf{
	Descr: "turn pwm on",
	F: func(c *cli.CLI, args []string) {
		app := c.User.(*slam)
		app.pwm.Set(0.5)
	},
}

// pwm submenu items
var pwm_menu = cli.Menu{
	{"off", pwm_off},
	{"on", pwm_on},
}

//-----------------------------------------------------------------------------

// root menu
var menu_root = cli.Menu{
	{"exit", cmd_exit},
	{"help", cmd_help},
	{"history", cmd_history, cli.HistoryHelp},
	{"lidar", lidar_menu, "lidar functions"},
	{"pwm", pwm_menu, "pwm functions"},
}

//-----------------------------------------------------------------------------

type slam struct {
	lidar *lidar.LIDAR
	pwm   *pwm.PWM
}

func NewSlam() *slam {
	app := slam{}
	return &app
}

func (app *slam) Put(s string) {
	fmt.Printf("%s", s)
}

//-----------------------------------------------------------------------------

func mainx() {

	// open the logfile
	logfile, err := os.OpenFile("slamx.log", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Printf("error opening log file: %v", err)
		os.Exit(1)
	}
	log.SetOutput(logfile)

	// setup the user application object
	app := NewSlam()

	// setup the LIDAR
	l, err := lidar.NewLIDAR("lidar0", lidar_serial, lidar_pwm)
	if err != nil {
		log.Fatal("unable to open lidar device")
	}
	app.lidar = l

	// global quit channel for all goroutines
	quit := make(chan bool)
	// wait group to wait for child goroutine completion
	wg := &sync.WaitGroup{}

	// Start the LIDAR goroutine
	wg.Add(1)
	go app.lidar.Process(quit, wg)

	hpath := "history.txt"
	c := cli.NewCLI(app)
	c.HistoryLoad(hpath)
	c.SetRoot(menu_root)
	c.SetPrompt("slamx> ")
	for c.Running() {
		select {
		case scan := <-app.lidar.Scan:
			_ = scan
		default:
			c.Run()
		}
	}
	c.HistorySave(hpath)

	// stop all go routines
	close(quit)
	wg.Wait()

	logfile.Close()
	os.Exit(0)
}

//-----------------------------------------------------------------------------

func main() {

	// open the logfile
	logfile, err := os.OpenFile("slamx.log", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Printf("error opening log file: %v", err)
		os.Exit(1)
	}
	log.SetOutput(logfile)

	// setup the user application object
	app := NewSlam()

	// setup the PWM
	p, err := pwm.Open("pwm", "21", 0)
	if err != nil {
		log.Fatal("unable to open pwm channel")
	}
	app.pwm = p

	hpath := "history.txt"
	c := cli.NewCLI(app)
	c.HistoryLoad(hpath)
	c.SetRoot(menu_root)
	c.SetPrompt("slamx> ")
	for c.Running() {
		c.Run()
	}
	c.HistorySave(hpath)

	logfile.Close()
	os.Exit(0)
}

//-----------------------------------------------------------------------------
