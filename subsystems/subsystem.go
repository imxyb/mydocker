package subsystems

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
)

var subsystems = []SubSystem{
	&MemorySubsystem{},
	&CpuSubsystem{},
	&CpuSetSubsystem{},
}

type SubSystem interface {
	Name() string
	Set(cpath string, config *ResourceConfig) error
	Apply(cpath string, pid int) error
	Remove(cpath string) error
}

type ResourceConfig struct {
	MemoryLimit string
	CpuShare    string
	CpuSet      string
}

type CgroupManager struct {
	Path   string
	Config *ResourceConfig
}

func NewCgroupManager(cpath string) *CgroupManager {
	return &CgroupManager{
		Path: cpath,
	}
}

func (c *CgroupManager) Apply(pid int) error {
	for _, subsystem := range subsystems {
		if err := subsystem.Apply(c.Path, pid); err != nil {
			return err
		}
	}
	return nil
}

func (c *CgroupManager) Set(config *ResourceConfig) error {
	for _, subsystem := range subsystems {
		if err := subsystem.Set(c.Path, config); err != nil {
			return err
		}
	}
	return nil
}

func (c *CgroupManager) Destroy() error {
	for _, subsystem := range subsystems {
		if err := subsystem.Remove(c.Path); err != nil {
			return err
		}
	}
	return nil
}

func GetCgroupPathInfo(subsystem string, cgroupRoot string, autoCreate bool) (string, error) {
	cpath, err := findCgroupPathInfo(subsystem)
	if cpath == "" {
		return "", fmt.Errorf("can not found %s cgroup", subsystem)
	}
	if err != nil {
		return "", err
	}
	fullPath := path.Join(cpath, cgroupRoot)
	if _, err = os.Stat(fullPath); err == nil || (autoCreate && os.IsNotExist(err)) {
		if os.IsNotExist(err) {
			if err = os.Mkdir(fullPath, 0755); err != nil {
				return "", err
			}
		}
		return fullPath, nil
	}
	return "", fmt.Errorf("cpath err:%s", err.Error())
}

func findCgroupPathInfo(subsystem string) (path string, err error) {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		txt := scanner.Text()
		fields := strings.Split(txt, " ")
		for _, opt := range strings.Split(fields[len(fields)-1], ",") {
			if opt == subsystem {
				path = fields[4]
				return
			}
		}
	}
	if err = scanner.Err(); err != nil {
		return
	}
	return
}
