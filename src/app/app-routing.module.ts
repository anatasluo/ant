import { NgModule } from '@angular/core';
import { Routes, RouterModule } from '@angular/router';
import { LocalDownloadComponent } from './components/local-download/local-download.component';
import { PageNotFoundComponent } from './components/page-not-found/page-not-found.component';
import { PlayerComponent } from './components/player/player.component';
import { SearchComponent } from './components/search/search.component';
import { SettingsComponent } from './components/settings/settings.component';
import { AboutComponent } from './components/about/about.component';

const routes: Routes = [
    { path: 'total', component: LocalDownloadComponent },
    { path: 'running', component: LocalDownloadComponent },
    { path: 'completed', component: LocalDownloadComponent },
    { path: 'player/:hexString', component: PlayerComponent },
    { path: 'search', component: SearchComponent },
    { path: 'settings', component: SettingsComponent },
    { path: 'about', component: AboutComponent },
    { path: '', redirectTo: '/total', pathMatch: 'full' },
    { path: '**', component: PageNotFoundComponent }
];
@NgModule({
    imports: [RouterModule.forRoot(routes, {
        useHash: true,
        anchorScrolling: 'enabled',
        scrollPositionRestoration: 'enabled',
    })],
    exports: [RouterModule]
})
export class AppRoutingModule { }
