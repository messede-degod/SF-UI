import { Component,Input } from '@angular/core';
import { Config } from 'src/app/config/config';
import { DomSanitizer, SafeUrl } from '@angular/platform-browser';

@Component({
  selector: 'desktop-view',
  templateUrl: './desktop-view.component.html',
  styleUrls: ['./desktop-view.component.css']
})
export class DesktopViewComponent {
  IframeURL: SafeUrl
  @Input() ShowFrame: boolean = false

  constructor(private sanitizer: DomSanitizer) {
    let secret = localStorage.getItem("secret");
    let wsPath = "%2Fxpraws%3Fsecret%3D"+secret

    this.IframeURL = sanitizer.bypassSecurityTrustResourceUrl(Config.ApiEndpoint
      + "/assets/xpra_client/html5/index.html?path="+wsPath+"&password=abc");  
  }
}
