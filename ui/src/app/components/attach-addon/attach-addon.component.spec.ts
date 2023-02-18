import { ComponentFixture, TestBed } from '@angular/core/testing';

import { AttachAddonComponent } from './attach-addon.component';

describe('AttachAddonComponent', () => {
  let component: AttachAddonComponent;
  let fixture: ComponentFixture<AttachAddonComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ AttachAddonComponent ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(AttachAddonComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
