import { Component } from '@angular/core';
import { MatDialog } from '@angular/material/dialog';
import { HelpDialogComponent } from 'src/app/components/help-dialog/help-dialog.component';
import { MatSnackBar } from '@angular/material/snack-bar';
import { Router } from '@angular/router';
import { Config } from '../../config/config';

@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.css']
})
export class LoginComponent {
  router!: Router

  constructor(public dialog: MatDialog, private snackBar: MatSnackBar, router: Router) {
    this.router = router

    if (localStorage.getItem('theme') == 'dark') {
      this.setTheme('dark')
    }
    if (localStorage.getItem('intro-shown') != 'true') {
      this.openHelpDialog()
    }
  }

  curTheme: string | null = null
  setTheme(theme: string) {
    this.curTheme = theme
    document.documentElement.setAttribute('data-theme', theme);
    localStorage.setItem('theme', theme)
  }

  openHelpDialog() {
    const dialogRef = this.dialog.open(HelpDialogComponent);
    dialogRef.afterClosed().subscribe(result => {
      localStorage.setItem('intro-shown', 'true')
    });
  }

  secret: string = ""
  async login() {
    if (this.secret == "") {
      var logginInMsg = this.snackBar.open("Please Enter Your Secret !", "OK", {
        duration: 2 * 1000
      });
      return
    }

    var logginInMsg = this.snackBar.open("Loggin You In ....", "OK", {
      duration: 5 * 1000
    });

    let response = fetch(Config.ApiEndpoint + "/secret", {
      "method": "POST",
      "body": JSON.stringify({ "secret": this.secret })
    })
    let rdata = await response
    if (rdata.status == 200) {
      logginInMsg.dismiss()
      localStorage.setItem('secret',this.secret)
      this.router.navigate(['/dashboard'])
      return
    }

    logginInMsg.dismiss()
    this.snackBar.open("Invalid Secret !", "OK", {
      duration: 5 * 1000
    });
  }

}
