package pkg

import (
	"context"
	"crypto/md5"
	"encoding/json"
	log "github.com/mhchlib/logger"
	"github.com/mhchlib/mconfig-api/api/v1/sdk"
)

type MConfigSDK struct {
}

func NewMConfigSDK() *MConfigSDK {
	return &MConfigSDK{}
}

func (m *MConfigSDK) GetVStream(ctx context.Context, request *sdk.GetVRequest, stream sdk.MConfig_GetVStreamStream) error {
	localConfiCacheMd5 := ""
	defer func() {
		_ = stream.Close()
	}()
	appId := AppId(request.AppId)
	configsCache, err := GetConfigFromCache(appId, request.Filters)
	if err != nil {
		log.Error(err)
		return err
	}
	if configsCache == nil {
		//no cache
		// pull mconfig from store
		configsCache, err = GetConfigFromStore(appId, request.Filters)
		if err != nil {
			log.Error(err)
			return err
		}
	}
	err = sendConfig(stream, configsCache)
	if err != nil {
		return err
	}
	client, err := NewClient()
	clientChanMap.AddClient(client.Id, appId, client.MsgChan)
	defer func() {
		clientChanMap.RemoveClient(client.Id, appId)
	}()
	clietnStreamMsg := make(chan interface{})
	go func() {
		msg := &struct{}{}
		err := stream.RecvMsg(&msg)
		if err != nil {
			log.Error("client id：", client.Id, err)
		}
		clietnStreamMsg <- msg
	}()

	for {
		select {
		case <-client.MsgChan:
			log.Info("client: ", client.Id, " get msg event, appId: ", appId)
			configsCache, err = GetConfigFromCache(appId, request.Filters)
			if err != nil {
				log.Error(err)
				return err
			}
			if ok, md5 := checkNeedNotifyClient(localConfiCacheMd5, configsCache); ok {
				err := sendConfig(stream, configsCache)
				if err != nil {
					log.Error(err)
					return err
				}
				localConfiCacheMd5 = md5
			}
		case <-clietnStreamMsg:
			return nil
		}
	}
}

func checkNeedNotifyClient(localConfiCacheMd5 string, cache []*sdk.Config) (bool, string) {
	hash := md5.New()
	bs, _ := json.Marshal(cache)
	hash.Write(bs)
	sum := hash.Sum(nil)
	if localConfiCacheMd5 == string(sum) {
		return false, ""
	}
	return true, string(sum)
}

func sendConfig(stream sdk.MConfig_GetVStreamStream, configs []*sdk.Config) error {
	err := stream.Send(&sdk.GetVResponse{
		Configs: configs,
	})
	if err != nil {
		return err
	}
	return nil
}