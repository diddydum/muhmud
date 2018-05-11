
/**
 * Renders a message from the server into html.
 *
 * @param msg message to render
 */
export function renderMessage(msg: string): string {
  let esc = escapeHTML(msg);
  // THIS IS WRONG
  esc = esc
      .replace('%r', '<span class="c-r">')
      .replace('%c', '</span>');
  return esc;
}

function escapeHTML(s: string): string {
  return s
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#039;');
}
