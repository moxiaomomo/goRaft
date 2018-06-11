# Docker container test

## dockfile

```shell
# Version 0.1

# 基础镜像
FROM ubuntu:16.04

# 维护者信息
MAINTAINER moguang

# 镜像操作命令
# RUN apt-get -y update
# RUN apt-get -y install git
RUN mkdir -p /www/web/ 

COPY ./raftsvr /www/web/

RUN chmod 777 /www/web/raftsvr

CMD ["/www/web/raftsvr"]  
```

## start container

```shell
docker network create --subnet=172.19.0.0/16 shadownet

docker run -d -p 3000:3000 -p 4000:4000 -e "DOCKER_CONTAINER=1" -e "MEMBER_NODE_SRV=172.19.0.2:3000" -e "MBER_NODE_CLI=172.19.0.2:4000" --net shadownet --ip 172.19.0.2 --name raftnode1 moxiaomomo:regsvr1.0

docker run -d -p 3001:3000 -p 4001:4000 -e "DOCKER_CONTAINER=1" -e "MEMBER_NODE_SRV=172.19.0.3:3000" -e "MBER_NODE_CLI=172.19.0.3:4000" --net shadownet --ip 172.19.0.3 --name raftnode2 moxiaomomo:regsvr1.0

docker run -d -p 3002:3000 -p 4002:4000 -e "DOCKER_CONTAINER=1" -e "MEMBER_NODE_SRV=172.19.0.4:3000" -e "MBER_NODE_CLI=172.19.0.4:4000" --net shadownet --ip 172.19.0.4 --name raftnode3 moxiaomomo:regsvr1.0
```