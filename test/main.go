package main

import (
	"context"
	"fmt"
	"github.com/qudj/fly_lib/tools"
	"time"
)

func main() {
	ctx := context.Background()
	sHost := "localhost:50053"
	sProjectKey := "project1"
	sGroupKey := "group1"
	starling := tools.InitStarlingTool(sHost, sProjectKey, sGroupKey)
	starling.SetExpireTime(5*time.Second)

	fHost := "localhost:50052"
	fProjectKey := "project1"
	fGroupKey := "grou"
	fcc := tools.InitFccConfTool(fHost, fProjectKey, fGroupKey)

	pre := time.Now().UnixNano() / 1e6
	for i := 0; i < 10; i++ {
		lang, err := starling.GetTrans(ctx, "zh", []string{"lang1"})
		fmt.Println(lang, err)
		//time.Sleep(time.Second)
	}

	for i := 0; i < 10; i++ {
		conf, err := fcc.GetValue(ctx, "test_one")
		fmt.Println(conf, err)
	}
	cur := time.Now().UnixNano() / 1e6
	fmt.Println("time taste", cur-pre)

}
