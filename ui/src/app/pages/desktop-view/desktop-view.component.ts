import { Component, Input, SimpleChanges } from '@angular/core';
import { Config } from 'src/app/config/config';
import { DomSanitizer, SafeUrl } from '@angular/platform-browser';
import { MatSnackBar } from '@angular/material/snack-bar';

@Component({
  selector: 'desktop-view',
  templateUrl: './desktop-view.component.html',
  styleUrls: ['./desktop-view.component.css']
})
export class DesktopViewComponent {
  IframeURL: SafeUrl
  @Input() ShowFrame: boolean = false
  DesktopStarted: boolean = false
  XpraClientReady: boolean = false

  constructor(private sanitizer: DomSanitizer, private snackBar: MatSnackBar) {
    let secret = localStorage.getItem("secret");
    let wsPath = "%2Fxpraws%3Fsecret%3D" + secret

    this.IframeURL = sanitizer.bypassSecurityTrustResourceUrl(Config.ApiEndpoint
      + "/assets/xpra_client/html5/index.html?path=" + wsPath + "&password=abc");

  }

  ngOnChanges() {
    if (this.ShowFrame) {
      this.startDesktop()
    }
  }

  async startDesktop() {
    let clientSecret = localStorage.getItem("secret")

    let data = {
      desktop_type: "xpra",
      client_secret: clientSecret
    }

    let response = fetch(Config.ApiEndpoint + "/desktop", {
      "method": "POST",
      "body": JSON.stringify(data)
    })
    let rdata = await response

    if (rdata.status != 200) {
      this.snackBar.open("Could not start desktop!", "OK", {
        duration: 5 * 1000
      });
    }

    // Wait for Xpra to start on remote instance
    await new Promise(f => setTimeout(f, 5000));

    this.DesktopStarted = true
  }

  onXpraClientLoad() {
    this.XpraClientReady = true
  }
}
