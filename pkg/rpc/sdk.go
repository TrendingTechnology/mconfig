package rpc

import (
	log "github.com/mhchlib/logger"
	"github.com/mhchlib/mconfig-api/api/v1/sdk"
	"github.com/mhchlib/mconfig/pkg/client"
	"github.com/mhchlib/mconfig/pkg/mconfig"
)

// MConfigSDK ...
type MConfigSDK struct {
}

// NewMConfigSDK ...
func NewMConfigSDK() *MConfigSDK {
	return &MConfigSDK{}
}

// GetVStream ...
func (m *MConfigSDK) GetVStream(stream sdk.MConfig_GetVStreamServer) error {
	request := &sdk.GetVRequest{}
	err := stream.RecvMsg(request)
	if err != nil {
		log.Error(err)
		return err
	}
	appKey := "appKey"
	configKey := "configKey"
	env := "dev"

	c, err := client.NewClient(&client.MetaData{}, send(stream), recv(stream))
	if err != nil {
		return err
	}
	err = c.BuildClientConfigRelation(mconfig.Appkey(appKey), []mconfig.ConfigKey{mconfig.ConfigKey(configKey)}, mconfig.ConfigEnv(env))
	if err != nil {
		return err
	}
	c.Hold()
	return nil

	//localConfiCacheMd5 := ""
	//configsCache, err := config.GetConfigFromCache(appKey, request.Filters)
	//if err != nil {
	//	log.Error(err)
	//	return err
	//}
	//if configsCache == nil {
	//	//no cache
	//	// pull pkg from store
	//	configsCache, err = config.GetConfigFromStore(appKey, request.Filters)
	//	if err != nil {
	//		log.Error(appKey, request.Filters, err)
	//		return err
	//	}
	//}
	//err = sendConfig(stream, configsCache)
	//if err != nil {
	//	return err
	//}
	//client, err := client2.NewClient()
	//pkg.ClientChans.AddClient(client.Id, appKey, client.MsgChan)
	//defer func() {
	//	pkg.ClientChans.RemoveClient(client.Id, appKey)
	//}()
	//clietnStreamMsg := make(chan interface{})
	//go func() {
	//	msg := &struct{}{}
	//	err := stream.RecvMsg(&msg)
	//	log.Error(err)
	//	if err != nil {
	//		log.Error("client id：", client.Id, err)
	//	}
	//	clietnStreamMsg <- msg
	//}()
	//
	//for {
	//	select {
	//	case <-client.MsgChan:
	//		log.Info("client: ", client.Id, " get msg event, appId: ", appKey)
	//		configsCache, err = config.GetConfigFromCache(appKey, request.Filters)
	//		if err != nil {
	//			log.Error(err)
	//			return err
	//		}
	//		if ok, md5 := checkNeedNotifyClient(localConfiCacheMd5, configsCache); ok {
	//			err := sendConfig(stream, configsCache)
	//			if err != nil {
	//				log.Error(err)
	//				return err
	//			}
	//			localConfiCacheMd5 = md5
	//		}
	//	case <-clietnStreamMsg:
	//		return nil
	//	}
}

func recv(stream sdk.MConfig_GetVStreamServer) client.ClientRecvFunc {
	return func() interface{} {
		return nil
	}
}

func send(stream sdk.MConfig_GetVStreamServer) client.ClientSendFunc {
	return func(data interface{}) error {
		response := sdk.GetVResponse{
			Configs: nil,
		}
		return stream.Send(&response)
	}
}

//func checkNeedNotifyClient(localConfiCacheMd5 string, cache []*sdk.Config) (bool, string) {
//	//avoid affect md5 val
//	for _, v := range cache {
//		v.CreateTime = 0
//		v.UpdateTime = 0
//	}
//	hash := md5.New()
//	bs, _ := json.Marshal(cache)
//	hash.Write(bs)
//	sum := hash.Sum(nil)
//	if localConfiCacheMd5 == string(sum) {
//		return false, ""
//	}
//	return true, string(sum)
//}
//
//func sendConfig(stream sdk.MConfig_GetVStreamServer, configs []*sdk.Config) error {
//	err := stream.Send(&sdk.GetVResponse{
//		Configs: configs,
//	})
//	if err != nil {
//		return err
//	}
//	return nil
//}
