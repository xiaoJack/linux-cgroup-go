package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"syscall"
)

const (
	// 挂载了 memory subsystem的hierarchy的根目录位置
	cgroupMemoryHierarchyMount = "/sys/fs/cgroup/memory"
	cgroupCPUHierarchyMount    = "/sys/fs/cgroup/cpu"
)

func main() {

	if os.Args[0] == "/proc/self/exe" {
		//容器进程
		fmt.Printf("current pid %d \n", syscall.Getpid())

		cmd := exec.Command("sh", "-c", "stress --vm-bytes 500m --vm-keep -m 1")
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			panic(err)
		}
	}

	cmd := exec.Command("/proc/self/exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		panic(err)
	}
	// 得到 fork出来进程映射在外部命名空间的pid
	fmt.Printf("%+v", cmd.Process.Pid)

	cgroupMenory(cmd.Process.Pid)
	cgroupCPU(cmd.Process.Pid)

	cmd.Process.Wait()
}

func cgroupCPU(pid int) {
	// 创建子cgroup
	newCgroupCPU := path.Join(cgroupCPUHierarchyMount, "cgroup-demo-cpu")
	os.Mkdir(newCgroupCPU, 0755)

	// 将容器进程放到子cgroup中
	if err := ioutil.WriteFile(path.Join(newCgroupCPU, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
		panic(err)
	}
	// 限制cgroup的内存使用
	if err := ioutil.WriteFile(path.Join(newCgroupCPU, "cpu.cfs_quota_us"), []byte("20000"), 0644); err != nil {
		panic(err)
	}
}

func cgroupMenory(pid int) {
	// 创建子cgroup
	newCgroupMemory := path.Join(cgroupMemoryHierarchyMount, "cgroup-demo-memory")
	os.Mkdir(newCgroupMemory, 0755)

	// 将容器进程放到子cgroup中
	if err := ioutil.WriteFile(path.Join(newCgroupMemory, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
		panic(err)
	}
	// 限制cgroup的内存使用
	if err := ioutil.WriteFile(path.Join(newCgroupMemory, "memory.limit_in_bytes"), []byte("100m"), 0644); err != nil {
		panic(err)
	}
}
