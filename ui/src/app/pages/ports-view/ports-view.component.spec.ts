import { ComponentFixture, TestBed } from '@angular/core/testing';

import { PortsViewComponent } from './ports-view.component';

describe('PortsViewComponent', () => {
  let component: PortsViewComponent;
  let fixture: ComponentFixture<PortsViewComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ PortsViewComponent ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(PortsViewComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
