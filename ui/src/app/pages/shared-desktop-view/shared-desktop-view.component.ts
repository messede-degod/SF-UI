import { Component } from '@angular/core';
import { DomSanitizer, SafeUrl } from '@angular/platform-browser';
import { ActivatedRoute } from '@angular/router';
import { Config } from 'src/environments/environment';

@Component({
  selector: 'app-shared-desktop-view',
  templateUrl: './shared-desktop-view.component.html',
  styleUrls: ['./shared-desktop-view.component.css']
})
export class SharedDesktopViewComponent {
  shareSecret: string = ""
  clientId: string = ""
  shareExpired: boolean = false
  shareAvailable: boolean = false
  serverError: boolean = false
  loading: boolean = false
  NoVNCClientReady: boolean = false
  IframeURL!: SafeUrl
  shouldEncrypt: string
  desktopType: string = "novnc"
  secretRegex: RegExp = /^[a-zA-Z0-9]+$/

  constructor(private route: ActivatedRoute, private sanitizer: DomSanitizer,) {
    this.shouldEncrypt = document.location.protocol == 'https:' ? 'true' : 'false'
  }

  ngOnInit() {
    let secret = String(this.route.snapshot.params['secret']);
    let secretsParts = secret.split(":")

    if (this.secretRegex.test(secretsParts[0])) {
      this.shareSecret = secretsParts[0]
    }

    if (this.secretRegex.test(secretsParts[1])) {
      this.clientId = secretsParts[1]
    }
  }

  async loadSharedDesktop() {
    this.loading = true

    this.shareExpired = false
    this.shareAvailable = false
    this.serverError = false

    let data = {
      action: "verify",
      secret: this.shareSecret,
      client_id: this.clientId
    }

    let response = fetch(Config.ApiEndpoint + "/desktop/share", {
      "method": "POST",
      "body": JSON.stringify(data)
    })

    let rdata = await response
    switch (rdata.status) {
      case 200:
        this.shareAvailable = true
        let wsPath = "sharedDesktopWs%3Fsecret%3D" + this.shareSecret + "%26type%3D" + this.desktopType + "%26client%5Fid%3D" + this.clientId
        this.IframeURL = this.sanitizer.bypassSecurityTrustResourceUrl("/assets/novnc_client/vnc.html?path=" + wsPath
          + "&host=" + Config.ApiHost + "&port=" + Config.ApiPort + "&encrypt=" + this.shouldEncrypt
          + "&autoconnect=true&shared=true&reconnect=false&logging=error&resize=scale");
        break
      case 410:
      case 403:
        this.shareExpired = true
        break
      default:
        this.serverError = true
    }
    this.loading = false
  }

  stateChange() {
    this.NoVNCClientReady = true
  }
}
