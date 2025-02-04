package registry

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"

	fatomic "github.com/natefinch/atomic"
)

//Registry holds the registry
var Registry map[string]string

func init() {
	Registry = make(map[string]string)
	LoadRegistry(&Registry)

}

func LoadRegistry(registry *map[string]string) {

	regmap := *registry

	registryFile := os.Getenv("BROCADE_REGISTRY")
	if registryFile == "" {
		regmap["error"] = "BROCADE_REGISTRY environment variable is not defined"
		return
	}
	info, err := os.Stat(registryFile)
	if err == nil && info.IsDir() {
		regmap["error"] = fmt.Sprintf("BROCADE_REGISTRY `%s` points to a directory. It should be a file.", registryFile)
		return
	}
	b := make([]byte, 0)
	if !errors.Is(err, fs.ErrNotExist) {
		b, err = os.ReadFile(registryFile)
		if err != nil {
			regmap["error"] = fmt.Sprintf("Cannot read file '%s' (BROCADE_REGISTRY environment variable)", registryFile)
			return
		}
	}
	if len(b) == 0 {
		b = []byte("{}")
		err = fatomic.WriteFile(registryFile, bytes.NewReader(b))
		if err != nil {
			regmap["error"] = fmt.Sprintf("Cannot initialise file '%s' (BROCADE_REGISTRY environment variable)", registryFile)
			return
		}
	}
	err = json.Unmarshal(b, &regmap)
	if err != nil {
		regmap["error"] = fmt.Sprintf("registry file '%s' does not contain valid JSON.\nUse http://jsonlint.com/", registryFile)
		return
	}
	delete(regmap, "error")
	if regmap["brocade-registry-file"] != registryFile {
		regmap["brocade-registry-file"] = registryFile
		SetRegistry("brocade-registry-file", regmap["brocade-registry-file"])
	}
	if regmap["$schema"] == "" {
		regmap["$schema"] = "https://dev.anet.be/brocade/schema/registry.schema.json"
		SetRegistry("$schema", regmap["$schema"])
	}
	return
}

//SetRegistry set a value to a key in the registry
func SetRegistry(key, value string) error {
	registryFile := os.Getenv("BROCADE_REGISTRY")
	if registryFile == "" {
		return fmt.Errorf("BROCADE_REGISTRY environment variable is not defined")
	}
	b, err := os.ReadFile(registryFile)
	if err != nil {
		return fmt.Errorf("cannot read file `%s` (BROCADE_REGISTRY environment variable): %s", registryFile, err.Error())
	}
	if len(b) == 0 {
		b = []byte("{}")
	}
	err = json.Unmarshal(b, &Registry)
	if err != nil {
		return fmt.Errorf("registry file `%s` does not contain valid JSON.\nUse http://jsonlint.com/", registryFile)
	}

	ovalue, ok := Registry[key]
	if ok && value == ovalue {
		return nil
	}
	Registry[key] = value
	r, err := json.Marshal(Registry)
	if err != nil {
		return fmt.Errorf("cannot marshal to valid JSON: %s", err.Error())
	}
	err = fatomic.WriteFile(registryFile, bytes.NewReader(r))
	if err != nil {
		return fmt.Errorf("cannot write file `%s` (BROCADE_REGISTRY environment variable): %s", registryFile, err.Error())
	}
	return nil
}

//InitRegistry set a value to a key in the registry if it does not exist
func InitRegistry(key, value string) error {
	registryFile := os.Getenv("BROCADE_REGISTRY")
	if registryFile == "" {
		return fmt.Errorf("BROCADE_REGISTRY environment variable is not defined")
	}
	b, err := os.ReadFile(registryFile)
	if err != nil {
		return fmt.Errorf("cannot read file `%s` (BROCADE_REGISTRY environment variable): %s", registryFile, err.Error())
	}
	if len(b) == 0 {
		b = []byte("{}")
	}
	err = json.Unmarshal(b, &Registry)
	if err != nil {
		return fmt.Errorf("registry file `%s` does not contain valid JSON.\nUse http://jsonlint.com/", registryFile)
	}
	_, ok := Registry[key]
	if ok {
		return nil
	}
	Registry[key] = value
	r, err := json.Marshal(Registry)
	if err != nil {
		return fmt.Errorf("cannot marshal to valid JSON: %s", err.Error())
	}
	err = fatomic.WriteFile(registryFile, bytes.NewReader(r))
	if err != nil {
		return fmt.Errorf("cannot write file `%s` (BROCADE_REGISTRY environment variable): %s", registryFile, err.Error())
	}
	return nil
}
