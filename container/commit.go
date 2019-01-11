package container

import (
	"os"
	"os/exec"
)

func CommitContainer(imageName string) error {
	mntURL := "/root/test1/mnt"
	imageTar := "/root/test1/" + imageName + ".tar"
	cmd := exec.Command("tar", "-czf", imageTar, "-C", mntURL, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
