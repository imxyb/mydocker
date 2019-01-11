package container

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"math/rand"
	"mydocker/util"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	Running             = "running"
	Stop                = "stop"
	Exit                = "exit"
	DefaultInfoLocation = "/var/run/mydocker/%s/"
	ConfigName          = "config.json"
)

type ContainerInfo struct {
	Pid        string `json:"pid"`
	Id         string `json:"id"`
	Name       string `json:"name"`
	Command    string `json:"command"`
	CreateTime string `json:"createTime"`
	Status     string `json:"status"`
}

func NewContainerProcess(tty bool, volume string) (cmd *exec.Cmd, writePipe *os.File, err error) {
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
	//rootURL := "/root/test1/"
	//mntURL := "/root/test1/mnt/"
	//if err = NewWorkspace(rootURL, mntURL, volume); err != nil {
	//	return
	//}
	//cmd.Dir = mntURL
	cmd.Dir = "/root/test1/busybox"

	// 如果需要tty则把目前的标准输入、标准输出、标准错误赋予给新的子进程
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return
}

func NewWorkspace(rootURL, mnt, volume string) error {
	if err := CreateReadOnly(rootURL); err != nil {
		return err
	}
	if err := CreateWriteLayer(rootURL); err != nil {
		return err
	}
	if err := CreateMountPoint(rootURL, mnt); err != nil {
		return err
	}
	if volume != "" {
		volumeURLs := volumeExtract(volume)
		if len(volumeURLs) == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			if err := MountVolume(rootURL, mnt, volumeURLs); err != nil {
				return err
			}
			logrus.Info("mount the volume:%+v", volumeURLs)
		}
	}
	return nil
}

// 创建只读层，这一层只是把对应的文件系统复制过来
func CreateReadOnly(rootURL string) error {
	target := filepath.Join(rootURL, "busybox")
	source := filepath.Join(rootURL, "busybox.tar")
	exist, err := pathExist(target)
	if err != nil {
		return err
	}
	if !exist {
		if err = os.Mkdir(target, 0777); err != nil {
			return err
		}
		if _, err = exec.Command("tar", "-xvf", source, "-C", target).CombinedOutput(); err != nil {
			return err
		}
	}
	return nil
}

// make a write layer
func CreateWriteLayer(rootURL string) error {
	writeDir := filepath.Join(rootURL, "writeLayer")
	if err := os.Mkdir(writeDir, 0777); err != nil {
		return err
	}
	return nil
}

// use aufs mount write and read layer
func CreateMountPoint(rootURL string, mnt string) error {
	if err := os.Mkdir(mnt, 0777); err != nil {
		return err
	}
	dirs := "dirs=" + filepath.Join(rootURL+"writeLayer") + ":" + filepath.Join(rootURL+"busybox")
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mnt)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func MountVolume(rootURL, mntURL string, volume []string) error {
	parentURL := volume[0]
	exist, err := pathExist(parentURL)
	if err != nil {
		return err
	}
	if !exist {
		if err = os.Mkdir(parentURL, 0777); err != nil {
			return err
		}
	}
	containerURL := filepath.Join(mntURL, volume[1])
	if err := os.Mkdir(containerURL, 0777); err != nil {
		return err
	}
	dirs := "dirs=" + parentURL
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", containerURL)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err = cmd.Run(); err != nil {
		return err
	}
	return nil
}

func DeleteWorkSpace(rootURL, mntURL, volume string) error {
	if volume != "" {
		volumeURLs := volumeExtract(volume)
		if len(volumeURLs) == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			if err := DeleteMountPointWithVolume(mntURL, volumeURLs); err != nil {
				return err
			}
			logrus.Info("umount the volume:%+v", volumeURLs)
		}
	} else {
		if err := DeleteMountPoint(mntURL); err != nil {
			return err
		}
	}
	if err := DeleteWriteLayer(rootURL); err != nil {
		return err
	}
	return nil
}

func DeleteMountPoint(mntURL string) error {
	if err := umountMnt(mntURL); err != nil {
		return err
	}
	if err := os.RemoveAll(mntURL); err != nil {
		return err
	}
	return nil
}

func DeleteMountPointWithVolume(mntURL string, volumeURLs []string) error {
	containerURL := filepath.Join(mntURL, volumeURLs[1])
	cmd := exec.Command("umount", containerURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	if err := umountMnt(mntURL); err != nil {
		return err
	}
	if err := os.RemoveAll(mntURL); err != nil {
		return err
	}
	return nil
}

func DeleteWriteLayer(rootURL string) error {
	cpath := filepath.Join(rootURL, "writeLayer")
	if err := os.RemoveAll(cpath); err != nil {
		return err
	}
	return nil
}

func umountMnt(mntURL string) error {
	cmd := exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func pathExist(pathStr string) (bool, error) {
	_, err := os.Stat(pathStr)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func volumeExtract(volume string) []string {
	return strings.Split(volume, ":")
}

func generateContainerID(n int) string {
	letterBytes := "1234567890"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func RecordContainerInfo(pid int, command []string, containerName string) (string, error) {
	id := generateContainerID(10)
	createTime := time.Now().Format("2006-01-02 15:04:05")
	commandStr := strings.Join(command, " ")
	if containerName == "" {
		containerName = id
	}
	containerInfo := ContainerInfo{
		Pid:        strconv.Itoa(pid),
		Command:    commandStr,
		Id:         id,
		CreateTime: createTime,
		Name:       containerName,
		Status:     Running,
	}
	b, err := json.Marshal(containerInfo)
	if err != nil {
		return "", err
	}
	jsonStr := string(b)
	dirURL := fmt.Sprintf(DefaultInfoLocation, containerName)
	exist, err := pathExist(dirURL)
	if err != nil {
		return "", err
	}
	if !exist {
		if err := os.MkdirAll(dirURL, 0755); err != nil {
			return "", err
		}
	}
	fileName := filepath.Join(dirURL, ConfigName)
	file, err := os.Create(fileName)
	defer file.Close()
	if err != nil {
		return "", err
	}
	if _, err := file.WriteString(jsonStr); err != nil {
		return "", err
	}
	return containerName, nil
}

func DeleteContainerInfo(containerName string) error {
	dirURL := fmt.Sprintf(DefaultInfoLocation, containerName)
	return os.RemoveAll(dirURL)
}
