//-----------------------------------------------------------------------------

//-----------------------------------------------------------------------------

package main

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/deadsy/go-cli"
	"github.com/deadsy/slamx/gpio"
	"github.com/deadsy/slamx/lidar"
	"github.com/deadsy/slamx/motor"
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
		rows = append(rows, []string{"serial port", l.PortName})
		rows = append(rows, []string{"motor", l.Motor.Name})
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
		app.motor.Set(0.0)
	},
}

var pwm_on = cli.Leaf{
	Descr: "turn pwm on",
	F: func(c *cli.CLI, args []string) {
		app := c.User.(*slam)
		app.motor.Set(0.4)
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
	motor *motor.Motor
}

func NewSlam() *slam {
	app := slam{}
	return &app
}

func (app *slam) Put(s string) {
	fmt.Printf("%s", s)
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
	defer logfile.Close()

	// setup the user application object
	app := NewSlam()

	// gpio subsystem
	gpio, err := gpio.NewGPIO("gpio0")
	if err != nil {
		log.Fatal("unable to create gpio device")
	}
	defer gpio.Close()

	// pwm output for motor control
	pwm, err := gpio.NewPWM(xv11_pwm, 0)
	if err != nil {
		log.Fatal("unable to create pwm output")
	}
	defer pwm.Close()

	// standby (on/off) control for motor driver
	stby, err := gpio.NewOutput(xv11_stby, 0)
	if err != nil {
		log.Fatal("unable to create gpio output")
	}
	defer stby.Close()

	// setup the driver for the lidar motor
	motor, err := motor.NewMotor("motor0", pwm, stby)
	if err != nil {
		log.Fatal("unable to create motor control")
	}
	defer motor.Close()

	// setup the xv11 lidar
	lidar, err := lidar.NewLIDAR("lidar0", xv11_serial, motor)
	if err != nil {
		log.Fatal("unable to open lidar device")
	}
	defer lidar.Close()
	app.lidar = lidar

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

	os.Exit(0)
}

//-----------------------------------------------------------------------------
