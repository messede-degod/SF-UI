# Deploying using Nginx
Deploying  UI files (html,css,js etc) with nginx is recommended for production deployments.


## Copy UI files
- Run `make UI` if you have not already.
- Copy the UI files to web root, `cp - r ui/dist/sf-ui/*  /var/www/html/`
- Adjust your nginx configurations  web root config  if necessary.

## Enable websocket proxypass
  - Edit `/etc/nginx.conf`
  - Add the following `map` within the `http` block
  ```
       map $http_upgrade $connection_upgrade {
                default upgrade;
                '' close;
        }
```

## Proxy pass  control endpoints to sf-ui
- Edit `/etc/nginx/sites-enabled/default` or any other site you wish.
- Make sure you have SF-UI running on a appropriate address that nginx can reach.
- Add the following location directives within the `server` block, (in the same order !).
```
	location ~* .(png|ico|gif|jpg|jpeg|css|js|svg|html)$ {
                # First attempt to serve request as file, then
                # as directory, then fall back to displaying a 404.
                try_files $uri $uri/ =404;
        }

        location ~ /(ws|xpraws) {
            proxy_pass http://127.0.0.1:7171;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection $connection_upgrade;
            proxy_set_header Host $host;
        }

        location / {
                proxy_pass http://127.0.0.1:7171;
        }
```
- Adjust the listening ip and port to match  SF-UIs listening address.
- Run `sudo nginx -s reload` to reload nginx and apply the settings.


UI files will now be served by nginx and the requests to other endpoints should be proxy passed to SF-UI.
