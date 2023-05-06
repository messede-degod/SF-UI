import { Component, Input, Output, EventEmitter } from '@angular/core';
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
  MaxTerminalsOpen = false
  closedTermTab: number = -1
  noOfTerminals: Number = 1;
  @Output() noOfTerminalsChange = new EventEmitter<number>();

  newTerminal() {
    if (this.terminalWindows.length == Config.MaxOpenTerminals) {
      this.MaxTerminalsOpen = true
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
    this.noOfTerminalsChange.emit(this.terminalWindows.length)
  }

  async removeTerminal(termId: number) {
    this.MaxTerminalsOpen = false
    this.closedTermTab = termId
    let removedTermPos = 0
    await new Promise(f => setTimeout(f, 150));
    for (let i = 0; i < this.terminalWindows.length; i++) {
      if (this.terminalWindows[i].id == termId) {
        removedTermPos = i
        this.terminalWindows.splice(i, 1)
        break
      }
    }

    if (this.activeTerminalId == termId && this.terminalWindows.length != 0) {
      if(removedTermPos>this.terminalWindows.length-1){
        this.setTerminalActive(this.terminalWindows[this.terminalWindows.length-1].id)  
      }else{
        this.setTerminalActive(this.terminalWindows[removedTermPos].id)
      }
    }
    this.closedTermTab = -1
    this.noOfTerminalsChange.emit(this.terminalWindows.length)
  }

  setTerminalActive(termId: number) {
    this.activeTerminalId = termId
    this.activeTerminalIdChange.emit(this.activeTerminalId)
  }

  getTerminalName(termId: number): string {
    // let termTitle = window.frames[termId].document.title
    // termTitle =  termTitle.slice(0,6) + "..."
    return "Terminal " + termId
  }
}
