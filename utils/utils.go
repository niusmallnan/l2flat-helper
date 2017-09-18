package utils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/docker/engine-api/client"
	"github.com/pkg/errors"
)

const (
	netnsDir = "/var/run/netns/"
)

func BroadcastArp(dockerID string, iface string, count int) {
	logrus.Infof("Broadcast container %s mac address", dockerID)
	out, _ := exec.Command("ip", "netns", "exec", dockerID, "arping", "-i", iface, "-B", "-c", strconv.Itoa(count)).CombinedOutput()
	logrus.Debugf("arping output: %s", out)
}

func LinkNS(dc *client.Client, dockerID string) error {
	logrus.Infof("Link container %s netns", dockerID)
	_, err := os.Stat(path.Dir(netnsDir))
	if os.IsNotExist(err) {
		os.Mkdir(path.Dir(netnsDir), os.FileMode(0777))
	}
	inspect, err := dc.ContainerInspect(context.Background(), dockerID)
	if err != nil {
		return errors.Wrapf(err, "Inspecting container: %v", dockerID)
	}

	err = CleanNS(dockerID)
	if err != nil {
		return err
	}

	containerNSStr := fmt.Sprintf("/proc/%v/ns/net", inspect.State.Pid)
	out, err := exec.Command("ln", "-s", containerNSStr, path.Join(netnsDir, dockerID)).CombinedOutput()
	logrus.Debugf("link netns output: %s", out)
	return err
}

func CleanNS(dockerID string) error {
	logrus.Debugf("Try to clean %s", path.Join(netnsDir, dockerID))
	if _, err := os.Lstat(path.Join(netnsDir, dockerID)); err == nil {
		logrus.Infof("Clean container %s netns", dockerID)
		return os.Remove(path.Join(netnsDir, dockerID))
	}
	return nil
}
