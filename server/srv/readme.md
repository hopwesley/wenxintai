upstream app_backend {
ip_hash; # 如果你想保证业务+SSE 在同一台机器，就加这个
server 10.0.0.1:8080;
server 10.0.0.2:8080;
}

location /api/ {
proxy_pass http://app_backend;

    # 业务接口的超时（请求 & 响应）
    proxy_read_timeout  15s;
    proxy_send_timeout  15s;
}



location /api/sub/ {
proxy_pass http://app_backend;

    # SSE 要关闭 buffer，直接逐行转发
    proxy_buffering off;

    # 浏览器到 Nginx 之间保持连接
    proxy_http_version 1.1;
    proxy_set_header Connection "";

    # 关键：这里的超时时间要放大很多（比如 1 小时、甚至几小时）
    # proxy_read_timeout 表示：Nginx 在「从后端读取数据」时，如果在 N 秒内啥都收不到，就认为超时。
    proxy_read_timeout  1h;
    proxy_send_timeout  1h;
}
