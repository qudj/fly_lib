package tools

import (
	"context"
	"fmt"
	"github.com/qudj/fly_lib/models/proto/fcc_serv"
	"golang.org/x/sync/singleflight"
	"google.golang.org/grpc"
	"sync"
	"time"
)

const (
	defaultFccExpireTime = 60*5
)

type FccConfTool struct {
	client     fcc_serv.FccServiceClient
	expireTime int64
	projectKey string
	groupKey   string

	gsf     singleflight.Group
	mu      sync.Mutex
	confMap map[string]*FccConfValue
}

type FccConfValue struct {
	Value      string
	UpdateTime int64
}

func InitFccConfTool(host string, projectKey, groupKey string) *FccConfTool {
	conn, err := grpc.Dial(host, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	client := fcc_serv.NewFccServiceClient(conn)
	ret := &FccConfTool{
		client:     client,
		expireTime: defaultFccExpireTime,
		projectKey: projectKey,
		groupKey:   groupKey,
		confMap:    map[string]*FccConfValue{},
	}
	return ret
}

func (f *FccConfTool) SetExpireTime(expireTime int64) {
	f.expireTime = expireTime
}

func (f *FccConfTool) GetValue(ctx context.Context, key string) (string, error) {
	curTime := time.Now().Unix()
	has, ok := f.confMap[key]
	if ok {
		if curTime-has.UpdateTime < f.expireTime {
			return has.Value, nil
		}
	}
	gRes, err, _ := f.gsf.Do(key, func() (interface{}, error) {
		req := &fcc_serv.FetchConfigRequest{
			ProjectKey: f.projectKey,
			GroupKey:   f.groupKey,
			ConfKey:    key,
		}
		res, err := f.client.FetchConfig(ctx, req)
		if err != nil {
			return nil, err
		}
		return res.Data.Value, nil
	})
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	value, _ := gRes.(string)
	f.mu.Lock()
	f.confMap[key] = &FccConfValue{
		Value:      value,
		UpdateTime: curTime,
	}
	f.mu.Unlock()
	return value, nil
}
