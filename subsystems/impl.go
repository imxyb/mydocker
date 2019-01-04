package subsystems

import (
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

type MemorySubsystem struct {
}

func (s *MemorySubsystem) Name() string {
	return "memory"
}

func (s *MemorySubsystem) Set(cpath string, config *ResourceConfig) error {
	if config.MemoryLimit != "" {
		cpath, err := GetCgroupPathInfo(s.Name(), cpath, true)
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(path.Join(cpath, "memory.limit_in_bytes"), []byte(config.MemoryLimit), 0644); err != nil {
			return err
		}
	}
	return nil
}

func (s *MemorySubsystem) Apply(cpath string, pid int) error {
	cpath, err := GetCgroupPathInfo(s.Name(), cpath, false)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path.Join(cpath, "tasks"), []byte(strconv.Itoa(pid)), 0644);
}

func (s *MemorySubsystem) Remove(cpath string) error {
	cpath, err := GetCgroupPathInfo(s.Name(), cpath, false)
	if err != nil {
		return err
	}
	return os.RemoveAll(cpath)
}

type CpuSubsystem struct {
}

func (s *CpuSubsystem) Name() string {
	return "cpu"
}

func (s *CpuSubsystem) Set(cpath string, config *ResourceConfig) error {
	if config.CpuShare != "" {
		cpath, err := GetCgroupPathInfo(s.Name(), cpath, true)
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(path.Join(cpath, "cpu.shares"), []byte(config.MemoryLimit), 0644); err != nil {
			return err
		}
	}
	return nil
}

func (s *CpuSubsystem) Apply(cpath string, pid int) error {
	cpath, err := GetCgroupPathInfo(s.Name(), cpath, false)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path.Join(cpath, "tasks"), []byte(strconv.Itoa(pid)), 0644);
}

func (s *CpuSubsystem) Remove(cpath string) error {
	cpath, err := GetCgroupPathInfo(s.Name(), cpath, false)
	if err != nil {
		return err
	}
	return os.RemoveAll(cpath)
}

type CpuSetSubsystem struct {
}

func (s *CpuSetSubsystem) Name() string {
	return "cpuset"
}

func (s *CpuSetSubsystem) Set(cpath string, config *ResourceConfig) error {
	if config.CpuSet != "" {
		cpath, err := GetCgroupPathInfo(s.Name(), cpath, true)
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(path.Join(cpath, "cpuset.cpus"), []byte(config.MemoryLimit), 0644); err != nil {
			return err
		}
	}
	return nil
}

func (s *CpuSetSubsystem) Apply(cpath string, pid int) error {
	cpath, err := GetCgroupPathInfo(s.Name(), cpath, false)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path.Join(cpath, "tasks"), []byte(strconv.Itoa(pid)), 0644);
}

func (s *CpuSetSubsystem) Remove(cpath string) error {
	cpath, err := GetCgroupPathInfo(s.Name(), cpath, false)
	if err != nil {
		return err
	}
	return os.RemoveAll(cpath)
}
