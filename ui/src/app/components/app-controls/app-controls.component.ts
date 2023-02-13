import { Component } from '@angular/core';
import { Router } from '@angular/router';

@Component({
  selector: 'app-controls',
  templateUrl: './app-controls.component.html',
  styleUrls: ['./app-controls.component.css']
})
export class AppControlsComponent {
  router!: Router

  constructor(router:  Router){
    this.router = router
    if(localStorage.getItem('theme')=='dark'){
      this.setTheme('dark')
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

  setTheme(theme: string){
    this.curTheme = theme
    document.documentElement.setAttribute('data-theme', theme);
    localStorage.setItem('theme',theme)
  }

  fullScreen: boolean = false
  async toggleFullscreen() {
    if (this.fullScreen) {
      await document.exitFullscreen();
    } else {
      document.body.requestFullscreen();
    }
    this.fullScreen = !this.fullScreen
  }

  logout(){
    this.router.navigate(['/login'])
  }

}
