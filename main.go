package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/samuel/go-zookeeper/zk"
)

var (
	zookeeper *zk.Conn
	setFlags  = []cli.Flag{
		cli.IntFlag{Name: "v,version", Value: 0, Usage: "node version"},
	}
	permMap = map[string]int32{
		"all":    zk.PermAll,
		"write":  zk.PermWrite,
		"read":   zk.PermRead,
		"create": zk.PermCreate,
		"delete": zk.PermDelete,
		"admin":  zk.PermAdmin,
	}
)

// getPath ensures that the path provided on the cli has "/" prepended.
func getPath(context *cli.Context) string {
	return filepath.Join("/", context.Args().Get(0))
}

// getACL returns the correct zk.ACL values specified on the cli.
// it defaults to PermAll if none are provided.
func getACL(context *cli.Context) []zk.ACL {
	var aclFlag int32
	for _, perm := range context.StringSlice("acl") {
		aclFlag |= permMap[perm]
	}
	if aclFlag == 0 {
		aclFlag = zk.PermAll
	}
	return zk.WorldACL(aclFlag)
}

var createCommand = cli.Command{
	Name:  "create",
	Usage: "create a node",
	Flags: append(setFlags,
		cli.BoolFlag{Name: "world,w", Usage: "set world ACL"},
		cli.StringSliceFlag{Name: "acl,a", Value: &cli.StringSlice{}, Usage: "ACL used on create. default is all. all|create|write|read|delete|admin"},
	),
	Action: func(context *cli.Context) {
		if _, err := zookeeper.Create(
			getPath(context),
			[]byte(context.Args().Get(1)),
			int32(context.Int("version")),
			getACL(context),
		); err != nil {
			logrus.Error(err)
			zookeeper.Close()
			os.Exit(1)
		}
	},
}

var setCommand = cli.Command{
	Name:  "set",
	Usage: "set a key's value",
	Flags: append(setFlags, cli.BoolFlag{Name: "auto", Usage: "automatically set the version based on the current node"}),
	Action: func(context *cli.Context) {
		var (
			path    = getPath(context)
			version = int32(context.Int("version"))
		)
		// if auto fetch the current value for the node so that we
		// get the stats and version to use when setting the new value.
		if context.Bool("auto") {
			_, s, err := zookeeper.Get(path)
			if err != nil {
				logrus.Error(err)
				zookeeper.Close()
				os.Exit(1)
			}
			version = s.Version
		}
		if _, err := zookeeper.Set(
			path,
			[]byte(context.Args().Get(1)),
			version,
		); err != nil {
			logrus.Error(err)
			zookeeper.Close()
			os.Exit(1)
		}
	},
}

var lsCommand = cli.Command{
	Name:  "ls",
	Usage: "list nodes",
	Action: func(context *cli.Context) {
		path := getPath(context)
		children, _, err := zookeeper.Children(path)
		if err != nil {
			logrus.Error(err)
			zookeeper.Close()
			os.Exit(1)
		}
		for _, c := range children {
			fmt.Println(filepath.Join(path, c))
		}
	},
}

var getCommand = cli.Command{
	Name:  "get",
	Usage: "get a key's value",
	Action: func(context *cli.Context) {
		r, s, err := zookeeper.Get(getPath(context))
		if err != nil {
			logrus.Error(err)
			zookeeper.Close()
			os.Exit(1)
		}
		logrus.WithFields(logrus.Fields{"version": s.Version}).Debug("node stats")
		fmt.Printf("%s", r)
	},
}

var deleteCommand = cli.Command{
	Name:  "delete",
	Usage: "delete a key",
	Flags: append(setFlags,
		cli.BoolFlag{Name: "auto", Usage: "automatically set the version based on the current node"},
		cli.BoolFlag{Name: "recursive,r", Usage: "recursively delete keys within the path"},
	),
	Action: func(context *cli.Context) {
		var (
			err     error
			path    = getPath(context)
			version = int32(context.Int("version"))
		)
		if context.Bool("recursive") {
			if err := deleteRecursive(path); err != nil {
				logrus.Error(err)
				zookeeper.Close()
				os.Exit(1)
			}
			return
		}
		// if auto fetch the current value for the node so that we
		// get the stats and version to use when setting the new value.
		if context.Bool("auto") {
			if version, err = getVersion(path); err != nil {
				logrus.Error(err)
				zookeeper.Close()
				os.Exit(1)
			}
		}
		if err := zookeeper.Delete(path, version); err != nil {
			logrus.Error(err)
			zookeeper.Close()
			os.Exit(1)
		}
	},
}

var watchCommand = cli.Command{
	Name:  "watch",
	Usage: "watch a key for changes",
	Flags: []cli.Flag{
		cli.BoolFlag{Name: "follow,f", Usage: "keep watching forever"},
	},
	Action: func(context *cli.Context) {
	forever:
		_, _, ev, err := zookeeper.GetW(context.Args().First())
		if err != nil {
			logrus.Error(err)
			zookeeper.Close()
			os.Exit(1)
		}
		for e := range ev {
			if e.Err != nil {
				logrus.Warning(err)
				continue
			}
			fmt.Printf("%s: %s\n", e.Type, e.Path)
		}
		if context.Bool("follow") {
			goto forever
		}
	},
}

func getVersion(path string) (int32, error) {
	_, s, err := zookeeper.Get(path)
	if err != nil {
		return -1, err
	}
	return s.Version, nil
}

func deleteRecursive(path string) error {
	keys, _, err := zookeeper.Children(path)
	if err != nil {
		return err
	}
	for _, k := range keys {
		if err := deleteRecursive(filepath.Join(path, k)); err != nil {
			return err
		}
	}
	version, err := getVersion(path)
	if err != nil {
		return err
	}
	return zookeeper.Delete(path, version)
}

func main() {
	app := cli.NewApp()
	app.Name = "zkctl"
	app.Usage = "cli application for zookeeper"
	app.Version = "1"
	app.Author = "@crosbymichael"
	app.Email = "crosbymichael@gmail.com"
	app.Flags = []cli.Flag{
		cli.BoolFlag{Name: "debug,D", Usage: "enable debug output"},
		cli.DurationFlag{Name: "t,timeout", Value: 2 * time.Second, Usage: "connection timeout"},
		cli.StringSliceFlag{Name: "zk", Value: &cli.StringSlice{}, Usage: "zookeeper node addresses"},
	}
	app.Commands = []cli.Command{
		createCommand,
		deleteCommand,
		getCommand,
		lsCommand,
		setCommand,
		watchCommand,
	}
	app.Before = func(context *cli.Context) error {
		if context.GlobalBool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
		}
		servers := context.GlobalStringSlice("zk")
		if len(servers) == 0 {
			servers = []string{"127.0.0.1"}
		}
		conn, _, err := zk.Connect(servers, context.Duration("timeout"))
		zookeeper = conn
		return err
	}
	app.After = func(context *cli.Context) error {
		zookeeper.Close()
		return nil
	}
	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
