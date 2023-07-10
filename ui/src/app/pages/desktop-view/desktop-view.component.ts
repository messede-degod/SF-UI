import { Component } from '@angular/core';
import { Config } from 'src/environments/environment';
import { DomSanitizer, SafeUrl } from '@angular/platform-browser';
import { MatSnackBar } from '@angular/material/snack-bar';
import { MatDialog } from '@angular/material/dialog';
import { ShareDesktopDialogComponent } from 'src/app/components/share-desktop-dialog/share-desktop-dialog.component';

@Component({
  selector: 'desktop-view',
  templateUrl: './desktop-view.component.html',
  styleUrls: ['./desktop-view.component.css']
})
export class DesktopViewComponent {
  IframeURL: SafeUrl

  DesktopRequested: boolean = false
  NoVNCClientReady: boolean = false

  LastPage: string = ""

  constructor(private sanitizer: DomSanitizer, private snackBar: MatSnackBar, public dialog: MatDialog) {
    let secret = localStorage.getItem("secret");
    let shouldEncrypt = document.location.protocol == 'https:' ? 'true' : 'false'
    let desktopType = "novnc"
    let wsPath = "desktopws%3Fsecret%3D" + secret + "%26type%3D" + desktopType
    // switch to remote scaling for larger screens since it provides better resolution
    let resize = (window.screen.width > 1920 && window.screen.height > 1080) ? "remote" : "scale"

    this.IframeURL = sanitizer.bypassSecurityTrustResourceUrl("/assets/novnc_client/vnc.html?path=" + wsPath
      + "&host=" + Config.ApiHost + "&port=" + Config.ApiPort + "&encrypt=" + shouldEncrypt
      + "&autoconnect=true&shared=true&logging=error&resize=" + resize + "&reconnect=true&max_reconnects=3");
  }

  requestDesktop() {
    this.DesktopRequested = true
  }

  stateChange() {
    this.NoVNCClientReady = true
  }

  openShareDialog() {
    this.dialog.open(ShareDesktopDialogComponent);
  }
}
