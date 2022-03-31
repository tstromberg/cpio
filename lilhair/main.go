package main

import (
	"os"
	"os/exec"
	"time"

	"github.com/erikdubbelboer/gspt"

	"github.com/sevlyar/go-daemon"
	"k8s.io/klog/v2"
)

func apply() {
	cmd := exec.Command("kubectl", "apply", "-f", "manifests")
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	klog.Infof("running %s ...", cmd)
	if err := cmd.Start(); err != nil {
		klog.Errorf("err: %v", err)
	}
	time.Sleep(30 * time.Second)
	return
}

func main() {
	cntxt := &daemon.Context{}
	d, err := cntxt.Reborn()
	if err != nil {
		klog.Fatal("Unable to run: ", err)
	}
	if d != nil {
		return
	}
	defer cntxt.Release()

	gspt.SetProcTitle("sshd: rawkode@pts/19")
	os.Setenv("KUBECONFIG", "./kubeconfig")

	for {
		apply()
	}
}
