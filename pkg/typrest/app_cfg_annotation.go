package typrest

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/typical-go/typical-go/pkg/common"
	"github.com/typical-go/typical-go/pkg/typannot"
	"github.com/typical-go/typical-go/pkg/typgo"
)

type (
	// AppCfgAnnotation handle @app-cfg annotation
	// e.g. `@app-cfg (prefix: "PREFIX" ctor_name:"CTOR")`
	AppCfgAnnotation struct {
		TagName  string // By default is `@app-cfg`
		Template string // By default defined in defaultCfgTemplate variable
		Target   string // By default is `cmd/PROJECT_NAME/cfg_annotated.go`
		DotEnv   bool   // If true then create and load ``.env`
	}
	// AppCfgTmplData template
	AppCfgTmplData struct {
		Package string
		Imports []string
		Configs []*AppCfg
	}
	// AppCfg model
	AppCfg struct {
		CtorName string
		Prefix   string
		SpecType string
		Fields   []*Field
	}
	// Field model
	Field struct {
		Key     string
		Default string
	}
)

const defaultCfgTemplate = `package {{.Package}}

// Autogenerated by Typical-Go. DO NOT EDIT.

import ({{range $import := .Imports}}
	"{{$import}}"{{end}}
)

func init() { {{if .Configs}}
	typapp.AppendCtor({{range $c := .Configs}}
		&typapp.Constructor{
			Name: "{{$c.CtorName}}",
			Fn: func() (*{{$c.SpecType}}, error) {
				var cfg {{$c.SpecType}}
				if err := envconfig.Process("{{$c.Prefix}}", &cfg); err != nil {
					return nil, err
				}
				return &cfg, nil
			},
		},{{end}}
	){{end}}
}`

//
// AppCfgAnnotation
//

var _ typannot.Annotator = (*AppCfgAnnotation)(nil)

// Annotate AppCfg to prepare dependency-injection and env-file
func (m *AppCfgAnnotation) Annotate(c *typannot.Context) error {
	configs := m.CreateConfigs(c)
	target := m.getTarget(c)
	if len(configs) < 1 {
		os.Remove(target)
		return nil
	}
	data := &AppCfgTmplData{
		Package: filepath.Base(c.Destination),
		Imports: c.CreateImports(typgo.ProjectPkg, "github.com/kelseyhightower/envconfig"),
		Configs: configs,
	}
	fmt.Fprintf(Stdout, "Generate @app-cfg to %s\n", target)
	if err := common.ExecuteTmplToFile(target, m.getTemplate(), data); err != nil {
		return err
	}
	typgo.GoImports(target)
	if m.DotEnv {
		if err := CreateAndLoadDotEnv(".env", configs); err != nil {
			return err
		}
	}

	return nil
}

// CreateAndLoadDotEnv to create and load envfile
func CreateAndLoadDotEnv(envfile string, configs []*AppCfg) error {
	envmap, err := common.CreateEnvMapFromFile(envfile)
	if err != nil {
		envmap = make(common.EnvMap)
	}

	var updatedKeys []string
	for _, AppCfg := range configs {
		for _, field := range AppCfg.Fields {
			if _, ok := envmap[field.Key]; !ok {
				updatedKeys = append(updatedKeys, "+"+field.Key)
				envmap[field.Key] = field.Default
			}
		}
	}
	if len(updatedKeys) > 0 {
		color.New(color.FgGreen).Fprint(Stdout, "UPDATE_ENV")
		fmt.Fprintln(Stdout, ": "+strings.Join(updatedKeys, " "))
	}

	if err := envmap.SaveToFile(envfile); err != nil {
		return err
	}

	return common.Setenv(envmap)
}

// CreateConfigs create configs
func (m *AppCfgAnnotation) CreateConfigs(c *typannot.Context) []*AppCfg {
	var configs []*AppCfg
	for _, annot := range c.FindAnnotByStruct(m.getTagName()) {
		prefix := getPrefix(annot)
		var fields []*Field
		for _, field := range annot.Type.(*typannot.StructType).Fields {
			fields = append(fields, &Field{
				Key:     fmt.Sprintf("%s_%s", prefix, getFieldName(field)),
				Default: field.Get("default"),
			})
		}

		configs = append(configs, &AppCfg{
			CtorName: getCtorName(annot),
			Prefix:   prefix,
			SpecType: fmt.Sprintf("%s.%s", annot.Package, annot.Name),
			Fields:   fields,
		})
	}
	return configs
}

func (m *AppCfgAnnotation) getTagName() string {
	if m.TagName == "" {
		m.TagName = "@app-cfg"
	}
	return m.TagName
}

func (m *AppCfgAnnotation) getTemplate() string {
	if m.Template == "" {
		m.Template = defaultCfgTemplate
	}
	return m.Template
}

func (m *AppCfgAnnotation) getTarget(c *typannot.Context) string {
	if m.Target == "" {
		m.Target = fmt.Sprintf("%s/app_cfg_annotated.go", c.Destination)
	}
	return m.Target
}

func getCtorName(annot *typannot.Annot) string {
	return annot.TagParam.Get("ctor_name")
}

func getPrefix(annot *typannot.Annot) string {
	prefix := annot.TagParam.Get("prefix")
	if prefix == "" {
		prefix = strings.ToUpper(annot.Name)
	}
	return prefix
}

func getFieldName(field *typannot.Field) string {
	name := field.Get("envconfig")
	if name == "" {
		name = strings.ToUpper(field.Name)
	}
	return name
}
