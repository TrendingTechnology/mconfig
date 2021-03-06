package main

import (
	log "github.com/mhchlib/logger"
	"github.com/mhchlib/mconfig-api/api/v1/server"
	"github.com/mhchlib/mconfig/cmd/mconfig-server/internal"
	"github.com/mhchlib/mconfig/core"
	"github.com/mhchlib/mconfig/core/mconfig"
	"github.com/mhchlib/mconfig/core/store"
	_ "github.com/mhchlib/mconfig/core/store/plugin/etcd"
	"github.com/mhchlib/mconfig/rpc"
	"github.com/mhchlib/register"
	"github.com/mhchlib/register/regutils"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

var m *mconfig.MConfig

func init() {
	m = mconfig.NewMConfig()
	internal.ShowBanner()
	internal.ParseFlag(m)
}

const SERVICE_NAME = "mconfig-server"

func main() {
	log.SetDebugLogLevel()
	done := make(chan os.Signal, 1)
	defer core.InitMconfig(m)()
	if m.EnableRegistry {
		var registerTypeChoose register.Option
		if m.RegistryType == "etcd" {
			registerTypeChoose = register.SelectEtcdRegister()
		}
		if m.ServerIp == "" {
			ip, err := regutils.GetClientIp()
			if err != nil {
				log.Fatal("get client ip error")
			}
			m.ServerIp = ip
		}
		regClient, err := register.InitRegister(
			registerTypeChoose,
			register.ResgisterAddress(strings.Split(m.RegistryAddress, ",")),
			register.Namespace(m.Namspace),
			register.Instance(m.ServerIp+":"+strconv.Itoa(m.ServerPort)),
			register.Metadata("mode", store.GetStorePlugin().Mode),
		)
		if err != nil {
			log.Fatal(err)
		}
		demandSync := store.CheckNeedSyncData()
		if demandSync {
			err := store.SyncOtherMconfigData(regClient, SERVICE_NAME)
			if err != nil {
				log.Fatal("sync store data fail:", err)
			}
		}
		err = regClient.RegisterService(SERVICE_NAME, nil)
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			err = regClient.UnRegisterService(SERVICE_NAME)
			if err != nil {
				log.Error(err)
			}
		}()
	}
	listener, err := net.Listen("tcp", "0.0.0.0"+":"+strconv.Itoa(m.ServerPort))
	log.Info("mconfig-server listen :" + strconv.Itoa(m.ServerPort) + " success")

	if err != nil {
		log.Fatal(err)
	}
	s := grpc.NewServer()
	defer func() {
		_ = listener.Close()
		s.Stop()
	}()
	server.RegisterMConfigServer(s, rpc.NewMConfigServer())
	go func() {
		err = s.Serve(listener)
		if err != nil {
			log.Error(err)
			done <- syscall.SIGTERM
			return
		}
	}()
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done
}
