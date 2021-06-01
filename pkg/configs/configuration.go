package configs

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

//Server - Dados do servidor
type Server struct {
	Port string
}

//Configuration - Configuração geral da aplicação
type Configuration struct {
	App    App
	Server Server
}

//App struct
type App struct {
	Rabbit Rabbit
	Dbms   Dbms
}

//Rabbit struct
type Rabbit struct {
	URI string
}

//Dbms struct
type Dbms struct {
	URI string
}

//ReadConfig inicia as configurações
func ReadConfig(configFile string) (*Configuration, error) {

	var configuration Configuration

	viper.SetConfigName(configFile)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/")
	viper.AddConfigPath("/tmp/")
	viper.AddConfigPath("$HOME/")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("Error reading config file, %s", err)
	}
	err := viper.Unmarshal(&configuration)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshal config file, %s", err)
	}
	return &configuration, nil
}
