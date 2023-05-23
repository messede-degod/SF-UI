import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { SharedDesktopViewComponent } from './shared-desktop-view.component';

const routes: Routes = [{ path: '', component: SharedDesktopViewComponent }];

@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: [RouterModule]
})
export class SharedDesktopViewRoutingModule { }
