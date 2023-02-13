export class Config{
    // public static WSServerUrl = "ws://localhost:8080/ws";
    public static WSServerUrl = "ws://127.0.0.1:5555/v1.24/containers/1bdd154f59ba/attach/ws?logs=0&stream=1&stdin=1&stdout=1&stderr=1";
    public static WSServerProtocol = "tty";
    public static MaxOpenTerminals = 5;
}