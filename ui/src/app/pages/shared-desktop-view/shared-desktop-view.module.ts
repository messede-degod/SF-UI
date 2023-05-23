import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { SharedDesktopViewComponent } from './shared-desktop-view.component';
import { SharedDesktopViewRoutingModule } from './shared-desktop-view-routing.module';



@NgModule({
  declarations: [
    SharedDesktopViewComponent
  ],
  imports: [
    CommonModule,
    SharedDesktopViewRoutingModule
  ]
})
export class SharedDesktopViewModule { }
