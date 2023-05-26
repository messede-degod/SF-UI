import { Component, AfterViewInit, ViewEncapsulation, Input } from '@angular/core';
import { ITheme, ITerminalOptions, Terminal } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';
import { AttachAddonComponent } from '../attach-addon/attach-addon.component';
import { WebglAddon } from 'xterm-addon-webgl';
import { Config } from 'src/app/config/config';
import { TerminalService } from 'src/app/services/terminal.service';


@Component({
  selector: 'app-terminal',
  templateUrl: './terminal.component.html',
  styleUrls: ['./terminal.component.css', './xterm.css'],
  encapsulation: ViewEncapsulation.None,
})
export class TerminalComponent implements AfterViewInit {
  @Input() TermId: number = 0;

  constructor(private terminalService: TerminalService) {
    this.terminalService = terminalService
  }

  ngAfterViewInit(): void {
    this.terminalService.createNewTerminal(this.TermId)
  }

  ngOnDestroy() {
    this.terminalService.removeTerminal(this.TermId)
  }
}
