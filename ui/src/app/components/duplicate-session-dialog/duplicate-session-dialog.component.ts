import { Component, Inject } from '@angular/core';
import { MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';
import { Config } from 'src/environments/environment';

export interface DialogData {
  secret: string
}

@Component({
  selector: 'app-duplicate-session-dialog',
  templateUrl: './duplicate-session-dialog.component.html',
  styleUrls: ['./duplicate-session-dialog.component.css']
})
export class DuplicateSessionDialogComponent {
  LogginOut: boolean = false
  constructor(@Inject(MAT_DIALOG_DATA) private data: DialogData, private dialogRef: MatDialogRef<DuplicateSessionDialogComponent>) {
  }

  async onLogout() {
    this.LogginOut = true
    this.dialogRef.disableClose = true
    await this.logout()
    this.dialogRef.close({
      Logout: true
    })
  }

  onCancel() {
    this.dialogRef.close({
      Logout: false
    })
  }

  async logout() {
    let response = fetch(Config.ApiEndpoint + "/logout", {
      "method": "POST",
      "body": JSON.stringify({
        "secret": this.data.secret,
      })
    })
    let lrdata = await response
  }

}
