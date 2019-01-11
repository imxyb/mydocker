package container

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/tabwriter"
)

func ListContainers() error {
	dirURL := fmt.Sprintf(DefaultInfoLocation, "")
	dirURL = dirURL[:len(dirURL)-1]
	fileList, err := ioutil.ReadDir(dirURL)
	if err != nil {
		return err
	}
	var containInfos []*ContainerInfo
	for _, fileInfo := range fileList {
		info, err := getContainerInfoByFile(fileInfo)
		if err != nil {
			logrus.Errorf("get container info failed:%s", err.Error())
			continue
		}
		containInfos = append(containInfos, info)
	}
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(w, "ID\tNAME\tPID\tSTATUS\tCOMMAND\tCREATED\n")
	for _, item := range containInfos {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Id, item.Name, item.Pid, item.Status, item.Command, item.CreateTime)
	}
	if err = w.Flush(); err != nil {
		return err
	}
	return nil
}

func getContainerInfoByFile(file os.FileInfo) (*ContainerInfo, error) {
	containerName := file.Name()
	dirURL := fmt.Sprintf(DefaultInfoLocation, containerName)
	configURL := filepath.Join(dirURL, ConfigName)
	config, err := ioutil.ReadFile(configURL)
	if err != nil {
		return nil, err
	}
	var info ContainerInfo
	if err = json.Unmarshal(config, &info); err != nil {
		return nil, err
	}
	return &info, nil
}
