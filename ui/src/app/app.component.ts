import { Component } from '@angular/core';
import { Config } from './config/config';
import { MatSnackBar } from '@angular/material/snack-bar';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent {
  title = 'Segfault';

  constructor(private snackBar: MatSnackBar) {
    this.fetchConfig()
  }

  async fetchConfig() {
    let response = fetch(Config.ApiEndpoint + "/config", { "method": "GET" })
    let rdata = await response
    if (rdata.status == 200) {
      let config = await rdata.json()
      Config.MaxOpenTerminals = config.max_terminals
    } else {
      this.snackBar.open("Failed to fetch config from server !", "OK", {
        duration: 2 * 1000
      });
    }
  }

}
