SCRIPT_AUTHOR=Danil Pismenny <danil@brandymint.ru>
SCRIPT_VERSION=0.0.1
HOST=rz.iptv2022.com

all: clean build

build: bin/limehd-syslog-server

clean:
	rm -f bin/limehd-syslog-server

deploy:
	scp limehd-syslog-server root@${HOST}:/root/

bin/limehd-syslog-server:
	go build -o ./bin

shell:
	ssh root@rz.iptv2022.com
