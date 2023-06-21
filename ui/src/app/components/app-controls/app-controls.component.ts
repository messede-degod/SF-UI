import { Component } from '@angular/core';
import { MatDialog } from '@angular/material/dialog';
import { Router } from '@angular/router';
import { Config } from 'src/environments/environment';
import { DashboardComponent } from 'src/app/pages/dashboard/dashboard.component';
import { TerminalViewComponent } from 'src/app/pages/terminal-view/terminal-view.component'
import { TerminalService } from 'src/app/services/terminal.service';
import { ChangeFontSizeDialogComponent } from '../change-font-size-dialog/change-font-size-dialog.component';

@Component({
  selector: 'app-controls',
  templateUrl: './app-controls.component.html',
  styleUrls: ['./app-controls.component.css'],
})
export class AppControlsComponent {
  router!: Router

  constructor(router: Router,
    private dashboardComponent: DashboardComponent,
    private terminalViewComponent: TerminalViewComponent,
    private terminalService: TerminalService,
    public dialog: MatDialog) {
    this.router = router
    if (localStorage.getItem('theme') == 'dark') {
      this.setTheme('dark')
    }
    if (localStorage.getItem('sidebarVisible') == 'false') {
      this.toggleSidebar()
    }
  }

  curTheme: string | null = null
  toggleTheme() {
    this.curTheme = document.documentElement.getAttribute('data-theme')
    var theme = 'light'
    if (this.curTheme == null || this.curTheme == 'light') {
      theme = 'dark'
    }
    this.setTheme(theme)
  }

  setTheme(theme: string) {
    this.curTheme = theme
    document.documentElement.setAttribute('data-theme', theme);
    localStorage.setItem('theme', theme)
  }

  fullScreen: boolean = false
  toggleFullscreen() {
    if (this.fullScreen) {
      document.exitFullscreen();
    } else {
      document.body.requestFullscreen();
    }
    this.fullScreen = !this.fullScreen
  }

  async logout() {
    var logoutData = {
      "secret": localStorage.getItem("secret"),
    }

    localStorage.removeItem("secret")
    Config.LoggedIn = false
    this.router.navigate(['/login'])

    this.terminalService.disconnectAllTerminals()

    let response = fetch(Config.ApiEndpoint + "/logout", {
      "method": "POST",
      "body": JSON.stringify(logoutData)
    })
    let rdata = await response

  }

  sidebarVisible: boolean = true
  async toggleSidebar() {
    this.sidebarVisible = !this.sidebarVisible
    await new Promise(f => setTimeout(f, 150));
    this.dashboardComponent.sidebarVisible = this.sidebarVisible
    this.dashboardComponent.sidebarFirstLoad = false
    this.terminalViewComponent.showLogo = !this.sidebarVisible
    localStorage.setItem('sidebarVisible', this.sidebarVisible + "")
  }

  changeFontSize() {
    const dialogRef = this.dialog.open(ChangeFontSizeDialogComponent);
    dialogRef.afterClosed().subscribe(() => {
      this.terminalService.saveFontSize()
    })
  }

}
