package tunnel_client

import (
	"github.com/iwind/TeaGo/Tea"
	"os"
	"testing"
	"time"
)

func TestManager_Start(t *testing.T) {
	pwd, _ := os.Getwd()
	Tea.Root = pwd + "/main"

	err := SharedManager.Start()
	if err != nil {
		t.Fatal(err)
	}

	for {
		time.Sleep(24 * time.Hour)
	}
}
