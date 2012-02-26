package main

import "errors"
import "flag"
import "fmt"
import "github.com/str1ngs/unix"
import "log"
import "os"
import "os/exec"
import "path/filepath"
import "runtime"

const (
	REG_NAME = "GO"
	REG_DONE = "/proc/sys/fs/binfmt_misc/GO"
	REG_FILE = "/proc/sys/fs/binfmt_misc/register"
)

var (
	// flags
	fregister   = flag.Bool("register", false, "register .go extenstions with binfmt")
	funregister = flag.Bool("unregister", false, "unregister .go extenstions with binfmt")

	// errors
	ErrorOsNotSupported = errors.New("binfmt is only supported on linux")
	ErrorPermissions    = errors.New("you need to be root to register or unregister with binfmt")
)

func init() {
	if runtime.GOOS != "linux" {
		log.Fatal(ErrorOsNotSupported)
	}
	log.SetFlags(log.Lshortfile)
}

func main() {
	flag.Parse()
	switch {
	case *fregister:
		register()
	case *funregister:
		unregister()
	default:
		run()
	}
}

func run() {
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}
	dir, file := filepath.Split(flag.Arg(0))
	goRun := exec.Command("go", "run", file)
	goRun.Stdout = os.Stdout
	goRun.Stderr = os.Stderr
	goRun.Dir = dir
	err := goRun.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func register() {
	if os.Geteuid() != 0 {
		log.Fatal(ErrorPermissions)
	}
	unregister()
	bin, err := filepath.Abs(os.Args[0])
	if err != nil {
		log.Fatal(err)
	}
	binFmt := fmt.Sprintf(":%s:E::go::%s:", REG_NAME, bin)
	fd, err := os.OpenFile(REG_FILE, os.O_WRONLY, 0)
	if err != nil {
		log.Fatal(err)
	}
	defer fd.Close()
	_, err = fd.WriteString(binFmt)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Go registered with binfmt")
}

func unregister() {
	if !unix.FileExists(REG_DONE) {
		fmt.Println(".go extenstions are not registered with binfmt, skipping")
		return
	}
	if os.Geteuid() != 0 {
		log.Fatal(ErrorPermissions)
	}
	fd, err := os.OpenFile(REG_DONE, os.O_WRONLY, 0)
	if err != nil {
		log.Fatal(err)
	}
	defer fd.Close()
	_, err = fd.WriteString("-1")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Go unregistered with binfmt")
}
