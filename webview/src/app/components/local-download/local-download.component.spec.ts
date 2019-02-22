import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { LocalDownloadComponent } from './local-download.component';

describe('LocalDownloadComponent', () => {
  let component: LocalDownloadComponent;
  let fixture: ComponentFixture<LocalDownloadComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ LocalDownloadComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(LocalDownloadComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
