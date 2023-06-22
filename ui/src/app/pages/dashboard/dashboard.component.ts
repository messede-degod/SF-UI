import { Component } from '@angular/core';
import { Router } from '@angular/router';
import { Config } from 'src/environments/environment';

@Component({
  selector: 'app-dashboard',
  templateUrl: './dashboard.component.html',
  styleUrls: ['./dashboard.component.css']
})
export class DashboardComponent {
  title = 'Segfault';
  activeMenu: string = "terminal"
  sidebarVisible: boolean = true
  sidebarFirstLoad: boolean = true

  menuItems: Array<any> = []

  router!: Router
  desktopRequested: boolean = false
  filesRequested: boolean = false
  noOfTerminals: number = 1

  constructor(router: Router) {
    this.router = router
    this.menuItems.push({ ilink: '../assets/icons/term.svg', name: "terminal" })
    if (!Config.DesktopDisabled) {
      this.menuItems.push({ ilink: '../assets/icons/desk.svg', name: "desktop" })
    }
    this.menuItems.push({ ilink: '../assets/icons/files.svg', name: "files" })
  }

  setActiveMenu(name: string) {
    this.activeMenu = name
    if (this.activeMenu == "files" && !this.filesRequested) {
      this.filesRequested = true
    }
  }

  setNoOfTerminals(termNos: number){
    this.noOfTerminals = termNos
  }
}
