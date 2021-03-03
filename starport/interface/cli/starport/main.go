package main

import (
	"context"
	"fmt"
	"os"

	starportcmd "github.com/zhigui-projects/zeus-onestop/starport/interface/cli/starport/cmd"
	"github.com/zhigui-projects/zeus-onestop/starport/pkg/clictx"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
			os.Exit(1)
		}
	}()

	ctx := clictx.From(context.Background())
	err := starportcmd.New().ExecuteContext(ctx)

	if err == context.Canceled {
		fmt.Println("aborted")
		return
	}
	if err != nil {
		fmt.Println()
		panic(err)
	}
}
