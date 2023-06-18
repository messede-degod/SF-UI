import { NgModule } from '@angular/core';
import { ActivatedRouteSnapshot, CanActivateFn, RouterModule, RouterStateSnapshot, Routes, createUrlTreeFromSnapshot } from '@angular/router';
import { Config } from 'src/environments/environment';

const canActivateRoute: CanActivateFn =
  (route: ActivatedRouteSnapshot, state: RouterStateSnapshot) => {
    let secret = localStorage.getItem("secret")
    if (secret == "" || secret == null || !Config.LoggedIn) {
      return createUrlTreeFromSnapshot(route, ['/login']);

    }
    return true
  };


const routes: Routes = [
  { path: '', redirectTo: '/login', pathMatch: 'full' },
  {
    path: 'dashboard',
    loadChildren: () => import('./pages/dashboard/dashboard.module').then(m => m.DashboardModule),
    canActivate: [canActivateRoute],
  },
  { path: 'shared-desktop/:secret', loadChildren: () => import('./pages/shared-desktop-view/shared-desktop-view.module').then(m => m.SharedDesktopViewModule) },
  { path: 'login', loadChildren: () => import('./pages/login/login.module').then(m => m.LoginModule) },
];

@NgModule({
  imports: [RouterModule.forRoot(routes, { useHash: true })],
  exports: [RouterModule]
})
export class AppRoutingModule { }
