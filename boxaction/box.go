package boxaction

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"syscall"

	"github.com/oswaldoooo/app/box"
	"github.com/oswaldoooo/app/internal/linux"
	"github.com/oswaldoooo/app/internal/network"
	"github.com/oswaldoooo/app/internal/utils"
	"github.com/oswaldoooo/app/parser"
)

const (
	BoxRootDir = "/app"
	RepoDir    = "repo"
	VMDir      = "vm"
)

var (
	Parent_NETNS_ID = "1"
	Parent_MNTNS_ID = "1"
	Parent_PIDNS_ID = "1"
	Parent_UTSNS_ID = "1"
)

func BuildBoxAction(hashvalue string) (err error) {
	hashpath := path.Join(BoxRootDir, RepoDir, hashvalue)
	obj := parser.ParseObject(hashpath)
	if obj.Err != nil {
		err = obj.Err
		return
	}
	// copy the resources to app zone
	var paths = path_join(hashpath, obj.Resources...)
	vmpath := path.Join(BoxRootDir, VMDir)
	paths = append(paths, vmpath)
	err = utils.CopyAll(paths...)
	if err != nil {
		return
	}
	//chroot
	err = syscall.Chroot(vmpath)
	return
}
func BoxExec(ctx context.Context, ID string, cmd ...string) {
	err := linux.SetNsWithFile("/proc/"+ID+"/ns/mnt", syscall.CLONE_FS)
	if err != nil {
		return
	}
}
func validate(cnf *box.BoxConfig) {
	if len(cnf.LinkFs) > 0 {
		Parent_MNTNS_ID = cnf.LinkFs
	}
	if len(cnf.LinkNet) > 0 {
		Parent_NETNS_ID = cnf.LinkNet
	}
	if len(cnf.LinkPid) > 0 {
		Parent_PIDNS_ID = cnf.LinkPid
	}
	if len(cnf.LinkUts) > 0 {
		Parent_UTSNS_ID = cnf.LinkUts
	}
}
func BoxBuild(ctx context.Context, bcnf box.BoxConfig) error {
	bnscnf := bcnf.NsConfig()
	validate(&bcnf)
	err := linux.Unshare(bnscnf.Flags, bnscnf.MountProc).Then(func(hook *linux.Hook, a any) {
		cnf := a.(box.BoxNsConfig)
		var err error
		err = linux.SetNsWithFile("/proc/1/ns/net", 0)
		_pid := hook.Data().(int)
		println("box pid", _pid)
		if err != nil {
			fmt.Fprintln(os.Stderr, "set parent ns error", err)
			syscall.Kill(_pid, syscall.SIGKILL)
			return
		}
		err = linux.SetNsWithFile("/proc/"+Parent_MNTNS_ID+"/ns/mnt", 0)
		if err != nil {
			fmt.Fprintln(os.Stderr, "set parent ns error", err)
			syscall.Kill(_pid, syscall.SIGKILL)
			return
		}
		if cnf.Flags&syscall.CLONE_NEWNET > 0 {
			//config new ns to ip netns
			pidstr := strconv.Itoa(_pid)
			os.Mkdir("/var/run/netns", 0755)
			err = os.Symlink("/proc/"+pidstr+"/ns/net", "/var/run/netns/"+pidstr)
			if err != nil {
				hook.Err = errors.New("link network to ip netns error" + err.Error())
				syscall.Kill(_pid, syscall.SIGKILL)
				return
			}
		}
		child_pid := hook.Data().(int)
		if cnf.Flags&syscall.CLONE_NEWNET > 0 {
			err := config_network(child_pid)
			if err != nil {
				syscall.Kill(child_pid, syscall.SIGINT)
				hook.Err = errors.New("config network error " + err.Error())
				return
			}
		}
	}, bnscnf).End(1)
	if err != nil {
		return errors.New("make new appbox error " + err.Error())
	}
	go notify_close()
	err = prepare_ns(bcnf.LinkNet, bcnf.LinkFs, bcnf.LinkPid, bcnf.LinkUts)
	if err != nil {
		return errors.New("herit ns error" + err.Error())
	}
	err = herit_sys_dep()
	if err != nil {
		return errors.New("herit system dependencies to appbox error " + err.Error())
	}
	if bcnf.LinkMode {
		err = link_app_resouces(rdonly, bcnf.Path)
	} else {
		err = copy_app_resources(bcnf.Path)
	}
	if err != nil {
		return errors.New("move path resource to appbox error " + err.Error())
	}
	err = link_app_resouces(rw_both, bcnf.StaticPath)
	if err != nil {
		return errors.New("move static path resource to appbox error " + err.Error())
	}
	err = linux.Chroot("/app")
	// err = syscall.Chroot("/app")
	if err != nil {
		return errors.New("chroot to appbox error " + err.Error())
	}
	dents, _ := os.ReadDir("/")
	var ds []string = make([]string, len(dents))
	for i, d := range dents {
		ds[i] = d.Name()
	}
	cmd := linux.Execute(ctx, bcnf.Run[0], bcnf.Run[1:]...).Start()
	if cmd.Err != nil {
		return errors.New("run command error " + cmd.Err.Error() + "\n" + strings.Join(bcnf.Run, " "))
	}
	println("[debug] pid", cmd.Process.Pid)
	return cmd.Wait().Err
}

func copy_app_resources(paths []string) error {
	if len(paths) == 0 {
		return nil
	}
	paths = append(paths, "/app")
	return utils.CopyAll(paths...)
}

// link app resource to app box
func link_app_resouces(rwflag _rwmode_t, paths []string) error {
	var __rdonly bool
	if rwflag&wdonly == 0 {
		__rdonly = true
	}
	for _, p := range paths {
		err := mount_bind_sys_dir(__rdonly, p)
		if err != nil {
			return err
		}
	}
	return nil
}

func config_network(child_pid int) error {
	err := network.Lock()
	if err != nil {
		return errors.New("net config is locked")
	}
	defer func() {
		err := network.Unlock()
		if err != nil {
			fmt.Fprintln(os.Stderr, "unlock net config file error", err, " but you can remove it by yourself. /etc/appbox/appbox-net.json.lock")
		}
		//unlink netns
		syscall.Unlink("/var/run/netns/" + strconv.Itoa(child_pid))
		if Parent_NETNS_ID != "1" {
			syscall.Unlink("/var/run/netns/" + Parent_NETNS_ID)
		}
	}()
	root_net, err := network.GetNetConfig()
	if err != nil {
		return err
	}
	subnet, err := network.NewSubnet(&root_net)
	if err != nil {
		return err
	}
	subnet.VethAttr.PairA.NsPid = strconv.Itoa(child_pid)
	if Parent_NETNS_ID != "1" {
		subnet.VethAttr.PairB.NsPid = Parent_NETNS_ID
	}
	cnf, err := network.GetNetConfig()
	if err != nil {
		return err
	}
	//create network interface
	err = network.NewInterface(&subnet, strconv.Itoa(child_pid))
	if err != nil {
		return err
	}
	//master interface to bridge interface
	if subnet.IsVeth() {
		err = network.IfaceMaster(subnet.VethAttr.PairB.Name, cnf.Name, "")
		if err != nil {
			return err
		}
		subnet.NsPid = strconv.Itoa(child_pid)
		root_net.SubNet = append(root_net.SubNet, subnet)
		err = network.Dump(&root_net)
		if err != nil {
			fmt.Fprintln(os.Stderr, "dump to appbox-net config error", err)
		}
	}

	return nil
}
func dump(rpath string, v any) error {
	f, err := os.OpenFile(rpath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "    ")
	return encoder.Encode(v)
}

//herit the system bin and lib

func herit_sys_dep() error {
	err := os.Mkdir("/app/dev", 0755)
	if err != nil {
		return err
	}
	// println("[debug] pid", os.Getpid())
	// time.Sleep(time.Second * 100)
	// cmd := linux.Execute(context.Background(), "mknod", "/app/dev/null", "c", "1", "3").Run()
	// err = cmd.Err
	err = linux.Mknod("/app/dev/null", 0666)
	if err != nil {
		return err
	}
	// exec.Command("mknod", "/app/dev/null", "c", "1", "3").Run()
	err = os.Chmod("/app/dev/null", 0666)
	if err != nil {
		return err
	}
	err = os.Mkdir("/app/proc", 0555)
	if err != nil {
		return err
	}
	err = syscall.Mount("proc", "/app/proc", "proc", 0, "")
	if err != nil {
		return err
	}
	err = mount_bind_sys_dir(true, "/bin")
	if err != nil {
		return err
	}
	err = mount_bind_sys_dir(true, "/lib")
	if err != nil {
		return err
	}
	_, err = os.Stat("/lib64")
	if err == nil {
		err = mount_bind_sys_dir(true, "/lib64")
		if err != nil {
			return err
		}
	}
	_, err = os.Stat("/lib32")
	if err == nil {
		err = mount_bind_sys_dir(true, "/lib32")
		if err != nil {
			return err
		}
	}
	return nil
}

func mount_bind_sys_dir(rdonly bool, name string) error {
	info, err := os.Stat(name)
	if err != nil {
		return err
	}
	to := path.Join("/app", name)
	err = os.MkdirAll(to, info.Mode())
	if err != nil {
		return err
	}
	_flag := 0
	if rdonly {
		_flag |= syscall.MS_RDONLY
	}
	return linux.MountBind(_flag, name, to)
}

func prepare_ns(link_net, link_fs, link_pid, link_uts string) error {
	// fd, err := syscall.Open("/proc/"+link_net+"/ns/net", syscall.O_RDONLY, 0644)
	// if err != nil {
	// 	return err
	// }
	// defer syscall.Close(fd)
	// ok, _, err := syscall.Syscall(syscall.SYS_SETNS, uintptr(fd), 0, 0)
	// if int(ok) < 0 {
	// 	return err
	// }
	var err error
	if len(link_net) > 0 {
		err = linux.SetNsWithFile("/proc/"+link_net+"/ns/net", 0)
		if err != nil {
			return err
		}
	}
	if len(link_fs) > 0 {
		err = linux.SetNsWithFile("/proc/"+link_fs+"/ns/mnt", 0)
		if err != nil {
			return err
		}
	}
	if len(link_pid) > 0 {
		err = linux.SetNsWithFile("/proc/"+link_pid+"/ns/pid", 0)
		if err != nil {
			return err
		}
	}
	if len(link_uts) > 0 {
		err = linux.SetNsWithFile("/proc/"+link_uts+"/ns/uts", 0)
		if err != nil {
			return err
		}
	}
	return nil
}

type _destroyed_node struct {
	_Func func(any)
	_Arg  any
}

var _end_heap []_destroyed_node

func notify_close() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	println("receive exit singal,start exit")
	for i := len(_end_heap) - 1; i >= 0; i-- {
		_end_heap[i]._Func(_end_heap[i]._Arg)
	}
	println("exit finished")
}
