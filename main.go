package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

//go run main.go run <cmd> <args>
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

func run() {
	fmt.Printf("Running %v\n", os.Args[2:])

	// func Command returns the Cmd struct to execute the named program with the
	// given arguments.

	cmd := exec.Command("/proc/self/exe",
		append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Each of these flags is a namespace flag...
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS | syscall.CLONE_NEWUSER,
		Credential: &syscall.Credential{Uid: 0, Gid: 0},
		UidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getuid(), Size: 1},
		},
		GidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getgid(), Size: 1},
		},
	}

	must(cmd.Run())

}

func child() {
	fmt.Printf("Running %v\n", os.Args[2:])

	//cg()

	// func Command returns the Cmd struct to execute the named program with the
	// given arguments.

	cmd := exec.Command(os.Args[2], os.Args[3:]...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	must(syscall.Sethostname([]byte("container")))
	must(syscall.Chroot("/var/chroot"))
	must(os.Chdir("/"))
	must(syscall.Mount("proc", "proc", "proc", 0, ""))

	// The following line owuld create an in-memory filesystem.
	must(syscall.Mount("something", "mytemp", "tmpfs", 0, ""))
	must(cmd.Run())

	must(syscall.Unmount("proc", 0))
	must(syscall.Unmount("mytemp", 0))

}

func cg() {
	cgroups := "/sys/fs/cgroup/"
	//pids := filepath.Join(cgroups, "memory")
	mem := filepath.Join(cgroups, "memory")

	// Create a subdirectory called michael (if it doesn't exist)
	os.Mkdir(filepath.Join(mem, "michael..."), 0755)
	must(ioutil.WriteFile(filepath.Join(mem, "michael/memory.limit_in_bytes"), []byte("9999424"), 0700))
	must(ioutil.WriteFile(filepath.Join(mem, "michael/memory.kmem.limit_in_bytes"), []byte("9999424"), 0700))
	must(ioutil.WriteFile(filepath.Join(mem, "michael/notify_on_release"), []byte("1"), 0700))

	//must(ioutil.WriteFile(filepath.Join(pids, "liz/pids.max"), []byte("20"), 0700))
	// Removes the new cgroup in place after the container exits
	//must(ioutil.WriteFile(filepath.Join(pids, "liz/notify_on_release"), []byte("1"), 0700))
	//must(ioutil.WriteFile(filepath.Join(pids, "liz/cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700))

	pid := strconv.Itoa(os.Getpid())
	must(ioutil.WriteFile(filepath.Join(mem, "michael/cgroup.procs"), []byte(pid), 0700))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
