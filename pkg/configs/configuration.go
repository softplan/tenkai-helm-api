package configs

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

//Configuration - Configuração geral da aplicação
type Configuration struct {
	App    App
}

//App struct
type App struct {
	Passkey string
	Elastic Elastic
	Rabbit  Rabbit
}

//Rabbit struct
type Rabbit struct {
	URI string
}

//Elastic Config Structure
type Elastic struct {
	URL      string
	Username string
	Password string
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