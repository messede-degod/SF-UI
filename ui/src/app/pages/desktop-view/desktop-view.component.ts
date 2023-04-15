import { Component, Input, ViewChild, ElementRef } from '@angular/core';
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
  DesktopDisconnected: boolean = false
  XpraClientReady: boolean = false
  LastPage: string = ""
  DesktopConnected: boolean = false
  FirstStart: boolean = true


  constructor(private sanitizer: DomSanitizer, private snackBar: MatSnackBar) {
    let secret = localStorage.getItem("secret");
    let wsPath = "%2Fxpraws%3Fsecret%3D" + secret

    this.IframeURL = sanitizer.bypassSecurityTrustResourceUrl("/assets/xpra_client/html5/index.html?server=" + Config.ApiHost
      + "&port=" + Config.ApiPort + "&path=" + wsPath + "&password=abc");
  }

  ngOnChanges() {
    if (this.ShowFrame && this.FirstStart) {
      this.startDesktop()
      this.FirstStart = false
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

    if (rdata.status == 200) {
      // Desktop Service was started, need to start xpra
      this.DesktopStarted = false
      this.XpraClientReady = false
      // Wait for Xpra to start on remote instance
      await new Promise(f => setTimeout(f, 5000));
      this.DesktopStarted = true
    } else if (rdata.status == 406) {
      // no active connection reload xpra
      this.DesktopStarted = false
      this.XpraClientReady = false
      // Trigger reload of xpra
      await new Promise(f => setTimeout(f, 1000));
      this.DesktopStarted = true

    } else if (rdata.status == 201) {
      // connection is active do nothing
    } else {
      this.snackBar.open("Could not start desktop!", "OK", {
        duration: 5 * 1000
      });
    }

  }

  reconnectToDesktop() {
    this.DesktopDisconnected = false
    this.DesktopStarted = false
    this.startDesktop()
  }

  XpraStateChange = () =>  {
    this.XpraClientReady = true


    let iw = document.getElementById("DesktopFrame") as HTMLIFrameElement

    if (iw != null) {
      if (iw.contentWindow != null) {
        let pn = iw.contentWindow.location.pathname
        let ps = pn.split("/")
        let currentPage = ps[ps.length - 1]

        if (this.LastPage == "index.html" && currentPage == "connect.html") {
          this.DesktopDisconnected = true
          this.XpraClientReady = false
        }


        this.LastPage = currentPage
      }
    }
  }
}
