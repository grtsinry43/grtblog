# 使用docker-compose快速部署

您只需修改项目根目录中的`application.example.yml`文件（ 将其重命名为`application.yml`），然后运行以下命令即可快速部署。

```bash
docker-compose up -d
```

这将启动后端，数据库，Redis和推荐服务。无需您手动配置数据库和搜索/推荐服务。

现在，您跳转到可以[配置Nginx反向代理](/quick-start/nginx-config)，然后启动前端和管理端。
