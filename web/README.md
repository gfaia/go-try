# web

go语言实现web服务，并部署在k8s集群中。

## instructions

```sh
# 编译程序
env GOOS=linux GOARCH=386 go build app/main.go

# docker编译程序
docker build -t gfaia/go-web .

# kubectl的相关命令
kubectl -n feedback-app-tianfei delete deployment go-web \
  && kubectl -n feedback-app-tianfei delete service go-web\
  && kubectl -n feedback-app-tianfei apply -f ./deployment.yaml\
  && kubectl -n feedback-app-tianfei apply -f ./service.yaml
```