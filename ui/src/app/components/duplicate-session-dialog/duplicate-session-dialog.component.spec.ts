import { ComponentFixture, TestBed } from '@angular/core/testing';

import { DuplicateSessionDialogComponent } from './duplicate-session-dialog.component';

describe('DuplicateSessionDialogComponent', () => {
  let component: DuplicateSessionDialogComponent;
  let fixture: ComponentFixture<DuplicateSessionDialogComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [DuplicateSessionDialogComponent]
    });
    fixture = TestBed.createComponent(DuplicateSessionDialogComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
