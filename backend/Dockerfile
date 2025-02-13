FROM openjdk:17-jdk-slim AS build

# 安装 Maven
RUN apt-get update && apt-get install -y curl \
    && curl -O https://archive.apache.org/dist/maven/maven-3/3.8.6/binaries/apache-maven-3.8.6-bin.tar.gz \
    && tar -xzf apache-maven-3.8.6-bin.tar.gz -C /usr/local \
    && ln -s /usr/local/apache-maven-3.8.6/bin/mvn /usr/bin/mvn \
    && rm apache-maven-3.8.6-bin.tar.gz

WORKDIR /app
COPY . .
RUN mvn clean package -DskipTests

# 第二阶段：运行环境
FROM openjdk:21
WORKDIR /app
# 从构建阶段复制生成的 JAR 文件
COPY --from=build /app/target/*.jar app.jar
# 暴露端口（与 Spring Boot 配置的端口一致）
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "app.jar", "--spring.config.location=application.yml"]
