GO=/usr/bin/go
RM=/bin/rm
BIN=gosubsonic
PATH=src/
GOPATH=${PWD}

${BIN}:
	${GO} install github.com/mdlayher/${BIN}
	${GO} build -o bin/${BIN} ${PATH}test.go

run:
	${GO} install ${BIN}
	${GO} run ${PATH}test.go

clean:
	${RM} -r bin/ pkg/
