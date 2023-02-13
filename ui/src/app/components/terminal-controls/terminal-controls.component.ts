import { Component,Input,Output,EventEmitter } from '@angular/core';
import { Config } from 'src/app/config/config';

@Component({
  selector: 'terminal-controls',
  templateUrl: './terminal-controls.component.html',
  styleUrls: ['./terminal-controls.component.css']
})
export class TerminalControlsComponent {
  @Input() activeTerminalId: number = 1
  @Output() activeTerminalIdChange = new EventEmitter<number>();
  @Input() terminalWindows: Array<any> = []
  @Output() terminalWindowsChange = new EventEmitter<Array<any>>();

  MaxOpenTerminals = Config.MaxOpenTerminals

  newTerminal() {
    if (this.terminalWindows.length == Config.MaxOpenTerminals){
      return
    }

    let newTermId = 1
    if (this.terminalWindows.length != 0) { // Find Largest Term Id
      const newTerm = this.terminalWindows.reduce(function (p, v) {
        return (p.id > v.id ? p : v);
      });
      newTermId = newTerm.id + 1
    }

    this.terminalWindows.push({ id: newTermId, name: "" })
    this.setTerminalActive(newTermId)
  }

  removeTerminal(termId: number) {
    for (let i = 0; i < this.terminalWindows.length; i++) {
      if (this.terminalWindows[i].id == termId) {
        this.terminalWindows.splice(i, 1)
      }
    }
    if (this.activeTerminalId == termId && this.terminalWindows.length != 0) {
      this.setTerminalActive(this.terminalWindows[0].id)
    }
    console.log(this.terminalWindows,this.activeTerminalId)
  }

  setTerminalActive(termId: number) {
    this.activeTerminalId = termId
  }

  getTerminalName(termId: number): string {
    // let termTitle = window.frames[termId].document.title
    // termTitle =  termTitle.slice(0,6) + "..."
    return "Terminal " + termId
  }
}
