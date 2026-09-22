import { Routes } from '@angular/router';
import { MainLayoutComponent } from './layouts/main-layout.component';
import { authGuard, roleGuard } from './guards';
import { ROLE } from '../constants/enums';

export const routes: Routes = [
  { path: 'login', loadComponent: () => import('./pages/login/login.component').then(m => m.LoginComponent) },
  {
    path: '',
    component: MainLayoutComponent,
    canActivate: [authGuard],
    children: [
      { path: '', pathMatch: 'full', redirectTo: 'dashboard' },
      { path: 'dashboard', loadComponent: () => import('./pages/dashboard/dashboard.component').then(m => m.DashboardComponent) },
      { path: 'devices', loadComponent: () => import('./pages/devices/devices.component').then(m => m.DevicesComponent) },
      { path: 'purchases', loadComponent: () => import('./pages/purchases/purchases.component').then(m => m.PurchasesComponent) },
      { path: 'maintenance', loadComponent: () => import('./pages/maintenance/maintenance.component').then(m => m.MaintenanceComponent) },
      { path: 'calibrations', loadComponent: () => import('./pages/calibrations/calibrations.component').then(m => m.CalibrationsComponent) },
      { path: 'transfers', loadComponent: () => import('./pages/transfers/transfers.component').then(m => m.TransfersComponent) },
      { path: 'scraps', loadComponent: () => import('./pages/scraps/scraps.component').then(m => m.ScrapsComponent) },
      { path: 'audits', canActivate: [roleGuard(ROLE.SUPER_ADMIN, ROLE.DEVICE_ADMIN)], loadComponent: () => import('./pages/audits/audits.component').then(m => m.AuditsComponent) },
    ],
  },
  { path: '**', redirectTo: 'dashboard' },
];
