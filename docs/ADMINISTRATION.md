## Administration
This doc covers the necessary utilities and administration guidelines for SFUI.

#### Deployment With Docker
-   Clone the repo
    -   `git clone https://github.com/messede-degod/sf-ui`

-   Essential Configuration<br>
 Create a config file `cp config_example.yaml config.yaml` .
    -   Setting up segfault endpoints: <br>
        `yaml-key: sf_endpoints`. Segfault endpoints to which SFUI must connects must be specified here.
        -   SFUI uses the subdomain part of the specified endpoints as the ***endpoint name*** .
        -   During login SFUI checks if the secret is in the following format `EndpointName-SECRETXXXXXXX`, the endpoint name is then checked against the specified set of endpoints, if a match is found ssh connection is established to said endpoint else a error is thrown.
        -   New account creations are load balanced across all available enpoints in a round robin fashion.  
            
    -   Setting up SSH-key based auth:<br>
        By default SFUI uses `segfault_ssh_username` and `segfault_ssh_password` to establish SSH connections. An alternative is to use a SSH key based authentication following keys must be populated in that case: <br>
        1.  `segfault_use_ssh_key` - Set to true
        2.  `segfault_ssh_key_path` - Path to the key            
        
    -   Enabling metric logging:<br>
        SFUI can log events like logins, logouts and new account creations to a elasticsearch or openobserve, which can later be visualized using kibana / openobserve-ui.
        -   Set  `enable_metric_logging` to true
        -   Set `elastic_server_host` to hostname of elasticsearch server (only hostname dont specify protocol scheme).
        -   Set `elastic_index_name`,`elastic_username` and `elastic_password` to appropriate values.
        -   Set `open_observe_compatible` to true if using openobserve.
    
    -   Setting the maintenance secret:<br>
        SFUI provides a bunch of cmdline utilities that can be used to list, kill and ban clients, these utils are reliant on a administration api which required a predefined secret to work.
        - Set `maintenance_secret` to long and random value
        - Add the following lines to your `.bashrc`.<br> `export SF_MT_SECRET=<your_secret_here>`<br>
        `export SF_HOST=127.0.0.1:7171`<br>
        all utilities read the above environment variables.

    -   Downloading and placing the geoip mmdb:<br>
    SFUI uses the geoip mmdb to associate ip addresses with countries of origin, this information is logged to elasticsearch and also used by some of the admin utils.
        - Download the geoip lite mmdb from maxmind and place it in others/db/geoip/geo.mmdb.
        - It is recommended to update the geo ip db every 30 days.(perhaps a crontab with maxmind permanent download url can help.) 

    - Other configuration:<br>
        - Set `use_x_forwarded_for_header` to true if SFUI is behind a proxy like nginx.
        - SFUI by default listens on 127.0.0.1.7171, the listen address can be specified in the `server_bind_address` key
        - Maximum number of terminals that can be opened can be specified in `max_ws_terminals`
        - Desktop can be disabled with `disable_desktop`
        - Set `sf_ui_origin` to the the site address. ex: https://shell.segfault.net and also set `disable_origin_check` to false

    -   Building image
        - `cd sf-ui`
        - `sudo docker build -t sfui . `

    -   Starting Container
        - `sudo docker compose up -d`

    -   Proxying with nginx
        -  see `other/nginx/Readme.md`
        -  a sample configuration has been provided in `other/nginx/sample.conf`     


#### Checking Logs
`sudo docker container logs -f sfui`

#### Utilities
-  Installation:<br>
    Make sure go is installed.
    -   `cd other/admin-utils`
    -   `make all`
    -   `make install`

- Utilities:
    -   sf_clients: Lists all clients currently interacting with SFUI. Output Format is as follows.
    ```
    <client-id> <client-ip> <country> <no-of-terminals> <if-desktop-active> <session-duration>
    ```
    -   sf_kill: Kill a client
    ```
    sf_kill <client-id>
    ```
    -   sf_ban: Ban clients from using SFUI
    ```
    sf_ban <client_ip>
    ```
    -   sf_unban
    ```
    sf_unban <client_ip>
    ```
    -   sf_ban_list: List all banned client adresses.



