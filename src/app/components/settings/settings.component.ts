import { Component, OnInit } from '@angular/core';
import { Settings } from '../../classes/settings';
import { SettingsService } from '../../providers/settings.service';
import { TranslateService } from '@ngx-translate/core';

import { remote } from 'electron';

@Component({
    selector: 'app-settings',
    templateUrl: './settings.component.html',
    styleUrls: ['./settings.component.scss']
})
export class SettingsComponent implements OnInit {

    settings: Settings;
    currentLanguage: string;
    localeTolanguage: Map<string, string>;
    constructor(
        private settingsService: SettingsService,
        public translate: TranslateService
    ) { }

    ngOnInit() {
        // To avoid error of undefined
        this.settings = new Settings;
        this.currentLanguage = localStorage.getItem('locale');
        if (this.currentLanguage === null) {
            this.currentLanguage = this.translate.getDefaultLang();
        }
        this.localeTolanguage = new Map<string, string>();
        this.localeTolanguage.set('en', 'English');
        this.localeTolanguage.set('zh', '中文');
        this.loadSettings();
    }

    changeLanguage(aimLanguage: string) {
        localStorage.setItem('locale', aimLanguage);
        this.currentLanguage = aimLanguage;
        this.translate.use(aimLanguage);
    }

    private loadSettings() {
        this.settingsService.getSettings()
            .subscribe((data: Settings) => {
                // console.log(data);
                this.settings = data;
            }, error => {
                console.log(error);
            });
    }
    private chooseDirectory() {
        const directoryPath = remote.dialog.showOpenDialog({ properties: ['openDirectory']});
        return directoryPath;
    }

    selectDataDir() {
        alert('Change of download directory will lead to restart of all unfinished tasks.');
        const directories = this.chooseDirectory();
        if (directories !== null && directories !== undefined) {
            this.settings.DataDir = directories[0];
        }
    }

    selectTmpDir() {
        const directories = this.chooseDirectory();
        if (directories !== null && directories !== undefined) {
            this.settings.Tmpdir = directories[0];
        }
    }

    resetSetting() {
        this.loadSettings();
    }

    applySetting() {
        this.settingsService.applySettings(this.settings)
            .subscribe((isApplied: boolean) => {
                console.log(isApplied);
                location.reload();
            }, error => {
                console.log(error);
            });
    }

}
