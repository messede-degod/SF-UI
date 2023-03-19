import { Component, AfterViewInit, ViewEncapsulation, Input } from '@angular/core';
import { ITheme, ITerminalOptions, Terminal } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';
import { AttachAddonComponent } from '../attach-addon/attach-addon.component';
import { WebglAddon } from 'xterm-addon-webgl';
import { Config } from 'src/app/config/config';


@Component({
  selector: 'app-terminal',
  templateUrl: './terminal.component.html',
  styleUrls: ['./terminal.component.css', './xterm.css'],
  encapsulation: ViewEncapsulation.None,
})
export class TerminalComponent implements AfterViewInit {

  terminal: Terminal;
  fitAddon: FitAddon;
  textEncoder: TextEncoder;
  textDecoder: TextDecoder;
  socket!: WebSocket;
  webglAddon: WebglAddon;
  termEle!: HTMLElement | null; // Html element within which we render the terminal

  SF_RESIZE: number = 1
  SF_AUTHENTICATE: number = 4


  @Input() TermId: number = 0;
  @Input() AuthToken: string = ""; // To authenticate against ttyd


  terminalOptions: ITerminalOptions = {
    fontSize: 20,
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

  constructor() {
    this.terminal = new Terminal(this.terminalOptions);
    this.fitAddon = new FitAddon();
    this.textEncoder = new TextEncoder();
    this.textDecoder = new TextDecoder();
    this.webglAddon = new WebglAddon();
  }

  ngAfterViewInit() {
    this.termEle = document.getElementById(this.TermId + "")
    this.fitAddon.activate(this.terminal)

    if (this.termEle != null) {
      this.terminal.open(this.termEle)
      this.fitAddon.fit()
      this.enableWebglRenderer()

      this.socket = new WebSocket(this.getWSURL(), Config.WSServerProtocol);
      
      // Attach The Sockets I/O to the terminal
      const attachAddon = new AttachAddonComponent(this.socket,{bidirectional: true});
      this.terminal.loadAddon(attachAddon);

      //Authenticate using Secret
      this.socket.onopen = () => {
        const termSecret = {
          secret: localStorage.getItem('secret') 
        }
        this.socket?.send(this.SF_AUTHENTICATE+JSON.stringify(termSecret))
      }

      window.onresize = () => {
        this.fitAddon.fit();
      };

      this.terminal.onResize(({cols,rows})=>{
        const terminal_size = {
          cols: cols,
          rows: rows,
        };
        this.socket.binaryType= 'blob'
        
        this.socket.send(this.SF_RESIZE+JSON.stringify(terminal_size));
      })

    }
  }

  getWSURL = () :string =>{
    // Determine whether to use ws or wss
    let wsProto = "ws"
    if (location.protocol=="https:"){
      wsProto = "wss"
    }

    return wsProto+Config.WSServerUrl
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

  ngOnDestroy(){
    this.socket.close()
    this.termEle?.remove()
    this.webglAddon.dispose()
    this.terminal.clear()
  }
}
