package main

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	ipamapi "github.com/docker/go-plugins-helpers/ipam"
	netapi "github.com/docker/go-plugins-helpers/network"
	"github.com/docker/go-plugins-helpers/sdk"
	"github.com/nicescale/qingcloud-docker-network/drivers/ipam"
	"github.com/nicescale/qingcloud-docker-network/drivers/network"
	"github.com/nicescale/qingcloud-docker-network/qcsdk"
	"github.com/nicescale/qingcloud-docker-network/util"
	"github.com/urfave/cli"
)

const (
	version  = "0.1.0-alpha"
	sockPath = "/var/run/docker/plugins/qingcloud.sock"
)

func main() {
	app := cli.NewApp()
	app.Name = "qingcloud-docker-network"
	app.Usage = "Qingcloud Docker network plugin"
	app.Version = version
	app.Action = Run
	app.Author = "Shijiang Wei"
	app.Email = "mountkin@gmail.com"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "access-key-id,ak",
			Usage:  "The access key ID created from qingcloud console.",
			EnvVar: "ACCESS_KEY_ID",
		},
		cli.StringFlag{
			Name:   "secret-key,sk",
			Usage:  "The secret key of the corresponding access key.",
			EnvVar: "SECRET_KEY",
		},
		cli.StringFlag{
			Name:   "zone",
			Usage:  "The zone that the instance lies in.",
			EnvVar: "ZONE",
		},
		cli.StringFlag{
			Name:   "data-dir,d",
			Usage:  "The directory to store network related files.",
			EnvVar: "DATA_DIR",
			Value:  "/var/lib/docker/qingcloud-network",
		},
		cli.BoolFlag{
			Name:   "debug",
			Usage:  "Whether to print verbose debug log.",
			EnvVar: "DEBUG",
		},
	}
	app.Run(os.Args)
}

// Run initializes the driver
func Run(c *cli.Context) {
	util.Init()
	debug := c.Bool("debug")
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	ak := mustGetStringVar(c, "access-key-id")
	sk := mustGetStringVar(c, "secret-key")
	zone := mustGetStringVar(c, "zone")
	dir := mustGetStringVar(c, "data-dir")
	api := qcsdk.NewApi(ak, sk, zone)
	api.SetDebug(debug)
	dn, err := network.New(api, dir)
	if err != nil {
		errExit(1, err.Error())
	}
	gid := 0
	group, err := user.LookupGroup("docker")
	if err == nil {
		gid, _ = strconv.Atoi(group.Gid)
	}

	di := ipam.New(api)
	h := sdk.NewHandler()
	netapi.RegisterDriver(dn, h)
	ipamapi.RegisterDriver(di, h)
	if err = h.ServeUnix(sockPath, gid); err != nil {
		errExit(2, err.Error())
	}
}

func errExit(code int, format string, val ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", val...)
	os.Exit(code)
}

func mustGetStringVar(c *cli.Context, key string) string {
	v := strings.TrimSpace(c.String(key))
	if v == "" {
		errExit(1, "%s must be provided", key)
	}
	return v
}
