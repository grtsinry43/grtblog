# Nginx 反代配置

这部分将介绍如何配置 Nginx 反代，以便将前端、管理端和后端服务反代到同一个域名下。

其实无脑粘贴下面的配置就可以了，但是为了更好地理解，我们还是会解释一下。

```nginx
server {
	server_name next.blog.grtsinry43.com; # 替换为你的域名
	client_max_body_size 10m;
	location /admin {
    alias /home/grtsinry43/admin/; # 替换为你的管理端目录
    try_files $uri $uri/ /admin/index.html;
}

    location / {
        proxy_pass http://127.0.0.1:3000;  # 后端地址
        proxy_set_header Host $http_host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /api/v1 {
        rewrite ^/api/v1/(.*)$ /$1 break;
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $http_host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /uploads {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $http_host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /captcha {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $http_host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /socket.io { # socket.io 配置
      proxy_pass http://127.0.0.1:9091;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_set_header REMOTE-HOST $remote_addr;
      proxy_set_header Upgrade $http_upgrade;
      proxy_set_header Connection "upgrade";
      proxy_buffering off;
      proxy_http_version 1.1;
      add_header Cache-Control no-cache;
    }

    # 注意这个部分是使用 Certbot 生成的 SSL 配置，只需 sudo certbot --nginx 即可

    listen [::]:443 ssl ipv6only=on; # managed by Certbot
    listen 443 ssl; # managed by Certbot
    ssl_certificate /etc/letsencrypt/live/next.blog.grtsinry43.com/fullchain.pem; # managed by Certbot
    ssl_certificate_key /etc/letsencrypt/live/next.blog.grtsinry43.com/privkey.pem; # managed by Certbot
    include /etc/letsencrypt/options-ssl-nginx.conf; # managed by Certbot
    ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem; # managed by Certbot

}


server {
    if ($host = next.blog.grtsinry43.com) {
        return 301 https://$host$request_uri;
    } # managed by Certbot



	listen 80 default_server;
	listen [::]:80 default_server;

	server_name next.blog.grtsinry43.com;
    return 404; # managed by Certbot


}
```

这个配置文件的作用是将前端、管理端和后端服务反代到同一个域名下，同时配置了 SSL 证书。

其中，`server_name` 部分是你的域名，`alias` 部分是管理端的目录，`proxy_pass` 部分是后端服务的地址。

## 完成

恭喜你，Nginx 反代配置已经完成，接下来请继续 [启动前端](/quick-start/frontend.md)。
