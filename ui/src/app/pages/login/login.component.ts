import { Component } from '@angular/core';
import { MatDialog } from '@angular/material/dialog';
import { HelpDialogComponent } from 'src/app/components/help-dialog/help-dialog.component';
import { MatSnackBar, MatSnackBarRef, TextOnlySnackBar } from '@angular/material/snack-bar';
import { Router } from '@angular/router';
import { Config } from 'src/environments/environment';
import { SaveSecretDialogComponent } from 'src/app/components/save-secret-dialog/save-secret-dialog.component';
import { DuplicateSessionDialogComponent } from 'src/app/components/duplicate-session-dialog/duplicate-session-dialog.component';

@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.css']
})
export class LoginComponent {
  router!: Router
  rippleColor: string = "primary"
  loadingDashBoard: boolean = false
  buildHash: string = Config.BuildHash
  server: string = Config.SfEndpoint
  loginDisabled: boolean = false

  constructor(public dialog: MatDialog, private snackBar: MatSnackBar, router: Router) {
    this.router = router

    if (localStorage.getItem('theme') == 'dark') {
      this.setTheme('dark')
    } else {
      this.setTheme('light')
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
    this.dialog.open(HelpDialogComponent);
  }

  openSource() {
    window.open("https://github.com/messede-degod/SF-UI", "_blank")
  }

  openDonations() {
    window.open("https://www.thc.org/segfault/upgrade/", "_blank")
  }

  openBuildInfo() {
    window.open("https://github.com/messede-degod/SF-UI/commit/" + this.buildHash, "_blank")
  }

  showSaveSecretDialog(secret: string) {
    this.dialog.open(SaveSecretDialogComponent, {
      data: {
        secret: secret
      }
    });
  }

  secret: string = ""
  logginInMsg!: MatSnackBarRef<TextOnlySnackBar>

  async login() {
    if (this.loginDisabled) {
      return
    }

    this.loginDisabled = true

    let loginData = {
      "secret": this.secret,
      "new_instance": false,
      "tab_id": Config.TabId
    }

    if (this.LoginWithSecret) {
      let secretValid = this.secret.match('^[a-zA-Z0-9]{6,}$')
      if (secretValid == null || secretValid[0] == '') {
        this.logginInMsg = this.snackBar.open("Please Enter A Valid Secret !", "OK", {
          duration: 2 * 1000
        });
        this.loginDisabled = false
        return
      }
      this.logginInMsg = this.snackBar.open("Loggin You In ....", "OK", {
        duration: 8 * 1000
      });
    } else {
      loginData.new_instance = true
      this.logginInMsg = this.snackBar.open("Creating A New Instance ....", "OK", {
        duration: 8 * 1000
      });
    }

    let response = fetch(Config.ApiEndpoint + "/secret", {
      "method": "POST",
      "body": JSON.stringify(loginData)
    })
    let rdata = await response
    if (rdata.status == 200) {
      this.logginInMsg.dismiss()

      let response = await rdata.json()


      if (response.is_duplicate_session) {
        // Prompt if session is duplicate
        let LoggedOutOfAllSessionsPromise = this.handleDuplicateSession()
        let LoggedOutOfAllSessions = await LoggedOutOfAllSessionsPromise

        if (!LoggedOutOfAllSessions) { // dont go to dashboard
          this.loginDisabled = false
          return
        } else {  // fresh login after killing all previous sessions
          this.loginDisabled = false
          this.login()
        }
      }

      this.loadingDashBoard = true

      if (this.LoginWithSecret) {
        localStorage.setItem('secret', this.secret)
      } else {
        localStorage.setItem('secret', response.secret)
        // We are creating a new instance, prompt the user to save the secret
        this.showSaveSecretDialog(response.secret)
      }

      Config.LoggedIn = true
      this.router.navigate(['/dashboard'])
      this.loginDisabled = false
      return
    } else {
      localStorage.removeItem('secret')
    }

    this.logginInMsg.dismiss()
    this.snackBar.open("Invalid Secret !", "OK", {
      duration: 5 * 1000
    });
    this.loginDisabled = false
  }

  async handleDuplicateSession(): Promise<boolean> {
    return new Promise((resolve, reject) => {
      const dialogRef = this.dialog.open(DuplicateSessionDialogComponent, {
        data: {
          secret: this.secret
        }
      });
      dialogRef.afterClosed().subscribe(result => {
        if (result != undefined) {
          resolve(result.Logout)
        }
        resolve(false) // return false if undefined
      })
    })
  }

  LoginWithSecret: boolean = false

  async toggleLoginWithSecret() {
    this.LoginWithSecret = !this.LoginWithSecret
  }

}
