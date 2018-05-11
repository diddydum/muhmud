import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterModule, Routes } from '@angular/router';
import { HttpClientModule } from '@angular/common/http';

import { AppComponent } from './app.component';
import { PageNotFoundComponent } from './page-not-found/page-not-found.component';
import { MudClientComponent } from './mud-client/mud-client.component';
import { LoginComponent } from './login/login.component';
import { LoginGuard } from './login-guard.service';
import { AuthService } from './auth.service';
import { LogOutComponent } from './log-out/log-out.component';

const appRoutes: Routes = [
  { path: 'login', component: LoginComponent },
  { path: 'logout', component: LogOutComponent },
  { path: 'mud', canActivate: [LoginGuard], component: MudClientComponent },
  { path: '', redirectTo: 'mud', pathMatch: 'full'},
  { path: '**', component: PageNotFoundComponent},
];

@NgModule({
  declarations: [
    AppComponent,
    PageNotFoundComponent,
    MudClientComponent,
    LoginComponent,
    LogOutComponent
  ],
  imports: [
    RouterModule.forRoot(appRoutes),
    BrowserModule,
    FormsModule,
    HttpClientModule
  ],
  providers: [LoginGuard, AuthService],
  bootstrap: [AppComponent]
})
export class AppModule { }
