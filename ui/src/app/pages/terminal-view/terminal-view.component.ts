import { Component, EventEmitter,Output } from '@angular/core';

@Component({
  selector: 'terminal-view',
  templateUrl: './terminal-view.component.html',
  styleUrls: ['./terminal-view.component.css'],
})
export class TerminalViewComponent {
  activeTerminalId: number = 1
  showLogo!: boolean
  terminalWindows: Array<any> = [
    { id: 1, name: "" }, // Default tab that is open on first launch
  ]
  noOfTerminals: Number = 1;
  @Output() noOfTerminalsChange = new EventEmitter<Number>();

  setActiveTerminal(termId: number){
    this.activeTerminalId = termId
  }

  setNoOfTerminals(termNos: Number){
    this.noOfTerminalsChange.emit(termNos)
  }
}
