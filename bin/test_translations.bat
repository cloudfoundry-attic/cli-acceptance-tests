DEL c:\Users\Administrator\go\src\github.com\cloudfoundry\cli-acceptance-tests\gcf.exe
bitsadmin.exe /transfer "DownloadStableCLI" https://s3.amazonaws.com/go-cli/builds/cf-windows-amd64.exe c:\Users\Administrator\go\src\github.com\cloudfoundry\cli-acceptance-tests\cf.exe

go get -u github.com/cloudfoundry/cli-acceptance-tests/...

SET GATSPATH=%GOPATH%\src\github.com\cloudfoundry\cli-acceptance-tests
SET PATH=%GATSPATH%;C:\Program Files\cURL\bin;%PATH%
SET CONFIG=%CD%\config.json

go install -v github.com/onsi/ginkgo/ginkgo
ginkgo.exe -r -slowSpecThreshold=120 ./translations
