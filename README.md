# knative 镜像自动构建工具
knative镜像ko build自动执行

## 主要功能
### 1. 自动检测指定目录下 `./cmd` 目录中main.go，并自动完成`ko build`
### 2. 对于构建过程中选择`-local` 参数加载到本地的knative镜像，自动重新tag并push到指定仓库；支持自动替换 / 为 - 以适应docker hub个人仓库的名称限制。

## 注意
- eventing-broker-kafka 中 data-plane 镜像需要进行Java构建，本工具无法完成，参考一下docker build命令完成构建：
```bash
 docker buildx build --platform linux/amd64,linux/arm64 \
  -f ./data-plane/docker/Dockerfile \
  --build-arg JAVA_IMAGE=eclipse-temurin:17.0.2_8-jdk-centos7 \
  --build-arg BASE_IMAGE=eclipse-temurin:17.0.2_8-jre-centos7 \
  --build-arg APP_JAR=receiver-1.0-SNAPSHOT.jar \
  --build-arg APP_DIR=receiver \
  -t "linkinghack/knative-eventing-kafka-broker-receiver:1.2" \
  ./data-plane --push;

docker buildx build --platform linux/amd64,linux/arm64 \
  -f ./data-plane/docker/Dockerfile \
  --build-arg JAVA_IMAGE=eclipse-temurin:17.0.2_8-jdk-centos7 \
  --build-arg BASE_IMAGE=eclipse-temurin:17.0.2_8-jre-centos7 \
  --build-arg APP_JAR=dispatcher-1.0-SNAPSHOT.jar \
  --build-arg APP_DIR=dispatcher \
  -t "linkinghack/knative-eventing-kafka-broker-dispatcher:1.2" \
  ./data-plane --push
```