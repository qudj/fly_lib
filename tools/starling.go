package tools

import (
	"context"
	"fmt"
	"github.com/bluele/gcache"
	servbp "github.com/qudj/fly_lib/models/proto/fly_starling_serv"
	"golang.org/x/sync/singleflight"
	"google.golang.org/grpc"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	defaultStarlingExpireTime = 60*10
	defaultCapacity = 300
)

type StarlingTool struct {
	client     servbp.StarlingServiceClient
	expireTime int64
	projectKey string
	groupKey   string

	gsf   singleflight.Group
	mu    sync.Mutex
	cache gcache.Cache
}

func InitStarlingTool(host string, projectKey, groupKey string) *StarlingTool {
	conn, err := grpc.Dial(host, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	client := servbp.NewStarlingServiceClient(conn)
	ret := &StarlingTool{
		client:     client,
		expireTime: defaultStarlingExpireTime,
		projectKey: projectKey,
		groupKey:   groupKey,
		cache:      gcache.New(defaultCapacity).LFU().Build(),
	}
	return ret
}

func (f *StarlingTool) SetExpireTime (expireTime int64) {
	f.expireTime = expireTime
}

func (f *StarlingTool) SetCacheCapacity (capacity int) {
	f.cache = gcache.New(capacity).LFU().Build()
}

func (f *StarlingTool) GetTrans(ctx context.Context, lang string, keys []string) (map[string]string, error) {
	sort.Strings(keys)
	ret := make(map[string]string, 0)
	hasNot := make([]string, 0)
	for _, key := range keys {
		cacheRet, err := f.cache.Get(key)
		if err == nil && cacheRet != nil {
			ret[key] = cacheRet.(string)
			continue
		}
		hasNot = append(hasNot, key)
	}
	if len(hasNot) == 0 {
		return ret, nil
	}
	gsfKey := strings.Join(hasNot, "|")
	gsfRes, err, _ := f.gsf.Do(gsfKey, func() (interface{}, error) {
		req := &servbp.FetchTransLgsByKeyRequest{
			ProjectKey: f.projectKey,
			GroupKey:   f.groupKey,
			LangKeys:   hasNot,
			Lang:       lang,
		}
		return f.client.FetchTransLgsByKey(ctx, req)
	})
	if err != nil {
		fmt.Println(fmt.Sprintf("%v", err))
		return ret, err
	}
	res, _ := gsfRes.(*servbp.FetchTransLgsByKeyResponse)
	if res != nil && len(res.Data) > 0 {
		for _, v := range res.Data {
			ret[v.LangKey] = v.TranslateText
			expire := time.Duration(f.expireTime) * time.Second
			f.cache.SetWithExpire(v.LangKey, v.TranslateText, expire)
		}
	}
	return ret, nil
}
