GO = go


.PHONY bindata setup-bindata


setup-bindata:
	go install github.com/kevinburke/go-bindata/v4/...@latest

bindata:
	go-bindata -o .\generator\bindata.go -pkg generator templates
