import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ShareDesktopDialogComponent } from './share-desktop-dialog.component';

describe('ShareDesktopDialogComponent', () => {
  let component: ShareDesktopDialogComponent;
  let fixture: ComponentFixture<ShareDesktopDialogComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ ShareDesktopDialogComponent ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(ShareDesktopDialogComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
