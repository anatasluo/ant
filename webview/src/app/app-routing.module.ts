import { NgModule } from '@angular/core';
import { Routes, RouterModule } from '@angular/router';
import { LocalDownloadComponent } from './local-download/local-download.component'

const routes: Routes = [
  { path: '', redirectTo: '/localDownload', pathMatch: 'full' },
  { path: 'localDownload', component: LocalDownloadComponent }
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
