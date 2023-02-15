import { Component, Injectable } from '@angular/core';

@Component({
  selector: 'app-dashboard',
  templateUrl: './dashboard.component.html',
  styleUrls: ['./dashboard.component.css']
})
export class DashboardComponent {
  title = 'Segfault';
  activeMenu: String = "terminal"
  sidebarVisible: boolean = true

  menuItems: Array<any> = [
    { ilink: '../assets/icons/term.svg', name: "terminal" },
    { ilink: '../assets/icons/desk.svg', name: "desktop" },
    { ilink: '../assets/icons/ports.svg', name: "ports" },
    { ilink: '../assets/icons/web.svg', name: "web" },
  ]

  setActiveMenu(name: string) {
    this.activeMenu = name
  }
}
