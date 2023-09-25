package main

import (
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/siderolabs/go-procfs/procfs"
	"github.com/sirupsen/logrus"
)

const (
	TailscaleAuthKeyArg   = "tailscale.authkey"
	TailscaleHortnameArg  = "tailscale.hortname"
	TailscaleAcceptDNSArg = "tailscale.accept-dns"
	TailscaleRouteArg     = "tailscale.route"
	TailscaleAuthOnceArg  = "tailscale.authonce"
)

type TailscaleConfig struct {
	AuthKey   string `kernelarg:"tailscale.authkey" env:"TS_AUTHKEY"`
	Hostname  string `kernelarg:"tailscale.hostname" env:"TS_HOSTNAME"`
	AcceptDNS bool   `kernelarg:"tailscale.accept-dns" env:"TS_ACCEPT_DNS"`
	AuthOnce  bool   `kernelarg:"tailscale.authonce" env:"TS_AUTH_ONCE"`
}

func loadProcCmdline(cfg *TailscaleConfig, c *procfs.Cmdline) error {
	rvp := reflect.ValueOf(cfg)
	rv := rvp.Elem()
	t := rvp.Elem().Type()
	for i := 0; i < t.NumField(); i++ {
		if k, ok := t.Field(i).Tag.Lookup("kernelarg"); ok {
			if k == "" || k == "-" {
				continue
			}
			vp := c.Get(k).First()
			if vp == nil {
				continue
			}
			v := *vp
			f := rv.Field(i)
			if f.CanSet() {
				if f.Kind() == reflect.String {
					f.SetString(v)
				} else if f.Kind() == reflect.Bool {
					if b, err := strconv.ParseBool(v); err == nil {
						f.SetBool(b)
					}
				}
			}
		}
	}
	return nil
}

func dumpEnv(cfg *TailscaleConfig) (string, error) {
	env := ""
	rvp := reflect.ValueOf(cfg)
	rv := rvp.Elem()
	t := rvp.Elem().Type()
	for i := 0; i < t.NumField(); i++ {
		if k, ok := t.Field(i).Tag.Lookup("env"); ok {
			if k == "" || k == "-" {
				continue
			}
			f := rv.Field(i)
			if f.Kind() == reflect.String {
				env += k + "=" + f.String() + "\n"
			} else if f.Kind() == reflect.Bool {
				env += k + "=" + strconv.FormatBool(f.Bool()) + "\n"
			}
		}
	}
	return env, nil
}

func writeEnv(env string) error {
	filePath := "/var/etc/tailscale/auth.env"
	if envFilePath := os.Getenv("TS_ENV_FILE"); envFilePath != "" {
		filePath = envFilePath
	}
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(env)
	if err != nil {
		return err
	}
	return nil
}

func processAll() (err error) {
	logrus.Info("Parsing arguments from /proc/cmdline")
	cfg := &TailscaleConfig{}
	err = loadProcCmdline(cfg, procfs.ProcCmdline())
	if err != nil {
		logrus.Fatal("failed to load /proc/cmdline: ", err)
	}
	env, err := dumpEnv(cfg)
	if err != nil {
		logrus.Fatal("failed to dump env: ", err)
	}
	logrus.Info("Writing env to /var/etc/tailscale/auth.env")
	logrus.Debug("env: ", env)
	err = writeEnv(env)
	if err != nil {
		logrus.Fatal("failed to write env: ", err)
	}
	logrus.Info("Done writing env to /var/etc/tailscale/auth.env")
	return err
}

func main() {
	signals := make(chan os.Signal, 1)

	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logrus.Fatal("failed to rtart file watcher: ", err)
	}
	defer watcher.Close()

	done := make(chan bool, 1)
	go func() {
		defer close(done)

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				logrus.Infof("%s %s\n", event.Name, event.Op)
				// re-process /proc/cmdline
				processAll()
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logrus.Error("error:", err)
			case sig := <-signals:
				logrus.Infof("received signal %s, exiting", sig)
				done <- true
			}
		}
	}()

	processAll()

	err = watcher.Add("/proc/cmdline")
	if err != nil {
		logrus.Fatal("failed to watch /proc/cmdline: ", err)
	}
	<-done
}
