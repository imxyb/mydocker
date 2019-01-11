package container

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func InitContainerProcess() (err error) {
	// 读取fd为3，也就是附加的read管道
	readPipe := os.NewFile(uintptr(3), "pipe")
	// 读取管道的数据
	b, err := ioutil.ReadAll(readPipe)
	if err != nil {
		return
	}
	msgStr := string(b)
	commandArr := strings.Split(msgStr, " ")
	if len(commandArr) == 0 || commandArr[0] == "" {
		err = fmt.Errorf("command %s is not found", commandArr[0])
		return
	}

	// 获取实际路径
	command, err := exec.LookPath(commandArr[0])
	if err != nil {
		return err
	}
	logrus.Infof("command %s", command)

	if err = setUpMount(); err != nil {
		return err
	}
	// 使用系统调用execve来替换当前的init程序为传入的command
	if err := syscall.Exec(command, commandArr[0:], os.Environ()); err != nil {
		return err
	}
	return nil
}

func pivotRoot(root string) error {
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return err
	}
	pivotDir := filepath.Join(root, ".pivot_root")
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		return err
	}
	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return err
	}
	if err := syscall.Chdir("/"); err != nil {
		return err
	}
	pivotDir = filepath.Join("/", ".pivot_root")
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return err
	}
	return os.Remove(pivotDir)
}

func setUpMount() error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := pivotRoot(pwd); err != nil {
		return err
	}

	// 设置mount的flag
	// MS_NOEXEC 表示不执行任何程序
	// MS_NOSUID 不允许set uid
	// MS_NODEV 默认设定
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	// 挂载proc
	if err := syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), ""); err != nil {
		return err
	}
	if err := syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755"); err != nil {
		return err
	}

	return nil
}
