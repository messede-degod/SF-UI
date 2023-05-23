import { Component } from '@angular/core';
import { MatSnackBar } from '@angular/material/snack-bar';
import { ShareDesktopService } from 'src/app/services/sharedesktop.service';

export interface ShareDesktopDialogData {
  isActive: boolean
}

@Component({
  selector: 'app-share-desktop-dialog',
  templateUrl: './share-desktop-dialog.component.html',
  styleUrls: ['./share-desktop-dialog.component.css']
})
export class ShareDesktopDialogComponent {
  isActive: boolean = false
  viewOnly: boolean = true
  sharelink: string = ""
  enablingShare: boolean = false

  constructor(private shareDesktopService: ShareDesktopService, private snackBar: MatSnackBar) {
    this.shareDesktopService = shareDesktopService
    this.isActive = shareDesktopService.isactive
    this.sharelink = shareDesktopService.sharingLink
  }

  toggleViewOnly() {
    this.viewOnly = !this.viewOnly
  }

  async toggleSharing() {
    if (!this.isActive) {
      this.enablingShare = true
      // send enable request
      let response = await this.shareDesktopService.enableSharing(this.viewOnly)
      switch (response) {
        case 0:
          this.isActive = true
          this.sharelink = this.shareDesktopService.sharingLink
          break
        case 1:
          // desktop not active
          this.snackBar.open("Please Connect To Desktop First !", "OK", {
            duration: 5 * 1000
          });
          break
        case -1:
        // server error
      }
      this.enablingShare = false
    } else {
      let dresponse = await this.shareDesktopService.disableSharing()
      if (dresponse==0||dresponse==1) {
        this.isActive = false
      } else {
        this.isActive = false
        this.snackBar.open("Server Error !", "OK", {
          duration: 4 * 1000
        });
      }
    }
  }
}
