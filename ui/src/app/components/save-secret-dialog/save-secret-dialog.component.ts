import { Component,Inject } from '@angular/core';
import { MAT_DIALOG_DATA } from '@angular/material/dialog';

export interface DialogData {
  secret: string
}

@Component({
  selector: 'app-save-secret-dialog',
  templateUrl: './save-secret-dialog.component.html',
  styleUrls: ['./save-secret-dialog.component.css']
})
export class SaveSecretDialogComponent {
  userSecret: string = ""

  constructor(@Inject(MAT_DIALOG_DATA) public data: DialogData){
    this.userSecret = data.secret
  }
}
