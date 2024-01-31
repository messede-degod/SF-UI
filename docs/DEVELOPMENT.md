## Development
This doc covers the architecture and other components SFUI.


#### Architecture:


#### Components:
- Golang API:<br>
     Heart of SFUI, performs the following functions
    -   Accepts segfault secrets from client and establishes SSH connection to segfault servers
    -   Maintains a list of active SSH connections
    -   Runs commands on behalf of users to start services like filebrowser or startxvnc on the users
        segfault instance.
    -   Dynamically creates port forwards - currently port forwards are made for acessing VNC and filebrowser ports.
    -   Wraps websockets around raw tty and VNC-RFB which enables xterm.js and NoVNC to reach them, basically acts as a websocket proxy for protocols which the browser cant talk.   

- Angular UI:<br>

- Modified NoVNC html client:<br>
    -   SFUI required the secret to be sent to it whenever a request is made, the default NoVNC client had to be modified to send the secret whenever it makes a request.
    -   NoVNC by default does not support direct clipboard copy pasting, a fix which does not exist on the source branch had to be made. 

- Modified FileBrowser UI client:<br>
    -   SFUI required the secret to be sent to it whenever a request is made, the default fileBrowser client had to be modified to send the secret whenever it makes a request.
