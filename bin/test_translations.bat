DEL c:\Users\Administrator\go\src\github.com\pivotal-cf-experimental\GATS\gcf.exe
bitsadmin.exe /transfer "DownloadStableCLI" https://s3.amazonaws.com/go-cli/builds/cf-windows-amd64.exe c:\Users\Administrator\go\src\github.com\pivotal-cf-experimental\GATS\gcf.exe

go get -u github.com/pivotal-cf-experimental/GATS/...

SET GATSPATH=%GOPATH%\src\github.com\pivotal-cf-experimental\GATS
SET PATH=%GATSPATH%;C:\Program Files\cURL\bin;%PATH%
SET CONFIG=%CD%\config.json
SET LOCAL_GOPATH=%GATSPATH%\Godeps\_workspace
MKDIR %LOCAL_GOPATH%\bin
SET GOPATH=%LOCAL_GOPATH%;%GOPATH%
SET PATH=%LOCAL_GOPATH%\bin;%PATH%

go install -v github.com/onsi/ginkgo/ginkgo
ginkgo.exe -r -slowSpecThreshold=120 ./translations
