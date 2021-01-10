package pkg

import (
	"context"
	"errors"
	log "github.com/mhchlib/logger"
	"github.com/mhchlib/mconfig"
	"github.com/mhchlib/mconfig-api/api/v1/sdk"
)

// InitMconfig ...
func InitMconfig(mconfig *mconfig.MConfig) func() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	Cancel = cancelFunc
	InitStore(*mconfig.StoreType)
	configChan, _ := ConfigStore.WatchAppConfigs(ctx)
	go handleEventMsg(configChan, ctx)
	go dispatchMsgToClient(ctx)
	return EndMconfig()
}

func dispatchMsgToClient(ctx context.Context) {
	for {
		select {
		case AppId, ok := <-ConfigChangeChan:
			if !ok {
				return
			}
			log.Info("app: ", AppId, "is changed, notify event to clients")
			notifyClients(AppId)
		case <-ctx.Done():
			log.Info("the function dispatch msg to client is done")
			return
		}
	}
}

func notifyClients(id Appkey) {
	clientsChans := ClientChans.GetClientsChan(id)
	if clientsChans != nil {
		for _, v := range clientsChans {
			v <- &struct{}{}
		}
	}
	log.Info("notify app config change info to ", len(clientsChans), " clients")
}

// GetConfigFromStore ...
func GetConfigFromStore(key Appkey, filters *sdk.ConfigFilters) ([]*sdk.Config, error) {
	appConfigs, err := ConfigStore.GetAppConfigs(key)
	//paser config str to ob
	if err != nil {
		return nil, err
	}
	go func() {
		err = mconfigCache.putConfigCache(key, appConfigs)
		if err != nil {
			log.Error(err)
		}
	}()
	configsForClient, err := filterConfigsForClient(&AppConfigsMap{AppConfigs: appConfigs}, filters, key)
	if err != nil {
		return nil, err
	}
	return configsForClient, nil
}

// GetConfigFromCache ...
func GetConfigFromCache(key Appkey, filters *sdk.ConfigFilters) ([]*sdk.Config, error) {
	cache, err := mconfigCache.getConfigCache(key)
	if err != nil {
		if errors.Is(err, Error_AppConfigNotFound) {
			return nil, nil
		} else {
			return nil, err
		}
	}
	configsForClient, err := filterConfigsForClient(cache, filters, key)
	if err != nil {
		return nil, err
	}
	return configsForClient, nil
}

// EndMconfig ...
func EndMconfig() func() {
	return func() {
		Cancel()
	}
}

func handleEventMsg(configChan chan *ConfigEvent, ctx context.Context) {
	log.Info("receive app config change event is started ")
	defer func() {
		log.Error("receive app config change event is closed ")
	}()
	for {
		select {
		case v, ok := <-configChan:
			if !ok {
				return
			}
			log.Info("receive app ", v.Key, " config change event ")
			//config 2 cache
			appConfigs := v.AppConfigs
			err := mconfigCache.putConfigCache(v.Key, appConfigs)
			if err != nil {
				log.Error(err)
				break
			}
			//notify client
			ConfigChangeChan <- v.Key
			log.Info("push app ", v.Key, " config change event to cache ")
		case <-ctx.Done():
			return
		}
	}
}
