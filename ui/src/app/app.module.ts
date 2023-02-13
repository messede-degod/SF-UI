import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';

import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { TerminalComponent } from './components/terminal/terminal.component';
import { AppControlsComponent } from './components/app-controls/app-controls.component';
import { TerminalControlsComponent } from './components/terminal-controls/terminal-controls.component';
import { TerminalViewComponent } from './pages/terminal-view/terminal-view.component';
import { DashboardComponent } from './pages/dashboard/dashboard.component';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import {MatDialogModule} from '@angular/material/dialog';
import {MatButtonModule} from '@angular/material/button';
import {MatSnackBarModule} from '@angular/material/snack-bar';
import { LoginComponent } from './pages/login/login.component';
import { HelpDialogComponent } from './components/help-dialog/help-dialog.component';
import { FormsModule } from '@angular/forms';

@NgModule({
  declarations: [
    AppComponent,
    TerminalComponent,
    AppControlsComponent,
    TerminalControlsComponent,
    TerminalViewComponent,
    DashboardComponent,
    LoginComponent,
    HelpDialogComponent,
  ],
  imports: [
    BrowserModule,
    AppRoutingModule,
    BrowserAnimationsModule,
    MatDialogModule,
    MatButtonModule,
    MatSnackBarModule,
    FormsModule
  ],
  providers: [],
  bootstrap: [AppComponent]
})
export class AppModule { }
