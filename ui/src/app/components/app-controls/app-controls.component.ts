import { Component } from '@angular/core';
import { Router } from '@angular/router';
import { Config } from 'src/app/config/config';
import { DashboardComponent } from 'src/app/pages/dashboard/dashboard.component';
import { TerminalViewComponent } from 'src/app/pages/terminal-view/terminal-view.component'
import { TerminalService } from 'src/app/services/terminal.service';

@Component({
  selector: 'app-controls',
  templateUrl: './app-controls.component.html',
  styleUrls: ['./app-controls.component.css'],
})
export class AppControlsComponent {
  router!: Router

  constructor(router: Router,
    private dashboardComponent: DashboardComponent,
    private terminalViewComponent: TerminalViewComponent) {
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

    let response = fetch(Config.ApiEndpoint + "/logout", {
      "method": "POST",
      "body": JSON.stringify(logoutData)
    })
    let rdata = await response

    localStorage.removeItem("secret")
    this.router.navigate(['/login'])
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

}
