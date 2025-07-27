package experiences

import (
	"os"

	"github.com/typical-developers/goblox/opencloud"
)

var (
	Opencloud = opencloud.NewClient().WithAPIKey(os.Getenv("OPENCLOUD_API_KEY"))
)
