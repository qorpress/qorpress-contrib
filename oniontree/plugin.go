package main

import (
	"context"
	"fmt"
	"net/http"

	// move to core
	"github.com/qorpress/qorpress-contrib/oniontree/controllers"
	"github.com/qorpress/qorpress-contrib/oniontree/models"
	"github.com/qorpress/qorpress-contrib/oniontree/utils/funcmapmaker"
	"github.com/qorpress/qorpress/core/render"
	"github.com/qorpress/qorpress/pkg/config/application"
	plug "github.com/qorpress/qorpress/pkg/plugins"
)

var Tables = []interface{}{
	&models.OnionPublicKey{},
	&models.OnionService{},
	&models.OnionLink{},
	&models.OnionTag{},
	&models.OnionCategory{},
	&models.OnionSetting{},
}

var Resources = []interface{}{
	&models.OnionService{},
	&models.OnionTag{},
}

type onionTreePlugin string

func (o onionTreePlugin) Name() string      { return string(o) }
func (o onionTreePlugin) Section() string   { return `OnionTree` }
func (o onionTreePlugin) Usage() string     { return `hello` }
func (o onionTreePlugin) ShortDesc() string { return `prints greeting "hello there"` }
func (o onionTreePlugin) LongDesc() string  { return o.ShortDesc() }

func (o onionTreePlugin) Migrate() []interface{} {
	return Tables
}

func (o onionTreePlugin) Resources() []interface{} {
	return Resources
}

func (o onionTreePlugin) Routes() []http.HandlerFunc {
	h := make([]http.HandlerFunc, 0)
	return h
}

func (o onionTreePlugin) Application() application.MicroAppInterface {
	return controllers.New(&controllers.Config{})
}

func (o onionTreePlugin) FuncMapMaker(view *render.Render) *render.Render {
	return funcmapmaker.AddFuncMapMaker(view)
}

// func (o onionTreePlugin) Settings() {
// }

// func (o onionTreePlugin) Import() {} {
// }

// func (o onionTreePlugin) Export() {} {
// }

// func (o onionTreePlugin) Backup() {} {
// }

type onionTreeCommands struct{}

func (t *onionTreeCommands) Init(ctx context.Context) error {
	// to set your splash, modify the text in the println statement below, multiline is supported
	fmt.Println(`
--------------------------------------------------------------------------------------------
:'#######::'##::: ##:'####::'#######::'##::: ##::::'########:'########::'########:'########:
'##.... ##: ###:: ##:. ##::'##.... ##: ###:: ##::::... ##..:: ##.... ##: ##.....:: ##.....::
 ##:::: ##: ####: ##:: ##:: ##:::: ##: ####: ##::::::: ##:::: ##:::: ##: ##::::::: ##:::::::
 ##:::: ##: ## ## ##:: ##:: ##:::: ##: ## ## ##::::::: ##:::: ########:: ######::: ######:::
 ##:::: ##: ##. ####:: ##:: ##:::: ##: ##. ####::::::: ##:::: ##.. ##::: ##...:::: ##...::::
 ##:::: ##: ##:. ###:: ##:: ##:::: ##: ##:. ###::::::: ##:::: ##::. ##:: ##::::::: ##:::::::
. #######:: ##::. ##:'####:. #######:: ##::. ##::::::: ##:::: ##:::. ##: ########: ########:
:.......:::..::::..::....:::.......:::..::::..::::::::..:::::..:::::..::........::........::
`)

	return nil
}

func (t *onionTreeCommands) Registry() map[string]plug.Plugin {
	return map[string]plug.Plugin{
		"oniontree": onionTreePlugin("oniontree"), //OP
	}
}

var Plugins onionTreeCommands
