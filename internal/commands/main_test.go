package commands

import (
	"os"
	"testing"

	"github.com/neetozone/neeto-cli-commons/cli"
	"github.com/neetozone/neeto-cli-commons/config"
	product "github.com/neetozone/neeto-invoice-cli"
)

func TestMain(m *testing.M) {
	cfg, err := config.Parse(product.ConfigYAML)
	if err != nil {
		panic(err)
	}
	Register(cli.New(*cfg))
	os.Exit(m.Run())
}
