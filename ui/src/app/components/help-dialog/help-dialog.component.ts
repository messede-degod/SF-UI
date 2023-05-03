import { Component } from '@angular/core';
import { Config } from 'src/app/config/config';

@Component({
  selector: 'app-help-dialog',
  templateUrl: './help-dialog.component.html',
  styleUrls: ['./help-dialog.component.css']
})
export class HelpDialogComponent {
  SfEndpoint: string = Config.SfEndpoint
}
