package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func PathExists(path string) (bool, error) {

	_, err := os.Stat(path)

	if err == nil {

		return true, nil

	}

	if os.IsNotExist(err) {

		return false, nil

	}

	return false, err

}

func main() {
	var basePath string

	var cmd = &cobra.Command{
		Use:     "goserve",
		Short:   "serve static assets",
		Long:    "like serve, can serve base path",
		Example: "goserve <path(if null will be current dir)> -p <basePath>",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			var err error
			location := "."
			if len(args) == 1 {
				location, err = filepath.Abs(args[0])
				if err != nil {
					return errors.WithStack(err)
				}
			}

			has, err := PathExists(location)
			if err != nil {
				return errors.WithStack(err)
			}
			if !has {
				return fmt.Errorf("path %v not exist", location)
			}

			if basePath != "/" {
				basePath = strings.TrimPrefix(basePath, "/")
				basePath = strings.TrimSuffix(basePath, "/")
				basePath = "/" + basePath + "/"
			}

			fs := http.FileServer(http.Dir(location))

			http.Handle(basePath, http.StripPrefix(basePath, fs))

			port := 8080
			go func() {

				var hosts []string

				addrs, err := net.InterfaceAddrs()
				if err != nil {
					fmt.Println(err)
					return
				}
				for _, address := range addrs {
					if ipnet, ok := address.(*net.IPNet); ok {
						if ipnet.IP.To4() != nil {
							hosts = append(hosts, ipnet.IP.String())
						}
					}
				}

				idleTimeout := time.NewTimer(time.Millisecond * 300)
				defer idleTimeout.Stop()
				select {
				case <-idleTimeout.C:
					fmt.Println("serve at:")
					for _, host := range hosts {
						fmt.Printf("\thttp://%v:%v%v\n", host, port, basePath)
					}
				}
			}()

			for {
				err := http.ListenAndServe(":"+strconv.Itoa(port), nil)
				if err != nil {
					err1 := syscall.Errno(syscall.EADDRINUSE)
					if errors.As(err, &err1) {
						port++
						continue
					}

					return errors.WithStack(err)
				}
			}
		},
	}

	cmd.Flags().StringVarP(&basePath, "base_relative_path", "p", "/", "config base path")

	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
