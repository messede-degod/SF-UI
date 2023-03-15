import { Component, Injectable } from '@angular/core';
import { Router } from '@angular/router';

@Component({
  selector: 'app-dashboard',
  templateUrl: './dashboard.component.html',
  styleUrls: ['./dashboard.component.css']
})
export class DashboardComponent {
  title = 'Segfault';
  activeMenu: String = "terminal"
  sidebarVisible: boolean = true
  sidebarFirstLoad: boolean = true

  menuItems: Array<any> = [
    { ilink: '../assets/icons/term.svg', name: "terminal" },
    { ilink: '../assets/icons/desk.svg', name: "desktop" },
    { ilink: '../assets/icons/ports.svg', name: "ports" },
    { ilink: '../assets/icons/web.svg', name: "web" },
  ]

  router!: Router
  desktopRequested: boolean = false

  constructor(router: Router) {
    this.router = router
  }

  setActiveMenu(name: string) {
    this.activeMenu = name
    if(this.activeMenu=="desktop" && !this.desktopRequested){
      this.desktopRequested = true
    }
  }
}
