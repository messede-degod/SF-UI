import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ChangeFontSizeDialogComponent } from './change-font-size-dialog.component';

describe('ChangeFontSizeDialogComponent', () => {
  let component: ChangeFontSizeDialogComponent;
  let fixture: ComponentFixture<ChangeFontSizeDialogComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ ChangeFontSizeDialogComponent ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(ChangeFontSizeDialogComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
