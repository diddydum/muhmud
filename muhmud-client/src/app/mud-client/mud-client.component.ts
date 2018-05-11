import { Component, OnInit, AfterViewChecked, ViewEncapsulation } from '@angular/core';
import { AuthService } from '../auth.service';
import { DomSanitizer } from '@angular/platform-browser';
import { SecurityContext } from '@angular/core';
import { renderMessage } from './render-message';
import { environment } from '../../environments/environment';

@Component({
  selector: 'app-mud-client',
  templateUrl: './mud-client.component.html',
  styleUrls: ['./mud-client.component.css'],
  encapsulation: ViewEncapsulation.None,
})
export class MudClientComponent implements OnInit, AfterViewChecked {
  constructor(private authService: AuthService, private sanitizer: DomSanitizer) {}

  websocket: WebSocket;
  messages = [];
  command = '';
  authExpired = false;

  messagesScrolledToBottom = true;

  ngOnInit() {
    // TODO: put ws stuff in its own class
    const wsServer = new URL(environment.server);
    if (wsServer.protocol === 'https:') {
      wsServer.protocol = 'wss:';
    } else {
      wsServer.protocol = 'ws:';
    }
    this.websocket = new WebSocket(wsServer.toString() + 'ws?token='
        + encodeURIComponent(this.authService.token));
    console.log(this.websocket.url);
    this.websocket.onclose = (ev) => {
      console.log(`DEBUG onclose occurred: ${ev.code}`);
      if (ev.code === 4000) {
        // Auth must be expired
        this.authService.clearToken();
        this.authExpired = true;
      }
    };
    this.websocket.onopen = () => console.log('DEBUG onopen occurred');
    this.websocket.onerror = () => console.log('DEBUG onerror occurred');
    this.websocket.onmessage = (event) => {
      this.messages.push(renderMessage(event.data));
    };
  }

  ngAfterViewChecked(): void {
    if (this.messagesScrolledToBottom) {
      const el = document.getElementById('messages');
      el.scrollTop = el.scrollHeight;
    }
  }

  messagesScrolled(event): void {
    // check to see if we are still scrolled at bottom
    const el = event.target;
    this.messagesScrolledToBottom = (el.scrollHeight - el.clientHeight) === el.scrollTop;
  }

  onCommandSubmit() {
    this.websocket.send(this.command);

    this.command = '';
  }

  isConnected() {
    return this.websocket && this.websocket.readyState === WebSocket.OPEN;
  }

  isCommandEmpty() {
    return this.command.length === 0;
  }
}
