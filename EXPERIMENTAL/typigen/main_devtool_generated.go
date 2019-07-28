package typigen

import (
	"io/ioutil"

	"github.com/typical-go/typical-rest-server/EXPERIMENTAL/internal/bash"
	"github.com/typical-go/typical-rest-server/EXPERIMENTAL/typictx"
	"github.com/typical-go/typical-rest-server/EXPERIMENTAL/typienv"
	"github.com/typical-go/typical-rest-server/EXPERIMENTAL/typigen/generated"

	log "github.com/sirupsen/logrus"
)

// MainDevToolGenerated to generate code in typical package
func MainDevToolGenerated(t typictx.Context) (err error) {
	filename := typienv.TypicalDevToolMainPackage() + "/generated.go"
	log.Infof("Main Dev Tool Generated Code: %s", filename)

	recipe := generated.SourceRecipe{
		PackageName: "main",
		Imports:     make(map[string]string),
	}

	for _, lib := range devtoolSideEffects(t) {
		recipe.Imports[lib] = "_"
	}

	err = ioutil.WriteFile(filename, []byte(recipe.String()), 0644)
	if err != nil {
		return
	}

	bash.GoFmt(filename)
	return
}

func devtoolSideEffects(t typictx.Context) (sideEffects []string) {
	for _, module := range t.Modules {
		for _, sideEffect := range module.SideEffects {
			if sideEffect.TypicalDevToolFlag {
				sideEffects = append(sideEffects, sideEffect.Library)
			}
		}
	}
	return
}