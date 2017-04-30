// Copyright 2017 NDP Systèmes. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"go/build"
	"os"
	"os/exec"
	"path"
	"text/template"

	"github.com/gin-gonic/gin"
	"github.com/inconshreveable/log15"
	"github.com/npiganeau/yep/yep/actions"
	"github.com/npiganeau/yep/yep/controllers"
	"github.com/npiganeau/yep/yep/menus"
	"github.com/npiganeau/yep/yep/models"
	"github.com/npiganeau/yep/yep/server"
	"github.com/npiganeau/yep/yep/tools/generate"
	"github.com/npiganeau/yep/yep/tools/logging"
	"github.com/npiganeau/yep/yep/views"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const startFileName = "start.go"

var log log15.Logger

var serverCmd = &cobra.Command{
	Use:   "server [projectDir]",
	Short: "Start the YEP server",
	Long: `Start the YEP server of the project in 'projectDir'.
If projectDir is omitted, defaults to the current directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		projectDir := "."
		if len(args) > 0 {
			projectDir = args[0]
		}
		generateAndRunFile(projectDir, startFileName, startFileTemplate)
	},
}

// generateAndRunFile creates the startup file of the project and runs it.
func generateAndRunFile(projectDir, fileName string, tmpl *template.Template) {
	projectPack, err := build.ImportDir(path.Join(projectDir, "config"), 0)
	if err != nil {
		panic(fmt.Errorf("Error while importing project path: %s", err))
	}

	tmplData := struct {
		Imports []string
		Config  string
	}{
		Imports: projectPack.Imports,
		Config:  fmt.Sprintf("%#v", viper.AllSettings()),
	}
	startFileName := path.Join(projectDir, fileName)
	generate.CreateFileFromTemplate(startFileName, tmpl, tmplData)
	cmd := exec.Command("go", "run", startFileName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

// StartServer starts the YEP server. It is meant to be called from
// a project start file which imports all the project's module.
func StartServer(config map[string]interface{}) {
	setupConfig(config)
	connectToDB()
	models.BootStrap()
	server.LoadInternalResources()
	views.BootStrap()
	actions.BootStrap()
	controllers.BootStrap()
	menus.BootStrap()
	server.PostInit()
	srv := server.GetServer()
	log.Info("YEP is up and running")
	srv.Run()
}

// setupConfig takes the given config map and stores it into the viper configuration
// It also initializes the logger
func setupConfig(config map[string]interface{}) {
	for key, value := range config {
		viper.Set(key, value)
	}
	if !viper.GetBool("Debug") {
		gin.SetMode(gin.ReleaseMode)
	}
	logging.Initialize()
	log = logging.GetLogger("init")
}

// connectToDB creates the connection to the database
func connectToDB() {
	connectString := fmt.Sprintf("dbname=%s sslmode=disable", viper.GetString("DB.Name"))
	if viper.GetString("DB.User") != "" {
		connectString += fmt.Sprintf(" user=%s", viper.GetString("DB.User"))
	}
	if viper.GetString("DB.Password") != "" {
		connectString += fmt.Sprintf(" password=%s", viper.GetString("DB.Password"))
	}
	if viper.GetString("DB.Host") != "" {
		connectString += fmt.Sprintf(" host=%s", viper.GetString("DB.Host"))
	}
	if viper.GetString("DB.Port") != "5432" {
		connectString += fmt.Sprintf(" port=%s", viper.GetString("DB.Port"))
	}
	models.DBConnect(viper.GetString("DB.Driver"), connectString)
}

func initServer() {
	YEPCmd.AddCommand(serverCmd)
}

var startFileTemplate = template.Must(template.New("").Parse(`
// This file is autogenerated by yep-server
// DO NOT MODIFY THIS FILE - ANY CHANGES WILL BE OVERWRITTEN

package main

import (
	"github.com/npiganeau/yep/cmd"
{{ range .Imports }}	_ "{{ . }}"
{{ end }}
)

func main() {
	cmd.StartServer({{ .Config }})
}
`))