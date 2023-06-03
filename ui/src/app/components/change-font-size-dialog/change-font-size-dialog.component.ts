import { Component } from '@angular/core';
import { TerminalService } from 'src/app/services/terminal.service';

@Component({
  selector: 'app-change-font-size-dialog',
  templateUrl: './change-font-size-dialog.component.html',
  styleUrls: ['./change-font-size-dialog.component.css']
})
export class ChangeFontSizeDialogComponent {
  fontSize: number = 16
  constructor(private terminalService: TerminalService) {
    this.fontSize = terminalService.fontSize
  }

  fontSizeChanged(fontSize: number) {
    this.terminalService.changeTerminalFontSize(fontSize)
  }
}
