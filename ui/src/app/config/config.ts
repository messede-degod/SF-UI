import { environment } from "src/environment/environment";

export class Config {
    public static ApiEndpoint = environment.Proto + "//" + environment.Host
    public static WSServerProto = environment.WSServerProto
    public static WSServerProtocol = this.WSServerProto
    public static WSServerUrl = this.WSServerProto + "://" + environment.Host + "/ws";
    public static MaxOpenTerminals = environment.MaxOpenTerminals
    public static ClientSecret = ""
}