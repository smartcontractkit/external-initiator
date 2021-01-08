package static

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

func Get(platform string) ([]byte, error) {
	wd, _ := os.Getwd()
	if !strings.HasSuffix(wd, "/blockchain") {
		wd += "/blockchain"
	}
	responsesPath := path.Join(wd, fmt.Sprintf("static/%s.json", platform))
	responsesFile, err := os.Open(responsesPath)
	if err != nil {
		return nil, err
	}
	defer responsesFile.Close()

	return ioutil.ReadAll(responsesFile)
}
