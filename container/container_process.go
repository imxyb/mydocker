package container

import (
	"fmt"
	"mydocker/util"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func NewContainerProcess(tty bool) (cmd *exec.Cmd, writePipe *os.File, err error) {
	readPipe, writePipe, err := util.NewPipe()
	if err != nil {
		return
	}
	// proc/self/exec 表示执行自己的init方法
	cmd = exec.Command("/proc/self/exe", "init")
	// 为进程创建对应的namespace
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	cmd.ExtraFiles = []*os.File{readPipe}
	rootURL := "./"
	mntURL := "./mnt"
	if err = NewWorkspace(rootURL, mntURL); err != nil {
		return
	}
	cmd.Dir = mntURL

	// 如果需要tty则把目前的标准输入、标准输出、标准错误赋予给新的子进程
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return
}

func NewWorkspace(rootURL string, mnt string) error {
	if err := CreateReadOnly(rootURL); err != nil {
		return err
	}
	if err := CreateWriteLayer(rootURL); err != nil {
		return err
	}
	if err := CreateMountPoint(rootURL, mnt); err != nil {
		return err
	}
	return nil
}

func CreateReadOnly(rootURL string) error {
	target := filepath.Join(rootURL, "busybox")
	source := filepath.Join(rootURL, "oldbusybox")
	exist, err := PathExist(target)
	if err != nil {
		return err
	}
	if !exist {
		if err = os.Mkdir(target, 0777); err != nil {
			return err
		}
		if err = exec.Command("cp", "-r", source, target).Run(); err != nil{
			return err
		}
	}
	return nil
}

func CreateWriteLayer(rootURL string) error {
	writeDir := filepath.Join(rootURL, "writeLayer")
	if err := os.Mkdir(writeDir, 0777); err != nil {
		return err
	}
	return nil
}

func CreateMountPoint(rootURL string, mnt string) error {
	if err := os.Mkdir(mnt, 0777); err != nil {
		return err
	}
	dirs := "dirs=" + rootURL + "writeLayer:" + rootURL + "busybox"
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mnt)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func PathExist(pathStr string) (bool, error) {
	_, err := os.Stat(pathStr)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
