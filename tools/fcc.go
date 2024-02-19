package tools

import (
	"context"
	"errors"
	"fmt"
	"github.com/qudj/fly_lib/models/proto/fcc_serv"
	"golang.org/x/sync/singleflight"
	"google.golang.org/grpc"
	"sync"
	"time"
)

const (
	defaultFccExpireTime = 5 * time.Minute
)

type FccConfTool struct {
	client     fcc_serv.FccServiceClient
	expireTime time.Duration
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

func (f *FccConfTool) SetExpireTime(expireTime time.Duration) {
	f.expireTime = expireTime
}

func (f *FccConfTool) GetValue(ctx context.Context, key string) (string, error) {
	return f.getValue(ctx, key, f.expireTime)
}

func (f *FccConfTool) GetValueWithExpire(ctx context.Context, key string, expireTime time.Duration) (string, error) {
	return f.getValue(ctx, key, expireTime)
}

func (f *FccConfTool) getValue(ctx context.Context, key string, expireTime time.Duration) (string, error) {
	curTime := time.Now().Unix()
	f.mu.Lock()
	has, ok := f.confMap[key]
	f.mu.Unlock()
	if ok {
		if curTime-has.UpdateTime < int64(expireTime/time.Second) {
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
			Logger().Errorf("FetchConfig req=%v, error=%+v", req, err)
			return nil, err
		}
		if res == nil || res.BaseRet == nil || res.BaseRet.Code != 0 {
			Logger().Errorf("FetchConfig req=%v, res=%v", req, res)
			return nil, errors.New(fmt.Sprintf("%v", res))
		}

		f.mu.Lock()
		f.confMap[key] = &FccConfValue{
			Value:      res.Data.Value,
			UpdateTime: curTime,
		}
		f.mu.Unlock()

		return res.Data.Value, nil
	})
	if err != nil {
		if ok {
			return has.Value, err
		}
		return "", err
	}
	value, _ := gRes.(string)
	return value, nil
}
