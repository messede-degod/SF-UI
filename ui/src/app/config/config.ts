export class Config {
    public static ApiHost = location.hostname
    public static ApiPort = location.port
    public static ApiEndpoint =  location.protocol + "//" + location.host
    public static WSServerProtocol = "sftty"
    public static WSServerUrl = "://" + location.host + "/ws";
    public static MaxOpenTerminals = 5
    public static ClientSecret = ""
    public static DesktopDisabled = false
}


// export class Config {
//     public static ApiHost = "127.0.0.1"
//     public static ApiPort = "7171"
//     public static ApiEndpoint = "http://" + this.ApiHost + ":" + this.ApiPort
//     public static WSServerProtocol = "sftty"
//     public static WSServerUrl = "://127.0.0.1:7171/ws";
//     public static MaxOpenTerminals = 5
//     public static ClientSecret = ""
//     public static DesktopDisabled = false
// }