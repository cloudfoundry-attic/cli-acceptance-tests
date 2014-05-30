cd %GOPATH%\src\github.com\cloudfoundry\cf-acceptance-tests

SET GATSPATH=%GOPATH%\src\github.com\pivotal-cf-experimental\GATS

DEL c:\Users\Administrator\go\src\github.com\pivotal-cf-experimental\GATS\gcf.exe
bitsadmin.exe /transfer "DownloadStableCLI" https://s3.amazonaws.com/go-cli/builds/cf-windows-amd64.exe c:\Users\Administrator\go\src\github.com\pivotal-cf-experimental\GATS\gcf.exe

SET PATH=%PATH%;%GATSPATH%;C:\Program Files\cURL\bin

SET CONFIG=%GATSPATH%\config.json

SET LOCAL_GOPATH=%GATSPATH%\Godeps\_workspace
MKDIR %LOCAL_GOPATH%\bin

SET GOPATH=%LOCAL_GOPATH%;%GOPATH%
SET PATH=%LOCAL_GOPATH%\bin;%PATH%

go install -v github.com/onsi/ginkgo/ginkgo
SET PATH=%PATH$;

Godeps\_workspace\bin\ginkgo.exe -r -slowSpecThreshold=120 ./quotas
