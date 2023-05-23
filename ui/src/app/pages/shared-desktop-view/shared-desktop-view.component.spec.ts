import { ComponentFixture, TestBed } from '@angular/core/testing';

import { SharedDesktopViewComponent } from './shared-desktop-view.component';

describe('SharedDesktopViewComponent', () => {
  let component: SharedDesktopViewComponent;
  let fixture: ComponentFixture<SharedDesktopViewComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ SharedDesktopViewComponent ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(SharedDesktopViewComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
