package k8s

import (
	"io"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/deprecated/scheme"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

func ExecCmd(podName, containerName, namespace string,
	command string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	kubeCfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)
	restCfg, err := kubeCfg.ClientConfig()
	if err != nil {
		return err
	}
	coreClient, err := corev1.NewForConfig(restCfg)
	if err != nil {
		return err
	}

	cmd := []string{
		"sh",
		"-c",
		command,
	}
	req := coreClient.RESTClient().Post().Resource("pods").Name(podName).
		Namespace(namespace).SubResource("exec").Param("container", containerName)
	option := &v1.PodExecOptions{
		Command: cmd,
		Stdin:   true,
		Stdout:  true,
		Stderr:  true,
		TTY:     true,
	}
	if stdin == nil {
		option.Stdin = false
	}
	req.VersionedParams(
		option,
		scheme.ParameterCodec,
	)
	exec, err := remotecommand.NewSPDYExecutor(restCfg, "POST", req.URL())
	if err != nil {
		return err
	}
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
	})
	if err != nil {
		return err
	}

	return nil
}
