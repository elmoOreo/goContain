package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	cgroupsv2 "github.com/containerd/cgroups/v3/cgroup2"
)

func main() {
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("help")

	}

}

func cgroupCreate() {
	res := cgroupsv2.Resources{}

	pidMax := int64(5)

	pids := cgroupsv2.Pids{Max: pidMax}

	res = cgroupsv2.Resources{Pids: &pids}

	_, err := cgroupsv2.NewManager("/sys/fs/cgroup/", "/mycontainer", &res)
	if err != nil {
		fmt.Printf("Error creating cgroup: %v\n", err)
		return
	} else {
		fmt.Println("The group created successfully")
	}
	//cgroupManager.Delete()
}

func run() {
	fmt.Printf("Running %v as pid %d\n", os.Args[2:], os.Getpid())

	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | //give me an isolation around hostname
			syscall.CLONE_NEWPID, //give me an isolation around processes

	}

	must(cmd.Run())
}

func child() {
	fmt.Printf("Running %v as pid %d\n", os.Args[2:], os.Getpid())

	cgroupCreate()

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	must(syscall.Sethostname([]byte("mycontainer")))
	must(syscall.Chroot("/home/datasigntist/ubuntu_fs"))

	must(syscall.Mount("proc", "/proc", "proc", 0, ""))

	if _, err := os.Stat("myContainerTemp"); os.IsNotExist(err) {
		must(os.Mkdir("myContainerTemp", os.ModePerm))
	}
	must(syscall.Mount("something", "myContainerTemp", "tmpfs", 0, ""))

	must(syscall.Chdir("/myContainerTemp"))

	must(cmd.Run())

	must(syscall.Unmount("/proc", 0))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
