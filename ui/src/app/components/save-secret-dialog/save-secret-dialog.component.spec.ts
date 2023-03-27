import { ComponentFixture, TestBed } from '@angular/core/testing';

import { SaveSecretDialogComponent } from './save-secret-dialog.component';

describe('SaveSecretDialogComponent', () => {
  let component: SaveSecretDialogComponent;
  let fixture: ComponentFixture<SaveSecretDialogComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ SaveSecretDialogComponent ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(SaveSecretDialogComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
