import { EventEmitter, Injectable, Output } from '@angular/core';
import { ITheme, ITerminalOptions, Terminal } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';
import { AttachAddonComponent } from '../components/attach-addon/attach-addon.component';
import { WebglAddon } from 'xterm-addon-webgl';
import { Config } from 'src/environments/environment';


class SfTerminal {
    termId: number;
    terminal: Terminal;
    fitAddon: FitAddon;
    textEncoder: TextEncoder;
    textDecoder: TextDecoder;
    socket!: WebSocket;
    webglAddon: WebglAddon;
    termEle!: HTMLElement | null; // Html element within which we render the terminal

    SF_RESIZE: number = 1
    SF_AUTHENTICATE: number = 4
    SF_PING: number = 5

    keepAliveInterval!: NodeJS.Timer

    connected: EventEmitter<any> = new EventEmitter();
    disconnected: EventEmitter<any> = new EventEmitter();

    terminalOptions: ITerminalOptions = {
        fontSize: 16,
        cursorBlink: true,
        fontFamily: 'Consolas,Liberation Mono,Menlo,Courier,monospace',
        theme: {
            foreground: '#d2d2d2',
            background: '#2b2b2b',
            cursor: '#adadad',
            black: '#000000',
            red: '#d81e00',
            green: '#5ea702',
            yellow: '#cfae00',
            blue: '#427ab3',
            magenta: '#89658e',
            cyan: '#00a7aa',
            white: '#dbded8',
            brightBlack: '#686a66',
            brightRed: '#f54235',
            brightGreen: '#99e343',
            brightYellow: '#fdeb61',
            brightBlue: '#84b0d8',
            brightMagenta: '#bc94b7',
            brightCyan: '#37e6e8',
            brightWhite: '#f1f1f0',
        } as ITheme
    };

    constructor(termId: number) {
        this.termId = termId
        this.terminal = new Terminal(this.terminalOptions);
        this.fitAddon = new FitAddon();
        this.textEncoder = new TextEncoder();
        this.textDecoder = new TextDecoder();
        this.webglAddon = new WebglAddon();
    }


    create() {
        this.termEle = document.getElementById(String(this.termId))
        this.fitAddon.activate(this.terminal)

        if (this.termEle != null) {
            this.terminal.open(this.termEle)
            this.fitAddon.fit()
            this.enableWebglRenderer()

            this.socket = new WebSocket(this.getWSURL(), Config.WSServerProtocol);

            // Attach The Sockets I/O to the terminal
            const attachAddon = new AttachAddonComponent(this.socket, { bidirectional: true });
            this.terminal.loadAddon(attachAddon);

            this.terminal.writeln("Connecting to SFUI Socket...")

            //Authenticate using Secret
            this.socket.onopen = () => {
                this.terminal.clear()
                this.terminal.writeln("Connecting to instance...")
                const termSecret = {
                    secret: localStorage.getItem('secret')
                }
                this.socket?.send(this.SF_AUTHENTICATE + JSON.stringify(termSecret))
                // Resize Terminal for the first time
                this.fitAddon.fit();
                this.connected.emit(true)
            }

            window.onresize = () => {
                this.fitAddon.fit();
            };

            // Send Pings at regular interval to prevent socket disconnection
            let keepAliveInterval = setInterval(() => {
                this.socket.send(String(this.SF_PING))
            }, Config.WSPingInterval * 1000) // 40 secs

            this.socket.onclose = () => {
                clearInterval(keepAliveInterval)
                this.disconnected.emit(true)
                this.terminal.writeln("Terminal Disconnected!")
            }

            this.terminal.onResize(({ cols, rows }) => {
                const terminal_size = {
                    cols: cols,
                    rows: rows,
                };
                this.socket.binaryType = 'blob'

                this.socket.send(this.SF_RESIZE + JSON.stringify(terminal_size));
            })

        }
    }


    removeTerminal() {
        this.socket.close()
        this.termEle?.remove()
        this.webglAddon.dispose()
        this.terminal.clear()
    }


    getWSURL = (): string => {
        // Determine whether to use ws or wss
        let wsProto = "ws"
        if (location.protocol == "https:") {
            wsProto = "wss"
        }

        return wsProto + Config.WSServerUrl
    }

    enableWebglRenderer = () => {
        try {
            this.webglAddon.onContextLoss(() => {
                this.webglAddon?.dispose();
            });
            this.terminal.loadAddon(this.webglAddon);
            console.log('WebGL renderer loaded');
        } catch (e) {
            console.log('WebGL renderer could not be loaded', e);
        }
    };

}


@Injectable({
    providedIn: 'root',
})
export class TerminalService {

    terminals: Map<number, SfTerminal> = new Map()
    activeTerms: number = 0
    isactive: EventEmitter<any> = new EventEmitter();
    fontSize: number = 16

    constructor() {
        let localFontSize = localStorage.getItem("font-size")
        if (window.innerWidth < 800) {
            this.fontSize = 14
            return
        }
        this.fontSize = localFontSize === null ? 16 : Number(localFontSize)
    }

    createNewTerminal(termId: number) {
        this.removeTerminal(termId)
        let sfTerminal = new SfTerminal(termId)
        sfTerminal.terminal.options.fontSize = this.fontSize
        sfTerminal.create()

        sfTerminal.disconnected.subscribe(() => {
            this.handleTerminalClose()
        })

        sfTerminal.connected.subscribe(() => {
            this.handleTerminalOpen()
        })

        this.terminals.set(termId, sfTerminal)
    }

    removeTerminal(termId: number) {
        let sfTerminal = this.terminals.get(termId)
        if (sfTerminal != undefined) {
            sfTerminal.removeTerminal()
            this.terminals.delete(termId)
        }
    }

    changeTerminalFontSize(fontSize: number) {
        this.terminals.forEach(sfTerminal => {
            sfTerminal.terminal.options.fontSize = fontSize
            sfTerminal.terminal.refresh(0, sfTerminal.terminal.rows - 1)
            sfTerminal.fitAddon.fit()
        });
        this.fontSize = fontSize
    }

    saveFontSize(){
        localStorage.setItem("font-size", String(this.fontSize))
    }

    refresh(termId: number) {
        let sfTerminal = this.terminals.get(termId)
        if (sfTerminal != undefined) {
            sfTerminal.terminal.options.fontSize = this.fontSize
            sfTerminal.terminal.refresh(0, sfTerminal.terminal.rows - 1)
            sfTerminal.fitAddon.fit()
        }
    }

    private handleTerminalClose() {
        this.activeTerms -= 1
        if (this.activeTerms < 1) {
            this.isactive.emit(false)
        }
    }

    private handleTerminalOpen() {
        this.activeTerms += 1
        if (this.activeTerms > 0) {
            this.isactive.emit(true)
        }
    }

}