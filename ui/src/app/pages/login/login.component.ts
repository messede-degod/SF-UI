import { Component } from '@angular/core';
import { MatDialog } from '@angular/material/dialog';
import { HelpDialogComponent } from 'src/app/components/help-dialog/help-dialog.component';
import { MatSnackBar, MatSnackBarRef, TextOnlySnackBar } from '@angular/material/snack-bar';
import { Router } from '@angular/router';
import { Config } from '../../config/config';
import { SaveSecretDialogComponent } from 'src/app/components/save-secret-dialog/save-secret-dialog.component';

@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.css']
})
export class LoginComponent {
  router!: Router
  rippleColor: string = "primary"

  constructor(public dialog: MatDialog, private snackBar: MatSnackBar, router: Router) {
    this.router = router

    if (localStorage.getItem('theme') == 'dark') {
      this.setTheme('dark')
    }
    if (localStorage.getItem('intro-shown') != 'true') {
      this.openHelpDialog()
    }
    let storedSecret = localStorage.getItem('secret')
    if (storedSecret != null) {
      this.LoginWithSecret = true
      this.secret = storedSecret
      this.login()
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

  showSaveSecretDialog(secret: string) {
    const dialogRef = this.dialog.open(SaveSecretDialogComponent,{
      data:{
        secret: secret
      }
    });
  }

  secret: string = ""
  logginInMsg!: MatSnackBarRef<TextOnlySnackBar>

  async login() {
    var loginData = {
      "secret": this.secret,
      "new_instance": false
    }

    if (this.LoginWithSecret) {
      let secretValid = this.secret.match('[a-zA-Z]*$')
      if (secretValid == null || secretValid[0] == '') {
        this.logginInMsg = this.snackBar.open("Please Enter A Valid Secret !", "OK", {
          duration: 2 * 1000
        });
        return
      }
      this.logginInMsg = this.snackBar.open("Loggin You In ....", "OK", {
        duration: 5 * 1000
      });
    } else {
      loginData.new_instance = true
      this.logginInMsg = this.snackBar.open("Creating A New Instance ....", "OK", {
        duration: 5 * 1000
      });
    }

    let response = fetch(Config.ApiEndpoint + "/secret", {
      "method": "POST",
      "body": JSON.stringify(loginData)
    })
    let rdata = await response
    if (rdata.status == 200) {
      this.logginInMsg.dismiss()

      if (this.LoginWithSecret) {
        localStorage.setItem('secret', this.secret)
      } else {
        let respBody = await rdata.json()
        localStorage.setItem('secret', respBody.secret)
        // We are creating a new instance, prompt the user to save the secret
        this.showSaveSecretDialog(respBody.secret)
      }

      this.router.navigate(['/dashboard'])
      return
    } else {
      localStorage.removeItem('secret')
    }

    this.logginInMsg.dismiss()
    this.snackBar.open("Invalid Secret !", "OK", {
      duration: 5 * 1000
    });
  }

  LoginWithSecret: boolean = false

  async loginWithSecret() {
    this.LoginWithSecret = true
  }

}
