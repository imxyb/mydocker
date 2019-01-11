package main

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"mydocker/container"
	"mydocker/subsystems"
	"os"
	"strings"
)

var initCmd = cli.Command{
	Name:  "init",
	Usage: "init some env",
	Action: func(ctx *cli.Context) error {
		log.Infof("init execute")
		if err := container.InitContainerProcess(); err != nil {
			return err
		}
		return nil
	},
}

var runCmd = cli.Command{
	Name:  "run",
	Usage: "create container",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name: "ti",
		},
		cli.StringFlag{
			Name:  "m",
			Usage: "memory limit",
		},
		cli.StringFlag{
			Name:  "cpushare",
			Usage: "cpushare limit",
		},
		cli.StringFlag{
			Name:  "cpuset",
			Usage: "cpuset limit",
		},
		cli.StringFlag{
			Name:  "v",
			Usage: "create volume",
		},
		cli.BoolFlag{
			Name:  "d",
			Usage: "detach",
		},
		cli.StringFlag{
			Name:  "name",
			Usage: "create container with name",
		},
	},
	Action: func(ctx *cli.Context) error {
		if len(ctx.Args()) < 1 {
			return errors.New("at lease one arg on run")
		}
		commandArr := ctx.Args()
		tty := ctx.Bool("ti")
		detach := ctx.Bool("d")
		if tty && detach {
			return errors.New("can not get tty when detach")
		}
		resConfig := &subsystems.ResourceConfig{
			MemoryLimit: ctx.String("m"),
			CpuShare:    ctx.String("cpushare"),
			CpuSet:      ctx.String("cpuset"),
		}
		volume := ctx.String("v")
		containerName := ctx.String("name")
		// 实际运行的命令
		if err := Run(tty, commandArr, resConfig, volume, containerName); err != nil {
			log.Fatal(err)
		}
		return nil
	},
}

var commitCommand = cli.Command{
	Name:  "commit",
	Usage: "commit a change for image",
	Action: func(ctx *cli.Context) error {
		if len(ctx.Args()) < 1 {
			return errors.New("must at lease one arg")
		}
		imageName := ctx.Args().Get(0)
		if err := container.CommitContainer(imageName); err != nil {
			return err
		}
		return nil
	},
}

var listCommand = cli.Command{
	Name:  "ps",
	Usage: "list container list",
	Action: func(ctx *cli.Context) error {
		if err := container.ListContainers(); err != nil {
			return err
		}
		return nil
	},
}

// 实际运行的命令
func Run(tty bool, commandArr []string, resConfig *subsystems.ResourceConfig, volume, containerName string) (err error) {
	// 创建命令环境,并且返回一个写管道，用于写入命令字符串
	parent, writePipe, err := container.NewContainerProcess(tty, volume)
	if err != nil {
		return err
	}
	if err = parent.Start(); err != nil {
		return err
	}
	containerName, err = container.RecordContainerInfo(parent.Process.Pid, commandArr, containerName)
	if err != nil {
		return err
	}
	// 设置容器资源限制
	cgroupManager := subsystems.NewCgroupManager("mydocker")
	// 命令结束时候清理容器限制
	defer cgroupManager.Destroy()
	// 设置对应的资源
	if err = cgroupManager.Set(resConfig); err != nil {
		return err
	}
	// 把对应的进程pid写入cgroup
	cgroupManager.Apply(parent.Process.Pid)
	// 发送命令到管道
	if err = sendCommand(commandArr, writePipe); err != nil {
		return err
	}
	if tty {
		if err = parent.Wait(); err != nil {
			return err
		}
		//rootURL := "/root/test1/"
		//mntURL := "/root/test1/mnt/"
		//if err = container.DeleteWorkSpace(rootURL, mntURL, volume); err != nil {
		//	return err
		//}
		if err = container.DeleteContainerInfo(containerName); err != nil {
			return err
		}
	}
	return
}

func sendCommand(commandArr []string, writePipe *os.File) (err error) {
	command := strings.Join(commandArr, " ")
	log.Infof("command is %s", command)
	if _, err = writePipe.WriteString(command); err != nil {
		return err
	}
	writePipe.Close()
	return
}

func main() {
	app := cli.NewApp()
	app.Name = "mydocker"
	app.Usage = "use mydocker"
	// 定义两个命令
	app.Commands = []cli.Command{
		initCmd,
		runCmd,
		commitCommand,
		listCommand,
	}
	app.Before = func(context *cli.Context) error {
		log.SetFormatter(&log.JSONFormatter{})
		log.SetOutput(os.Stdout)
		return nil
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
