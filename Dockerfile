FROM golang:alpine

LABEL MAINTAINER="2390647320@qq.com"
# 为我们的镜像设置必要的环境变量
ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct

# 移动到工作目录
WORKDIR /go/src/athena-agent
# 将代码复制到容器中
COPY . .
RUN go mod tidy

# go generate 编译前自动执行代码
# go env 查看go的环境变量
# go build -o athena-agent . 打包项目生成文件名为athena-agent的二进制文件
RUN go generate && go env && go build -o athena-agent .
# 移动到工作目录
WORKDIR /go/src/athena-agent/test
RUN go mod tidy
RUN go generate && go env && go build -o poseidon .


FROM alpine:latest

#设置容器时区为上海
RUN apk add tzdata && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone \
    && apk del tzdata

LABEL MAINTAINER="2390647320@qq.com"
WORKDIR /go/src/athena-agent

# 把/go/src/athena-agent 整个文件夹的文件到当前工作目录
COPY --from=0 /go/src/athena-agent ./

ENTRYPOINT ./athena-agent $0 $@