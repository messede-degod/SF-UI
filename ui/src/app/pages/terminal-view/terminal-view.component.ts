import { Component, Input } from '@angular/core';

@Component({
  selector: 'terminal-view',
  templateUrl: './terminal-view.component.html',
  styleUrls: ['./terminal-view.component.css'],
})
export class TerminalViewComponent {
  activeTerminalId: number = 1
  terminalWindows: Array<any> = [
    { id: 1, name: "" }, // Default tab that is open on first launch
  ]

  setActiveTerminal(termId: number){
    this.activeTerminalId = termId
  }

}
