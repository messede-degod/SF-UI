import { Component, Input } from '@angular/core';
import { DashboardComponent } from '../dashboard/dashboard.component';

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

  setActiveTerminal(termId: number){
    this.activeTerminalId = termId
  }

}
