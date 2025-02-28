package utils

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

func LoadEnvConfig() {

	log.Println("Loading env config")

	viper.SetConfigType("env")
	viper.SetConfigFile("../../../" + "test.env")

	//viper.ReadInConfig()
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		log.Fatalf("Fatal error config file: %w \n", err)
	}

	log.Println("Loaded the env config file > " + getTestEnv())
	log.Println(viper.Get("registerpostendpoint"))
}

func getTestEnv() string {
	return os.Getenv("env")
}
