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
  FbIframeURL: SafeUrl
  @Input() ShowFrame: boolean = false
  FileBrowserActive: boolean = false
  FileBrowserDisconnected: boolean = false
  FirstStart: boolean = true


  constructor(private sanitizer: DomSanitizer, private snackBar: MatSnackBar) {
    this.FbIframeURL = sanitizer.bypassSecurityTrustResourceUrl(
      "/assets/filebrowser_client/#/" + localStorage.getItem("secret")
      + ',' + localStorage.getItem("theme")
      + ',' + Config.ApiEndpoint
    );
  }

  ngOnChanges() {
    if (this.ShowFrame && this.FirstStart) {
      this.startFileBrowser()
      this.FirstStart = false
    }
  }

  async startFileBrowser() {
    this.FileBrowserDisconnected = false
    let clientSecret = localStorage.getItem("secret")

    let data = {
      client_secret: clientSecret
    }

    let response = fetch(Config.ApiEndpoint + "/filebrowser", {
      "method": "POST",
      "body": JSON.stringify(data)
    })
      .then((rdata) => {
        if (rdata.status == 200 || rdata.status == 201) {
          this.FileBrowserActive = true
          this.FileBrowserDisconnected = false
        } else {
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
