package utils

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/containernetworking/cni/pkg/ns"
	"github.com/docker/engine-api/client"
	"github.com/pkg/errors"
)

func BroadcastArp(iface string, count int) error {
	arpingPath, err := exec.LookPath("arping")
	if err != nil {
		return errors.Wrap(err, "Failed to lookup arping")
	}
	out, err := exec.Command(arpingPath, "-i", iface, "-B", "-c", strconv.Itoa(count)).CombinedOutput()
	logrus.Debugf("arping output: %s", out)
	return err
}

func EnterNS(dc *client.Client, dockerID string, f func(ns.NetNS) error) error {
	inspect, err := dc.ContainerInspect(context.Background(), dockerID)
	if err != nil {
		return errors.Wrapf(err, "Inspecting container: %v", dockerID)
	}

	containerNSStr := fmt.Sprintf("/proc/%v/ns/net", inspect.State.Pid)
	netns, err := ns.GetNS(containerNSStr)
	if err != nil {
		return errors.Wrapf(err, "Failed to open netns %v", containerNSStr)
	}
	defer netns.Close()

	err = netns.Do(func(n ns.NetNS) error {
		return f(n)
	})
	if err != nil {
		return errors.Wrapf(err, "In name ns for container %s", dockerID)
	}

	return nil
}
