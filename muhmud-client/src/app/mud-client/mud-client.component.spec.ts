import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { MudClientComponent } from './mud-client.component';

describe('MudClientComponent', () => {
  let component: MudClientComponent;
  let fixture: ComponentFixture<MudClientComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ MudClientComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(MudClientComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
