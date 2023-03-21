export class Config {
    public static ApiEndpoint =  location.protocol + "//" + location.host
    public static WSServerProtocol = "sftty"
    public static WSServerUrl = "://" + location.host + "/ws";
    public static MaxOpenTerminals = 5
    public static ClientSecret = ""
    public static DesktopDisabled = false
}


// export class Config {
//     public static ApiEndpoint =  "http://127.0.0.1:7171"
//     public static WSServerProtocol = "sftty"
//     public static WSServerUrl = "://127.0.0.1:7171/ws";
//     public static MaxOpenTerminals = 5
//     public static ClientSecret = ""
//     public static DesktopDisabled = false
// }