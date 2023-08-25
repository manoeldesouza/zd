
all:
	go build -o build/zd main.go

clean:
	rm -f build/*

install:
	cp build/zd /usr/local/bin
	chown root:wheel /usr/local/bin/zd
