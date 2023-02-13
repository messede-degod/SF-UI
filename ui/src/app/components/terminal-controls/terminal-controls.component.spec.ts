import { ComponentFixture, TestBed } from '@angular/core/testing';

import { TerminalControlsComponent } from './terminal-controls.component';

describe('TerminalControlsComponent', () => {
  let component: TerminalControlsComponent;
  let fixture: ComponentFixture<TerminalControlsComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ TerminalControlsComponent ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(TerminalControlsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
