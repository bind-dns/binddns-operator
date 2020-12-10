package version

import "fmt"

var (
	APPNAME = "Unknown"

	BRANCH = "Unknown"

	TAG = "Unknown"

	REVISION = "Unknown"

	BUILDTIME = "Unknown"

	GOVERSION = "Unknown"

	BINDVERSION = "Unknown"
)

func String() string {
	return fmt.Sprintf(`-----------------------------------------
-----------------------------------------
AppName:      %v
Branch:       %v
Tag:          %v
BindVersion:  %v
Revision:     %v
Go:           %v
BuildTime:    %v
-----------------------------------------
-----------------------------------------
`, APPNAME, BRANCH, TAG, BINDVERSION, REVISION, GOVERSION, BUILDTIME)
}
