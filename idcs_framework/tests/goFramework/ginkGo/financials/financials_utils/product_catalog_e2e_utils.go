package financials_utils

import (
	"fmt"
	"io/ioutil"
	"os/exec"
)

func ExecShellCommand(command string) error {
	cmd := exec.Command("/bin/sh", "-c", command)
	err := cmd.Run()
	return err
}

func GetProductsFromJson() []byte {
	const CONFIG_FILE = "../../../data/product_catalog_e2e.json"
	jsonFile, err := ioutil.ReadFile(CONFIG_FILE)

	if err != nil {
		fmt.Print("ERROR...", err.Error())
	}

	return jsonFile
}
