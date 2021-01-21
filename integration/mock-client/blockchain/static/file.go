package static

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/smartcontractkit/chainlink/core/logger"
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
	defer logger.ErrorIfCalling(responsesFile.Close)

	return ioutil.ReadAll(responsesFile)
}
