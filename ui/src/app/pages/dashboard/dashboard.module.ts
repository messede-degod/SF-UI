import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';

import { DashboardRoutingModule } from './dashboard-routing.module';
import { DashboardComponent } from './dashboard.component';

import { TerminalComponent } from '../../components/terminal/terminal.component';
import { AppControlsComponent } from '../../components/app-controls/app-controls.component';
import { TerminalControlsComponent } from '../../components/terminal-controls/terminal-controls.component';
import { TerminalViewComponent } from '../terminal-view/terminal-view.component';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { MatDialogModule } from '@angular/material/dialog';
import { MatButtonModule } from '@angular/material/button';
import { MatSnackBarModule } from '@angular/material/snack-bar';
import { HelpDialogComponent } from '../../components/help-dialog/help-dialog.component';
import { FormsModule } from '@angular/forms';
import { DesktopViewComponent } from '../desktop-view/desktop-view.component';
import { PortsViewComponent } from '../ports-view/ports-view.component';
import { WebViewComponent } from '../web-view/web-view.component';
import { SaveSecretDialogComponent } from '../../components/save-secret-dialog/save-secret-dialog.component';
import { FilesViewComponent } from '../files-view/files-view.component';


@NgModule({
  declarations: [
    DashboardComponent,
    TerminalComponent,
    AppControlsComponent,
    TerminalControlsComponent,
    TerminalViewComponent,
    HelpDialogComponent,
    DesktopViewComponent,
    PortsViewComponent,
    WebViewComponent,
    SaveSecretDialogComponent,
    FilesViewComponent,
  ],
  imports: [
    CommonModule,
    DashboardRoutingModule,
    // BrowserAnimationsModule,
    MatDialogModule,
    MatButtonModule,
    MatSnackBarModule,
    FormsModule
  ]
})
export class DashboardModule { }
