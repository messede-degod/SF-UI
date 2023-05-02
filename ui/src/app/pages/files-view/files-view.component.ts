import { Component, Input } from '@angular/core';
import { DomSanitizer, SafeUrl } from '@angular/platform-browser';
import { Config } from 'src/app/config/config';
import { MatSnackBar } from '@angular/material/snack-bar';

@Component({
  selector: 'files-view',
  templateUrl: './files-view.component.html',
  styleUrls: ['./files-view.component.css']
})
export class FilesViewComponent {
  FbIframeURL!: SafeUrl
  @Input() ShowFrame: boolean = false
  @Input() InView: boolean = false
  FileBrowserActive: boolean = false
  FileBrowserDisconnected: boolean = false
  FileBrowserNeedsTerminal: boolean = false
  FirstStart: boolean = true
  CurrentTheme: string | null = ""
  DOMsanitizer!: DomSanitizer


  constructor(private sanitizer: DomSanitizer, private snackBar: MatSnackBar) {
    this.DOMsanitizer = sanitizer
    this.setFrameUrl()
  }

  ngOnChanges() {
    if (this.ShowFrame && this.FirstStart) {
      this.startFileBrowser()
      this.FirstStart = false
    }
    // Update Filebrowser if theme has changed
    if (this.InView) {
      if (this.CurrentTheme != localStorage.getItem("theme")) {
        this.setFrameUrl()
      }
    }
  }

  async setFrameUrl() {
    let indexFile = "index.html"
    this.CurrentTheme = localStorage.getItem("theme")
    if (this.CurrentTheme == "dark") {
      indexFile = "index-dark.html"
    }

    this.FbIframeURL = this.DOMsanitizer.bypassSecurityTrustResourceUrl(
      "/assets/filebrowser_client/" + indexFile + "#/" + localStorage.getItem("secret")
      + ',' + Config.ApiEndpoint
    );
  }

  async startFileBrowser() {
    this.FileBrowserDisconnected = false
    let clientSecret = localStorage.getItem("secret")

    let data = {
      client_secret: clientSecret
    }

    fetch(Config.ApiEndpoint + "/filebrowser", {
      "method": "POST",
      "body": JSON.stringify(data)
    })
      .then((rdata) => {
        if (rdata.status == 200) {
          this.FileBrowserActive = true
          this.FileBrowserDisconnected = false
        }
        else if (rdata.status == 451) {
          this.FileBrowserActive = false
          this.FileBrowserDisconnected = false
          this.FileBrowserNeedsTerminal = true
        }
        else {
          this.FileBrowserDisconnected = true
          this.snackBar.open("Could not start filebrowser!", "OK", {
            duration: 5 * 1000
          });
        }
      })
      .catch(() => {
        this.FileBrowserDisconnected = true
        this.snackBar.open("Could not start filebrowser!", "OK", {
          duration: 5 * 1000
        });
      })

  }
}
