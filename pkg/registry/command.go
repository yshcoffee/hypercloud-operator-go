package registry

import (
	"bytes"
	k8sCmd "hypercloud-operator-go/internal/common/k8s"
)

const (
	RegistryContainerName = "registry"
)

type Exec struct {
	Command string
	Outbuf  *bytes.Buffer
	Errbuf  *bytes.Buffer
}

func NewExec() *Exec {
	return &Exec{Outbuf: &bytes.Buffer{}, Errbuf: &bytes.Buffer{}}
}

type Commander struct {
	pod, ns string
}

func NewCommander(podName, namespace string) *Commander {
	return &Commander{pod: podName, ns: namespace}
}

func (c *Commander) GarbageCollect() (out *Exec, err error) {
	out = NewExec()
	out.Command = "/bin/registry garbage-collect /etc/docker/registry/config.yml"
	if err := c.execCmd(out); err != nil {
		return out, err
	}

	return out, nil
}

func (c *Commander) execCmd(e *Exec) error {
	if err := k8sCmd.ExecCmd(c.pod, RegistryContainerName, c.ns, e.Command, nil, e.Outbuf, e.Errbuf); err != nil {
		return err
	}

	return nil
}
