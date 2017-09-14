package macsync

import (
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/containernetworking/cni/pkg/ns"
	"github.com/docker/engine-api/client"
	"github.com/pkg/errors"
	"github.com/rancher/go-rancher-metadata/metadata"
	"github.com/rancher/l2flat-helper/utils"
)

const (
	syncLabelKey  = "io.rancher.flat.macsync"
	syncInterface = "eth0"
	syncCount     = 2
)

func Watch(mc metadata.Client, dockerClient *client.Client) error {
	w := &watcher{
		mc: mc,
		dc: dockerClient,
	}
	go mc.OnChange(5, w.onChangeNoError)
	return nil
}

type watcher struct {
	mu sync.Mutex
	dc *client.Client
	mc metadata.Client
}

func (w *watcher) onChangeNoError(version string) {
	if err := w.onChange(version); err != nil {
		logrus.Errorf("Failed to sync macaddress: %v", err)
	}
}

func (w *watcher) onChange(version string) error {
	logrus.Debug("Syncing mac address")
	w.mu.Lock()
	defer w.mu.Unlock()

	host, err := w.mc.GetSelfHost()
	if err != nil {
		return errors.Wrap(err, "Failed to get self host from metadata")
	}

	containers, err := w.mc.GetContainers()
	if err != nil {
		return errors.Wrap(err, "Failed to get containers from metadata")
	}

	for _, c := range containers {
		if (c.State != "running" && c.State != "starting") || c.HostUUID != host.UUID {
			continue
		}

		if c.Labels[syncLabelKey] == "true" {
			return utils.EnterNS(w.dc, c.ExternalId, func(n ns.NetNS) error {
				logrus.Debugf("Broadcast container %s mac address", c.ExternalId)
				return utils.BroadcastArp(syncInterface, syncCount)
			})
		}
	}

	return nil
}
