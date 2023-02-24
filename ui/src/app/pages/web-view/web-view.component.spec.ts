import { ComponentFixture, TestBed } from '@angular/core/testing';

import { WebViewComponent } from './web-view.component';

describe('WebViewComponent', () => {
  let component: WebViewComponent;
  let fixture: ComponentFixture<WebViewComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ WebViewComponent ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(WebViewComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
